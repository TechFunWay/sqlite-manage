package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sqlite-manager/auth"
)

type AddRecentRequest struct {
	Path string `json:"path" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// GetRecentDatabases 获取最近打开的数据库
func GetRecentDatabases(c *gin.Context) {
	records, err := auth.GetRecentDatabases(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

// AddRecentDatabase 添加最近打开的数据库
func AddRecentDatabase(c *gin.Context) {
	var req AddRecentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := auth.AddRecentDatabase(req.Path, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// ClearRecentDatabases 清空最近打开的数据库
func ClearRecentDatabases(c *gin.Context) {
	if err := auth.ClearRecentDatabases(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
