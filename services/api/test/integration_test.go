package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthCheck 測試健康檢查端點
func TestHealthCheck(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)
	
	// 創建路由
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "Service is healthy",
		})
	})
	
	// 創建測試請求
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)
	
	// 創建響應記錄器
	w := httptest.NewRecorder()
	
	// 執行請求
	router.ServeHTTP(w, req)
	
	// 檢查響應
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 解析響應體
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Service is healthy", response["message"])
}

// TestUserRegistration 測試用戶註冊
func TestUserRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.POST("/api/v1/users/register", func(c *gin.Context) {
		var request struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}
		
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		
		// 模擬成功註冊
		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"user": gin.H{
				"username": request.Username,
				"email":    request.Email,
			},
		})
	})
	
	// 測試有效註冊
	validData := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	
	jsonData, _ := json.Marshal(validData)
	req, _ := http.NewRequest("POST", "/api/v1/users/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User registered successfully", response["message"])
	
	// 測試無效註冊（缺少密碼）
	invalidData := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}
	
	jsonData, _ = json.Marshal(invalidData)
	req, _ = http.NewRequest("POST", "/api/v1/users/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUserLogin 測試用戶登入
func TestUserLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.POST("/api/v1/users/login", func(c *gin.Context) {
		var request struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}
		
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		
		// 模擬登入驗證
		if request.Email == "test@example.com" && request.Password == "password123" {
			c.JSON(http.StatusOK, gin.H{
				"message": "Login successful",
				"token":   "mock-jwt-token",
				"user": gin.H{
					"email": request.Email,
				},
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
			})
		}
	})
	
	// 測試有效登入
	validData := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	
	jsonData, _ := json.Marshal(validData)
	req, _ := http.NewRequest("POST", "/api/v1/users/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Login successful", response["message"])
	assert.Equal(t, "mock-jwt-token", response["token"])
	
	// 測試無效登入
	invalidData := map[string]interface{}{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	
	jsonData, _ = json.Marshal(invalidData)
	req, _ = http.NewRequest("POST", "/api/v1/users/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestLiveStreamAPI 測試直播串流 API
func TestLiveStreamAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/api/v1/lives", func(c *gin.Context) {
		// 模擬返回直播列表
		lives := []map[string]interface{}{
			{
				"id":          1,
				"title":       "測試直播",
				"description": "這是一個測試直播",
				"status":      "live",
				"user": map[string]interface{}{
					"id":       1,
					"username": "testuser",
				},
			},
		}
		
		c.JSON(http.StatusOK, gin.H{
			"data": lives,
		})
	})
	
	req, _ := http.NewRequest("GET", "/api/v1/lives", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)
	
	live := data[0].(map[string]interface{})
	assert.Equal(t, float64(1), live["id"])
	assert.Equal(t, "測試直播", live["title"])
}

// TestMiddleware 測試中間件
func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	
	// 添加 CORS 中間件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	})
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// 測試 CORS 預檢請求
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	
	// 測試正常請求
	req, _ = http.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

// BenchmarkAPI 性能測試
func BenchmarkAPI(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	req, _ := http.NewRequest("GET", "/health", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
} 