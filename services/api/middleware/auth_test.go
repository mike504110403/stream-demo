package middleware

import (
	"net/http"
	"net/http/httptest"
	"stream-demo/backend/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := RequireRole("admin")

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "權限不足")
}

func TestRequireRole_CorrectRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := RequireRole("admin")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestRequireRole_WrongRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := RequireRole("admin")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	})
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "權限不足")
}

// 測試 JWT 工具的基本功能
func TestJWTUtil_Basic(t *testing.T) {
	jwtUtil := utils.NewJWTUtil("test-secret")

	// 測試生成令牌
	token, err := jwtUtil.GenerateToken(1, "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 測試驗證令牌
	claims, err := jwtUtil.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "user", claims.Role)
}

// 測試無效令牌
func TestJWTUtil_InvalidToken(t *testing.T) {
	jwtUtil := utils.NewJWTUtil("test-secret")

	// 測試無效令牌
	_, err := jwtUtil.ValidateToken("invalid-token")
	assert.Error(t, err)
}
