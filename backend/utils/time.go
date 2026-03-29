package utils

import (
	"os"
	"time"
)

// GetBeijingTime 获取北京时间
func GetBeijingTime() time.Time {
	// 设置时区为亚洲/上海（北京时间）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载时区失败，使用固定偏移
		loc = time.FixedZone("CST", 8*60*60)
	}
	return time.Now().In(loc)
}

// GetBeijingTimeString 获取北京时间字符串（RFC3339格式）
func GetBeijingTimeString() string {
	return GetBeijingTime().Format(time.RFC3339)
}

// GetBeijingTimeStringCustom 获取自定义格式的北京时间字符串
func GetBeijingTimeStringCustom(layout string) string {
	return GetBeijingTime().Format(layout)
}

// InitTimezone 初始化时区设置
func InitTimezone() {
	// 设置环境变量
	os.Setenv("TZ", "Asia/Shanghai")
}
