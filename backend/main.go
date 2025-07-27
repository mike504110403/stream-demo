package main

import (
	"flag"
	"fmt"
	"os"
	"stream-demo/backend/api"
	"stream-demo/backend/config"
	"stream-demo/backend/database"
	"stream-demo/backend/middleware"
	postgresqlRepo "stream-demo/backend/repositories/postgresql"
	"stream-demo/backend/services"
	"stream-demo/backend/utils"
	"stream-demo/backend/ws"
	"time"

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
		return
	}

	// 初始化日誌系統
	utils.InitLogger()

	// 初始化配置（傳入資料庫類型參數）
	cfg := config.NewConfig(*configFile, *env, *dbType)

	// 顯示當前配置信息
	utils.LogInfo("🚀 串流平台後端服務啟動")
	utils.LogInfo("📂 配置文件: %s", *configFile)
	utils.LogInfo("🌍 運行環境: %s", *env)
	utils.LogInfo("🗄️  當前資料庫: %s", cfg.ActiveDatabase)
	utils.LogInfo("📋 可用資料庫: %v", cfg.GetAvailableDatabases())

	// 執行資料庫遷移
	if err := database.MigratePostgreSQL(cfg); err != nil {
		utils.LogError("資料庫遷移失敗: %v", err)
	} else {
		utils.LogInfo("資料庫遷移完成")
	}

	// 初始化緩存系統
	var cache interface{}
	if cfg.Cache.Type == "redis" {
		cache = utils.NewRedisCache(cfg.Cache.DB, cfg.Cache.KeyPrefix, time.Duration(cfg.Cache.DefaultExpiration)*time.Second)
		utils.LogInfo("Redis緩存已初始化 (DB: %d)", cfg.Cache.DB)
	} else {
		cache = utils.NewPostgreSQLCache(cfg.DB["master"], cfg.Cache.TableName, time.Duration(cfg.Cache.DefaultExpiration)*time.Second, time.Duration(cfg.Cache.CleanupInterval)*time.Second)
		utils.LogInfo("PostgreSQL緩存已初始化")
	}

	// 初始化訊息系統（僅支援 Redis）
	var messaging *utils.RedisMessaging
	if cfg.Messaging.Type == "redis" {
		redisMessaging, err := utils.NewRedisMessaging(cfg.Messaging.DB)
		if err != nil {
			utils.LogError("初始化Redis訊息系統失敗: %v", err)
			messaging = nil
		} else {
			messaging = redisMessaging
			utils.LogInfo("Redis訊息系統已初始化 (DB: %d)", cfg.Messaging.DB)
		}
	} else {
		utils.LogError("不支援的訊息系統類型: %s，請使用 'redis'", cfg.Messaging.Type)
		messaging = nil
	}

	// 初始化WebSocket Hub
	hub := ws.NewHub(messaging)
	utils.LogInfo("WebSocket Hub已初始化")

	// 初始化WebSocket Handler
	wsHandler := ws.NewHandler(hub)

	// 初始化 JWT 工具
	jwtUtil := utils.NewJWTUtil(cfg.JWT.Secret)

	// 初始化服務層
	userService := services.NewUserService(cfg)
	videoService := services.NewVideoService(cfg)
	liveService, err := services.NewLiveService(cfg)
	if err != nil {
		utils.LogError("初始化直播服務失敗: %v", err)
		// 如果直播服務初始化失敗，創建一個基本的服務實例
		liveService = &services.LiveService{
			Conf:      cfg,
			Repo:      postgresqlRepo.NewPostgreSQLRepo(cfg.DB["master"]),
			RepoSlave: postgresqlRepo.NewPostgreSQLRepo(cfg.DB["slave"]),
		}
	}

	// 初始化公開流服務
	var publicStreamService *services.PublicStreamService
	var publicStreamHandler *api.PublicStreamHandler
	if redisCache, ok := cache.(*utils.RedisCache); ok {
		publicStreamService, err = services.NewPublicStreamService(cfg, redisCache)
		if err != nil {
			utils.LogError("初始化公開流服務失敗: %v", err)
		} else {
			utils.LogInfo("公開流服務初始化成功")
			publicStreamHandler = api.NewPublicStreamHandler(publicStreamService)
		}
	} else {
		utils.LogError("Redis 緩存未初始化，跳過公開流服務")
	}

	paymentService := services.NewPaymentService(cfg)

	// 初始化背景轉碼工作服務
	transcodeWorker := services.NewTranscodeWorker(videoService)
	transcodeWorker.Start()
	utils.LogInfo("背景轉碼工作服務已啟動")

	// 初始化Handler層
	userHandler := api.NewUserHandler(userService)
	videoHandler := api.NewVideoHandler(videoService)
	liveHandler := api.NewLiveHandler(liveService)
	paymentHandler := api.NewPaymentHandler(paymentService)

	// 初始化Gin引擎
	r := gin.Default()

	// 中間件
	r.Use(middleware.ErrorHandler())

	// 健康檢查端點
	r.GET("/health", func(c *gin.Context) {
		// 檢查資料庫連接
		dbOK := cfg.CheckDatabaseConnections()

		// 檢查緩存
		var cacheStats interface{}
		var cacheErr error
		if redisCache, ok := cache.(*utils.RedisCache); ok {
			cacheStats, cacheErr = redisCache.Stats()
		} else if pgCache, ok := cache.(*utils.PostgreSQLCache); ok {
			cacheStats, cacheErr = pgCache.Stats()
		}

		// 檢查Redis連接（如果使用Redis）
		var redisOK bool
		if cfg.Cache.Type == "redis" || cfg.Messaging.Type == "redis" {
			redisOK = utils.CheckRedisConnection() == nil
		} else {
			redisOK = true // 不使用Redis時默認為OK
		}

		// 檢查WebSocket統計
		roomStats := hub.GetRoomStats()

		// 檢查訊息系統統計
		var messagingStats interface{}
		if messaging != nil {
			messagingStats = messaging.GetStats()
		} else {
			messagingStats = map[string]interface{}{
				"type":   "none",
				"status": "not_initialized",
			}
		}

		// 獲取資料庫信息
		dbInfo := cfg.GetDatabaseInfo()

		status := "ok"
		if !dbOK || cacheErr != nil || !redisOK {
			status = "degraded"
		}

		c.JSON(200, gin.H{
			"status":          status,
			"timestamp":       time.Now().Format(time.RFC3339),
			"database":        dbOK,
			"database_info":   dbInfo,
			"redis":           redisOK,
			"cache_type":      cfg.Cache.Type,
			"cache_stats":     cacheStats,
			"messaging_type":  cfg.Messaging.Type,
			"messaging_stats": messagingStats,
			"room_stats":      roomStats,
		})
	})

	// 動態資料庫切換端點（僅開發環境使用）
	if *env == "local" || *env == "development" {
		r.POST("/admin/switch-database", func(c *gin.Context) {
			var req struct {
				DatabaseType string `json:"database_type" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
				return
			}

			utils.LogInfo("嘗試切換資料庫到: %s", req.DatabaseType)

			if err := cfg.SwitchDatabase(req.DatabaseType); err != nil {
				utils.LogError("資料庫切換失敗: %v", err)
				c.JSON(500, gin.H{"error": "Database switch failed", "details": err.Error()})
				return
			}

			// 重新初始化服務（使用新的資料庫連接）
			userService = services.NewUserService(cfg)
			videoService = services.NewVideoService(cfg)
			liveService, err = services.NewLiveService(cfg)
			if err != nil {
				utils.LogError("重新初始化直播服務失敗: %v", err)
				// 如果直播服務初始化失敗，創建一個基本的服務實例
				liveService = &services.LiveService{
					Conf:      cfg,
					Repo:      postgresqlRepo.NewPostgreSQLRepo(cfg.DB["master"]),
					RepoSlave: postgresqlRepo.NewPostgreSQLRepo(cfg.DB["slave"]),
				}
			}
			paymentService = services.NewPaymentService(cfg)

			// 重新初始化Handler
			userHandler = api.NewUserHandler(userService)
			videoHandler = api.NewVideoHandler(videoService)
			liveHandler = api.NewLiveHandler(liveService)
			paymentHandler = api.NewPaymentHandler(paymentService)

			utils.LogInfo("資料庫切換成功: %s", req.DatabaseType)

			c.JSON(200, gin.H{
				"message":         "Database switched successfully",
				"active_database": cfg.ActiveDatabase,
				"available":       cfg.GetAvailableDatabases(),
			})
		})

		// 獲取可用資料庫列表
		r.GET("/admin/databases", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"active":    cfg.ActiveDatabase,
				"available": cfg.GetAvailableDatabases(),
				"info":      cfg.GetDatabaseInfo(),
			})
		})
	}

	// 註冊路由
	api.RegisterRoutes(r, userHandler, videoHandler, liveHandler, paymentHandler, jwtUtil, publicStreamHandler)

	// 註冊WebSocket路由
	r.GET("/ws/:liveID", middleware.AuthMiddleware(jwtUtil), wsHandler.ServeWS)

	// 優雅關閉
	defer func() {
		if messaging != nil {
			messaging.Close()
		}
		if redisCache, ok := cache.(*utils.RedisCache); ok {
			redisCache.Close()
		}
		if hub != nil {
			hub.Close()
		}
		if cfg.Cache.Type == "redis" || cfg.Messaging.Type == "redis" {
			utils.CloseRedisClients()
		}
		utils.LogInfo("應用程式已關閉")
	}()

	// 啟動服務器
	port := fmt.Sprintf(":%d", cfg.Gin.Port)
	utils.LogInfo("🌐 服務器啟動在端口: %s", port)
	utils.LogInfo("📡 健康檢查: http://localhost%s/health", port)
	utils.LogInfo("🔌 WebSocket: ws://localhost%s/ws/{liveID}", port)
	if *env == "local" || *env == "development" {
		utils.LogInfo("🔧 管理端點:")
		utils.LogInfo("   GET  http://localhost%s/admin/databases", port)
		utils.LogInfo("   POST http://localhost%s/admin/switch-database", port)
	}

	if err := r.Run(port); err != nil {
		utils.LogError("服務器啟動失敗: %v", err)
	}
}
