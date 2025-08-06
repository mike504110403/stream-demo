package mocks

import (
	"context"
	"database/sql"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"stream-demo/backend/database/models"
	"stream-demo/backend/dto"
	"stream-demo/backend/pkg/storage"
)

// MockRedisClient 模擬 Redis 客戶端
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) TTL(ctx context.Context, key string) *redis.DurationCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.DurationCmd)
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Decr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) MGet(ctx context.Context, keys ...string) *redis.SliceCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.SliceCmd)
}

func (m *MockRedisClient) MSet(ctx context.Context, pairs ...interface{}) *redis.StatusCmd {
	args := m.Called(ctx, pairs)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockPubSub 模擬 Redis PubSub
type MockPubSub struct {
	mock.Mock
}

func (m *MockPubSub) Channel(opts ...redis.ChannelOption) <-chan *redis.Message {
	args := m.Called(opts)
	return args.Get(0).(<-chan *redis.Message)
}

func (m *MockPubSub) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockGormDB 模擬 GORM 數據庫
type MockGormDB struct {
	mock.Mock
}

func (m *MockGormDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Update(column string, value interface{}) *gorm.DB {
	args := m.Called(column, value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(value, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Begin() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Commit() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Rollback() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Error() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGormDB) RowsAffected() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

// MockHTTPClient 模擬 HTTP 客戶端
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req interface{}) (interface{}, error) {
	args := m.Called(req)
	return args.Get(0), args.Error(1)
}

// MockWebSocket 模擬 WebSocket 連接
type MockWebSocket struct {
	mock.Mock
}

func (m *MockWebSocket) WriteMessage(messageType int, data []byte) error {
	args := m.Called(messageType, data)
	return args.Error(0)
}

func (m *MockWebSocket) ReadMessage() (messageType int, p []byte, err error) {
	args := m.Called()
	return args.Get(0).(int), args.Get(1).([]byte), args.Error(2)
}

func (m *MockWebSocket) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockFileSystem 模擬文件系統
type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) Create(name string) (interface{}, error) {
	args := m.Called(name)
	return args.Get(0), args.Error(1)
}

func (m *MockFileSystem) Open(name string) (interface{}, error) {
	args := m.Called(name)
	return args.Get(0), args.Error(1)
}

func (m *MockFileSystem) Remove(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockFileSystem) Stat(name string) (interface{}, error) {
	args := m.Called(name)
	return args.Get(0), args.Error(1)
}

// MockStorage 模擬存儲服務
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Upload(file interface{}, key string) (string, error) {
	args := m.Called(file, key)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) Download(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockStorage) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockStorage) GetURL(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

// MockMediaProcessor 模擬媒體處理器
type MockMediaProcessor struct {
	mock.Mock
}

func (m *MockMediaProcessor) Transcode(input string, output string, options map[string]interface{}) error {
	args := m.Called(input, output, options)
	return args.Error(0)
}

func (m *MockMediaProcessor) GenerateThumbnail(input string, output string, time string) error {
	args := m.Called(input, output, time)
	return args.Error(0)
}

func (m *MockMediaProcessor) GetMediaInfo(input string) (map[string]interface{}, error) {
	args := m.Called(input)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// CreateMockGormDB 創建模擬 GORM 數據庫
func CreateMockGormDB() (*MockGormDB, sqlmock.Sqlmock, *sql.DB, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, nil, err
	}

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	_, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, nil, nil, err
	}

	mockGormDB := &MockGormDB{}
	return mockGormDB, mock, db, nil
}

// CreateMockRedisClient 創建模擬 Redis 客戶端
func CreateMockRedisClient() *MockRedisClient {
	return &MockRedisClient{}
}

// CreateMockPubSub 創建模擬 PubSub
func CreateMockPubSub() *MockPubSub {
	return &MockPubSub{}
}

// CreateMockWebSocket 創建模擬 WebSocket
func CreateMockWebSocket() *MockWebSocket {
	return &MockWebSocket{}
}

// CreateMockStorage 創建模擬存儲服務
func CreateMockStorage() *MockStorage {
	return &MockStorage{}
}

// CreateMockMediaProcessor 創建模擬媒體處理器
func CreateMockMediaProcessor() *MockMediaProcessor {
	return &MockMediaProcessor{}
}

// MockUserService 模擬用戶服務
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(username, email, password string) (*dto.UserDTO, error) {
	args := m.Called(username, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDTO), args.Error(1)
}

func (m *MockUserService) Login(username, password string) (string, *dto.UserDTO, time.Time, error) {
	args := m.Called(username, password)
	return args.String(0), args.Get(1).(*dto.UserDTO), args.Get(2).(time.Time), args.Error(3)
}

func (m *MockUserService) GetUserByID(userID uint) (*dto.UserDTO, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDTO), args.Error(1)
}

