package telemetry

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	// 统计地址
	Endpoint = "http://page.wycto.cc/api/apps.online/refresh"
	// 发送间隔 (60分钟)
	SendInterval = 60 * time.Minute
)

// TelemetryData 统计数据
type TelemetryData struct {
	AppName  string `json:"app_name"`
	Version  string `json:"version"`
	DeviceID string `json:"device_id"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Hostname string `json:"hostname"`
}

var (
	deviceID string
	hostname string
	appName  = "sqlite-manage"
	version  string
	stopChan chan struct{}
)

// Init 初始化统计模块
func Init(appVer string) {
	version = appVer

	// 生成设备ID
	deviceID = generateDeviceID()

	// 获取主机名
	hostname, _ = os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	// 取前12位
	if len(hostname) > 12 {
		hostname = hostname[:12]
	}

	stopChan = make(chan struct{})
}

// generateDeviceID 生成设备ID
func generateDeviceID() string {
	hostname, _ := os.Hostname()
	macAddr := getMacAddress()

	data := fmt.Sprintf("%s-%s-%s-%s",
		hostname,
		macAddr,
		runtime.GOOS,
		runtime.GOARCH,
	)

	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// getMacAddress 获取MAC地址
func getMacAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String()
		}
	}
	return "unknown"
}

// collectData 收集统计数据
func collectData() TelemetryData {
	return TelemetryData{
		AppName:  appName,
		Version:  "v" + version,
		DeviceID: deviceID,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: hostname,
	}
}

// send 发送统计数据（静默）
func send() {
	data := collectData()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(Endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// Start 启动定时统计
func Start() {
	go func() {
		// 启动时发送一次
		send()

		// 每60分钟发送一次
		ticker := time.NewTicker(SendInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				send()
			case <-stopChan:
				return
			}
		}
	}()
}

// Stop 停止统计
func Stop() {
	if stopChan != nil {
		close(stopChan)
	}
}
