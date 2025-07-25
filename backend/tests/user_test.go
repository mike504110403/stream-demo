package tests

import (
	"stream-demo/backend/database/models"
	"stream-demo/backend/services"
	"stream-demo/backend/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository 模擬使用者資料庫操作
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// ================================
// 🆕 使用新測試工具包的改進版測試
// ================================

func TestUserService_Register_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功註冊", func(t *testing.T) {
		// 改進後：簡化設置
		builder := testutils.NewServiceBuilder(t)
		// 設置用戶名和郵箱不存在的期望
		builder.UserRepo.On("FindByUsername", "testuser").Return((*models.User)(nil), assert.AnError)
		builder.UserRepo.On("FindByEmail", "test@example.com").Return((*models.User)(nil), assert.AnError)
		builder.UserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

		service := builder.BuildUserService()

		// Act
		user, err := service.Register("testuser", "test@example.com", "password123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：使用者名稱已存在", func(t *testing.T) {
		// 改進後：直接設置期望結果
		existingUser := &models.User{Username: "existinguser"}
		builder := testutils.NewServiceBuilder(t)
		builder.UserRepo.On("FindByUsername", "existinguser").Return(existingUser, nil)

		service := builder.BuildUserService()

		// Act
		user, err := service.Register("existinguser", "test@example.com", "password123")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：郵箱已存在", func(t *testing.T) {
		// 改進後：鏈式設置多個期望
		existingUser := &models.User{Email: "existing@example.com"}
		builder := testutils.NewServiceBuilder(t)
		builder.UserRepo.On("FindByUsername", "testuser").Return((*models.User)(nil), assert.AnError)
		builder.UserRepo.On("FindByEmail", "existing@example.com").Return(existingUser, nil)

		service := builder.BuildUserService()

		// Act
		user, err := service.Register("testuser", "existing@example.com", "password123")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)

		builder.AssertAllExpectations()
	})
}

func TestUserService_Login_WithToolkit(t *testing.T) {
	// 準備測試用戶
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	validUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	t.Run("🟢 改進版：成功登入", func(t *testing.T) {
		// 改進後：直接設置用戶
		builder := testutils.NewServiceBuilder(t)
		builder.UserRepo.On("FindByEmail", "test@example.com").Return(validUser, nil)

		service := builder.BuildUserService()

		// Act
		token, user, expiresAt, err := service.Login("test@example.com", "password123")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, user)
		assert.False(t, expiresAt.IsZero())
		assert.True(t, expiresAt.After(time.Now()))
		assert.Equal(t, "test@example.com", user.Email)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：使用者不存在", func(t *testing.T) {
		// 改進後：使用工具包的便利方法
		builder := testutils.NewServiceBuilder(t)
		builder.UserRepo.On("FindByEmail", "nonexistent@example.com").Return((*models.User)(nil), assert.AnError)

		service := builder.BuildUserService()

		// Act
		token, user, expiresAt, err := service.Login("nonexistent@example.com", "password123")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
		assert.True(t, expiresAt.IsZero())

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：密碼錯誤", func(t *testing.T) {
		// 改進後：重用用戶設置
		builder := testutils.NewServiceBuilder(t)
		builder.UserRepo.On("FindByEmail", "test@example.com").Return(validUser, nil)

		service := builder.BuildUserService()

		// Act
		token, user, expiresAt, err := service.Login("test@example.com", "wrongpassword")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
		assert.True(t, expiresAt.IsZero())

		builder.AssertAllExpectations()
	})
}

func TestUserService_GetUserByID_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功獲取用戶", func(t *testing.T) {
		// 改進後：使用工具包的便利方法
		testUser := &models.User{ID: 1, Username: "testuser", Email: "test@example.com"}
		builder := testutils.NewServiceBuilder(t).
			WithUser(testUser)

		service := builder.BuildUserService()

		// Act
		user, err := service.GetUserByID(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, uint(1), user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：用戶不存在", func(t *testing.T) {
		// 改進後：使用工具包的便利方法
		builder := testutils.NewServiceBuilder(t).
			WithUserNotFound(999)

		service := builder.BuildUserService()

		// Act
		user, err := service.GetUserByID(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)

		builder.AssertAllExpectations()
	})
}

// ================================
// 🚀 多資料庫測試
// ================================

func TestUserService_MultiDatabase(t *testing.T) {
	// 準備測試數據
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	testCases := []struct {
		name      string
		dbType    testutils.DatabaseType
		setupTest func(builder *testutils.ServiceBuilder) *services.UserService
		runTest   func(service *services.UserService) error
		wantError bool
	}{
		{
			name:   "PostgreSQL 用戶註冊",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.UserService {
				builder.UserRepo.On("FindByUsername", "newuser").Return((*models.User)(nil), assert.AnError)
				builder.UserRepo.On("FindByEmail", "new@example.com").Return((*models.User)(nil), assert.AnError)
				builder.UserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
				return builder.BuildUserService()
			},
			runTest: func(service *services.UserService) error {
				_, err := service.Register("newuser", "new@example.com", "password123")
				return err
			},
			wantError: false,
		},
		{
			name:   "MySQL 用戶登入",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.UserService {
				builder.UserRepo.On("FindByEmail", "test@example.com").Return(testUser, nil)
				return builder.BuildUserService()
			},
			runTest: func(service *services.UserService) error {
				_, _, _, err := service.Login("test@example.com", "password123")
				return err
			},
			wantError: false,
		},
		{
			name:   "PostgreSQL 用戶查詢",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.UserService {
				return builder.WithUser(testUser).BuildUserService()
			},
			runTest: func(service *services.UserService) error {
				_, err := service.GetUserByID(1)
				return err
			},
			wantError: false,
		},
		{
			name:   "MySQL 用戶查詢",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.UserService {
				return builder.WithUser(testUser).BuildUserService()
			},
			runTest: func(service *services.UserService) error {
				_, err := service.GetUserByID(1)
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
// 🔧 測試工具配置驗證
// ================================

func TestServiceBuilder_Configuration(t *testing.T) {
	t.Run("預設 PostgreSQL 配置", func(t *testing.T) {
		builder := testutils.NewServiceBuilder(t)

		assert.NoError(t, builder.ValidateConfig())
		assert.Equal(t, testutils.PostgreSQLTest, builder.GetDatabaseType())

		configInfo := builder.GetConfigInfo()
		assert.Equal(t, "postgresql", configInfo["active_database"])
		assert.Contains(t, configInfo["available"], "postgresql")
		assert.Contains(t, configInfo["available"], "mysql")
	})

	t.Run("MySQL 配置", func(t *testing.T) {
		builder := testutils.NewMySQLServiceBuilder(t)

		assert.NoError(t, builder.ValidateConfig())
		assert.Equal(t, testutils.MySQLTest, builder.GetDatabaseType())

		configInfo := builder.GetConfigInfo()
		assert.Equal(t, "mysql", configInfo["active_database"])
	})

	t.Run("動態切換資料庫類型", func(t *testing.T) {
		builder := testutils.NewServiceBuilder(t)

		// 初始為 PostgreSQL
		assert.Equal(t, testutils.PostgreSQLTest, builder.GetDatabaseType())

		// 切換到 MySQL
		builder.WithDatabase(testutils.MySQLTest)
		assert.Equal(t, testutils.MySQLTest, builder.GetDatabaseType())
		assert.NoError(t, builder.ValidateConfig())
	})
}
