package testutils

import (
	"errors"
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/services"
	"testing"

	"github.com/stretchr/testify/mock"
)

// DatabaseType 測試支援的資料庫類型
type DatabaseType string

const (
	PostgreSQLTest DatabaseType = "postgresql"
	MySQLTest      DatabaseType = "mysql"
)

// ServiceBuilder 用於構建測試服務的工具
type ServiceBuilder struct {
	t            *testing.T
	UserRepo     *MockUserRepository
	VideoRepo    *MockVideoRepository
	PaymentRepo  *MockPaymentRepository
	LiveRepo     *MockLiveRepository
	databaseType DatabaseType
	configPath   string
}

// NewServiceBuilder 創建新的服務構建器（默認使用 PostgreSQL）
func NewServiceBuilder(t *testing.T) *ServiceBuilder {
	return &ServiceBuilder{
		t:            t,
		UserRepo:     NewMockUserRepository(),
		VideoRepo:    NewMockVideoRepository(),
		PaymentRepo:  NewMockPaymentRepository(),
		LiveRepo:     NewMockLiveRepository(),
		databaseType: PostgreSQLTest, // 默認使用 PostgreSQL
		configPath:   "../config/config.test.yaml",
	}
}

// NewServiceBuilderWithDB 創建指定資料庫類型的服務構建器
func NewServiceBuilderWithDB(t *testing.T, dbType DatabaseType) *ServiceBuilder {
	return &ServiceBuilder{
		t:            t,
		UserRepo:     NewMockUserRepository(),
		VideoRepo:    NewMockVideoRepository(),
		PaymentRepo:  NewMockPaymentRepository(),
		LiveRepo:     NewMockLiveRepository(),
		databaseType: dbType,
		configPath:   "../config/config.test.yaml",
	}
}

// NewServiceBuilderWithConfig 創建使用自定義配置的服務構建器
func NewServiceBuilderWithConfig(t *testing.T, configPath string, dbType DatabaseType) *ServiceBuilder {
	return &ServiceBuilder{
		t:            t,
		UserRepo:     NewMockUserRepository(),
		VideoRepo:    NewMockVideoRepository(),
		PaymentRepo:  NewMockPaymentRepository(),
		LiveRepo:     NewMockLiveRepository(),
		databaseType: dbType,
		configPath:   configPath,
	}
}

// WithDatabase 設定資料庫類型
func (sb *ServiceBuilder) WithDatabase(dbType DatabaseType) *ServiceBuilder {
	sb.databaseType = dbType
	return sb
}

// WithConfig 設定配置文件路徑
func (sb *ServiceBuilder) WithConfig(configPath string) *ServiceBuilder {
	sb.configPath = configPath
	return sb
}

// Mock Repository 接口
type MockUserRepository struct{ mock.Mock }
type MockVideoRepository struct{ mock.Mock }
type MockPaymentRepository struct{ mock.Mock }
type MockLiveRepository struct{ mock.Mock }

// UserRepository Mock 方法
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

// VideoRepository Mock 方法
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

