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

	data, total, err := database.GetData(tableName, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     data,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
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

	upperSQL := req.SQL
	for i := 0; i < len(upperSQL); i++ {
		if upperSQL[i] >= 'a' && upperSQL[i] <= 'z' {
			upperSQL = upperSQL[:i] + string(upperSQL[i]-32) + upperSQL[i+1:]
		}
	}

	if len(upperSQL) > 0 && (upperSQL[:6] == "SELECT" || upperSQL[:6] == "PRAGMA" || upperSQL[:4] == "WITH") {
		results, err := database.ExecuteQuery(req.SQL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"type": "select",
			"data": results,
		})
	} else {
		rowsAffected, err := database.ExecuteNonQuery(req.SQL)
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
