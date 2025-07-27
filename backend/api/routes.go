package api

import (
	"stream-demo/backend/middleware"
	"stream-demo/backend/utils"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 註冊所有路由
func RegisterRoutes(r *gin.Engine, userHandler *UserHandler, videoHandler *VideoHandler, liveHandler *LiveHandler, paymentHandler *PaymentHandler, jwtUtil *utils.JWTUtil, publicStreamHandler *PublicStreamHandler) {
	// 公開路由
	public := r.Group("/api")
	{
		public.POST("/users/register", userHandler.Register)
		public.POST("/users/login", userHandler.Login)

		// 公開流路由（不需要認證）
		if publicStreamHandler != nil {
			RegisterPublicStreamRoutes(public, publicStreamHandler)
		}
	}

	// 需要認證的路由
	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware(jwtUtil))
	{
		// 用戶相關
		auth.GET("/users/:id", userHandler.GetUser)
		auth.PUT("/users/:id", userHandler.UpdateUser)
		auth.DELETE("/users/:id", userHandler.DeleteUser)

		// 影片相關
		auth.GET("/videos", videoHandler.ListVideos)
		auth.POST("/videos/upload-url", videoHandler.GenerateUploadURL) // S3預簽名上傳URL
		auth.POST("/videos/confirm-upload", videoHandler.ConfirmUpload) // 確認上傳完成
		auth.POST("/videos", videoHandler.UploadVideo)                  // 傳統表單上傳（向下相容）
		auth.GET("/videos/:id", videoHandler.GetVideo)
		auth.GET("/videos/:id/transcode-status", videoHandler.GetVideoTranscodeStatus) // 轉碼狀態檢查
		auth.GET("/users/:id/videos", videoHandler.GetUserVideos)
		auth.PUT("/videos/:id", videoHandler.UpdateVideo)
		auth.DELETE("/videos/:id", videoHandler.DeleteVideo)
		auth.GET("/videos/search", videoHandler.SearchVideos)
		auth.POST("/videos/:id/like", videoHandler.LikeVideo)

		// 直播相關
		auth.GET("/lives", liveHandler.ListLives)
		auth.POST("/lives", liveHandler.CreateLive)
		auth.GET("/lives/:id", liveHandler.GetLive)
		auth.GET("/users/:id/lives", liveHandler.GetUserLives)
		auth.PUT("/lives/:id", liveHandler.UpdateLive)
		auth.DELETE("/lives/:id", liveHandler.DeleteLive)
		auth.POST("/lives/:id/start", liveHandler.StartLive)
		auth.POST("/lives/:id/end", liveHandler.EndLive)
		auth.GET("/lives/:id/stream-key", liveHandler.GetStreamKey)
		auth.POST("/lives/:id/chat/toggle", liveHandler.ToggleChat)

		// 支付相關
		auth.GET("/payments", paymentHandler.ListPayments)
		auth.POST("/payments", paymentHandler.CreatePayment)
		auth.GET("/payments/:id", paymentHandler.GetPayment)
		auth.GET("/users/:id/payments", paymentHandler.GetUserPayments)
		auth.POST("/payments/:id/process", paymentHandler.ProcessPayment)
		auth.POST("/payments/:id/refund", paymentHandler.RefundPayment)
	}
}
