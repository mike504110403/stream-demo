package main

import (
	"flag"
	"fmt"
	"os"
	"stream-demo/backend/api"
	"stream-demo/backend/config"
	"stream-demo/backend/database"
	"stream-demo/backend/di"
	"stream-demo/backend/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 命令行參數解析
	var (
		configFile = flag.String("config", "config/config.local.yaml", "配置文件路徑")
		env        = flag.String("env", "local", "運行環境")
		dbType     = flag.String("db", "", "資料庫類型 (mysql|postgresql)，不指定則使用配置文件默認值")
		showHelp   = flag.Bool("help", false, "顯示幫助信息")
	)
	flag.Parse()

	// 顯示幫助信息
	if *showHelp {
		showHelpInfo()
		return
	}

	// 初始化日誌系統
	utils.InitLogger()

	// 初始化配置
	cfg := config.NewConfig(*configFile, *env, *dbType)
	utils.LogInfo("🚀 串流平台後端服務啟動")
	utils.LogInfo("📂 配置文件: %s", *configFile)
	utils.LogInfo("🌍 運行環境: %s", *env)
	utils.LogInfo("🗄️  當前資料庫: %s", cfg.ActiveDatabase)

	// 執行資料庫遷移
	if err := database.MigratePostgreSQL(cfg); err != nil {
		utils.LogError("資料庫遷移失敗: %v", err)
		os.Exit(1)
	}
	utils.LogInfo("資料庫遷移完成")

	// 初始化依賴注入容器
	container, err := di.NewContainer(cfg)
	if err != nil {
		utils.LogError("初始化依賴注入容器失敗: %v", err)
		os.Exit(1)
	}
	utils.LogInfo("依賴注入容器初始化完成")

	// 啟動服務
	container.StartServices()
	utils.LogInfo("所有服務啟動完成")

	// 初始化 Gin 引擎
	r := gin.Default()

	// 初始化路由管理器
	router := api.NewRouter(
		r,
		container.UserHandler,
		container.VideoHandler,
		container.LiveHandler,
		container.PaymentHandler,
		container.PublicStreamHandler,
		container.JWTUtil,
	)

	// 設置路由
	router.SetupRoutes()

	// 設置 WebSocket 路由
	if container.WSHandler != nil {
		r.GET("/ws/:liveID", container.WSHandler.ServeWS)
	}

	// 啟動服務器
	addr := fmt.Sprintf(":%d", cfg.Gin.Port)
	utils.LogInfo("🌐 HTTP 服務器啟動在 %s", addr)

	if err := r.Run(addr); err != nil {
		utils.LogError("服務器啟動失敗: %v", err)
		os.Exit(1)
	}
}

// showHelpInfo 顯示幫助信息
func showHelpInfo() {
	fmt.Println("🚀 Stream Demo Backend - 串流平台後端服務")
	fmt.Println("")
	fmt.Println("用法:")
	fmt.Printf("  %s [選項]\n", os.Args[0])
	fmt.Println("")
	fmt.Println("選項:")
	fmt.Println("  -config string")
	fmt.Println("        配置文件路徑 (默認: config/config.local.yaml)")
	fmt.Println("  -env string")
	fmt.Println("        運行環境 (默認: local)")
	fmt.Println("  -db string")
	fmt.Println("        資料庫類型 mysql|postgresql (默認: 使用配置文件設定)")
	fmt.Println("  -help")
	fmt.Println("        顯示幫助信息")
	fmt.Println("")
	fmt.Println("範例:")
	fmt.Println("  go run main.go                    # 使用默認配置")
	fmt.Println("  go run main.go -db mysql          # 強制使用 MySQL")
	fmt.Println("  go run main.go -db postgresql     # 強制使用 PostgreSQL")
	fmt.Println("  go run main.go -config custom.yaml # 使用自定義配置文件")
	fmt.Println("")
	fmt.Println("環境變數:")
	fmt.Println("  DATABASE_TYPE=mysql|postgresql    # 設定資料庫類型")
	fmt.Println("")
}
