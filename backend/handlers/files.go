package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type FileInfo struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

type BrowseResponse struct {
	CurrentPath string     `json:"currentPath"`
	Parent      string     `json:"parent"`
	Files       []FileInfo `json:"files"`
}

// BrowseFiles 浏览服务器文件系统
func BrowseFiles(c *gin.Context) {
	requestPath := c.Query("path")
	if requestPath == "" {
		// Default to home directory or current directory
		home, err := os.UserHomeDir()
		if err != nil {
			requestPath = "."
		} else {
			requestPath = home
		}
	}

	// Clean and validate path
	absPath, err := filepath.Abs(requestPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的路径"})
		return
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "路径不存在"})
		return
	}

	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择目录"})
		return
	}

	// Read directory
	entries, err := os.ReadDir(absPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取目录"})
		return
	}

	// Build file list
	var files []FileInfo
	for _, entry := range entries {
		// Skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		file := FileInfo{
			Name:  entry.Name(),
			Path:  filepath.Join(absPath, entry.Name()),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		}
		files = append(files, file)
	}

	// Sort: directories first, then by name
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	// Filter to show only directories and .db/.sqlite files
	var filteredFiles []FileInfo
	for _, file := range files {
		if file.IsDir {
			filteredFiles = append(filteredFiles, file)
		} else if isSQLiteFile(file.Name) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	parent := filepath.Dir(absPath)
	if parent == absPath {
		parent = "" // Already at root
	}

	c.JSON(http.StatusOK, BrowseResponse{
		CurrentPath: absPath,
		Parent:      parent,
		Files:       filteredFiles,
	})
}

func isSQLiteFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".db") ||
		strings.HasSuffix(lower, ".sqlite") ||
		strings.HasSuffix(lower, ".sqlite3")
}
