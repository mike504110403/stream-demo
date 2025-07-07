package main

import (
	"fmt"
	"stream-demo/backend/api"
	"stream-demo/backend/config"
	"stream-demo/backend/database"
	"stream-demo/backend/middleware"
	"stream-demo/backend/services"
	"stream-demo/backend/utils"
	"stream-demo/backend/ws"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日誌
	utils.InitLogger()
	utils.LogInfo("啟動應用程式...")

	// 載入配置（使用PostgreSQL）
	cfg := config.NewPostgreSQLConfig("config/config.local.yaml", "local")
	utils.LogInfo("載入配置成功，環境: %s", cfg.Configurations.Gin.Mode)

	// 執行資料庫遷移
	if err := database.MigratePostgreSQL(cfg); err != nil {
		utils.LogError("資料庫遷移失敗: %v", err)
		return
	}
	utils.LogInfo("資料庫遷移完成")

	// 初始化PostgreSQL緩存
	cache := utils.NewPostgreSQLCache(
		cfg.DB["master"],
		cfg.Cache.TableName,
		time.Duration(cfg.Cache.DefaultExpiration)*time.Second,
		time.Duration(cfg.Cache.CleanupInterval)*time.Second,
	)
	utils.LogInfo("PostgreSQL緩存已初始化")

	// 初始化PostgreSQL訊息系統
	messaging, err := utils.NewPostgreSQLMessaging(cfg.DB["master"])
	if err != nil {
		utils.LogError("初始化PostgreSQL訊息系統失敗: %v", err)
		// 不中斷啟動，只記錄錯誤
		messaging = nil
	} else {
		utils.LogInfo("PostgreSQL訊息系統已初始化")
	}

	// 初始化WebSocket Hub
	hub := ws.NewHub(messaging)
	utils.LogInfo("WebSocket Hub已初始化")

	// 初始化 JWT 工具
	jwtUtil := utils.NewJWTUtil(cfg.JWT.Secret)

	// 初始化 Service (傳入 DatabaseManager)
	userService := services.NewUserService(cfg)
	videoService := services.NewVideoService(cfg)
	liveService := services.NewLiveService(cfg)
	paymentService := services.NewPaymentService(cfg)

	// 初始化 Handler
	userHandler := api.NewUserHandler(userService)
	videoHandler := api.NewVideoHandler(videoService)
	liveHandler := api.NewLiveHandler(liveService)
	paymentHandler := api.NewPaymentHandler(paymentService)
	wsHandler := ws.NewHandler(hub)

	// 初始化 Gin
	r := gin.New()

	// 使用中間件
	r.Use(middleware.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(gin.Logger())

	// 添加健康檢查端點
	r.GET("/health", func(c *gin.Context) {
		// 檢查資料庫連接
		dbOK := cfg.CheckPostgreSQLConnections()

		// 檢查緩存
		cacheStats, cacheErr := cache.Stats()

		// 檢查WebSocket統計
		roomStats := hub.GetRoomStats()

		status := "ok"
		if !dbOK || cacheErr != nil {
			status = "degraded"
		}

		c.JSON(200, gin.H{
			"status":      status,
			"timestamp":   time.Now().Format(time.RFC3339),
			"database":    dbOK,
			"cache_stats": cacheStats,
			"room_stats":  roomStats,
			"messaging":   messaging != nil,
		})
	})

	// 註冊路由
	api.RegisterRoutes(r, userHandler, videoHandler, liveHandler, paymentHandler, jwtUtil)

	// 註冊WebSocket路由
	r.GET("/ws/:liveID", middleware.AuthMiddleware(jwtUtil), wsHandler.ServeWS)

	// 優雅關閉處理
	defer func() {
		if messaging != nil {
			messaging.Close()
		}
		hub.Close()
		utils.LogInfo("應用程式已關閉")
	}()

	// 啟動伺服器
	addr := fmt.Sprintf("%s:%d", cfg.Configurations.Gin.Host, cfg.Configurations.Gin.Port)
	utils.LogInfo("伺服器啟動於 %s", addr)
	utils.LogInfo("WebSocket端點: ws://%s/ws/{liveID}", addr)
	utils.LogInfo("健康檢查端點: http://%s/health", addr)

	if err := r.Run(addr); err != nil {
		utils.LogError("伺服器啟動失敗: %v", err)
	}
}