func (m *MockUserService) UpdateUser(userID uint, username, email, avatar, bio string) (*dto.UserDTO, error) {
	args := m.Called(userID, username, email, avatar, bio)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDTO), args.Error(1)
}

func (m *MockUserService) DeleteUser(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

// MockVideoService 模擬影片服務
type MockVideoService struct {
	mock.Mock
}

func (m *MockVideoService) GenerateUploadURL(userID uint, filename string, fileSize int64) (*storage.PresignedUploadURL, error) {
	args := m.Called(userID, filename, fileSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.PresignedUploadURL), args.Error(1)
}

func (m *MockVideoService) CreateVideoRecord(userID uint, title, description, s3Key string) (*dto.VideoDTO, error) {
	args := m.Called(userID, title, description, s3Key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VideoDTO), args.Error(1)
}

func (m *MockVideoService) ConfirmUploadOnly(videoID uint, s3Key string) error {
	args := m.Called(videoID, s3Key)
	return args.Error(0)
}

func (m *MockVideoService) GetVideos(offset, limit int) ([]*dto.VideoDTO, int64, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*dto.VideoDTO), args.Get(1).(int64), args.Error(2)
}

func (m *MockVideoService) GetVideoByID(videoID uint) (*dto.VideoDTO, error) {
	args := m.Called(videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VideoDTO), args.Error(1)
}

func (m *MockVideoService) GetVideosByUserID(userID uint) ([]*dto.VideoDTO, int64, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*dto.VideoDTO), args.Get(1).(int64), args.Error(2)
}

func (m *MockVideoService) UpdateVideo(id uint, title string, description string, videoData *dto.VideoDTO) error {
	args := m.Called(id, title, description, videoData)
	return args.Error(0)
}

func (m *MockVideoService) DeleteVideo(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockVideoService) SearchVideos(query string, offset, limit int) ([]*dto.VideoDTO, int64, error) {
	args := m.Called(query, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*dto.VideoDTO), args.Get(1).(int64), args.Error(2)
}

func (m *MockVideoService) LikeVideo(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockVideoService) IncrementViews(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockVideoService) IncrementLikes(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockLiveService 模擬直播服務
type MockLiveService struct {
	mock.Mock
}

func (m *MockLiveService) ListLives(offset, limit int) ([]*dto.LiveDTO, int64, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*dto.LiveDTO), args.Get(1).(int64), args.Error(2)
}

func (m *MockLiveService) CreateLive(userID uint, title, description string, startTime time.Time) (*dto.LiveDTO, error) {
	args := m.Called(userID, title, description, startTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LiveDTO), args.Error(1)
}

func (m *MockLiveService) GetLiveByID(id uint) (*dto.LiveDTO, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LiveDTO), args.Error(1)
}

func (m *MockLiveService) GetLivesByUserID(userID uint) ([]*dto.LiveDTO, int64, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*dto.LiveDTO), args.Get(1).(int64), args.Error(2)
}

func (m *MockLiveService) UpdateLive(id uint, title, description string, startTime time.Time) (*dto.LiveDTO, error) {
	args := m.Called(id, title, description, startTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LiveDTO), args.Error(1)
}

func (m *MockLiveService) DeleteLive(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLiveService) StartLive(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLiveService) EndLive(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLiveService) GetStreamKey(id uint) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockLiveService) UpdateViewerCount(id uint, count int64) error {
	args := m.Called(id, count)
	return args.Error(0)
}

func (m *MockLiveService) ToggleChat(id uint, enabled bool) error {
	args := m.Called(id, enabled)
	return args.Error(0)
}

func (m *MockLiveService) GetActiveLives() ([]*dto.LiveDTO, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.LiveDTO), args.Error(1)
}

