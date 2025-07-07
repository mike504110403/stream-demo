package tests

import (
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/dto"
	"stream-demo/backend/services"
	"stream-demo/backend/tests/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockVideoRepository 模擬影片儲存庫
type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) Create(video *models.Video) error {
	args := m.Called(video)
	return args.Error(0)
}

func (m *MockVideoRepository) FindByID(id uint) (*models.Video, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Video), args.Error(1)
}

func (m *MockVideoRepository) FindByUserID(userID uint) ([]models.Video, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Video), args.Error(1)
}

func (m *MockVideoRepository) FindAll() ([]models.Video, error) {
	args := m.Called()
	return args.Get(0).([]models.Video), args.Error(1)
}

func (m *MockVideoRepository) Search(query string) ([]models.Video, error) {
	args := m.Called(query)
	return args.Get(0).([]models.Video), args.Error(1)
}

func (m *MockVideoRepository) Update(video *models.Video) error {
	args := m.Called(video)
	return args.Error(0)
}

func (m *MockVideoRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockVideoRepository) IncrementViews(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockVideoRepository) IncrementLikes(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockUserRepository for VideoService
type MockUserRepositoryForVideo struct {
	mock.Mock
}

func (m *MockUserRepositoryForVideo) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepositoryForVideo) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForVideo) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForVideo) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForVideo) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepositoryForVideo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestVideoService_UploadVideo(t *testing.T) {
	tests := []struct {
		name         string
		userID       uint
		title        string
		description  string
		videoURL     string
		thumbnailURL string
		mockSetup    func(*MockVideoRepository, *MockUserRepositoryForVideo)
		wantErr      bool
	}{
		{
			name:         "成功上傳影片",
			userID:       1,
			title:        "測試影片",
			description:  "這是一個測試影片",
			videoURL:     "/path/to/video.mp4",
			thumbnailURL: "/path/to/thumbnail.jpg",
			mockSetup: func(mockVideoRepo *MockVideoRepository, mockUserRepo *MockUserRepositoryForVideo) {
				mockUserRepo.On("FindByID", uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
				mockVideoRepo.On("Create", mock.AnythingOfType("*models.Video")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "用戶不存在",
			userID:       999,
			title:        "測試影片",
			description:  "這是一個測試影片",
			videoURL:     "/path/to/video.mp4",
			thumbnailURL: "/path/to/thumbnail.jpg",
			mockSetup: func(mockVideoRepo *MockVideoRepository, mockUserRepo *MockUserRepositoryForVideo) {
				mockUserRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:         "儲存庫錯誤",
			userID:       1,
			title:        "測試影片",
			description:  "這是一個測試影片",
			videoURL:     "/path/to/video.mp4",
			thumbnailURL: "/path/to/thumbnail.jpg",
			mockSetup: func(mockVideoRepo *MockVideoRepository, mockUserRepo *MockUserRepositoryForVideo) {
				mockUserRepo.On("FindByID", uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
				mockVideoRepo.On("Create", mock.AnythingOfType("*models.Video")).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVideoRepo := new(MockVideoRepository)
			mockUserRepo := new(MockUserRepositoryForVideo)
			cfg := config.NewPostgreSQLConfig("config.yaml", "local")
			service := services.NewVideoService(cfg)
			tt.mockSetup(mockVideoRepo, mockUserRepo)

			err := service.UpdateVideo(tt.userID, tt.title, tt.description, &dto.VideoDTO{
				Title:        tt.title,
				Description:  tt.description,
				OriginalURL:  tt.videoURL,
				ThumbnailURL: tt.thumbnailURL,
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockVideoRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestVideoService_GetVideoByID(t *testing.T) {
	mockVideoRepo := new(MockVideoRepository)
	mockUserRepo := new(MockUserRepositoryForVideo)
	cfg := config.NewPostgreSQLConfig("config.yaml", "local")
	service := services.NewVideoService(cfg)

	tests := []struct {
		name      string
		id        uint
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "成功獲取影片",
			id:   1,
			mockSetup: func() {
				mockVideoRepo.On("FindByID", uint(1)).Return(&models.Video{
					ID:          1,
					Title:       "測試影片",
					Description: "這是一個測試影片",
					UserID:      1,
				}, nil)
				mockUserRepo.On("FindByID", uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
			},
			wantErr: false,
		},
		{
			name: "影片不存在",
			id:   999,
			mockSetup: func() {
				mockVideoRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for each test
			mockVideoRepo.Mock = mock.Mock{}
			mockUserRepo.Mock = mock.Mock{}
			tt.mockSetup()

			video, err := service.GetVideoByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, video)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, video)
				assert.Equal(t, tt.id, video.ID)
			}
			mockVideoRepo.AssertExpectations(t)
		})
	}
}

func TestVideoService_GetVideosByUserID(t *testing.T) {
	mockVideoRepo := new(MockVideoRepository)
	mockUserRepo := new(MockUserRepositoryForVideo)
	cfg := config.NewPostgreSQLConfig("config.yaml", "local")
	service := services.NewVideoService(cfg)

	tests := []struct {
		name      string
		userID    uint
		mockSetup func()
		wantErr   bool
	}{
		{
			name:   "成功獲取用戶影片",
			userID: 1,
			mockSetup: func() {
				videos := []models.Video{
					{ID: 1, Title: "影片1", UserID: 1},
					{ID: 2, Title: "影片2", UserID: 1},
				}
				mockVideoRepo.On("FindByUserID", uint(1)).Return(videos, nil)
				mockUserRepo.On("FindByID", uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
			},
			wantErr: false,
		},
		{
			name:   "用戶不存在",
			userID: 999,
			mockSetup: func() {
				mockVideoRepo.On("FindByUserID", uint(999)).Return([]models.Video{}, nil)
				mockUserRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for each test
			mockVideoRepo.Mock = mock.Mock{}
			mockUserRepo.Mock = mock.Mock{}
			tt.mockSetup()

			videos, total, err := service.GetVideosByUserID(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, videos)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, videos)
				assert.Greater(t, total, int64(0))
			}
			mockVideoRepo.AssertExpectations(t)
		})
	}
}

func TestVideoService_UpdateVideo(t *testing.T) {
	tests := []struct {
		name        string
		id          uint
		title       string
		description string
		mockSetup   func(*MockVideoRepository, *MockUserRepositoryForVideo)
		wantErr     bool
	}{
		{
			name:        "成功更新影片",
			id:          1,
			title:       "更新後的標題",
			description: "更新後的描述",
			mockSetup: func(mockVideoRepo *MockVideoRepository, mockUserRepo *MockUserRepositoryForVideo) {
				mockVideoRepo.On("FindByID", uint(1)).Return(&models.Video{
					ID:          1,
					UserID:      1,
					Title:       "原始標題",
					Description: "原始描述",
				}, nil)
				mockUserRepo.On("FindByID", uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
				mockVideoRepo.On("Update", mock.AnythingOfType("*models.Video")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "影片不存在",
			id:          999,
			title:       "更新後的標題",
			description: "更新後的描述",
			mockSetup: func(mockVideoRepo *MockVideoRepository, mockUserRepo *MockUserRepositoryForVideo) {
				mockVideoRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVideoRepo := new(MockVideoRepository)
			mockUserRepo := new(MockUserRepositoryForVideo)
			cfg := config.NewPostgreSQLConfig("config.yaml", "local")
			service := services.NewVideoService(cfg)
			tt.mockSetup(mockVideoRepo, mockUserRepo)

			err := service.UpdateVideo(tt.id, tt.title, tt.description, &dto.VideoDTO{
				Title:       tt.title,
				Description: tt.description,
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockVideoRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// ================================
// 🆕 使用新測試工具包的改進測試
// ================================

func TestVideoService_UploadVideo_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功上傳影片", func(t *testing.T) {
		// 改進後：只需要 3 行設置
		testUser := &models.User{ID: 1, Username: "testuser"}
		builder := testutils.NewServiceBuilder(t).
			WithUser(testUser).
			WithCreateVideoSuccess()
		service := builder.BuildVideoService()

		// Act
		err := service.UpdateVideo(1, "測試影片", "這是一個測試影片", &dto.VideoDTO{
			Title:        "測試影片",
			Description:  "這是一個測試影片",
			OriginalURL:  "/video.mp4",
			ThumbnailURL: "/thumb.jpg",
		})

		// Assert
		assert.NoError(t, err)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：用戶不存在", func(t *testing.T) {
		// 改進後：只需要 2 行設置
		builder := testutils.NewServiceBuilder(t).
			WithUserNotFound(999)
		service := builder.BuildVideoService()

		// Act
		err := service.UpdateVideo(999, "測試影片", "描述", &dto.VideoDTO{
			Title:        "測試影片",
			Description:  "描述",
			OriginalURL:  "/video.mp4",
			ThumbnailURL: "/thumb.jpg",
		})

		// Assert
		assert.Error(t, err)

		builder.AssertAllExpectations()
	})
}

// ================================
// 📊 新舊測試對比展示
// ================================

/*
🔴 舊版測試複雜度：
- Mock 設置：8-12 行代碼
- 業務邏輯：2-3 行代碼
- 維護成本：高（每次 Interface 變更都要修改）

🟢 新版測試複雜度：
- Mock 設置：2-3 行代碼 (減少 70%)
- 業務邏輯：2-3 行代碼
- 維護成本：低（工具包統一管理）

💡 TDD 友好度：
- 舊版：先設計複雜 Mock → 寫測試 → 實現功能
- 新版：快速寫測試 → 實現功能 → 重構優化
*/
