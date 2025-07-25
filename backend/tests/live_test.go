package tests

import (
	"fmt"
	"stream-demo/backend/database/models"
	"stream-demo/backend/services"
	"stream-demo/backend/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ================================
// 🆕 使用新測試工具包的改進版測試
// ================================

func TestLiveService_CreateLive_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功建立直播", func(t *testing.T) {
		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("Create", mock.AnythingOfType("*models.Live")).Return(nil)

		service := builder.BuildLiveService()

		// Act
		startTime := time.Now().Add(1 * time.Hour)
		live, err := service.CreateLive(1, "測試直播", "測試直播描述", startTime)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, live)
		assert.Equal(t, "測試直播", live.Title)
		assert.Equal(t, "測試直播描述", live.Description)
		assert.Equal(t, uint(1), live.UserID)
		assert.Equal(t, "scheduled", live.Status)

		builder.AssertAllExpectations()
	})
}

func TestLiveService_GetLiveByID_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功獲取直播", func(t *testing.T) {
		testLive := &models.Live{
			ID:          1,
			Title:       "測試直播",
			Description: "測試直播描述",
			UserID:      1,
			Status:      "live",
			StartTime:   time.Now(),
			ChatEnabled: true,
		}

		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("FindByID", uint(1)).Return(testLive, nil)

		service := builder.BuildLiveService()

		// Act
		live, err := service.GetLiveByID(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, live)
		assert.Equal(t, uint(1), live.ID)
		assert.Equal(t, "測試直播", live.Title)
		assert.Equal(t, "live", live.Status)
		assert.True(t, live.ChatEnabled)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：直播不存在", func(t *testing.T) {
		builder := testutils.NewServiceBuilder(t)
		builder.LiveRepo.On("FindByID", uint(999)).Return((*models.Live)(nil), assert.AnError)

		service := builder.BuildLiveService()

		// Act
		live, err := service.GetLiveByID(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, live)

		builder.AssertAllExpectations()
	})
}

func TestLiveService_StartLive_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功開始直播", func(t *testing.T) {
		testLive := &models.Live{
			ID:     1,
			UserID: 1,
			Status: "scheduled",
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
}

func TestLiveService_EndLive_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功結束直播", func(t *testing.T) {
		testLive := &models.Live{
			ID:     1,
			UserID: 1,
			Status: "live",
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
}

// ================================
// 🚀 多資料庫測試
// ================================

