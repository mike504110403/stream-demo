package di

import (
	"fmt"
	"stream-demo/backend/api"
	"stream-demo/backend/config"
	postgresqlRepo "stream-demo/backend/repositories/postgresql"
	"stream-demo/backend/services"
	"stream-demo/backend/utils"
	"stream-demo/backend/ws"
	"time"
)

// Container 依賴注入容器
type Container struct {
	// 配置
	Config *config.Config

	// 工具
	JWTUtil   *utils.JWTUtil
	Cache     interface{}
	Messaging *utils.RedisMessaging

	// WebSocket
	Hub               *ws.Hub
	WSHandler         *ws.Handler
	LiveRoomWSHandler *ws.LiveRoomHandler

	// 倉儲層
	UserRepo    *postgresqlRepo.PostgreSQLRepo
	VideoRepo   *postgresqlRepo.PostgreSQLRepo
	LiveRepo    *postgresqlRepo.PostgreSQLRepo
	PaymentRepo *postgresqlRepo.PostgreSQLRepo

	// 服務層
	UserService         *services.UserService
	VideoService        *services.VideoService
	LiveService         *services.LiveService
	LiveRoomService     *services.LiveRoomService
	LiveRoomSyncService *services.LiveRoomSyncService
	PaymentService      *services.PaymentService
	PublicStreamService *services.PublicStreamService
	TranscodeWorker     *services.TranscodeWorker

	// 處理器層
	UserHandler         *api.UserHandler
	VideoHandler        *api.VideoHandler
	LiveHandler         *api.LiveHandler
	LiveRoomHandler     *api.LiveRoomHandler
	PaymentHandler      *api.PaymentHandler
	PublicStreamHandler *api.PublicStreamHandler

	// 路由
	Router *api.Router
}

// NewContainer 創建依賴注入容器
func NewContainer(cfg *config.Config) (*Container, error) {
	container := &Container{
		Config: cfg,
	}

	// 初始化工具
	if err := container.initUtils(); err != nil {
		return nil, err
	}

	// 初始化倉儲層
	if err := container.initRepositories(); err != nil {
		return nil, err
	}

	// 初始化服務層
	if err := container.initServices(); err != nil {
		return nil, err
	}

	// 初始化處理器層
	if err := container.initHandlers(); err != nil {
		return nil, err
	}

	// 初始化 WebSocket
	if err := container.initWebSocket(); err != nil {
		return nil, err
	}

	// 設置 WebSocket 處理器到服務中
	container.LiveRoomService.SetWSHandler(container.LiveRoomWSHandler)

	return container, nil
}

// initUtils 初始化工具
func (c *Container) initUtils() error {
	// 初始化 JWT 工具
	c.JWTUtil = utils.NewJWTUtil(c.Config.JWT.Secret)

	// 初始化 Redis 客戶端
	if err := utils.InitRedisClient(
		utils.RedisConfig{
			Host:           c.Config.Redis.Master.Host,
			Port:           c.Config.Redis.Master.Port,
			Password:       c.Config.Redis.Master.Password,
			DB:             c.Config.Redis.Master.DB,
			MaxActive:      c.Config.Redis.Pool.MaxActive,
			MaxIdle:        c.Config.Redis.Pool.MaxIdle,
			IdleTimeout:    c.Config.Redis.Pool.IdleTimeout,
			ConnectTimeout: c.Config.Redis.Pool.ConnectTimeout,
			ReadTimeout:    c.Config.Redis.Pool.ReadTimeout,
			WriteTimeout:   c.Config.Redis.Pool.WriteTimeout,
		},
		utils.RedisConfig{
			Host:           c.Config.Redis.Slave.Host,
			Port:           c.Config.Redis.Slave.Port,
			Password:       c.Config.Redis.Slave.Password,
			DB:             c.Config.Redis.Slave.DB,
			MaxActive:      c.Config.Redis.Pool.MaxActive,
			MaxIdle:        c.Config.Redis.Pool.MaxIdle,
			IdleTimeout:    c.Config.Redis.Pool.IdleTimeout,
			ConnectTimeout: c.Config.Redis.Pool.ConnectTimeout,
			ReadTimeout:    c.Config.Redis.Pool.ReadTimeout,
			WriteTimeout:   c.Config.Redis.Pool.WriteTimeout,
		},
	); err != nil {
		return fmt.Errorf("init Redis client failed: %v", err)
	}

	// 初始化緩存
	if c.Config.Cache.Type == "redis" {
		c.Cache = utils.NewRedisCache(
			c.Config.Cache.DB,
			c.Config.Cache.KeyPrefix,
			time.Duration(c.Config.Cache.DefaultExpiration)*time.Second,
		)
	} else {
		c.Cache = utils.NewPostgreSQLCache(
			c.Config.DB["master"],
			c.Config.Cache.TableName,
			time.Duration(c.Config.Cache.DefaultExpiration)*time.Second,
			time.Duration(c.Config.Cache.CleanupInterval)*time.Second,
		)
	}

	// 初始化訊息系統
	if c.Config.Messaging.Type == "redis" {
		messaging, err := utils.NewRedisMessaging(c.Config.Messaging.DB)
		if err != nil {
			return err
		}
		c.Messaging = messaging
	}

	return nil
}

