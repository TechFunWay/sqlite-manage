package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
	"sqlite-manager/auth"
	"sqlite-manager/config"
	"sqlite-manager/handlers"
	"sqlite-manager/telemetry"
	"sqlite-manager/upgrade"
	"sqlite-manager/utils"
)

// AppVersion 应用版本号
const AppVersion = upgrade.Version

// 全局配置
var (
	systemDB   *sql.DB
	serverPort string
)

func main() {
	// 解析命令行参数
	port := flag.String("port", "", "服务端口 (默认: 8080)")
	dataDir := flag.String("data-dir", "", "数据目录 (默认: ./data)")
	webDir := flag.String("web-dir", "", "静态资源目录 (默认: ./public)")
	uploadDir := flag.String("upload-dir", "", "上传目录 (默认: ./upload)")
	shareDirs := flag.String("share-dirs", "", "共享目录 (冒号分隔)")
	noBrowser := flag.Bool("no-browser", false, "不自动打开浏览器")
	flag.Parse()

	// 子命令处理
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		switch os.Args[1] {
		case "reset-password":
			parseConfig(port, dataDir, webDir, uploadDir)
			initSystemDB()
			resetPasswordCLI()
			return
		case "version":
			fmt.Printf("SQLite Manager v%s\n", AppVersion)
			return
		case "upgrade-status":
			parseConfig(port, dataDir, webDir, uploadDir)
			initSystemDB()
			showUpgradeStatus()
			return
		case "help":
			printHelp()
			return
		}
	}

	// 解析配置
	parseConfig(port, dataDir, webDir, uploadDir)

	// 初始化时区
	utils.InitTimezone()

	// 保存打开浏览器标志
	shouldOpenBrowser := !*noBrowser

	// 设置共享目录
	if *shareDirs != "" {
		config.SetShareDirs(*shareDirs)
	} else if envShare := os.Getenv("TRIM_DATA_SHARE_PATHS"); envShare != "" {
		config.SetShareDirs(envShare)
	}

	// 自动检测飞牛存储卷
	detectVolumes()

	// 获取本机IP地址
	localIPs := getLocalIPs()

	// 打印启动信息
	log.Println("==========================================")
	log.Printf("  SQLite Manager v%s", AppVersion)
	log.Println("==========================================")
	log.Printf("  时间: %s", utils.GetBeijingTimeString())
	log.Printf("  端口: %s", serverPort)
	log.Println("  访问地址:")
	log.Printf("    http://localhost:%s", serverPort)
	for _, ip := range localIPs {
		log.Printf("    http://%s:%s", ip, serverPort)
	}
	log.Printf("  数据目录: %s", config.DataDir)
	log.Printf("  静态资源: %s", config.PublicDir)
	log.Printf("  上传目录: %s", config.UploadDir)
	log.Printf("  共享目录: %v", config.ShareDirs)
	log.Printf("  系统数据库: %s", config.GetSystemDBPath())
	log.Printf("  打开浏览器: %v", shouldOpenBrowser)
	log.Println("==========================================")

	// 初始化系统数据库
	if err := initSystemDB(); err != nil {
		log.Fatalf("Failed to initialize system database: %v", err)
	}
	defer systemDB.Close()

	// 初始化用户表
	auth.SetDB(systemDB)
	if err := auth.InitTables(); err != nil {
		log.Fatalf("Failed to initialize auth tables: %v", err)
	}

	// 初始化升级表
	upgrade.SetDB(systemDB)
	if err := upgrade.InitTables(); err != nil {
		log.Fatalf("Failed to initialize upgrade tables: %v", err)
	}

	// 执行升级检查
	if err := upgrade.RunUpgrade(); err != nil {
		log.Printf("Warning: Upgrade failed: %v", err)
	}

	// 初始化并启动统计（静默，每60分钟上报一次）
	telemetry.Init(AppVersion)
	telemetry.Start()
	defer telemetry.Stop()

	// 启动服务器
	startServer(shouldOpenBrowser)
}

