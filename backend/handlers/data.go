package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"sqlite-manager/database"
)

func GetTableData(c *gin.Context) {
	tableName := c.Param("name")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "100")
	where := c.DefaultQuery("where", "")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 100
	}

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	// 设置当前表名，供 SQL 执行器使用
	database.SetCurrentTableName(tableName)

	data, total, err := database.GetData(tableName, page, pageSize, where)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     data,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"where":    where,
	})
}

type InsertRowRequest struct {
	Data map[string]interface{} `json:"data"`
}

func InsertRow(c *gin.Context) {
	tableName := c.Param("name")
	var req InsertRowRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	id, err := database.InsertRow(tableName, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Row inserted successfully"})
}

type UpdateRowRequest struct {
	PrimaryKey string                 `json:"primaryKey"`
	PkValue    interface{}            `json:"pkValue"`
	Data       map[string]interface{} `json:"data"`
}

func UpdateRow(c *gin.Context) {
	tableName := c.Param("name")
	var req UpdateRowRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	if err := database.UpdateRow(tableName, req.PrimaryKey, req.PkValue, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Row updated successfully"})
}

type DeleteRowRequest struct {
	PrimaryKey string      `json:"primaryKey"`
	PkValue    interface{} `json:"pkValue"`
}

func DeleteRow(c *gin.Context) {
	tableName := c.Param("name")
	var req DeleteRowRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	if err := database.DeleteRow(tableName, req.PrimaryKey, req.PkValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Row deleted successfully"})
}

func ExecuteQuery(c *gin.Context) {
	var req struct {
		SQL string `json:"sql"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	sql := req.SQL
	if sql == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SQL 语句不能为空"})
		return
	}

	// 获取当前表名
	currentTable := database.GetCurrentTableName()

	// 如果 SQL 中没有指定表名，自动添加当前表
	upperSQL := ""
	for i := 0; i < len(sql); i++ {
		if sql[i] >= 'a' && sql[i] <= 'z' {
			upperSQL += string(sql[i] - 32)
		} else {
			upperSQL += string(sql[i])
		}
	}

	// 检测 SQL 类型
	isSelect := false
	if len(upperSQL) >= 6 && upperSQL[:6] == "SELECT" {
		isSelect = true
	} else if len(upperSQL) >= 6 && upperSQL[:6] == "PRAGMA" {
		isSelect = true
	} else if len(upperSQL) >= 4 && upperSQL[:4] == "WITH" {
		isSelect = true
	}

	// 如果是 SELECT 但没有 FROM 子句，自动添加
	if isSelect && currentTable != "" {
		// 检查是否已有 FROM 子句（简单检测）
		hasFrom := false
		for i := 0; i < len(upperSQL)-4; i++ {
			if upperSQL[i:i+4] == "FROM" {
				hasFrom = true
				break
			}
		}
		
		if !hasFrom {
			// 在 WHERE 前插入 FROM 表名，如果没有 WHERE 则在语句末尾插入
			whereIdx := -1
			for i := 0; i < len(upperSQL)-5; i++ {
				if upperSQL[i:i+5] == "WHERE" {
					whereIdx = i
					break
				}
			}
			
			if whereIdx > 0 {
				sql = sql[:whereIdx] + "FROM \"" + currentTable + "\" " + sql[whereIdx:]
			} else {
				// 在 SELECT 后面找第一个空格，然后插入 FROM
				spaceIdx := 6
				for spaceIdx < len(sql) && sql[spaceIdx] == ' ' {
					spaceIdx++
				}
				// 找到下一个空格或关键字
				for spaceIdx < len(sql) {
					ch := uint8(sql[spaceIdx])
					if ch == ' ' || ch == 'W' || ch == 'O' || ch == 'G' || ch == 'L' || ch == 'O' || ch == 'R' || ch == 'D' || ch == 'B' || ch == 'E' || ch == 'N' || ch == 'T' || ch == 'C' || ch == 'A' || ch == 'S' {
						break
					}
					spaceIdx++
				}
				sql = sql[:spaceIdx] + " FROM \"" + currentTable + "\" " + sql[spaceIdx:]
			}
		}
	}

	if isSelect {
		results, err := database.ExecuteQuery(sql)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"type": "select",
			"data": results,
		})
	} else {
		rowsAffected, err := database.ExecuteNonQuery(sql)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"type":         "modify",
			"rowsAffected": rowsAffected,
		})
	}
}

func GetPrimaryKey(c *gin.Context) {
	tableName := c.Param("name")

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	pk, err := database.GetPrimaryKey(tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"primaryKey": pk})
}