// initRepositories 初始化倉儲層
func (c *Container) initRepositories() error {
	// 初始化 PostgreSQL 倉儲
	c.UserRepo = postgresqlRepo.NewPostgreSQLRepo(c.Config.DB["master"])
	c.VideoRepo = postgresqlRepo.NewPostgreSQLRepo(c.Config.DB["master"])
	c.LiveRepo = postgresqlRepo.NewPostgreSQLRepo(c.Config.DB["master"])
	c.PaymentRepo = postgresqlRepo.NewPostgreSQLRepo(c.Config.DB["master"])

	return nil
}

// initServices 初始化服務層
func (c *Container) initServices() error {
	// 初始化用戶服務
	c.UserService = services.NewUserService(c.Config)

	// 初始化影片服務
	c.VideoService = services.NewVideoService(c.Config)

	// 初始化直播服務
	liveService, err := services.NewLiveService(c.Config)
	if err != nil {
		return fmt.Errorf("init live service failed: %v", err)
	}
	c.LiveService = liveService

	// 初始化直播間服務
	c.LiveRoomService = services.NewLiveRoomService(c.Config, c.Config.DB["master"])

	// 初始化直播間同步服務
	c.LiveRoomSyncService = services.NewLiveRoomSyncService(c.LiveRoomService)

	// 初始化支付服務
	c.PaymentService = services.NewPaymentService(c.Config)

	// 初始化公開流服務
	if redisCache, ok := c.Cache.(*utils.RedisCache); ok {
		publicStreamService, err := services.NewPublicStreamService(c.Config, redisCache)
		if err != nil {
			return fmt.Errorf("init public stream service failed: %v", err)
		}
		c.PublicStreamService = publicStreamService
	}

	// 初始化轉碼工作服務
	c.TranscodeWorker = services.NewTranscodeWorker(c.VideoService)

	return nil
}

// initHandlers 初始化處理器層
func (c *Container) initHandlers() error {
	// 初始化用戶處理器
	c.UserHandler = api.NewUserHandler(c.UserService)

	// 初始化影片處理器
	c.VideoHandler = api.NewVideoHandler(c.VideoService)

	// 初始化直播處理器
	c.LiveHandler = api.NewLiveHandler(c.LiveService)

	// 初始化直播間處理器
	c.LiveRoomHandler = api.NewLiveRoomHandler(c.LiveRoomService)

	// 初始化支付處理器
	c.PaymentHandler = api.NewPaymentHandler(c.PaymentService)

	// 初始化公開流處理器
	if c.PublicStreamService != nil {
		c.PublicStreamHandler = api.NewPublicStreamHandler(c.PublicStreamService)
	}

	return nil
}

// initWebSocket 初始化 WebSocket
func (c *Container) initWebSocket() error {
	// 初始化 WebSocket Hub
	c.Hub = ws.NewHub(c.Messaging)

	// 初始化 WebSocket Handler
	c.WSHandler = ws.NewHandler(c.Hub)

	// 初始化直播間 WebSocket Handler
	c.LiveRoomWSHandler = ws.NewLiveRoomHandler(c.JWTUtil)

	return nil
}

// StartServices 啟動所有服務
func (c *Container) StartServices() {
	// 啟動轉碼工作服務
	if c.TranscodeWorker != nil {
		c.TranscodeWorker.Start()
	}

	// 啟動直播間同步服務
	if c.LiveRoomSyncService != nil {
		c.LiveRoomSyncService.Start()
	}

	// WebSocket Hub 不需要額外啟動，會在需要時自動創建房間
}

// StopServices 停止所有服務
func (c *Container) StopServices() {
	// 停止轉碼工作服務
	if c.TranscodeWorker != nil {
		c.TranscodeWorker.Stop()
	}

	// 停止直播間同步服務
	if c.LiveRoomSyncService != nil {
		c.LiveRoomSyncService.Stop()
	}

	// 關閉 WebSocket Hub
	if c.Hub != nil {
		c.Hub.Close()
	}
}