// PaymentRepository Mock 方法
func (m *MockPaymentRepository) Create(payment *models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByID(id uint) (*models.Payment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByUserID(userID uint) ([]models.Payment, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByTransactionID(transactionID string) (*models.Payment, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(payment *models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// LiveRepository Mock 方法
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

func (m *MockLiveRepository) FindActive() ([]*models.Live, error) {
	args := m.Called()
	return args.Get(0).([]*models.Live), args.Error(1)
}

func (m *MockLiveRepository) Update(live *models.Live) error {
	args := m.Called(live)
	return args.Error(0)
}

func (m *MockLiveRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLiveRepository) IncrementViewerCount(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLiveRepository) DecrementViewerCount(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// 工廠方法
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

func NewMockVideoRepository() *MockVideoRepository {
	return &MockVideoRepository{}
}

func NewMockPaymentRepository() *MockPaymentRepository {
	return &MockPaymentRepository{}
}

func NewMockLiveRepository() *MockLiveRepository {
	return &MockLiveRepository{}
}

// 構建器方法 - 鏈式調用
func (sb *ServiceBuilder) WithUser(user *models.User) *ServiceBuilder {
	sb.UserRepo.On("FindByID", user.ID).Return(user, nil)
	return sb
}

func (sb *ServiceBuilder) WithUserNotFound(userID uint) *ServiceBuilder {
	sb.UserRepo.On("FindByID", userID).Return((*models.User)(nil), errors.New("用戶不存在"))
	return sb
}

func (sb *ServiceBuilder) WithVideo(video *models.Video) *ServiceBuilder {
	sb.VideoRepo.On("FindByID", video.ID).Return(video, nil)
	return sb
}

func (sb *ServiceBuilder) WithVideoNotFound(videoID uint) *ServiceBuilder {
	sb.VideoRepo.On("FindByID", videoID).Return(nil, mock.AnythingOfType("error"))
	return sb
}

func (sb *ServiceBuilder) WithCreateVideoSuccess() *ServiceBuilder {
	sb.VideoRepo.On("Create", mock.AnythingOfType("*models.Video")).Return(nil)
	return sb
}

func (sb *ServiceBuilder) WithCreateVideoError() *ServiceBuilder {
	sb.VideoRepo.On("Create", mock.AnythingOfType("*models.Video")).Return(mock.AnythingOfType("error"))
	return sb
}

// 配置創建輔助方法
func (sb *ServiceBuilder) createTestConfig() *config.Config {
	return config.NewConfig(sb.configPath, "test", string(sb.databaseType))
}

// 服務創建方法（使用新的配置系統）
func (sb *ServiceBuilder) BuildVideoService() *services.VideoService {
	cfg := sb.createTestConfig()
	return services.NewVideoService(cfg)
}

func (sb *ServiceBuilder) BuildUserService() *services.UserService {
	cfg := sb.createTestConfig()
	return services.NewUserService(cfg)
}

func (sb *ServiceBuilder) BuildPaymentService() *services.PaymentService {
	cfg := sb.createTestConfig()
	return services.NewPaymentService(cfg)
}

func (sb *ServiceBuilder) BuildLiveService() *services.LiveService {
	cfg := sb.createTestConfig()
	liveService, err := services.NewLiveService(cfg)
	if err != nil {
		sb.t.Fatalf("Failed to create LiveService: %v", err)
	}
	return liveService
}

// 斷言輔助方法
func (sb *ServiceBuilder) AssertAllExpectations() {
	sb.UserRepo.AssertExpectations(sb.t)
	sb.VideoRepo.AssertExpectations(sb.t)
	sb.PaymentRepo.AssertExpectations(sb.t)
	sb.LiveRepo.AssertExpectations(sb.t)
}

// 便利方法：快速創建不同資料庫類型的構建器
func NewPostgreSQLServiceBuilder(t *testing.T) *ServiceBuilder {
	return NewServiceBuilderWithDB(t, PostgreSQLTest)
}

func NewMySQLServiceBuilder(t *testing.T) *ServiceBuilder {
	return NewServiceBuilderWithDB(t, MySQLTest)
}

// 測試輔助方法：檢查配置是否正確載入
func (sb *ServiceBuilder) ValidateConfig() error {
	cfg := sb.createTestConfig()
	if cfg == nil {
		return errors.New("配置創建失敗")
	}
	if cfg.ActiveDatabase != string(sb.databaseType) {
		return errors.New("資料庫類型配置不匹配")
	}
	return nil
}

// 測試輔助方法：獲取當前資料庫類型
func (sb *ServiceBuilder) GetDatabaseType() DatabaseType {
	return sb.databaseType
}

// 測試輔助方法：獲取配置信息
func (sb *ServiceBuilder) GetConfigInfo() map[string]interface{} {
	cfg := sb.createTestConfig()
	return map[string]interface{}{
		"active_database": cfg.ActiveDatabase,
		"available":       cfg.GetAvailableDatabases(),
		"config_path":     sb.configPath,
		"database_type":   sb.databaseType,
	}
}
