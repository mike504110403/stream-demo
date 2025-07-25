package tests

import (
	"fmt"
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

// ================================
// 🆕 使用新測試工具包的改進版測試
// ================================

func TestVideoService_CreateVideoRecord_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功建立影片記錄", func(t *testing.T) {
		// 測試用戶和影片資料
		testUser := &models.User{ID: 1, Username: "testuser"}

		builder := testutils.NewServiceBuilder(t).
			WithUser(testUser).
			WithCreateVideoSuccess()

		service := builder.BuildVideoService()

		// Act
		video, err := service.CreateVideoRecord(1, "測試影片", "測試描述", "test-s3-key")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, video)
		assert.Equal(t, "測試影片", video.Title)
		assert.Equal(t, "測試描述", video.Description)
		assert.Equal(t, uint(1), video.UserID)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：用戶不存在", func(t *testing.T) {
		builder := testutils.NewServiceBuilder(t).
			WithUserNotFound(999)

		service := builder.BuildVideoService()

		// Act
		video, err := service.CreateVideoRecord(999, "測試影片", "測試描述", "test-s3-key")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, video)

		builder.AssertAllExpectations()
	})
}

func TestVideoService_GetVideoByID_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功獲取影片", func(t *testing.T) {
		testVideo := &models.Video{
			ID:          1,
			Title:       "測試影片",
			Description: "測試描述",
			UserID:      1,
		}
		testUser := &models.User{ID: 1, Username: "testuser"}

		builder := testutils.NewServiceBuilder(t).
			WithVideo(testVideo).
			WithUser(testUser)

		service := builder.BuildVideoService()

		// Act
		video, err := service.GetVideoByID(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, video)
		assert.Equal(t, uint(1), video.ID)
		assert.Equal(t, "測試影片", video.Title)
		assert.Equal(t, "testuser", video.Username)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：影片不存在", func(t *testing.T) {
		builder := testutils.NewServiceBuilder(t).
			WithVideoNotFound(999)

		service := builder.BuildVideoService()

		// Act
		video, err := service.GetVideoByID(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, video)

		builder.AssertAllExpectations()
	})
}

func TestVideoService_GetVideosByUserID_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功獲取用戶影片", func(t *testing.T) {
		testUser := &models.User{ID: 1, Username: "testuser"}
		testVideos := []models.Video{
			{ID: 1, Title: "影片1", UserID: 1},
			{ID: 2, Title: "影片2", UserID: 1},
		}

		builder := testutils.NewServiceBuilder(t).
			WithUser(testUser)

		builder.VideoRepo.On("FindByUserID", uint(1)).Return(testVideos, nil)
		service := builder.BuildVideoService()

		// Act
		videos, total, err := service.GetVideosByUserID(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, videos)
		assert.Equal(t, int64(2), total)
		assert.Len(t, videos, 2)
		assert.Equal(t, "影片1", videos[0].Title)
		assert.Equal(t, "testuser", videos[0].Username)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：用戶不存在", func(t *testing.T) {
		builder := testutils.NewServiceBuilder(t).
			WithUserNotFound(999)

		service := builder.BuildVideoService()

		// Act
		videos, total, err := service.GetVideosByUserID(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, videos)
		assert.Equal(t, int64(0), total)

		builder.AssertAllExpectations()
	})
}

// ================================
// 🚀 多資料庫測試
// ================================

