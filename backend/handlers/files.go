package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"sqlite-manager/config"
)

type FileInfo struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

type ShareInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type BrowseResponse struct {
	CurrentPath string      `json:"currentPath"`
	Parent      string      `json:"parent"`
	Files       []FileInfo  `json:"files"`
	ShareDirs   []ShareInfo `json:"shareDirs"`
	CanGoBack   bool        `json:"canGoBack"`
}

// GetShareDirs 获取共享目录列表
func GetShareDirs(c *gin.Context) {
	shares := getAvailableShares()
	c.JSON(http.StatusOK, shares)
}

// getAvailableShares 获取可用的共享目录列表
func getAvailableShares() []ShareInfo {
	var shares []ShareInfo

	// 1. 添加用户配置的共享目录
	for _, dir := range config.GetShareDirs() {
		if dir != "" {
			name := filepath.Base(dir)
			shares = append(shares, ShareInfo{Name: name, Path: dir})
		}
	}

	// 2. 如果没有找到任何目录，添加用户主目录
	if len(shares) == 0 {
		home, err := os.UserHomeDir()
		if err == nil {
			shares = append(shares, ShareInfo{Name: "主目录", Path: home})
		}
	}

	return shares
}

// BrowseFiles 浏览服务器文件系统
func BrowseFiles(c *gin.Context) {
	requestPath := c.Query("path")

	shares := getAvailableShares()

	// 如果没有指定路径，返回第一个可用目录
	if requestPath == "" {
		if len(shares) > 0 {
			requestPath = shares[0].Path
		} else {
			requestPath = "/"
		}
	}

	absPath, err := filepath.Abs(requestPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的路径"})
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "路径不存在"})
		return
	}

	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择目录"})
		return
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取目录"})
		return
	}

	var files []FileInfo
	for _, entry := range entries {
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

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	var filteredFiles []FileInfo
	for _, file := range files {
		if file.IsDir {
			filteredFiles = append(filteredFiles, file)
		} else if isSQLiteFile(file.Name) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	parent := filepath.Dir(absPath)
	canGoBack := parent != absPath && absPath != "/"

	c.JSON(http.StatusOK, BrowseResponse{
		CurrentPath: absPath,
		Parent:      parent,
		Files:       filteredFiles,
		ShareDirs:   shares,
		CanGoBack:   canGoBack,
	})
}

func isSQLiteFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".db") ||
		strings.HasSuffix(lower, ".sqlite") ||
		strings.HasSuffix(lower, ".sqlite3")
}
