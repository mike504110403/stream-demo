package tests

import (
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/services"
	"stream-demo/backend/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ================================
// 🆕 使用測試工具包的改進版測試
// ================================

func TestLiveService_CreateLive_WithToolkit(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)

	t.Run("🟢 改進版：成功建立直播", func(t *testing.T) {
		// 改進後：簡化設置
		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("Create", mock.AnythingOfType("*models.Live")).Return(nil)

		service := builder.BuildLiveService()

		// Act
		live, err := service.CreateLive(1, "測試直播", "這是一個測試直播", startTime)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, live)
		assert.Equal(t, "測試直播", live.Title)
		assert.Equal(t, "這是一個測試直播", live.Description)
		assert.Equal(t, uint(1), live.UserID)
		assert.Equal(t, "scheduled", live.Status)
		assert.Equal(t, startTime, live.StartTime)
		assert.True(t, live.ChatEnabled)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：儲存庫錯誤", func(t *testing.T) {
		// 改進後：一行設置錯誤
		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("Create", mock.AnythingOfType("*models.Live")).Return(assert.AnError)

		service := builder.BuildLiveService()

		// Act
		live, err := service.CreateLive(1, "測試直播", "這是一個測試直播", startTime)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, live)

		builder.AssertAllExpectations()
	})
}

func TestLiveService_StartLive_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功開始直播", func(t *testing.T) {
		// 改進後：預設置直播對象
		testLive := &models.Live{
			ID:     1,
			UserID: 1,
			Status: "scheduled",
			Title:  "測試直播",
		}

		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("FindByID", uint(1)).Return(testLive, nil)
		builder.LiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)

		service := builder.BuildLiveService()

		// Act
		err := service.StartLive(1)

		// Assert
		assert.NoError(t, err)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：直播不存在", func(t *testing.T) {
		// 改進後：簡化錯誤設置
		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("FindByID", uint(999)).Return((*models.Live)(nil), assert.AnError)

		service := builder.BuildLiveService()

		// Act
		err := service.StartLive(999)

		// Assert
		assert.Error(t, err)

		builder.AssertAllExpectations()
	})
}

func TestLiveService_EndLive_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功結束直播", func(t *testing.T) {
		// 改進後：預設置直播對象
		testLive := &models.Live{
			ID:     1,
			UserID: 1,
			Status: "live",
			Title:  "進行中的直播",
		}

		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("FindByID", uint(1)).Return(testLive, nil)
		builder.LiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)

		service := builder.BuildLiveService()

		// Act
		err := service.EndLive(1)

		// Assert
		assert.NoError(t, err)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：直播不存在", func(t *testing.T) {
		// 改進後：一行設置錯誤
		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("FindByID", uint(999)).Return((*models.Live)(nil), assert.AnError)

		service := builder.BuildLiveService()

		// Act
		err := service.EndLive(999)

		// Assert
		assert.Error(t, err)

		builder.AssertAllExpectations()
	})
}

func TestLiveService_ToggleChat_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功開啟聊天", func(t *testing.T) {
		// 改進後：預設置聊天關閉的直播
		testLive := &models.Live{
			ID:          1,
			UserID:      1,
			Status:      "live",
			ChatEnabled: false,
		}

		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("FindByID", uint(1)).Return(testLive, nil)
		builder.LiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)

		service := builder.BuildLiveService()

		// Act
		err := service.ToggleChat(1, true)

		// Assert
		assert.NoError(t, err)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：成功關閉聊天", func(t *testing.T) {
		// 改進後：預設置聊天開啟的直播
		testLive := &models.Live{
			ID:          1,
			UserID:      1,
			Status:      "live",
			ChatEnabled: true,
		}

		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("FindByID", uint(1)).Return(testLive, nil)
		builder.LiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)

		service := builder.BuildLiveService()

		// Act
		err := service.ToggleChat(1, false)

		// Assert
		assert.NoError(t, err)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：直播不存在", func(t *testing.T) {
		// 改進後：簡化錯誤設置
		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("FindByID", uint(999)).Return((*models.Live)(nil), assert.AnError)

		service := builder.BuildLiveService()

		// Act
		err := service.ToggleChat(999, true)

		// Assert
		assert.Error(t, err)

		builder.AssertAllExpectations()
	})
}

// ================================
// 🔄 原有測試保留（向後兼容）
// ================================

// MockLiveRepository 模擬直播儲存庫
type MockLiveRepository struct {
	mock.Mock
}

func (m *MockLiveRepository) Create(live *models.Live) error {
	args := m.Called(live)
	return args.Error(0)
}

func (m *MockLiveRepository) FindByID(id uint) (*models.Live, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Live), args.Error(1)
}

func (m *MockLiveRepository) FindByUserID(userID uint) ([]*models.Live, int64, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.Live), args.Get(1).(int64), args.Error(2)
}

func (m *MockLiveRepository) FindByStreamKey(streamKey string) (*models.Live, error) {
	args := m.Called(streamKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Live), args.Error(1)
}

func (m *MockLiveRepository) IncrementViewerCount(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLiveRepository) DecrementViewerCount(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLiveRepository) Update(live *models.Live) error {
	args := m.Called(live)
	return args.Error(0)
}

