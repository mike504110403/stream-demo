package middleware

import (
	"net/http"
	"stream-demo/backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 認證中間件
func AuthMiddleware(jwtUtil *utils.JWTUtil) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供認證令牌"})
			c.Abort()
			return
		}

		// 檢查 Bearer 前綴
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "無效的認證令牌格式"})
			c.Abort()
			return
		}

		// 驗證令牌
		claims, err := jwtUtil.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "無效的認證令牌"})
			c.Abort()
			return
		}

		// 將用戶 ID 和角色存儲在上下文中
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// 檢查使用者角色
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func JWTAuthMiddleware(jwtUtil *utils.JWTUtil) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供有效的授權資訊"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtUtil.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 驗證失敗"})
			c.Abort()
			return
		}

		// 將 claims 存入 context，方便後續 handler 取得
		c.Set("claims", claims)
		c.Next()
	}
}
