package config

import "path/filepath"

// 全局配置
var (
	DataDir   = "./data"
	PublicDir = "./public"
	UploadDir = "./upload"
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

// GetUploadDBDir 获取上传数据库目录（带日期子目录）
func GetUploadDBDir(dateStr string) string {
	return filepath.Join(UploadDir, "db", dateStr)
}

// GetSystemDBPath 获取系统数据库路径
func GetSystemDBPath() string {
	return filepath.Join(DataDir, "sqlite-manage.db")
}

// SetDirs 设置目录
func SetDirs(data, public, upload string) {
	DataDir = data
	PublicDir = public
	UploadDir = upload
}
