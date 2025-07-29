package api

import (
	"stream-demo/backend/middleware"
	"stream-demo/backend/utils"

	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	engine *gin.Engine

	// 處理器
	userHandler         *UserHandler
	videoHandler        *VideoHandler
	liveHandler         *LiveHandler
	paymentHandler      *PaymentHandler
	publicStreamHandler *PublicStreamHandler

	// 工具
	jwtUtil *utils.JWTUtil
}

// NewRouter 創建路由管理器
func NewRouter(
	engine *gin.Engine,
	userHandler *UserHandler,
	videoHandler *VideoHandler,
	liveHandler *LiveHandler,
	paymentHandler *PaymentHandler,
	publicStreamHandler *PublicStreamHandler,
	jwtUtil *utils.JWTUtil,
) *Router {
	return &Router{
		engine:              engine,
		userHandler:         userHandler,
		videoHandler:        videoHandler,
		liveHandler:         liveHandler,
		paymentHandler:      paymentHandler,
		publicStreamHandler: publicStreamHandler,
		jwtUtil:             jwtUtil,
	}
}

// SetupRoutes 設置所有路由
func (r *Router) SetupRoutes() {
	// 設置中間件
	r.setupMiddleware()

	// 設置公開路由
	r.setupPublicRoutes()

	// 設置認證路由
	r.setupAuthRoutes()

	// 設置 WebSocket 路由
	r.setupWebSocketRoutes()
}

// setupMiddleware 設置中間件
func (r *Router) setupMiddleware() {
	r.engine.Use(middleware.ErrorHandler())
	// 注意：CORS、Logger、Recovery 中間件需要另外實現或使用 gin 內建的
}

// setupPublicRoutes 設置公開路由
func (r *Router) setupPublicRoutes() {
	public := r.engine.Group("/api")
	{
		// 健康檢查
		public.GET("/health", r.healthCheck)

		// 用戶認證
		public.POST("/users/register", r.userHandler.Register)
		public.POST("/users/login", r.userHandler.Login)

		// 公開流路由
		if r.publicStreamHandler != nil {
			r.setupPublicStreamRoutes(public)
		}
	}
}

// setupAuthRoutes 設置認證路由
func (r *Router) setupAuthRoutes() {
	auth := r.engine.Group("/api")
	auth.Use(middleware.AuthMiddleware(r.jwtUtil))
	{
		// 用戶相關路由
		r.setupUserRoutes(auth)

		// 視頻相關路由
		r.setupVideoRoutes(auth)

		// 直播相關路由
		r.setupLiveRoutes(auth)

		// 支付相關路由
		r.setupPaymentRoutes(auth)
	}
}

// setupUserRoutes 設置用戶路由
func (r *Router) setupUserRoutes(group *gin.RouterGroup) {
	users := group.Group("/users")
	{
		users.GET("/:id", r.userHandler.GetUser)
		users.PUT("/:id", r.userHandler.UpdateUser)
		users.DELETE("/:id", r.userHandler.DeleteUser)
	}
}

// setupVideoRoutes 設置視頻路由
func (r *Router) setupVideoRoutes(group *gin.RouterGroup) {
	videos := group.Group("/videos")
	{
		videos.GET("", r.videoHandler.ListVideos)
		videos.POST("/upload-url", r.videoHandler.GenerateUploadURL)
		videos.POST("/confirm-upload", r.videoHandler.ConfirmUpload)
		videos.POST("", r.videoHandler.UploadVideo)
		videos.GET("/:id", r.videoHandler.GetVideo)
		videos.GET("/:id/transcode-status", r.videoHandler.GetVideoTranscodeStatus)
		videos.PUT("/:id", r.videoHandler.UpdateVideo)
		videos.DELETE("/:id", r.videoHandler.DeleteVideo)
		videos.GET("/search", r.videoHandler.SearchVideos)
		videos.POST("/:id/like", r.videoHandler.LikeVideo)
	}

	// 用戶視頻路由
	group.GET("/users/:id/videos", r.videoHandler.GetUserVideos)
}

// setupLiveRoutes 設置直播路由
func (r *Router) setupLiveRoutes(group *gin.RouterGroup) {
	lives := group.Group("/lives")
	{
		lives.GET("", r.liveHandler.ListLives)
		lives.POST("", r.liveHandler.CreateLive)
		lives.GET("/:id", r.liveHandler.GetLive)
		lives.PUT("/:id", r.liveHandler.UpdateLive)
		lives.DELETE("/:id", r.liveHandler.DeleteLive)
		lives.POST("/:id/start", r.liveHandler.StartLive)
		lives.POST("/:id/end", r.liveHandler.EndLive)
		lives.GET("/:id/stream-key", r.liveHandler.GetStreamKey)
		lives.POST("/:id/chat/toggle", r.liveHandler.ToggleChat)
	}

	// 用戶直播路由
	group.GET("/users/:id/lives", r.liveHandler.GetUserLives)
}

// setupPaymentRoutes 設置支付路由
func (r *Router) setupPaymentRoutes(group *gin.RouterGroup) {
	payments := group.Group("/payments")
	{
		payments.GET("", r.paymentHandler.ListPayments)
		payments.POST("", r.paymentHandler.CreatePayment)
		payments.GET("/:id", r.paymentHandler.GetPayment)
		payments.POST("/:id/process", r.paymentHandler.ProcessPayment)
		payments.POST("/:id/refund", r.paymentHandler.RefundPayment)
	}

	// 用戶支付路由
	group.GET("/users/:id/payments", r.paymentHandler.GetUserPayments)
}

// setupPublicStreamRoutes 設置公開流路由
func (r *Router) setupPublicStreamRoutes(group *gin.RouterGroup) {
	streams := group.Group("/public-streams")
	{
		streams.GET("", r.publicStreamHandler.GetAvailableStreams)
		streams.GET("/:name", r.publicStreamHandler.GetStreamInfo)
		streams.GET("/:name/url", r.publicStreamHandler.GetStreamURL)
		streams.GET("/:name/urls", r.publicStreamHandler.GetStreamURLs)
		streams.GET("/:name/stats", r.publicStreamHandler.GetStreamStats)
	}
}

// setupWebSocketRoutes 設置 WebSocket 路由
func (r *Router) setupWebSocketRoutes() {
	// WebSocket 路由將在 main.go 中設置
	// 因為需要 hub 實例
}

// healthCheck 健康檢查
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "服務正常運行",
	})
}
