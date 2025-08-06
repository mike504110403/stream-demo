package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"stream-demo/backend/dto"
	"stream-demo/backend/dto/response"
	"stream-demo/backend/test/mocks"
)

func TestVideoHandler_ListVideos(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 測試用例
	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(*mocks.MockVideoService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "成功獲取影片列表",
			queryParams: map[string]string{
				"offset": "0",
				"limit":  "10",
			},
			mockSetup: func(mockService *mocks.MockVideoService) {
				mockService.On("GetVideos", 0, 10).Return([]*dto.VideoDTO{
					{
						ID:          1,
						Title:       "測試影片1",
						Description: "這是測試影片1",
						UserID:      1,
						Username:    "testuser",
						Status:      "ready",
						Views:       100,
						Likes:       10,
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
			mockSetup: func(mockService *mocks.MockVideoService) {
				mockService.On("GetVideos", 0, 10).Return(nil, int64(0), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的 mock
			mockVideoService := &mocks.MockVideoService{}

			// 創建處理器
			handler := &VideoHandler{
				videoService: mockVideoService,
			}

			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup(mockVideoService)
			}

			// 創建請求
			req, _ := http.NewRequest("GET", "/api/videos", nil)

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
			router.GET("/api/videos", handler.ListVideos)

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
			mockVideoService.AssertExpectations(t)
		})
	}
}

func TestVideoHandler_GetVideo(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 測試用例
	tests := []struct {
		name           string
		videoID        string
		mockSetup      func(*mocks.MockVideoService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "成功獲取影片",
			videoID: "1",
			mockSetup: func(mockService *mocks.MockVideoService) {
				mockService.On("GetVideoByID", uint(1)).Return(&dto.VideoDTO{
					ID:          1,
					Title:       "測試影片",
					Description: "這是測試影片",
					UserID:      1,
					Username:    "testuser",
					Status:      "ready",
					Views:       100,
					Likes:       10,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "影片不存在",
			videoID: "999",
			mockSetup: func(mockService *mocks.MockVideoService) {
				mockService.On("GetVideoByID", uint(999)).Return((*dto.VideoDTO)(nil), assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name:    "無效的影片ID",
			videoID: "invalid",
			mockSetup: func(mockService *mocks.MockVideoService) {
				// 不需要設置 mock，因為會因為無效ID而失敗
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的 mock
			mockVideoService := &mocks.MockVideoService{}

			// 創建處理器
			handler := &VideoHandler{
				videoService: mockVideoService,
			}

			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup(mockVideoService)
			}

			// 創建請求
			req, _ := http.NewRequest("GET", "/api/videos/"+tt.videoID, nil)

			// 創建響應記錄器
			w := httptest.NewRecorder()

			// 創建路由
			router := gin.New()
			router.GET("/api/videos/:id", handler.GetVideo)

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
			mockVideoService.AssertExpectations(t)
		})
	}
}

func TestVideoHandler_LikeVideo(t *testing.T) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 測試用例
	tests := []struct {
		name           string
		videoID        string
		mockSetup      func(*mocks.MockVideoService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "成功喜歡影片",
			videoID: "1",
			mockSetup: func(mockService *mocks.MockVideoService) {
				mockService.On("LikeVideo", uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "影片不存在",
			videoID: "999",
			mockSetup: func(mockService *mocks.MockVideoService) {
				mockService.On("LikeVideo", uint(999)).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:    "無效的影片ID",
			videoID: "invalid",
			mockSetup: func(mockService *mocks.MockVideoService) {
				// 不需要設置 mock，因為會因為無效ID而失敗
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的 mock
			mockVideoService := &mocks.MockVideoService{}

			// 創建處理器
			handler := &VideoHandler{
				videoService: mockVideoService,
			}

			// 設置 mock
			if tt.mockSetup != nil {
				tt.mockSetup(mockVideoService)
			}

			// 創建請求
			req, _ := http.NewRequest("POST", "/api/videos/"+tt.videoID+"/like", nil)

			// 創建響應記錄器
			w := httptest.NewRecorder()

			// 創建路由
			router := gin.New()
			router.POST("/api/videos/:id/like", handler.LikeVideo)

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
			mockVideoService.AssertExpectations(t)
		})
	}
}
