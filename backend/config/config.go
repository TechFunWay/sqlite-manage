package config

import (
	"path/filepath"
	"strings"
)

// 全局配置
var (
	DataDir   = "./data"
	PublicDir = "./public"
	UploadDir = "./upload"
	ShareDirs = []string{} // 共享目录列表
)

// GetDataDir 获取数据目录
func GetDataDir() string {
	return DataDir
}

// GetPublicDir 获取静态资源目录
func GetPublicDir() string {
	return PublicDir
}

// GetUploadDir 获取上传目录
func GetUploadDir() string {
	return UploadDir
}

// GetShareDirs 获取共享目录列表
func GetShareDirs() []string {
	return ShareDirs
}

// SetShareDirs 设置共享目录 (冒号分隔的路径)
func SetShareDirs(dirs string) {
	if dirs == "" {
		ShareDirs = []string{}
		return
	}

	parts := strings.Split(dirs, ":")
	ShareDirs = []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			ShareDirs = append(ShareDirs, p)
		}
	}
}

// GetUploadDBDir 获取上传数据库目录（带日期子目录）
func GetUploadDBDir(dateStr string) string {
	return filepath.Join(UploadDir, "db", dateStr)
}

// GetSystemDBPath 获取系统数据库路径
func GetSystemDBPath() string {
	return filepath.Join(DataDir, "db", "database.db")
}

// SetDirs 设置目录
func SetDirs(data, public, upload string) {
	DataDir = data
	PublicDir = public
	UploadDir = upload
}
