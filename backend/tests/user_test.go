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
// 🆕 使用測試工具包的改進版測試
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
// 🔄 原有測試保留（向後兼容）
// ================================

func TestUserService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := config.NewPostgreSQLConfig("config.yaml", "local")
	userService := services.NewUserService(cfg)

	tests := []struct {
		name          string
		username      string
		email         string
		password      string
		mockSetup     func()
		expectedError bool
	}{
		{
			name:     "成功註冊",
			username: "testuser",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				mockRepo.On("FindByUsername", "testuser").Return(nil, assert.AnError)
				mockRepo.On("FindByEmail", "test@example.com").Return(nil, assert.AnError)
				mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "使用者名稱已存在",
			username: "existinguser",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				mockRepo.On("FindByUsername", "existinguser").Return(&models.User{Username: "existinguser"}, nil)
			},
			expectedError: true,
		},
		{
			name:     "郵箱已存在",
			username: "testuser",
			email:    "existing@example.com",
			password: "password123",
			mockSetup: func() {
				mockRepo.On("FindByUsername", "testuser").Return(nil, assert.AnError)
				mockRepo.On("FindByEmail", "existing@example.com").Return(&models.User{Email: "existing@example.com"}, nil)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			mockRepo.Mock = mock.Mock{}
			tt.mockSetup()

			user, err := userService.Register(tt.username, tt.email, tt.password)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
				assert.Equal(t, tt.email, user.Email)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := config.NewPostgreSQLConfig("config.yaml", "local")
	userService := services.NewUserService(cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	validUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	tests := []struct {
		name          string
		email         string
		password      string
		mockSetup     func()
		expectedError bool
	}{
		{
			name:     "成功登入",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				mockRepo.On("FindByEmail", "test@example.com").Return(validUser, nil)
			},
			expectedError: false,
		},
		{
			name:     "使用者不存在",
			email:    "nonexistent@example.com",
			password: "password123",
			mockSetup: func() {
				mockRepo.On("FindByEmail", "nonexistent@example.com").Return(nil, assert.AnError)
			},
			expectedError: true,
		},
		{
			name:     "密碼錯誤",
			email:    "test@example.com",
			password: "wrongpassword",
			mockSetup: func() {
				mockRepo.On("FindByEmail", "test@example.com").Return(validUser, nil)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			mockRepo.Mock = mock.Mock{}
			tt.mockSetup()

			token, user, expiresAt, err := userService.Login(tt.email, tt.password)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, token)
				assert.Nil(t, user)
				assert.True(t, expiresAt.IsZero())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotNil(t, user)
				assert.False(t, expiresAt.IsZero())
				assert.True(t, expiresAt.After(time.Now()))
				assert.Equal(t, tt.email, user.Email)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := config.NewPostgreSQLConfig("config.yaml", "local")
	userService := services.NewUserService(cfg)

	tests := []struct {
		name          string
		id            uint
		mockSetup     func()
		expectedError bool
	}{
		{
			name: "成功獲取用戶",
			id:   1,
			mockSetup: func() {
				mockRepo.On("FindByID", uint(1)).Return(&models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "用戶不存在",
			id:   999,
			mockSetup: func() {
				mockRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			mockRepo.Mock = mock.Mock{}
			tt.mockSetup()

			user, err := userService.GetUserByID(tt.id)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.id, user.ID)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
