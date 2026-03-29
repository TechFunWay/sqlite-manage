package upgrade

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"sqlite-manager/utils"
)

// Version 当前程序版本号
const Version = "1.0.0"

var systemDB *sql.DB

// SetDB 设置数据库连接（由外部初始化后传入）
func SetDB(db *sql.DB) {
	systemDB = db
}

// UpgradeScript 升级脚本定义
type UpgradeScript struct {
	Version string
	Func    func(db *sql.DB) error
}

// UpgradeRecord 升级记录
type UpgradeRecord struct {
	ID        int    `json:"id"`
	Version   string `json:"version"`
	AppliedAt string `json:"appliedAt"`
	Success   bool   `json:"success"`
}

// 升级脚本列表，按版本号排序
// 每个版本的升级脚本会按顺序执行
var upgradeScripts = []UpgradeScript{
	{
		Version: "1.1.0",
		Func: func(db *sql.DB) error {
			log.Println("Upgrading to v1.1.0: Adding example field")
			// 示例：添加字段
			// _, err := db.Exec("ALTER TABLE users ADD COLUMN email TEXT")
			return nil
		},
	},
	{
		Version: "1.2.0",
		Func: func(db *sql.DB) error {
			log.Println("Upgrading to v1.2.0: Creating example table")
			// 示例：创建新表
			// _, err := db.Exec(`
			//   CREATE TABLE IF NOT EXISTS settings (
			//     key TEXT PRIMARY KEY,
			//     value TEXT
			//   )
			// `)
			return nil
		},
	},
	{
		Version: "1.3.0",
		Func: func(db *sql.DB) error {
			log.Println("Upgrading to v1.3.0: Example data migration")
			// 示例：数据迁移
			return nil
		},
	},
}

// InitTables 初始化升级记录表
func InitTables() error {
	if systemDB == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := systemDB.Exec(`
		CREATE TABLE IF NOT EXISTS upgrade_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			success INTEGER DEFAULT 1
		)
	`)
	return err
}

// GetCurrentDBVersion 获取数据库当前版本
func GetCurrentDBVersion() string {
	if systemDB == nil {
		return "0.0.0"
	}

	var version string
	err := systemDB.QueryRow("SELECT version FROM upgrade_history WHERE success = 1 ORDER BY id DESC LIMIT 1").Scan(&version)
	if err != nil {
		return "0.0.0"
	}
	return version
}

// GetAppliedVersions 获取已应用的版本列表
func GetAppliedVersions() ([]UpgradeRecord, error) {
	if systemDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := systemDB.Query("SELECT id, version, applied_at, success FROM upgrade_history ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []UpgradeRecord
	for rows.Next() {
		var r UpgradeRecord
		var success int
		err := rows.Scan(&r.ID, &r.Version, &r.AppliedAt, &success)
		if err != nil {
			return nil, err
		}
		r.Success = success == 1
		records = append(records, r)
	}
	return records, nil
}

// RecordUpgrade 记录升级
func RecordUpgrade(version string, success bool) error {
	if systemDB == nil {
		return fmt.Errorf("database not initialized")
	}

	successInt := 0
	if success {
		successInt = 1
	}

	_, err := systemDB.Exec("INSERT INTO upgrade_history (version, success) VALUES (?, ?)", version, successInt)
	return err
}

// CompareVersions 比较版本号
// 返回: -1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
func CompareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}
	return 0
}

// RunUpgrade 执行升级
func RunUpgrade() error {
	if systemDB == nil {
		log.Println("Database not initialized, skipping upgrade")
		return nil
	}

	currentDBVersion := GetCurrentDBVersion()
	appVersion := Version

	log.Printf("Current DB version: %s, App version: %s", currentDBVersion, appVersion)

	// 如果版本相同，无需升级
	if CompareVersions(currentDBVersion, appVersion) >= 0 {
		log.Println("Database is up to date")
		return nil
	}

	// 如果数据库版本是0.0.0，先记录当前版本
	if currentDBVersion == "0.0.0" {
		log.Println("First run, recording current version")
		return RecordUpgrade(appVersion, true)
	}

	// 收集需要执行的升级脚本
	var pendingScripts []UpgradeScript
	for _, script := range upgradeScripts {
		if CompareVersions(script.Version, currentDBVersion) > 0 &&
			CompareVersions(script.Version, appVersion) <= 0 {
			pendingScripts = append(pendingScripts, script)
		}
	}

	if len(pendingScripts) == 0 {
		log.Println("No upgrade scripts to execute")
		return RecordUpgrade(appVersion, true)
	}

	// 按版本号排序
	sort.Slice(pendingScripts, func(i, j int) bool {
		return CompareVersions(pendingScripts[i].Version, pendingScripts[j].Version) < 0
	})

	// 执行升级脚本
	log.Printf("Found %d upgrade scripts to execute", len(pendingScripts))

	// 开始事务
	tx, err := systemDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, script := range pendingScripts {
		log.Printf("Executing upgrade to v%s...", script.Version)

		// 执行升级脚本
		if err := script.Func(systemDB); err != nil {
			log.Printf("Upgrade to v%s failed: %v", script.Version, err)
			RecordUpgrade(script.Version, false)
			return fmt.Errorf("upgrade to v%s failed: %w", script.Version, err)
		}

		// 记录升级成功
		_, err := tx.Exec("INSERT INTO upgrade_history (version, success, applied_at) VALUES (?, 1, ?)",
			script.Version, utils.GetBeijingTimeString())
		if err != nil {
			return fmt.Errorf("failed to record upgrade: %w", err)
		}

		log.Printf("Upgrade to v%s completed successfully", script.Version)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit upgrade: %w", err)
	}

	log.Printf("All upgrades completed. Database is now at version %s", appVersion)
	return nil
}