// parseConfig 解析配置
func parseConfig(port, dataDirParam, webDirParam, uploadDirParam *string) {
	// 端口
	if *port != "" {
		serverPort = *port
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		serverPort = envPort
	} else {
		serverPort = "8903"
	}

	// 获取可执行文件所在目录
	execDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	// 数据目录
	dir := ""
	if *dataDirParam != "" {
		dir, _ = filepath.Abs(*dataDirParam)
	} else if envData := os.Getenv("SQLITE_DATA_DIR"); envData != "" {
		dir, _ = filepath.Abs(envData)
	} else {
		dir = filepath.Join(execDir, "data")
	}
	config.DataDir = dir

	// 静态资源目录
	if *webDirParam != "" {
		dir, _ = filepath.Abs(*webDirParam)
	} else if envWeb := os.Getenv("SQLITE_WEB_DIR"); envWeb != "" {
		dir, _ = filepath.Abs(envWeb)
	} else {
		dir = filepath.Join(execDir, "public")
	}
	config.PublicDir = dir

	// 上传目录
	if *uploadDirParam != "" {
		dir, _ = filepath.Abs(*uploadDirParam)
	} else if envUpload := os.Getenv("SQLITE_UPLOAD_DIR"); envUpload != "" {
		dir, _ = filepath.Abs(envUpload)
	} else {
		dir = filepath.Join(execDir, "upload")
	}
	config.UploadDir = dir
}

// getLocalIPs 获取本机所有非回环IP地址
func getLocalIPs() []string {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 只保留IPv4地址
			if ip != nil && ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}

	return ips
}

// detectVolumes 自动检测飞牛存储卷和共享目录
func detectVolumes() {
	// 如果已经有共享目录配置，跳过
	if len(config.ShareDirs) > 0 {
		return
	}

	var volumes []string

	// 检测飞牛存储卷 (/vol1, /vol2, ...)
	for i := 1; i <= 10; i++ {
		volPath := fmt.Sprintf("/vol%d", i)
		if info, err := os.Stat(volPath); err == nil && info.IsDir() {
			volumes = append(volumes, volPath)
		}
	}

	// 检测常见目录
	commonDirs := []string{
		"/mnt",
		"/media",
		"/share",
		"/shares",
	}

	for _, dir := range commonDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// 检查是否已添加
			exists := false
			for _, v := range volumes {
				if v == dir {
					exists = true
					break
				}
			}
			if !exists {
				volumes = append(volumes, dir)
			}
		}
	}

	// 如果找到目录，设置为共享目录
	if len(volumes) > 0 {
		config.SetShareDirs(strings.Join(volumes, ":"))
		log.Printf("Detected volumes: %v", volumes)
	}
}

// initSystemDB 初始化系统数据库
func initSystemDB() error {
	systemDBPath := config.GetSystemDBPath()
	os.MkdirAll(filepath.Dir(systemDBPath), 0755)

	var err error
	systemDB, err = sql.Open("sqlite", systemDBPath+"?_journal_mode=WAL")
	if err != nil {
		return err
	}

	return systemDB.Ping()
}

