package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware 跨域中間件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允許的來源
		c.Header("Access-Control-Allow-Origin", "*")
		// 允許的方法
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		// 允許的標頭
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-CSRF-Token, Cache-Control, Pragma")
		// 允許憑證
		c.Header("Access-Control-Allow-Credentials", "true")
		// 暴露的標頭
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		// 處理預檢請求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
