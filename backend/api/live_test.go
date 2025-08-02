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
	"github.com/stretchr/testify/mock"

	"stream-demo/backend/dto"
	"stream-demo/backend/dto/request"
	"stream-demo/backend/dto/response"
	"stream-demo/backend/test/mocks"
)

func TestLiveHandler_ListLives(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 測試用例
	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(*mocks.MockLiveService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "成功獲取直播列表",
			queryParams: map[string]string{
				"offset": "0",
				"limit":  "10",
			},
			mockSetup: func(mockService *mocks.MockLiveService) {
				mockService.On("ListLives", 0, 10).Return([]*dto.LiveDTO{
					{
						ID:          1,
						Title:       "測試直播1",
						Description: "這是測試直播1",
						UserID:      1,
						Status:      "active",
						StartTime:   time.Now(),
						ViewerCount: 100,
					},
				}, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "服務錯誤",
			queryParams: map[string]string{
				"offset": "0",
				"limit":  "10",
			},
			mockSetup: func(mockService *mocks.MockLiveService) {
				mockService.On("ListLives", 0, 10).Return(nil, int64(0), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的 mock
			mockLiveService := &mocks.MockLiveService{}
			
			// 創建處理器
			handler := &LiveHandler{
				liveService: mockLiveService,
			}

			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup(mockLiveService)
			}

			// 創建請求
			req, _ := http.NewRequest("GET", "/api/live", nil)
			
			// 添加查詢參數
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// 創建響應記錄器
			w := httptest.NewRecorder()

			// 創建路由
			router := gin.New()
			router.GET("/api/live", handler.ListLives)

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
			mockLiveService.AssertExpectations(t)
		})
	}
}

func TestLiveHandler_CreateLive(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 測試用例
	tests := []struct {
		name           string
		requestBody    request.CreateLiveRequest
		userID         uint
		mockSetup      func(*mocks.MockLiveService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "成功創建直播",
			requestBody: request.CreateLiveRequest{
				Title:       "測試直播",
				Description: "這是測試直播",
				StartTime:   time.Now().Add(time.Hour),
			},
			userID: 1,
			mockSetup: func(mockService *mocks.MockLiveService) {
				mockService.On("CreateLive", uint(1), "測試直播", "這是測試直播", mock.AnythingOfType("time.Time")).
					Return(&dto.LiveDTO{
						ID:          1,
						Title:       "測試直播",
						Description: "這是測試直播",
						UserID:      1,
						Status:      "scheduled",
						StartTime:   time.Now().Add(time.Hour),
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "未登入",
			requestBody: request.CreateLiveRequest{
				Title:       "測試直播",
				Description: "這是測試直播",
				StartTime:   time.Now().Add(time.Hour),
			},
			userID: 0, // 未設置用戶ID
			mockSetup: func(mockService *mocks.MockLiveService) {
				// 不需要設置 mock，因為會因為未登入而失敗
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "服務錯誤",
			requestBody: request.CreateLiveRequest{
				Title:       "測試直播",
				Description: "這是測試直播",
				StartTime:   time.Now().Add(time.Hour),
			},
			userID: 1,
			mockSetup: func(mockService *mocks.MockLiveService) {
				mockService.On("CreateLive", uint(1), "測試直播", "這是測試直播", mock.AnythingOfType("time.Time")).
					Return((*dto.LiveDTO)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的 mock
			mockLiveService := &mocks.MockLiveService{}
			
			// 創建處理器
			handler := &LiveHandler{
				liveService: mockLiveService,
			}

			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup(mockLiveService)
			}

			// 創建請求
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/live", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// 創建響應記錄器
			w := httptest.NewRecorder()

			// 創建路由
			router := gin.New()
			router.POST("/api/live", func(c *gin.Context) {
				if tt.userID > 0 {
					c.Set("user_id", tt.userID)
				}
				handler.CreateLive(c)
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
			mockLiveService.AssertExpectations(t)
		})
	}
}

func TestLiveHandler_GetLive(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 測試用例
	tests := []struct {
		name           string
		liveID         string
		mockSetup      func(*mocks.MockLiveService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "成功獲取直播",
			liveID: "1",
			mockSetup: func(mockService *mocks.MockLiveService) {
				mockService.On("GetLiveByID", uint(1)).Return(&dto.LiveDTO{
					ID:          1,
					Title:       "測試直播",
					Description: "這是測試直播",
					UserID:      1,
					Status:      "active",
					StartTime:   time.Now(),
					ViewerCount: 100,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "直播不存在",
			liveID: "999",
			mockSetup: func(mockService *mocks.MockLiveService) {
				mockService.On("GetLiveByID", uint(999)).Return((*dto.LiveDTO)(nil), assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name:   "無效的直播ID",
			liveID: "invalid",
			mockSetup: func(mockService *mocks.MockLiveService) {
				// 不需要設置 mock，因為會因為無效ID而失敗
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的 mock
			mockLiveService := &mocks.MockLiveService{}
			
			// 創建處理器
			handler := &LiveHandler{
				liveService: mockLiveService,
			}

			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup(mockLiveService)
			}

			// 創建請求
			req, _ := http.NewRequest("GET", "/api/live/"+tt.liveID, nil)

			// 創建響應記錄器
			w := httptest.NewRecorder()

			// 創建路由
			router := gin.New()
			router.GET("/api/live/:id", handler.GetLive)

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
			mockLiveService.AssertExpectations(t)
		})
	}
}

func TestLiveHandler_StartLive(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 測試用例
	tests := []struct {
		name           string
		liveID         string
		mockSetup      func(*mocks.MockLiveService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "成功開始直播",
			liveID: "1",
			mockSetup: func(mockService *mocks.MockLiveService) {
				mockService.On("StartLive", uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "直播不存在",
			liveID: "999",
			mockSetup: func(mockService *mocks.MockLiveService) {
				mockService.On("StartLive", uint(999)).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:   "無效的直播ID",
			liveID: "invalid",
			mockSetup: func(mockService *mocks.MockLiveService) {
				// 不需要設置 mock，因為會因為無效ID而失敗
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的 mock
			mockLiveService := &mocks.MockLiveService{}
			
			// 創建處理器
			handler := &LiveHandler{
				liveService: mockLiveService,
			}

			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup(mockLiveService)
			}

			// 創建請求
			req, _ := http.NewRequest("POST", "/api/live/"+tt.liveID+"/start", nil)

			// 創建響應記錄器
			w := httptest.NewRecorder()

			// 創建路由
			router := gin.New()
			router.POST("/api/live/:id/start", handler.StartLive)

			// 執行請求
			router.ServeHTTP(w, req)

			// 驗證結果
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				// StartLive 使用 c.Status() 而不是 c.JSON()，所以響應體為空
				assert.Empty(t, w.Body.String())
			}

			// 驗證 mock 調用
			mockLiveService.AssertExpectations(t)
		})
	}
} 