func (m *MockLiveRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLiveRepository) FindActive() ([]*models.Live, error) {
	args := m.Called()
	return args.Get(0).([]*models.Live), args.Error(1)
}

func TestLiveService_CreateLive(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		userID      uint
		title       string
		description string
		startTime   time.Time
		mockSetup   func(*MockLiveRepository)
		wantErr     bool
	}{
		{
			name:        "成功建立直播",
			userID:      1,
			title:       "測試直播",
			description: "這是一個測試直播",
			startTime:   startTime,
			mockSetup: func(mockLiveRepo *MockLiveRepository) {
				mockLiveRepo.On("Create", mock.AnythingOfType("*models.Live")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "儲存庫錯誤",
			userID:      1,
			title:       "測試直播",
			description: "這是一個測試直播",
			startTime:   startTime,
			mockSetup: func(mockLiveRepo *MockLiveRepository) {
				mockLiveRepo.On("Create", mock.AnythingOfType("*models.Live")).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLiveRepo := new(MockLiveRepository)
			cfg := config.NewPostgreSQLConfig("config.yaml", "local")
			service := services.NewLiveService(cfg)
			tt.mockSetup(mockLiveRepo)

			live, err := service.CreateLive(tt.userID, tt.title, tt.description, tt.startTime)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, live)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, live)
				assert.Equal(t, tt.title, live.Title)
				assert.Equal(t, tt.description, live.Description)
				assert.Equal(t, tt.userID, live.UserID)
				assert.Equal(t, "scheduled", live.Status)
				assert.Equal(t, tt.startTime, live.StartTime)
				assert.True(t, live.ChatEnabled)
			}
			mockLiveRepo.AssertExpectations(t)
		})
	}
}

func TestLiveService_StartLive(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		mockSetup func(*MockLiveRepository)
		wantErr   bool
	}{
		{
			name: "成功開始直播",
			id:   1,
			mockSetup: func(mockLiveRepo *MockLiveRepository) {
				mockLiveRepo.On("FindByID", uint(1)).Return(&models.Live{
					ID:     1,
					UserID: 1,
					Status: "scheduled",
				}, nil)
				mockLiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "直播不存在",
			id:   999,
			mockSetup: func(mockLiveRepo *MockLiveRepository) {
				mockLiveRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLiveRepo := new(MockLiveRepository)
			cfg := config.NewPostgreSQLConfig("config.yaml", "local")
			service := services.NewLiveService(cfg)
			tt.mockSetup(mockLiveRepo)

			err := service.StartLive(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockLiveRepo.AssertExpectations(t)
		})
	}
}

func TestLiveService_EndLive(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		mockSetup func(*MockLiveRepository)
		wantErr   bool
	}{
		{
			name: "成功結束直播",
			id:   1,
			mockSetup: func(mockLiveRepo *MockLiveRepository) {
				mockLiveRepo.On("FindByID", uint(1)).Return(&models.Live{
					ID:     1,
					UserID: 1,
					Status: "live",
				}, nil)
				mockLiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "直播不存在",
			id:   999,
			mockSetup: func(mockLiveRepo *MockLiveRepository) {
				mockLiveRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLiveRepo := new(MockLiveRepository)
			cfg := config.NewPostgreSQLConfig("config.yaml", "local")
			service := services.NewLiveService(cfg)
			tt.mockSetup(mockLiveRepo)

			err := service.EndLive(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockLiveRepo.AssertExpectations(t)
		})
	}
}

func TestLiveService_ToggleChat(t *testing.T) {
	tests := []struct {
		name          string
		id            uint
		mockSetup     func(*MockLiveRepository)
		wantErr       bool
		expectEnabled bool
	}{
		{
			name: "成功開啟聊天",
			id:   1,
			mockSetup: func(mockLiveRepo *MockLiveRepository) {
				mockLiveRepo.On("FindByID", uint(1)).Return(&models.Live{
					ID:          1,
					UserID:      1,
					Status:      "live",
					ChatEnabled: false,
				}, nil)
				mockLiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)
			},
			wantErr:       false,
			expectEnabled: true,
		},
		{
			name: "成功關閉聊天",
			id:   2,
			mockSetup: func(mockLiveRepo *MockLiveRepository) {
				mockLiveRepo.On("FindByID", uint(2)).Return(&models.Live{
					ID:          2,
					UserID:      1,
					Status:      "live",
					ChatEnabled: true,
				}, nil)
				mockLiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)
			},
			wantErr:       false,
			expectEnabled: false,
		},
		{
			name: "直播不存在",
			id:   999,
			mockSetup: func(mockLiveRepo *MockLiveRepository) {
				mockLiveRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLiveRepo := new(MockLiveRepository)
			cfg := config.NewPostgreSQLConfig("config.yaml", "local")
			service := services.NewLiveService(cfg)
			tt.mockSetup(mockLiveRepo)

			err := service.ToggleChat(tt.id, tt.expectEnabled)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockLiveRepo.AssertExpectations(t)
		})
	}
}