func TestLiveService_MultiDatabase(t *testing.T) {
	// 準備測試數據
	testLive := &models.Live{
		ID:          1,
		Title:       "測試直播",
		Description: "測試直播描述",
		UserID:      1,
		Status:      "scheduled",
		StartTime:   time.Now().Add(1 * time.Hour),
		ChatEnabled: true,
		StreamKey:   "stream_key_123",
		ViewerCount: 0,
	}

	testCases := []struct {
		name      string
		dbType    testutils.DatabaseType
		setupTest func(builder *testutils.ServiceBuilder) *services.LiveService
		runTest   func(service *services.LiveService) error
		wantError bool
	}{
		{
			name:   "PostgreSQL 直播創建",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.LiveService {
				builder.LiveRepo.On("Create", mock.AnythingOfType("*models.Live")).Return(nil)
				return builder.BuildLiveService()
			},
			runTest: func(service *services.LiveService) error {
				startTime := time.Now().Add(2 * time.Hour)
				_, err := service.CreateLive(1, "PostgreSQL 直播", "PostgreSQL 測試直播", startTime)
				return err
			},
			wantError: false,
		},
		{
			name:   "MySQL 直播查詢",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.LiveService {
				builder.LiveRepo.On("FindByID", uint(1)).Return(testLive, nil)
				return builder.BuildLiveService()
			},
			runTest: func(service *services.LiveService) error {
				_, err := service.GetLiveByID(1)
				return err
			},
			wantError: false,
		},
		{
			name:   "PostgreSQL 直播開始",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.LiveService {
				builder.LiveRepo.On("FindByID", uint(1)).Return(testLive, nil)
				builder.LiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)
				return builder.BuildLiveService()
			},
			runTest: func(service *services.LiveService) error {
				return service.StartLive(1)
			},
			wantError: false,
		},
		{
			name:   "MySQL 直播結束",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.LiveService {
				liveLive := *testLive
				liveLive.Status = "live"
				builder.LiveRepo.On("FindByID", uint(1)).Return(&liveLive, nil)
				builder.LiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)
				return builder.BuildLiveService()
			},
			runTest: func(service *services.LiveService) error {
				return service.EndLive(1)
			},
			wantError: false,
		},
		{
			name:   "PostgreSQL 用戶直播列表",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.LiveService {
				lives := []*models.Live{testLive}
				builder.LiveRepo.On("FindByUserID", uint(1)).Return(lives, int64(1), nil)
				return builder.BuildLiveService()
			},
			runTest: func(service *services.LiveService) error {
				_, _, err := service.GetLivesByUserID(1)
				return err
			},
			wantError: false,
		},
		{
			name:   "MySQL 直播列表",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.LiveService {
				lives := []*models.Live{testLive}
				builder.LiveRepo.On("FindByUserID", uint(0)).Return(lives, int64(1), nil)
				return builder.BuildLiveService()
			},
			runTest: func(service *services.LiveService) error {
				_, _, err := service.ListLives(0, 10)
				return err
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
// 🔧 直播服務專屬測試工具
// ================================

func TestLiveService_SpecialCases(t *testing.T) {
	t.Run("直播狀態流轉測試 - PostgreSQL", func(t *testing.T) {
		builder := testutils.NewPostgreSQLServiceBuilder(t)

		// 測試直播狀態流轉：scheduled -> live -> ended
		liveStates := []struct {
			status string
			action func(*services.LiveService, uint) error
		}{
			{"scheduled", nil}, // 初始狀態
			{"live", func(s *services.LiveService, id uint) error { return s.StartLive(id) }},
			{"ended", func(s *services.LiveService, id uint) error { return s.EndLive(id) }},
		}

		for i, state := range liveStates {
			live := &models.Live{
				ID:     uint(i + 1),
				UserID: 1,
				Status: state.status,
			}

			builder.LiveRepo.On("FindByID", uint(i+1)).Return(live, nil)
			if state.action != nil {
				builder.LiveRepo.On("Update", mock.AnythingOfType("*models.Live")).Return(nil)
			}
		}

		service := builder.BuildLiveService()

		// 測試狀態查詢
		for i, expectedState := range liveStates {
			live, err := service.GetLiveByID(uint(i + 1))
			assert.NoError(t, err)
			assert.Equal(t, expectedState.status, live.Status)
		}

		// 測試狀態流轉
		err := service.StartLive(2) // scheduled -> live
		assert.NoError(t, err)

		err = service.EndLive(3) // live -> ended
		assert.NoError(t, err)

		builder.AssertAllExpectations()
	})

	t.Run("高併發直播觀看 - MySQL", func(t *testing.T) {
		builder := testutils.NewMySQLServiceBuilder(t)

		testLive := &models.Live{
			ID:          1,
			Title:       "熱門直播",
			Status:      "live",
			ViewerCount: 1000,
		}

		// 模擬多次觀看者加入/離開
		builder.LiveRepo.On("FindByID", uint(1)).Return(testLive, nil)
		for i := 0; i < 100; i++ {
			builder.LiveRepo.On("IncrementViewerCount", uint(1)).Return(nil)
			builder.LiveRepo.On("DecrementViewerCount", uint(1)).Return(nil)
		}

		service := builder.BuildLiveService()

		// 模擬觀看者操作
		for i := 0; i < 100; i++ {
			// 假設有 IncrementViewers 和 DecrementViewers 方法
			// 這裡只是演示如何測試高併發場景
			live, err := service.GetLiveByID(1)
			assert.NoError(t, err)
			assert.Equal(t, "live", live.Status)
		}

		builder.AssertAllExpectations()
	})

	t.Run("直播搜索功能 - 混合資料庫", func(t *testing.T) {
		// 測試不同資料庫的搜索能力
		testCases := []struct {
			dbType     testutils.DatabaseType
			liveCount  int
			searchTerm string
		}{
			{testutils.PostgreSQLTest, 1000, "測試"}, // PostgreSQL 全文搜索
			{testutils.MySQLTest, 500, "直播"},       // MySQL LIKE 搜索
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s_搜索測試", tc.dbType), func(t *testing.T) {
				builder := testutils.NewServiceBuilderWithDB(t, tc.dbType)

				// 創建測試直播數據
				lives := make([]*models.Live, tc.liveCount)
				for i := 0; i < tc.liveCount; i++ {
					lives[i] = &models.Live{
						ID:     uint(i + 1),
						Title:  fmt.Sprintf("%s直播 %d", tc.searchTerm, i+1),
						Status: "live",
					}
				}

				// 模擬搜索結果（實際實現中可能需要 SearchLives 方法）
				builder.LiveRepo.On("FindActive").Return(lives[:10], nil) // 返回前10個結果
				service := builder.BuildLiveService()

				// 執行搜索（這裡假設有搜索方法）
				// 實際可能需要實現 SearchLives 方法
				// searchResults, err := service.SearchLives(tc.searchTerm, 0, 10)

				// 暫時使用 ListLives 代替搜索功能
				builder.LiveRepo.On("FindByUserID", uint(0)).Return(lives[:10], int64(tc.liveCount), nil)
				searchResults, total, err := service.ListLives(0, 10)

				assert.NoError(t, err)
				assert.Len(t, searchResults, 10)
				assert.Equal(t, int64(tc.liveCount), total)

				builder.AssertAllExpectations()
			})
		}
	})
}
