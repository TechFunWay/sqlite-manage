package fnos

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"sqlite-manager/auth"
)

type Handler struct {
	db *sql.DB
}

type bindRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func InitTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS fnos_bindings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fnos_user_id TEXT NOT NULL UNIQUE,
			fnos_username TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	return err
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/identity", h.identity)
	group.POST("/login", h.login)
	group.POST("/bind", h.bind)
	group.POST("/register", h.register)
}

func (h *Handler) identity(c *gin.Context) {
	identity, ok := identityFromRequest(c.Request)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "仅支持从飞牛应用中心入口登录"})
		return
	}

	var userID int
	err := h.db.QueryRow("SELECT user_id FROM fnos_bindings WHERE fnos_user_id = ?", identity.UserID).Scan(&userID)
	bound := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取飞牛账号绑定失败"})
		return
	}

	var count int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取本地账号失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fnosUsername": identity.Username,
		"isAdmin":      identity.IsAdmin,
		"bound":        bound,
		"hasAdmin":     count > 0,
	})
}

func (h *Handler) login(c *gin.Context) {
	identity, ok := identityFromRequest(c.Request)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "仅支持从飞牛应用中心入口登录"})
		return
	}

	user := &auth.User{}
	err := h.db.QueryRow(`
		SELECT u.id, u.username
		FROM fnos_bindings b
		JOIN users u ON u.id = b.user_id
		WHERE b.fnos_user_id = ?
	`, identity.UserID).Scan(&user.ID, &user.Username)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusConflict, gin.H{"error": "当前飞牛账号尚未关联本地管理员", "code": "FNOS_BINDING_REQUIRED"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "飞牛登录失败"})
		return
	}

	h.respondWithToken(c, user, "飞牛 NAS 登录成功")
}

func (h *Handler) bind(c *gin.Context) {
	identity, ok := identityFromRequest(c.Request)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "仅支持从飞牛应用中心入口登录"})
		return
	}

	var req bindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入本地管理员账号和密码"})
		return
	}

	user, err := auth.ValidateUser(strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "本地管理员账号或密码错误"})
		return
	}

	if err := h.insertBinding(identity, user.ID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	h.respondWithToken(c, user, "飞牛账号关联成功")
}

func (h *Handler) register(c *gin.Context) {
	identity, ok := identityFromRequest(c.Request)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "仅支持从飞牛应用中心入口登录"})
		return
	}

	var req bindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入本地管理员账号和密码"})
		return
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入本地管理员账号"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建本地管理员失败"})
		return
	}
	defer tx.Rollback()

	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil || count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "管理员账号已存在，请改为关联现有账号"})
		return
	}
	result, err := tx.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, auth.HashPassword(req.Password))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "创建本地管理员失败"})
		return
	}
	userID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建本地管理员失败"})
		return
	}
	if _, err := tx.Exec("INSERT INTO fnos_bindings (fnos_user_id, fnos_username, user_id) VALUES (?, ?, ?)", identity.UserID, identity.Username, userID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "飞牛账号已经关联过本地账号"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存账号关联失败"})
		return
	}

	h.respondWithToken(c, &auth.User{ID: int(userID), Username: username}, "管理员创建并关联成功")
}

func (h *Handler) insertBinding(identity gatewayIdentity, userID int) error {
	var existingUserID int
	err := h.db.QueryRow("SELECT user_id FROM fnos_bindings WHERE fnos_user_id = ?", identity.UserID).Scan(&existingUserID)
	if err == nil {
		if existingUserID == userID {
			_, updateErr := h.db.Exec("UPDATE fnos_bindings SET fnos_username = ?, updated_at = CURRENT_TIMESTAMP WHERE fnos_user_id = ?", identity.Username, identity.UserID)
			return updateErr
		}
		return errors.New("当前飞牛账号已关联其他本地账号")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return errors.New("读取飞牛账号关联失败")
	}
	_, err = h.db.Exec("INSERT INTO fnos_bindings (fnos_user_id, fnos_username, user_id) VALUES (?, ?, ?)", identity.UserID, identity.Username, userID)
	if err != nil {
		return errors.New("飞牛账号关联失败")
	}
	return nil
}

func (h *Handler) respondWithToken(c *gin.Context, user *auth.User, message string) {
	token, err := auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成登录凭证失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": message, "token": token, "user": user})
}