// MockS3Client S3客戶端mock
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) PutObject(input interface{}) (interface{}, error) {
	args := m.Called(input)
	return args.Get(0), args.Error(1)
}

func (m *MockS3Client) HeadObject(input interface{}) (interface{}, error) {
	args := m.Called(input)
	return args.Get(0), args.Error(1)
}

func (m *MockS3Client) DeleteObject(input interface{}) (interface{}, error) {
	args := m.Called(input)
	return args.Get(0), args.Error(1)
}



// MockWebSocketHub WebSocket Hub mock
type MockWebSocketHub struct {
	mock.Mock
}

func (m *MockWebSocketHub) GetRoom(roomID string) interface{} {
	args := m.Called(roomID)
	return args.Get(0)
}

func (m *MockWebSocketHub) GetRoomStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockWebSocketHub) Close() {
	m.Called()
}

func (m *MockWebSocketHub) PublishChatMessage(roomID string, message interface{}) {
	m.Called(roomID, message)
}

func (m *MockWebSocketHub) HandleChatMessage(roomID string, message interface{}) {
	m.Called(roomID, message)
}

func (m *MockWebSocketHub) HandleLiveUpdate(roomID string, update interface{}) {
	m.Called(roomID, update)
}

// MockPaymentService 支付服務mock
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreatePayment(userID uint, amount float64, currency string) (*dto.PaymentDTO, error) {
	args := m.Called(userID, amount, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentDTO), args.Error(1)
}

func (m *MockPaymentService) GetPaymentByID(paymentID uint) (*dto.PaymentDTO, error) {
	args := m.Called(paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentDTO), args.Error(1)
}

func (m *MockPaymentService) UpdatePaymentStatus(paymentID uint, status string) error {
	args := m.Called(paymentID, status)
	return args.Error(0)
}

func (m *MockPaymentService) GetPaymentsByUserID(userID uint) ([]*dto.PaymentDTO, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.PaymentDTO), args.Error(1)
}

// MockPublicStreamService 公共串流服務mock
type MockPublicStreamService struct {
	mock.Mock
}

func (m *MockPublicStreamService) GetPublicStreams() ([]*dto.PublicStreamDTO, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.PublicStreamDTO), args.Error(1)
}

func (m *MockPublicStreamService) GetPublicStreamByID(id uint) (*dto.PublicStreamDTO, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PublicStreamDTO), args.Error(1)
}

func (m *MockPublicStreamService) CreatePublicStream(name, description, streamURL string) (*dto.PublicStreamDTO, error) {
	args := m.Called(name, description, streamURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PublicStreamDTO), args.Error(1)
}

func (m *MockPublicStreamService) UpdatePublicStream(id uint, name, description, streamURL string) (*dto.PublicStreamDTO, error) {
	args := m.Called(id, name, description, streamURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PublicStreamDTO), args.Error(1)
}

func (m *MockPublicStreamService) DeletePublicStream(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockUserRepository 用戶倉儲mock
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindUserByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockVideoRepository 影片倉儲mock
type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) CreateVideo(video *models.Video) error {
	args := m.Called(video)
	return args.Error(0)
}

func (m *MockVideoRepository) FindVideoByID(id uint) (*models.Video, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Video), args.Error(1)
}

func (m *MockVideoRepository) FindVideos(offset, limit int) ([]*models.Video, int64, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Video), args.Get(1).(int64), args.Error(2)
}

func (m *MockVideoRepository) UpdateVideo(video *models.Video) error {
	args := m.Called(video)
	return args.Error(0)
}

func (m *MockVideoRepository) DeleteVideo(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockLiveRepository 直播倉儲mock
type MockLiveRepository struct {
	mock.Mock
}

func (m *MockLiveRepository) CreateLive(live *models.Live) error {
	args := m.Called(live)
	return args.Error(0)
}

func (m *MockLiveRepository) FindLiveByID(id uint) (*models.Live, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Live), args.Error(1)
}

func (m *MockLiveRepository) FindLives(offset, limit int) ([]*models.Live, int64, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Live), args.Get(1).(int64), args.Error(2)
}

func (m *MockLiveRepository) UpdateLive(live *models.Live) error {
	args := m.Called(live)
	return args.Error(0)
}

func (m *MockLiveRepository) DeleteLive(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