// startServer 启动服务器
func startServer(shouldOpenBrowser bool) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 静态资源
	publicDir := config.GetPublicDir()
	r.Static("/sqlite-web", filepath.Join(publicDir, "sqlite-web"))
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(publicDir, "index.html"))
	})

	// SPA 路由
	spaRoutes := []string{"/login", "/register", "/database"}
	for _, route := range spaRoutes {
		r.GET(route, func(c *gin.Context) {
			c.File(filepath.Join(publicDir, "index.html"))
		})
	}

	// Auth API (不需要认证)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", handlers.Login)
		authGroup.POST("/register", handlers.Register)
		authGroup.GET("/check", handlers.CheckAuth)
	}

	// 系统 API (不需要认证)
	api := r.Group("/api")
	api.GET("/version", handlers.GetVersion)
	api.GET("/upgrade-status", handlers.GetUpgradeStatus)

	// 需要认证的 API
	protectedAPI := r.Group("/api")
	protectedAPI.Use(auth.AuthMiddleware())
	{
		protectedAPI.POST("/auth/change-password", handlers.ChangePassword)
		protectedAPI.POST("/auth/logout", handlers.Logout)

		protectedAPI.GET("/recent-databases", handlers.GetRecentDatabases)
		protectedAPI.POST("/recent-databases", handlers.AddRecentDatabase)
		protectedAPI.DELETE("/recent-databases", handlers.ClearRecentDatabases)

		protectedAPI.GET("/files/browse", handlers.BrowseFiles)
		protectedAPI.GET("/files/shares", handlers.GetShareDirs)

		protectedAPI.POST("/database/open", handlers.OpenDatabase)
		protectedAPI.POST("/database/create", handlers.CreateDatabase)
		protectedAPI.POST("/database/upload", handlers.UploadDatabase)
		protectedAPI.GET("/databases", handlers.GetAllDatabases)
		protectedAPI.GET("/database/info", handlers.GetDatabaseInfo)
		protectedAPI.PUT("/databases/:id/activate", handlers.SetActiveDatabase)
		protectedAPI.DELETE("/databases/:id", handlers.CloseDatabase)

		protectedAPI.GET("/tables", handlers.GetTables)
		protectedAPI.GET("/tables/:name/schema", handlers.GetTableSchema)
		protectedAPI.POST("/tables", handlers.CreateTable)
		protectedAPI.DELETE("/tables/:name", handlers.DropTable)
		protectedAPI.PUT("/tables/rename", handlers.RenameTable)
		protectedAPI.POST("/tables/:name/columns", handlers.AddColumn)
		protectedAPI.DELETE("/tables/:name/columns/:column", handlers.DropColumn)

		protectedAPI.GET("/tables/:name/indexes", handlers.GetIndexes)
		protectedAPI.POST("/tables/:name/indexes", handlers.CreateIndex)
		protectedAPI.DELETE("/tables/:name/indexes/:name", handlers.DropIndex)

		protectedAPI.GET("/tables/:name/data", handlers.GetTableData)
		protectedAPI.POST("/tables/:name/data", handlers.InsertRow)
		protectedAPI.PUT("/tables/:name/data", handlers.UpdateRow)
		protectedAPI.DELETE("/tables/:name/data", handlers.DeleteRow)

		protectedAPI.GET("/tables/:name/primarykey", handlers.GetPrimaryKey)
		protectedAPI.POST("/query", handlers.ExecuteQuery)
	}

	// SPA fallback
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/sqlite-web") {
			filePath := filepath.Join(publicDir, path)
			if _, err := os.Stat(filePath); err == nil {
				c.File(filePath)
				return
			}
			c.File(filepath.Join(publicDir, "index.html"))
			return
		}
	})

	// 确保上传目录存在
	os.MkdirAll(filepath.Join(config.GetUploadDir(), "db"), 0755)

	log.Printf("SQLite Manager v%s starting on http://localhost:%s", AppVersion, serverPort)
	log.Printf("System database: %s", config.GetSystemDBPath())
	log.Printf("Admin exists: %v", auth.HasAdmin())

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		if systemDB != nil {
			systemDB.Close()
		}
		os.Exit(0)
	}()

	// 打开浏览器
	if shouldOpenBrowser {
		go func() {
			openBrowser("http://localhost:" + serverPort)
		}()
	}

	if err := r.Run(":" + serverPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// printHelp 打印帮助
func printHelp() {
	fmt.Printf(`SQLite Manager v%s

用法:
  sqlite-manager [选项] [命令]

命令:
  version          显示版本号
  reset-password   重置管理员密码
  upgrade-status   显示升级状态
  help             显示帮助信息

选项:
  -port string           服务端口 (默认: 8080，也可用 PORT 环境变量)
  -data-dir string       数据目录 (默认: ./data，也可用 SQLITE_DATA_DIR 环境变量)
  -web-dir string        静态资源目录 (默认: ./public，也可用 SQLITE_WEB_DIR 环境变量)
  -upload-dir string     上传目录 (默认: ./upload，也可用 SQLITE_UPLOAD_DIR 环境变量)
  -no-browser            不自动打开浏览器

示例:
  sqlite-manager                                 # 使用默认配置启动
  sqlite-manager -port 3000                      # 指定端口 3000
  sqlite-manager -data-dir /var/lib/sqlite       # 指定数据目录

环境变量:
  PORT                 服务端口
  SQLITE_DATA_DIR      数据目录
  SQLITE_WEB_DIR       静态资源目录
  SQLITE_UPLOAD_DIR    上传目录
`, AppVersion)
}

// showUpgradeStatus 显示升级状态
func showUpgradeStatus() {
	fmt.Printf("SQLite Manager v%s\n", AppVersion)
	fmt.Printf("System database: %s\n", config.GetSystemDBPath())

	upgrade.SetDB(systemDB)

	records, err := upgrade.GetAppliedVersions()
	if err != nil {
		fmt.Println("Failed to get upgrade history:", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("No upgrade history found")
		return
	}

	fmt.Println("\nUpgrade History:")
	fmt.Println("----------------")
	for _, r := range records {
		status := "✓"
		if !r.Success {
			status = "✗"
		}
		fmt.Printf("  %s v%s - %s\n", status, r.Version, r.AppliedAt)
	}
}

// resetPasswordCLI 终端重置密码
func resetPasswordCLI() {
	fmt.Println("=== SQLite Manager 密码重置工具 ===")
	fmt.Println()

	auth.SetDB(systemDB)
	auth.InitTables()

	users, err := auth.GetAllUsers()
	if err != nil {
		fmt.Println("获取用户列表失败:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	if len(users) == 0 {
		fmt.Println("当前没有任何管理员账号。")
		fmt.Print("是否创建新的管理员账号？(y/n): ")
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(confirm)

		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
			fmt.Println("已取消")
			return
		}

		fmt.Print("请输入用户名: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		if username == "" {
			fmt.Println("用户名不能为空")
			return
		}

		fmt.Print("请输入密码: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		if password == "" {
			fmt.Println("密码不能为空")
			return
		}

		fmt.Print("请再次输入密码: ")
		confirmPassword, _ := reader.ReadString('\n')
		confirmPassword = strings.TrimSpace(confirmPassword)

		if password != confirmPassword {
			fmt.Println("两次输入的密码不一致")
			return
		}

		md5Password := auth.MD5Hash(password)
		if err := auth.CreateUser(username, md5Password); err != nil {
			fmt.Println("创建用户失败:", err)
			return
		}

		fmt.Println()
		fmt.Println("管理员账号创建成功！")
		fmt.Printf("用户名: %s\n", username)
		return
	}

	fmt.Println("当前用户列表:")
	for i, user := range users {
		fmt.Printf("  %d. %s\n", i+1, user.Username)
	}
	fmt.Println()

	fmt.Print("请输入要重置密码的用户名: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Println("用户名不能为空")
		return
	}

	fmt.Print("请输入新密码: ")
	newPassword, _ := reader.ReadString('\n')
	newPassword = strings.TrimSpace(newPassword)

	if newPassword == "" {
		fmt.Println("密码不能为空")
		return
	}

	fmt.Print("请再次输入新密码: ")
	confirmPassword, _ := reader.ReadString('\n')
	confirmPassword = strings.TrimSpace(confirmPassword)

	if newPassword != confirmPassword {
		fmt.Println("两次输入的密码不一致")
		return
	}

	md5Password := auth.MD5Hash(newPassword)
	if err := auth.ResetPassword(username, md5Password); err != nil {
		fmt.Println("重置密码失败:", err)
		return
	}

	fmt.Println()
	fmt.Println("密码重置成功！")
	fmt.Println("请使用新密码登录。")
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}
