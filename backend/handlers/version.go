package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"sqlite-manager/upgrade"
)

// ProcessStartedAt 进程启动时间，由 main 包在启动时设置
var ProcessStartedAt = time.Now()

// GetVersion 获取当前版本号
func GetVersion(c *gin.Context) {
	currentDBVersion := upgrade.GetCurrentDBVersion()
	appliedVersions, _ := upgrade.GetAppliedVersions()

	c.JSON(http.StatusOK, gin.H{
		"appVersion":      upgrade.Version,
		"dbVersion":       currentDBVersion,
		"appliedVersions": appliedVersions,
		"startedAt":       ProcessStartedAt.Format(time.RFC3339),
	})
}

// GetUpgradeStatus 获取升级状态
func GetUpgradeStatus(c *gin.Context) {
	currentDBVersion := upgrade.GetCurrentDBVersion()
	appliedVersions, _ := upgrade.GetAppliedVersions()

	needsUpgrade := upgrade.CompareVersions(currentDBVersion, upgrade.Version) < 0

	c.JSON(http.StatusOK, gin.H{
		"appVersion":      upgrade.Version,
		"dbVersion":       currentDBVersion,
		"needsUpgrade":    needsUpgrade,
		"appliedVersions": appliedVersions,
	})
}
