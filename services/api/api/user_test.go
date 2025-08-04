package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"stream-demo/backend/dto"
	"stream-demo/backend/dto/request"
	"stream-demo/backend/dto/response"
	"stream-demo/backend/test/mocks"
)

func TestUserHandler_Register(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 創建模擬服務
	mockUserService := &mocks.MockUserService{}

	// 創建處理器
	handler := &UserHandler{
		userService: mockUserService,
	}

	// 測試用例
	tests := []struct {
		name           string
		requestBody    request.RegisterRequest
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "成功註冊",
			requestBody: request.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				mockUserService.On("Register", "testuser", "test@example.com", "password123").Return(&dto.UserDTO{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "無效的請求數據",
			requestBody: request.RegisterRequest{
				Username: "",
				Email:    "invalid-email",
				Password: "123",
			},
			mockSetup: func() {
				// 不需要設置 mock，因為驗證會失敗
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// 創建請求
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/users/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// 創建響應記錄器
			w := httptest.NewRecorder()

			// 創建路由
			router := gin.New()
			router.POST("/api/v1/users/register", handler.Register)

			// 執行請求
			router.ServeHTTP(w, req)

			// 驗證結果
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response response.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, 200, response.Code)
			}

			// 驗證 mock 調用
			mockUserService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 創建模擬服務
	mockUserService := &mocks.MockUserService{}

	// 創建處理器
	handler := &UserHandler{
		userService: mockUserService,
	}

	// 測試用例
	tests := []struct {
		name           string
		requestBody    request.LoginRequest
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "成功登入",
			requestBody: request.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockUserService.On("Login", "testuser", "password123").Return("test-token", &dto.UserDTO{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}, time.Now().Add(24*time.Hour), nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "無效的憑證",
			requestBody: request.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockUserService.On("Login", "testuser", "wrongpassword").Return("", (*dto.UserDTO)(nil), time.Time{}, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// 創建請求
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/users/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// 創建響應記錄器
			w := httptest.NewRecorder()

			// 創建路由
			router := gin.New()
			router.POST("/api/v1/users/login", handler.Login)

			// 執行請求
			router.ServeHTTP(w, req)

			// 驗證結果
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response response.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, 200, response.Code)
			}

			// 驗證 mock 調用
			mockUserService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 創建模擬服務
	mockUserService := &mocks.MockUserService{}

	// 創建處理器
	handler := &UserHandler{
		userService: mockUserService,
	}

	// 測試用例
	tests := []struct {
		name           string
		userID         uint
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "成功獲取用戶資料",
			userID: 1,
			mockSetup: func() {
				mockUserService.On("GetUserByID", uint(1)).Return(&dto.UserDTO{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "用戶不存在",
			userID: 999,
			mockSetup: func() {
				mockUserService.On("GetUserByID", uint(999)).Return((*dto.UserDTO)(nil), assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// 創建請求
			req, _ := http.NewRequest("GET", "/api/v1/users/profile", nil)

			// 設置用戶上下文（模擬認證中間件）
			req.Header.Set("Authorization", "Bearer test-token")

			// 創建響應記錄器
			w := httptest.NewRecorder()

			// 創建路由
			router := gin.New()
			router.GET("/api/v1/users/profile", func(c *gin.Context) {
				// 模擬認證中間件設置用戶 ID
				c.Set("user_id", tt.userID)
				handler.GetUser(c)
			})

			// 執行請求
			router.ServeHTTP(w, req)

			// 驗證結果
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response response.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, 200, response.Code)
			}

			// 驗證 mock 調用
			mockUserService.AssertExpectations(t)
		})
	}
}
