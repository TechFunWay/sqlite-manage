package telemetry

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// 默认统计地址（可用 STATS_ENDPOINT 环境变量覆盖，用于本地测试）
	defaultEndpoint = "https://techfunway.wycto.cn/api/apps.online/refresh"
	// 发送间隔 (60分钟)
	SendInterval = 60 * time.Minute
)

// TelemetryData 统计数据（仅设备级信息，不含任何用户数据与主机名原文）
type TelemetryData struct {
	AppName    string `json:"app_name"`
	Version    string `json:"version"`
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Event      string `json:"event,omitempty"` // 心跳为空；赞赏点击为 donate_support
}

var (
	deviceID   string
	deviceType string
	appName    = "sqlite-manage"
	version    string
	dataDir    string
	stopChan   chan struct{}
)

// Init 初始化统计模块（dataDir 用于持久化设备标识文件）
func Init(appVer string, devType string, dir string) {
	version = appVer
	deviceType = devType
	dataDir = dir

	// 生成设备ID：系统机器标识优先，回退数据目录持久化随机 ID
	deviceID = getDeviceID(dir)

	stopChan = make(chan struct{})
}

// Enabled 是否启用统计（DISABLE_STATS=1 或设 -disable-stats 关闭）
func Enabled() bool {
	return os.Getenv("DISABLE_STATS") == "" && !disabledByFlag
}

var disabledByFlag bool

// SetDisabled 由命令行参数调用，关闭统计上报
func SetDisabled() {
	disabledByFlag = true
}

// endpoint 统计上报地址
func endpoint() string {
	if e := os.Getenv("STATS_ENDPOINT"); e != "" {
		return e
	}
	return defaultEndpoint
}

// getDeviceID 设备唯一标识：系统机器标识（卸载重装不变）→ 数据目录持久化随机 ID
// 上报的只有 md5 哈希，不含任何原文
func getDeviceID(dir string) string {
	source := machineSignature()
	if source == "" {
		source = persistentDeviceID(dir)
	}
	hash := md5.Sum([]byte("sqlite-manage|" + source))
	return fmt.Sprintf("%x", hash)
}

// machineSignature 读取操作系统安装时生成的机器标识，读取失败返回空串
func machineSignature() string {
	switch runtime.GOOS {
	case "linux":
		for _, p := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
			if b, err := os.ReadFile(p); err == nil {
				if s := strings.TrimSpace(string(b)); s != "" {
					return s
				}
			}
		}
	case "darwin":
		out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if !strings.Contains(line, "IOPlatformUUID") {
					continue
				}
				if i := strings.Index(line, "= \""); i >= 0 {
					return strings.Trim(line[i+3:], "\" ")
				}
			}
		}
	case "windows":
		out, err := exec.Command("reg", "query", `HKLM\SOFTWARE\Microsoft\Cryptography`, "/v", "MachineGuid").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if strings.Contains(line, "MachineGuid") {
					fields := strings.Fields(line)
					if len(fields) >= 3 {
						return fields[len(fields)-1]
					}
				}
			}
		}
	}
	return ""
}

// persistentDeviceID 无系统机器标识时（容器等），在数据目录持久化随机 UUID，
// 容器重建、主机改名后身份保持稳定
func persistentDeviceID(dir string) string {
	path := filepath.Join(dir, "device.id")
	if b, err := os.ReadFile(path); err == nil {
		if s := strings.TrimSpace(string(b)); s != "" {
			return s
		}
	}
	id := uuid.NewString()
	if err := os.MkdirAll(dir, 0o755); err == nil {
		if err := os.WriteFile(path, []byte(id), 0o600); err == nil {
			return id
		}
	}
	// 数据目录不可写时的最终兜底
	hostname, _ := os.Hostname()
	return hostname + runtime.GOOS + runtime.GOARCH
}

// collectData 收集统计数据
func collectData(event string) TelemetryData {
	return TelemetryData{
		AppName:    appName,
		Version:    "v" + version,
		DeviceID:   deviceID,
		DeviceType: deviceType,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Event:      event,
	}
}

// send 发送统计数据，返回是否成功（统计服务返回非 2xx 视为失败）
func send(event string) bool {
	if !Enabled() {
		return false
	}

	jsonData, err := json.Marshal(collectData(event))
	if err != nil {
		return false
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(endpoint(), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// http.Post 不把 4xx/5xx 当作 error，必须校验状态码
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// SendEvent 发送一次事件上报（如赞赏点击 donate_support），同步返回是否成功
func SendEvent(event string) bool {
	return send(event)
}

// Start 启动定时统计
func Start() {
	if !Enabled() {
		return
	}
	go func() {
		// 启动时发送一次
		send("")

		// 每60分钟发送一次
		ticker := time.NewTicker(SendInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				send("")
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
