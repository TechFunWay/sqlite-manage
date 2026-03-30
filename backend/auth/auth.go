package auth

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "modernc.org/sqlite"
)

const (
	JWTSecret      = "sqlite-manager-secret-key-2024"
	JWTExpireHours = 24
	SaltValue      = "sqlite"
)

var systemDB *sql.DB

// SetDB 设置数据库连接（由外部初始化后传入）
func SetDB(db *sql.DB) {
	systemDB = db
}

// InitTables 初始化用户表和最近打开表
func InitTables() error {
	if systemDB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 用户表
	_, err := systemDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 最近打开数据库记录表
	_, err = systemDB.Exec(`
		CREATE TABLE IF NOT EXISTS recent_databases (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			name TEXT NOT NULL,
			last_opened DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

// HasAdmin 检查是否有管理员
func HasAdmin() bool {
	if systemDB == nil {
		return false
	}
	var count int
	systemDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count > 0
}

// MD5Hash 返回 MD5 哈希
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// HashPassword 哈希密码
func HashPassword(password string) string {
	return MD5Hash(password + SaltValue)
}

// CreateUser 创建用户
func CreateUser(username, password string) error {
	if systemDB == nil {
		return fmt.Errorf("database not initialized")
	}

	var count int
	systemDB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if count > 0 {
		return fmt.Errorf("user already exists")
	}

	hashedPassword := HashPassword(password)
	_, err := systemDB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
	return err
}

// ValidateUser 验证用户
func ValidateUser(username, password string) (*User, error) {
	if systemDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	user := &User{}
	hashedPassword := HashPassword(password)
	err := systemDB.QueryRow("SELECT id, username FROM users WHERE username = ? AND password = ?", username, hashedPassword).
		Scan(&user.ID, &user.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}
	return user, nil
}

// ChangePassword 修改密码
func ChangePassword(username, oldPassword, newPassword string) error {
	if systemDB == nil {
		return fmt.Errorf("database not initialized")
	}

	user, err := ValidateUser(username, oldPassword)
	if err != nil {
		return err
	}

	hashedPassword := HashPassword(newPassword)
	_, err = systemDB.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, user.ID)
	return err
}

// ResetPassword 重置密码（终端用）
func ResetPassword(username, newPassword string) error {
	if systemDB == nil {
		return fmt.Errorf("database not initialized")
	}

	var userID int
	err := systemDB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	hashedPassword := HashPassword(newPassword)
	_, err = systemDB.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, userID)
	return err
}

// GenerateToken 生成 JWT
func GenerateToken(user *User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * JWTExpireHours).Unix(),
	})
	return token.SignedString([]byte(JWTSecret))
}

// ValidateToken 验证 JWT
func ValidateToken(tokenString string) (*User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := &User{
			ID:       int(claims["user_id"].(float64)),
			Username: claims["username"].(string),
		}
		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetAllUsers 获取所有用户（终端用）
func GetAllUsers() ([]User, error) {
	if systemDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := systemDB.Query("SELECT id, username FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		rows.Scan(&user.ID, &user.Username)
		users = append(users, user)
	}
	return users, nil
}

// RecentDatabase 最近打开的数据库记录
type RecentDatabase struct {
	ID         int    `json:"id"`
	Path       string `json:"path"`
	Name       string `json:"name"`
	LastOpened string `json:"lastOpened"`
}

// AddRecentDatabase 添加最近打开的数据库
func AddRecentDatabase(path, name string) error {
	if systemDB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 先删除已存在的记录
	systemDB.Exec("DELETE FROM recent_databases WHERE path = ?", path)

	// 插入新记录
	_, err := systemDB.Exec("INSERT INTO recent_databases (path, name) VALUES (?, ?)", path, name)
	if err != nil {
		return err
	}

	// 只保留最近 20 条记录
	_, err = systemDB.Exec(`
		DELETE FROM recent_databases WHERE id NOT IN (
			SELECT id FROM recent_databases ORDER BY last_opened DESC LIMIT 20
		)
	`)
	return err
}

// GetRecentDatabases 获取最近打开的数据库列表
func GetRecentDatabases(limit int) ([]RecentDatabase, error) {
	if systemDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 10
	}

	rows, err := systemDB.Query("SELECT id, path, name, last_opened FROM recent_databases ORDER BY last_opened DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RecentDatabase
	for rows.Next() {
		var r RecentDatabase
		rows.Scan(&r.ID, &r.Path, &r.Name, &r.LastOpened)
		records = append(records, r)
	}
	return records, nil
}

// ClearRecentDatabases 清空最近打开的数据库记录
func ClearRecentDatabases() error {
	if systemDB == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := systemDB.Exec("DELETE FROM recent_databases")
	return err
}
