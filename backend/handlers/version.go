package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sqlite-manager/upgrade"
)

// GetVersion 获取当前版本号
func GetVersion(c *gin.Context) {
	currentDBVersion := upgrade.GetCurrentDBVersion()
	appliedVersions, _ := upgrade.GetAppliedVersions()

	c.JSON(http.StatusOK, gin.H{
		"appVersion":      upgrade.Version,
		"dbVersion":       currentDBVersion,
		"appliedVersions": appliedVersions,
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
