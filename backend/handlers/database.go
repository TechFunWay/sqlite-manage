package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sqlite-manager/config"
	"sqlite-manager/database"
)

type OpenDatabaseRequest struct {
	Path string `json:"path"`
}

func OpenDatabase(c *gin.Context) {
	var req OpenDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: path is required"})
		return
	}

	db, err := database.Open(req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, db)
}

func CreateDatabase(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: name is required"})
		return
	}

	// 确保 .db 扩展名
	fileName := req.Name
	if !strings.HasSuffix(fileName, ".db") && !strings.HasSuffix(fileName, ".sqlite") && !strings.HasSuffix(fileName, ".sqlite3") {
		fileName += ".db"
	}

	// 保存到 databases 目录（在 data 目录下）
	saveDir := filepath.Join(config.GetDataDir(), "databases")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory: " + err.Error()})
		return
	}

	savePath := filepath.Join(saveDir, fileName)
	db, err := database.OpenOrCreate(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, db)
}

func GetDatabaseInfo(c *gin.Context) {
	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	info, err := database.GetInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

func GetAllDatabases(c *gin.Context) {
	dbs := database.GetAllDatabases()
	c.JSON(http.StatusOK, dbs)
}

func SetActiveDatabase(c *gin.Context) {
	id := c.Param("id")
	db, err := database.SetActive(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, db)
}

func CloseDatabase(c *gin.Context) {
	id := c.Param("id")
	if err := database.Close(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Database closed"})
}

func GetTables(c *gin.Context) {
	if !database.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No database connected"})
		return
	}

	tables, err := database.GetTables()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type TableInfo struct {
		Name string `json:"name"`
		Rows int64  `json:"rows"`
	}

	tableInfos := make([]TableInfo, 0, len(tables))
	for _, name := range tables {
		count, _ := database.GetTableRowCount(name)
		tableInfos = append(tableInfos, TableInfo{
			Name: name,
			Rows: count,
		})
	}

	c.JSON(http.StatusOK, tableInfos)
}

func UploadDatabase(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 使用配置的上传目录
	dateStr := time.Now().Format("20060102")
	uploadDBDir := config.GetUploadDBDir(dateStr)
	if err := os.MkdirAll(uploadDBDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	savePath := filepath.Join(uploadDBDir, file.Filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	db, err := database.Open(savePath)
	if err != nil {
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, db)
}
