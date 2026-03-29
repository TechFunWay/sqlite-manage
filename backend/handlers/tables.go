package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sqlite-manager/database"
	"sqlite-manager/models"
)

func GetTableSchema(c *gin.Context) {
	tableName := c.Param("name")

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	schema, err := database.GetSchema(tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema)
}

func CreateTable(c *gin.Context) {
	var req models.CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	dbColumns := make([]database.Column, len(req.Columns))
	for i, col := range req.Columns {
		dbColumns[i] = database.Column{
			Name:         col.Name,
			Type:         col.Type,
			Nullable:     col.Nullable,
			DefaultValue: col.DefaultValue,
			PrimaryKey:   col.PrimaryKey,
		}
	}
	if err := database.CreateTable(req.Name, dbColumns); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Table created successfully"})
}

func DropTable(c *gin.Context) {
	tableName := c.Param("name")

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	if err := database.DropTable(tableName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Table dropped successfully"})
}

func RenameTable(c *gin.Context) {
	var req struct {
		OldName string `json:"oldName"`
		NewName string `json:"newName"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	if err := database.RenameTable(req.OldName, req.NewName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Table renamed successfully"})
}

func AddColumn(c *gin.Context) {
	tableName := c.Param("name")
	var reqColumn models.Column

	if err := c.ShouldBindJSON(&reqColumn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	column := database.Column{
		Name:         reqColumn.Name,
		Type:         reqColumn.Type,
		Nullable:     reqColumn.Nullable,
		DefaultValue: reqColumn.DefaultValue,
		PrimaryKey:   reqColumn.PrimaryKey,
	}
	if err := database.AddColumn(tableName, column); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Column added successfully"})
}

func DropColumn(c *gin.Context) {
	tableName := c.Param("name")
	columnName := c.Param("column")

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	if err := database.DropColumn(tableName, columnName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Column dropped successfully"})
}

func GetIndexes(c *gin.Context) {
	tableName := c.Param("name")

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	schema, err := database.GetSchema(tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema.Indexes)
}

func CreateIndex(c *gin.Context) {
	tableName := c.Param("name")
	var req models.CreateIndexRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	if err := database.CreateIndex(tableName, req.Name, req.Columns, req.Unique); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Index created successfully"})
}

func DropIndex(c *gin.Context) {
	indexName := c.Param("name")

	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	if err := database.DropIndex(indexName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Index dropped successfully"})
}