func TestVideoService_MultiDatabase(t *testing.T) {
	// 準備測試數據
	testUser := &models.User{ID: 1, Username: "testuser", Email: "test@example.com"}
	testVideo := &models.Video{
		ID:          1,
		Title:       "測試影片",
		Description: "測試描述",
		UserID:      1,
		Status:      "uploading",
	}

	testCases := []struct {
		name      string
		dbType    testutils.DatabaseType
		setupTest func(builder *testutils.ServiceBuilder) *services.VideoService
		runTest   func(service *services.VideoService) error
		wantError bool
	}{
		{
			name:   "PostgreSQL 影片創建",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.VideoService {
				return builder.WithUser(testUser).WithCreateVideoSuccess().BuildVideoService()
			},
			runTest: func(service *services.VideoService) error {
				_, err := service.CreateVideoRecord(1, "新影片", "描述", "s3-key")
				return err
			},
			wantError: false,
		},
		{
			name:   "MySQL 影片查詢",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.VideoService {
				return builder.WithVideo(testVideo).WithUser(testUser).BuildVideoService()
			},
			runTest: func(service *services.VideoService) error {
				_, err := service.GetVideoByID(1)
				return err
			},
			wantError: false,
		},
		{
			name:   "PostgreSQL 影片更新",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.VideoService {
				builder.VideoRepo.On("FindByID", uint(1)).Return(testVideo, nil)
				builder.VideoRepo.On("Update", mock.AnythingOfType("*models.Video")).Return(nil)
				builder.UserRepo.On("FindByID", uint(1)).Return(testUser, nil)
				return builder.BuildVideoService()
			},
			runTest: func(service *services.VideoService) error {
				return service.UpdateVideo(1, "更新標題", "更新描述", &dto.VideoDTO{})
			},
			wantError: false,
		},
		{
			name:   "MySQL 影片刪除",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.VideoService {
				builder.VideoRepo.On("Delete", uint(1)).Return(nil)
				return builder.BuildVideoService()
			},
			runTest: func(service *services.VideoService) error {
				return service.DeleteVideo(1)
			},
			wantError: false,
		},
		{
			name:   "PostgreSQL 影片搜索",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.VideoService {
				searchResults := []models.Video{*testVideo}
				builder.VideoRepo.On("Search", "測試").Return(searchResults, nil)
				builder.UserRepo.On("FindByID", uint(1)).Return(testUser, nil)
				return builder.BuildVideoService()
			},
			runTest: func(service *services.VideoService) error {
				_, _, err := service.SearchVideos("測試", 0, 10)
				return err
			},
			wantError: false,
		},
		{
			name:   "MySQL 影片點讚",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.VideoService {
				builder.VideoRepo.On("IncrementLikes", uint(1)).Return(nil)
				return builder.BuildVideoService()
			},
			runTest: func(service *services.VideoService) error {
				return service.LikeVideo(1)
			},
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 創建指定資料庫類型的構建器
			builder := testutils.NewServiceBuilderWithDB(t, tc.dbType)

			// 驗證配置
			assert.NoError(t, builder.ValidateConfig())
			assert.Equal(t, tc.dbType, builder.GetDatabaseType())

			// 設置測試
			service := tc.setupTest(builder)

			// 執行測試
			err := tc.runTest(service)

			// 檢查結果
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 驗證所有期望
			builder.AssertAllExpectations()
		})
	}
}

// ================================
// 🔧 影片服務專屬測試工具
// ================================

func TestVideoService_SpecialCases(t *testing.T) {
	t.Run("大型影片分頁查詢 - PostgreSQL", func(t *testing.T) {
		builder := testutils.NewPostgreSQLServiceBuilder(t)

		// 模擬大量影片數據
		videos := make([]models.Video, 100)
		for i := 0; i < 100; i++ {
			videos[i] = models.Video{
				ID:     uint(i + 1),
				Title:  fmt.Sprintf("影片 %d", i+1),
				UserID: 1,
			}
		}

		builder.VideoRepo.On("FindVideosWithPagination", 0, 20).Return(videos[:20], int64(100), nil)
		service := builder.BuildVideoService()

		// Act
		result, total, err := service.GetVideos(0, 20)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 20)
		assert.Equal(t, int64(100), total)

		builder.AssertAllExpectations()
	})

	t.Run("影片轉碼狀態檢查 - MySQL", func(t *testing.T) {
		builder := testutils.NewMySQLServiceBuilder(t)

		testVideo := &models.Video{
			ID:                 1,
			Title:              "轉碼測試影片",
			Status:             "processing",
			ProcessingProgress: 50,
		}
		testUser := &models.User{ID: 1, Username: "testuser"}

		builder.WithVideo(testVideo).WithUser(testUser)
		service := builder.BuildVideoService()

		// Act
		video, err := service.GetVideoByID(1)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "processing", video.Status)
		assert.Equal(t, 50, video.ProcessingProgress)

		builder.AssertAllExpectations()
	})
}
