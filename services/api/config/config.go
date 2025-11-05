package config

import (
	"fmt"
	"os"
	"strings"

	"stream-demo/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Configurations exported
type Configurations struct {
	Gin       GinConfigurations
	Image     ImageConfigurations
	Const     map[string]interface{}
	Swagger   SwaggerConfigurations
	JwtBearer JwtBearerConfigurations
	// 多資料庫配置（支援 MySQL 和 PostgreSQL）
	Databases map[string]DatabaseConfiguration `mapstructure:"databases"`
	Redis     RedisConfiguration               `mapstructure:"redis"`
	Cache     CacheConfiguration               `mapstructure:"cache"`
	Messaging MessagingConfiguration           `mapstructure:"messaging"`
	JWT       JWTConfiguration                 `mapstructure:"jwt"`
	// 其他配置字段
	Storage      StorageConfiguration      `mapstructure:"storage"`
	MediaConvert MediaConvertConfiguration `mapstructure:"media_convert"`
	Transcode    TranscodeConfiguration    `mapstructure:"transcode"` // 新增轉碼配置
	Video        VideoConfiguration        `mapstructure:"video"`
	// 直播配置
	Live LiveConfiguration `mapstructure:"live"`
}

type SwaggerConfigurations struct {
	Host string
	Path string
}

type JwtBearerConfigurations struct {
	SecurityKey         string
	Issuer              string
	Audience            string
	Sub                 string
	JwtExpires          int
	RefreshTokenExpires int
}

type GinConfigurations struct {
	Mode string
	Host string
	Port int
}

type ImageConfigurations struct {
	ImagePrefixUrl     string
	ImageSavePath      string
	OssAccessKeyID     string
	OssAccessKeySecret string
	OssEndpoint        string
}

// DatabaseConfiguration 資料庫配置（支援 MySQL 和 PostgreSQL）
type DatabaseConfiguration struct {
	Type   string                    `mapstructure:"type"` // "mysql" 或 "postgresql"
	Master DatabaseConnectionConfig  `mapstructure:"master"`
	Slave  DatabaseConnectionConfig  `mapstructure:"slave"`
	Pool   DatabasePoolConfiguration `mapstructure:"pool"`
}

// DatabaseConnectionConfig 資料庫連接配置
type DatabaseConnectionConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"` // PostgreSQL: disable/require, MySQL: true/false/skip-verify
}

// DatabasePoolConfiguration 資料庫連接池配置
type DatabasePoolConfiguration struct {
	MaxOpenConns    int `mapstructure:"max_open_conns"`
	MaxIdleConns    int `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"`
}

// RedisConfiguration Redis配置
type RedisConfiguration struct {
	Master RedisConnectionConfig  `mapstructure:"master"`
	Slave  RedisConnectionConfig  `mapstructure:"slave"`
	Pool   RedisPoolConfiguration `mapstructure:"pool"`
}

// RedisConnectionConfig Redis連接配置
type RedisConnectionConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// RedisPoolConfiguration Redis連接池配置
type RedisPoolConfiguration struct {
	MaxActive      int `mapstructure:"max_active"`
	MaxIdle        int `mapstructure:"max_idle"`
	IdleTimeout    int `mapstructure:"idle_timeout"`
	ConnectTimeout int `mapstructure:"connect_timeout"`
	ReadTimeout    int `mapstructure:"read_timeout"`
	WriteTimeout   int `mapstructure:"write_timeout"`
}

// GinConfiguration Gin框架配置
type GinConfiguration struct {
	Mode string `mapstructure:"mode"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// JWTConfiguration JWT配置
type JWTConfiguration struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn int    `mapstructure:"expires_in"`
}

// StorageConfiguration 儲存配置
type StorageConfiguration struct {
	Type string          `mapstructure:"type"`
	S3   S3Configuration `mapstructure:"s3"`
}

// S3Configuration S3配置
type S3Configuration struct {
	Region    string `mapstructure:"region"`
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Endpoint  string `mapstructure:"endpoint"`
	CDNDomain string `mapstructure:"cdn_domain"`
}

// MediaConvertConfiguration MediaConvert配置
type MediaConvertConfiguration struct {
	Enabled      bool   `mapstructure:"enabled"`
	Region       string `mapstructure:"region"`
	Endpoint     string `mapstructure:"endpoint"`
	RoleArn      string `mapstructure:"role_arn"`
	OutputBucket string `mapstructure:"output_bucket"`
}

// TranscodeConfiguration 轉碼配置
type TranscodeConfiguration struct {
	Type   string              `mapstructure:"type"` // "ffmpeg" 或 "mediaconvert"
	FFmpeg FFmpegConfiguration `mapstructure:"ffmpeg"`
}

// FFmpegConfiguration FFmpeg 轉碼配置
type FFmpegConfiguration struct {
	Enabled       bool   `mapstructure:"enabled"`
	ContainerName string `mapstructure:"container_name"`
}

// VideoConfiguration 影片處理配置
type VideoConfiguration struct {
	MaxFileSize      int64                   `mapstructure:"max_file_size"`
	MinFileSize      int64                   `mapstructure:"min_file_size"` // 最小轉檔檔案大小
	AllowedFormats   []string                `mapstructure:"allowed_formats"`
	TranscodePresets []TranscodePresetConfig `mapstructure:"transcode_presets"`
}

// TranscodePresetConfig 轉碼預設配置
type TranscodePresetConfig struct {
	Name    string `mapstructure:"name"`
	Width   int    `mapstructure:"width"`
	Height  int    `mapstructure:"height"`
	Bitrate int    `mapstructure:"bitrate"`
}

// CacheConfiguration 緩存配置（支援PostgreSQL和Redis）
type CacheConfiguration struct {
	Type string `mapstructure:"type"`
	// PostgreSQL 緩存配置
	TableName       string `mapstructure:"table_name"`
	CleanupInterval int    `mapstructure:"cleanup_interval"`
	// Redis 緩存配置
	DB        int    `mapstructure:"db"`
	KeyPrefix string `mapstructure:"key_prefix"`
	// 通用配置
	DefaultExpiration int `mapstructure:"default_expiration"`
}

// MessagingConfiguration 訊息佇列配置（支援PostgreSQL和Redis）
type MessagingConfiguration struct {
	Type     string   `mapstructure:"type"`
	Channels []string `mapstructure:"channels"`
	// Redis 特定配置
	DB int `mapstructure:"db"`
}

// LiveConfiguration 直播配置
type LiveConfiguration struct {
	Enabled bool                    `mapstructure:"enabled"`
	Type    string                  `mapstructure:"type"` // "local", "cloud", "hybrid"
	Local   LocalLiveConfiguration  `mapstructure:"local"`
	Cloud   CloudLiveConfiguration  `mapstructure:"cloud"`
	Hybrid  HybridLiveConfiguration `mapstructure:"hybrid"`
}

// LocalLiveConfiguration 本地直播配置
type LocalLiveConfiguration struct {
	Enabled           bool   `mapstructure:"enabled"`
	RTMPServer        string `mapstructure:"rtmp_server"`
	RTMPServerPort    int    `mapstructure:"rtmp_server_port"`
	TranscoderEnabled bool   `mapstructure:"transcoder_enabled"`
	HLSOutputDir      string `mapstructure:"hls_output_dir"`
	HTTPPort          int    `mapstructure:"http_port"`
}

// CloudLiveConfiguration 雲端直播配置
type CloudLiveConfiguration struct {
	Provider         string `mapstructure:"provider"` // "aws", "aliyun", "tencent"
	RTMPIngestURL    string `mapstructure:"rtmp_ingest_url"`
	HLSPlaybackURL   string `mapstructure:"hls_playback_url"`
	APIKey           string `mapstructure:"api_key"`
	APISecret        string `mapstructure:"api_secret"`
	TranscodeEnabled bool   `mapstructure:"transcode_enabled"`
}

// HybridLiveConfiguration 混合直播配置
type HybridLiveConfiguration struct {
	LocalEnabled    bool   `mapstructure:"local_enabled"`
	CloudEnabled    bool   `mapstructure:"cloud_enabled"`
	FallbackToLocal bool   `mapstructure:"fallback_to_local"`
	CloudProvider   string `mapstructure:"cloud_provider"`
}

type Config struct {
	*Configurations
	DB              map[string]*gorm.DB
	DatabaseFactory *DatabaseFactory
	ActiveDatabase  string // 當前使用的資料庫類型
}

// NewConfig 創建系統配置（支援 MySQL 和 PostgreSQL）
func NewConfig(env string, dbType string) *Config {
	var config Configurations

	// 載入 .env 檔案
	if err := godotenv.Load(); err != nil {
		utils.LogInfo("未找到 .env 檔案，使用系統環境變數")
	}

	// 啟用環境變數支援
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 設定環境變數前綴
	viper.SetEnvPrefix("STREAM_DEMO")

	// 綁定環境變數到配置結構
	bindEnvironmentVariables()

	// 設定預設值
	setDefaultValues(&config)

	// 從環境變數載入配置
	utils.LogInfo("開始從環境變數載入配置...")
	err := viper.Unmarshal(&config)
	if err != nil {
		utils.LogFatal("Unable to decode into struct: %v", err)
	}
	utils.LogInfo("環境變數載入完成")

	// 處理環境變數覆蓋
	utils.LogInfo("開始處理環境變數覆蓋...")
	overrideWithEnvironmentVariables(&config)
	utils.LogInfo("環境變數覆蓋處理完成")

	var conf Config
	conf.Configurations = &config
	gin.SetMode(config.Gin.Mode)

	// 決定使用哪個資料庫
	selectedDB := dbType
	if selectedDB == "" {
		selectedDB = "postgresql" // 默認使用 PostgreSQL
	}
	conf.ActiveDatabase = selectedDB

	// 驗證選擇的資料庫配置是否存在
	dbConfig, exists := config.Databases[selectedDB]
	if !exists {
		utils.LogFatal("Database configuration not found for type: %s", selectedDB)
	}

	// 確保資料庫類型正確設定
	dbConfig.Type = selectedDB

	// 驗證資料庫類型
	if err := ValidateDatabaseType(dbConfig.Type); err != nil {
		utils.LogFatal("Database configuration error: %v", err)
	}

	// 創建資料庫工廠
	conf.DatabaseFactory = NewDatabaseFactory(dbConfig)

	// 初始化資料庫連接
	utils.LogInfo("初始化資料庫連接，使用配置: %s (類型: %s)", selectedDB, dbConfig.Type)
	conf.DB = make(map[string]*gorm.DB)

	// 主資料庫
	masterDB, err := conf.DatabaseFactory.CreateDatabase(false)
	if err != nil {
		utils.LogFatal("Failed to create master database connection: %v", err)
	}
	conf.DB["master"] = masterDB

	// 從資料庫（如果配置不同的話）
	if dbConfig.Slave.Host != dbConfig.Master.Host ||
		dbConfig.Slave.DBName != dbConfig.Master.DBName {
		slaveDB, err := conf.DatabaseFactory.CreateDatabase(true)
		if err != nil {
			utils.LogWarn("Failed to create slave database connection, using master: %v", err)
			conf.DB["slave"] = masterDB
		} else {
			conf.DB["slave"] = slaveDB
		}
	} else {
		// 如果配置相同，使用同一個連接
		conf.DB["slave"] = masterDB
	}

	// 如果緩存類型是PostgreSQL，創建緩存表（僅對PostgreSQL）
	if config.Cache.Type == "postgresql" && dbConfig.Type == "postgresql" {
		InitCacheTable(masterDB, config.Cache.TableName)
	}

	utils.LogInfo("資料庫配置完成，使用: %s (類型: %s)", selectedDB, dbConfig.Type)

	// 初始化Redis連接
	if config.Cache.Type == "redis" || config.Messaging.Type == "redis" {
		InitRedis(config.Redis)
		utils.LogInfo("Redis configuration completed successfully")
	}

	return &conf
}

// bindEnvironmentVariables 綁定環境變數到配置
func bindEnvironmentVariables() {
	// 伺服器配置
	viper.BindEnv("gin.host", "STREAM_DEMO_HOST")
	viper.BindEnv("gin.port", "STREAM_DEMO_PORT")
	viper.BindEnv("gin.mode", "STREAM_DEMO_MODE")

	// 資料庫配置
	viper.BindEnv("databases.postgresql.master.host", "STREAM_DEMO_DB_HOST")
	viper.BindEnv("databases.postgresql.master.port", "STREAM_DEMO_DB_PORT")
	viper.BindEnv("databases.postgresql.master.username", "STREAM_DEMO_DB_USER")
	viper.BindEnv("databases.postgresql.master.password", "STREAM_DEMO_DB_PASSWORD")
	viper.BindEnv("databases.postgresql.master.dbname", "STREAM_DEMO_DB_NAME")
	viper.BindEnv("databases.postgresql.master.sslmode", "STREAM_DEMO_DB_SSL_MODE")

	// 資料庫連接池配置
	viper.BindEnv("databases.postgresql.pool.max_open_conns", "STREAM_DEMO_DB_MAX_OPEN_CONNS")
	viper.BindEnv("databases.postgresql.pool.max_idle_conns", "STREAM_DEMO_DB_MAX_IDLE_CONNS")
	viper.BindEnv("databases.postgresql.pool.conn_max_lifetime", "STREAM_DEMO_DB_CONN_MAX_LIFETIME")
	viper.BindEnv("databases.postgresql.pool.conn_max_idle_time", "STREAM_DEMO_DB_CONN_MAX_IDLE_TIME")

	// Redis 配置
	viper.BindEnv("redis.master.host", "STREAM_DEMO_REDIS_HOST")
	viper.BindEnv("redis.master.port", "STREAM_DEMO_REDIS_PORT")
	viper.BindEnv("redis.master.password", "STREAM_DEMO_REDIS_PASSWORD")
	viper.BindEnv("redis.master.db", "STREAM_DEMO_REDIS_DB")

	// Redis 連接池配置
	viper.BindEnv("redis.pool.max_active", "STREAM_DEMO_REDIS_MAX_ACTIVE")
	viper.BindEnv("redis.pool.max_idle", "STREAM_DEMO_REDIS_MAX_IDLE")
	viper.BindEnv("redis.pool.idle_timeout", "STREAM_DEMO_REDIS_IDLE_TIMEOUT")
	viper.BindEnv("redis.pool.connect_timeout", "STREAM_DEMO_REDIS_CONNECT_TIMEOUT")
	viper.BindEnv("redis.pool.read_timeout", "STREAM_DEMO_REDIS_READ_TIMEOUT")
	viper.BindEnv("redis.pool.write_timeout", "STREAM_DEMO_REDIS_WRITE_TIMEOUT")

	// 緩存配置
	viper.BindEnv("cache.type", "STREAM_DEMO_CACHE_TYPE")
	viper.BindEnv("cache.db", "STREAM_DEMO_CACHE_DB")
	viper.BindEnv("cache.default_expiration", "STREAM_DEMO_CACHE_DEFAULT_EXPIRATION")
	viper.BindEnv("cache.key_prefix", "STREAM_DEMO_CACHE_KEY_PREFIX")

	// 訊息佇列配置
	viper.BindEnv("messaging.type", "STREAM_DEMO_MESSAGING_TYPE")
	viper.BindEnv("messaging.db", "STREAM_DEMO_MESSAGING_DB")

	// JWT 配置
	viper.BindEnv("jwt.secret", "STREAM_DEMO_JWT_SECRET")
	viper.BindEnv("jwt.expires_in", "STREAM_DEMO_JWT_EXPIRES_IN")

	// S3 配置
	viper.BindEnv("storage.s3.region", "STREAM_DEMO_S3_REGION")
	viper.BindEnv("storage.s3.bucket", "STREAM_DEMO_S3_BUCKET")
	viper.BindEnv("storage.s3.access_key", "STREAM_DEMO_S3_ACCESS_KEY")
	viper.BindEnv("storage.s3.secret_key", "STREAM_DEMO_S3_SECRET_KEY")
	viper.BindEnv("storage.s3.endpoint", "STREAM_DEMO_S3_ENDPOINT")
	viper.BindEnv("storage.s3.cdn_domain", "STREAM_DEMO_S3_CDN_DOMAIN")

	// 轉碼配置
	viper.BindEnv("transcode.type", "STREAM_DEMO_TRANSCODE_TYPE")
	viper.BindEnv("transcode.ffmpeg.enabled", "STREAM_DEMO_FFMPEG_ENABLED")
	viper.BindEnv("transcode.ffmpeg.container_name", "STREAM_DEMO_FFMPEG_CONTAINER_NAME")

	// AWS MediaConvert 配置
	viper.BindEnv("media_convert.enabled", "STREAM_DEMO_MEDIACONVERT_ENABLED")
	viper.BindEnv("media_convert.region", "STREAM_DEMO_MEDIACONVERT_REGION")
	viper.BindEnv("media_convert.endpoint", "STREAM_DEMO_MEDIACONVERT_ENDPOINT")
	viper.BindEnv("media_convert.role_arn", "STREAM_DEMO_MEDIACONVERT_ROLE_ARN")
	viper.BindEnv("media_convert.output_bucket", "STREAM_DEMO_MEDIACONVERT_OUTPUT_BUCKET")

	// 影片配置
	viper.BindEnv("video.max_file_size", "STREAM_DEMO_VIDEO_MAX_FILE_SIZE")
	viper.BindEnv("video.min_file_size", "STREAM_DEMO_VIDEO_MIN_FILE_SIZE")
	viper.BindEnv("video.allowed_formats", "STREAM_DEMO_VIDEO_ALLOWED_FORMATS")

	// 直播配置
	viper.BindEnv("live.enabled", "STREAM_DEMO_LIVE_ENABLED")
	viper.BindEnv("live.type", "STREAM_DEMO_LIVE_TYPE")

	// 本地直播配置
	viper.BindEnv("live.local.enabled", "STREAM_DEMO_LIVE_LOCAL_ENABLED")
	viper.BindEnv("live.local.rtmp_server", "STREAM_DEMO_LIVE_LOCAL_RTMP_SERVER")
	viper.BindEnv("live.local.rtmp_server_port", "STREAM_DEMO_LIVE_LOCAL_RTMP_SERVER_PORT")
	viper.BindEnv("live.local.transcoder_enabled", "STREAM_DEMO_LIVE_LOCAL_TRANSCODER_ENABLED")
	viper.BindEnv("live.local.hls_output_dir", "STREAM_DEMO_LIVE_LOCAL_HLS_OUTPUT_DIR")
	viper.BindEnv("live.local.http_port", "STREAM_DEMO_LIVE_LOCAL_HTTP_PORT")

	// 雲端直播配置
	viper.BindEnv("live.cloud.provider", "STREAM_DEMO_LIVE_CLOUD_PROVIDER")
	viper.BindEnv("live.cloud.rtmp_ingest_url", "STREAM_DEMO_LIVE_CLOUD_RTMP_INGEST_URL")
	viper.BindEnv("live.cloud.hls_playback_url", "STREAM_DEMO_LIVE_CLOUD_HLS_PLAYBACK_URL")
	viper.BindEnv("live.cloud.api_key", "STREAM_DEMO_LIVE_API_KEY")
	viper.BindEnv("live.cloud.api_secret", "STREAM_DEMO_LIVE_API_SECRET")
	viper.BindEnv("live.cloud.transcode_enabled", "STREAM_DEMO_LIVE_CLOUD_TRANSCODE_ENABLED")

	// 混合直播配置
	viper.BindEnv("live.hybrid.local_enabled", "STREAM_DEMO_LIVE_HYBRID_LOCAL_ENABLED")
	viper.BindEnv("live.hybrid.cloud_enabled", "STREAM_DEMO_LIVE_HYBRID_CLOUD_ENABLED")
	viper.BindEnv("live.hybrid.fallback_to_local", "STREAM_DEMO_LIVE_HYBRID_FALLBACK_TO_LOCAL")
	viper.BindEnv("live.hybrid.cloud_provider", "STREAM_DEMO_LIVE_HYBRID_CLOUD_PROVIDER")
}

// setDefaultValues 設定預設配置值
func setDefaultValues(config *Configurations) {
	// 伺服器預設值
	if config.Gin.Host == "" {
		config.Gin.Host = "127.0.0.1"
	}
	if config.Gin.Port == 0 {
		config.Gin.Port = 8080
	}
	if config.Gin.Mode == "" {
		config.Gin.Mode = "debug"
	}

	// 資料庫預設值
	if config.Databases == nil {
		config.Databases = make(map[string]DatabaseConfiguration)
	}

	// PostgreSQL 預設值
	if _, exists := config.Databases["postgresql"]; !exists {
		config.Databases["postgresql"] = DatabaseConfiguration{
			Type: "postgresql",
			Master: DatabaseConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "stream_user",
				Password: "stream_password",
				DBName:   "stream_demo",
				SSLMode:  "disable",
			},
			Slave: DatabaseConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "stream_user",
				Password: "stream_password",
				DBName:   "stream_demo",
				SSLMode:  "disable",
			},
			Pool: DatabasePoolConfiguration{
				MaxOpenConns:    25,
				MaxIdleConns:    10,
				ConnMaxLifetime: 3600,
				ConnMaxIdleTime: 900,
			},
		}
	}

	// MySQL 預設值
	if _, exists := config.Databases["mysql"]; !exists {
		config.Databases["mysql"] = DatabaseConfiguration{
			Type: "mysql",
			Master: DatabaseConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "stream_user",
				Password: "stream_password",
				DBName:   "stream_demo",
				SSLMode:  "false",
			},
			Slave: DatabaseConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "stream_user",
				Password: "stream_password",
				DBName:   "stream_demo",
				SSLMode:  "false",
			},
			Pool: DatabasePoolConfiguration{
				MaxOpenConns:    25,
				MaxIdleConns:    10,
				ConnMaxLifetime: 3600,
				ConnMaxIdleTime: 900,
			},
		}
	}

	// Redis 預設值
	if config.Redis.Master.Host == "" {
		config.Redis.Master.Host = "localhost"
	}
	if config.Redis.Master.Port == 0 {
		config.Redis.Master.Port = 6379
	}
	if config.Redis.Pool.MaxActive == 0 {
		config.Redis.Pool.MaxActive = 100
	}
	if config.Redis.Pool.MaxIdle == 0 {
		config.Redis.Pool.MaxIdle = 20
	}

	// 緩存預設值
	if config.Cache.Type == "" {
		config.Cache.Type = "redis"
	}
	if config.Cache.DB == 0 {
		config.Cache.DB = 1
	}
	if config.Cache.DefaultExpiration == 0 {
		config.Cache.DefaultExpiration = 3600
	}
	if config.Cache.KeyPrefix == "" {
		config.Cache.KeyPrefix = "cache:"
	}

	// 訊息佇列預設值
	if config.Messaging.Type == "" {
		config.Messaging.Type = "redis"
	}
	if config.Messaging.DB == 0 {
		config.Messaging.DB = 2
	}

	// JWT 預設值
	if config.JWT.Secret == "" {
		config.JWT.Secret = "local_secret"
	}
	if config.JWT.ExpiresIn == 0 {
		config.JWT.ExpiresIn = 86400
	}

	// S3 預設值
	if config.Storage.S3.Region == "" {
		config.Storage.S3.Region = "us-east-1"
	}
	if config.Storage.S3.Bucket == "" {
		config.Storage.S3.Bucket = "stream-demo-videos"
	}
	if config.Storage.S3.AccessKey == "" {
		config.Storage.S3.AccessKey = "minioadmin"
	}
	if config.Storage.S3.SecretKey == "" {
		config.Storage.S3.SecretKey = "minioadmin"
	}
	if config.Storage.S3.Endpoint == "" {
		config.Storage.S3.Endpoint = "http://localhost:9000"
	}

	// 轉碼預設值
	if config.Transcode.Type == "" {
		config.Transcode.Type = "ffmpeg"
	}
	if !config.Transcode.FFmpeg.Enabled {
		config.Transcode.FFmpeg.Enabled = true
	}
	if config.Transcode.FFmpeg.ContainerName == "" {
		config.Transcode.FFmpeg.ContainerName = "stream-demo-converter"
	}

	// 影片預設值
	if config.Video.MaxFileSize == 0 {
		config.Video.MaxFileSize = 1073741824 // 1GB
	}
	if config.Video.MinFileSize == 0 {
		config.Video.MinFileSize = 1048576 // 1MB
	}
	if len(config.Video.AllowedFormats) == 0 {
		config.Video.AllowedFormats = []string{"mp4", "avi", "mov", "mkv", "webm"}
	}

	// 直播預設值
	if !config.Live.Enabled {
		config.Live.Enabled = true
	}
	if config.Live.Type == "" {
		config.Live.Type = "local"
	}
	if !config.Live.Local.Enabled {
		config.Live.Local.Enabled = true
	}
	if config.Live.Local.RTMPServer == "" {
		config.Live.Local.RTMPServer = "localhost"
	}
	if config.Live.Local.RTMPServerPort == 0 {
		config.Live.Local.RTMPServerPort = 1935
	}
	if !config.Live.Local.TranscoderEnabled {
		config.Live.Local.TranscoderEnabled = true
	}
	if config.Live.Local.HLSOutputDir == "" {
		config.Live.Local.HLSOutputDir = "/tmp/live"
	}
	if config.Live.Local.HTTPPort == 0 {
		config.Live.Local.HTTPPort = 8081
	}
}

// overrideWithEnvironmentVariables 用環境變數覆蓋配置
func overrideWithEnvironmentVariables(config *Configurations) {
	utils.LogInfo("進入環境變數覆蓋函數")
	
	// 列出所有 STREAM_DEMO_ 開頭的環境變數
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "STREAM_DEMO_") {
			utils.LogInfo("發現環境變數: %s", env)
		}
	}
	
	// 資料庫配置覆蓋
	if postgresqlConfig, exists := config.Databases["postgresql"]; exists {
		utils.LogInfo("找到 PostgreSQL 配置，當前主機: %s", postgresqlConfig.Master.Host)
		
		// 直接從環境變數讀取，不使用 viper 前綴
		if host := os.Getenv("STREAM_DEMO_DB_HOST"); host != "" {
			utils.LogInfo("覆蓋資料庫主機: %s", host)
			postgresqlConfig.Master.Host = host
			postgresqlConfig.Slave.Host = host // 同時更新 slave 配置
		} else {
			utils.LogInfo("未找到 STREAM_DEMO_DB_HOST 環境變數")
		}
		if portStr := os.Getenv("STREAM_DEMO_DB_PORT"); portStr != "" {
			if port := viper.GetInt("STREAM_DEMO_DB_PORT"); port != 0 {
				postgresqlConfig.Master.Port = port
				postgresqlConfig.Slave.Port = port
			}
		}
		if username := os.Getenv("STREAM_DEMO_DB_USER"); username != "" {
			utils.LogInfo("覆蓋資料庫用戶: %s", username)
			postgresqlConfig.Master.Username = username
			postgresqlConfig.Slave.Username = username
		}
		if password := os.Getenv("STREAM_DEMO_DB_PASSWORD"); password != "" {
			utils.LogInfo("覆蓋資料庫密碼: [已設置]")
			postgresqlConfig.Master.Password = password
			postgresqlConfig.Slave.Password = password
		}
		if dbname := os.Getenv("STREAM_DEMO_DB_NAME"); dbname != "" {
			utils.LogInfo("覆蓋資料庫名稱: %s", dbname)
			postgresqlConfig.Master.DBName = dbname
			postgresqlConfig.Slave.DBName = dbname
		}
		if sslmode := os.Getenv("STREAM_DEMO_DB_SSL_MODE"); sslmode != "" {
			utils.LogInfo("覆蓋 SSL 模式: %s", sslmode)
			postgresqlConfig.Master.SSLMode = sslmode
			postgresqlConfig.Slave.SSLMode = sslmode
		}
		config.Databases["postgresql"] = postgresqlConfig
	}

	// Redis 配置覆蓋
	if host := viper.GetString("STREAM_DEMO_REDIS_HOST"); host != "" {
		config.Redis.Master.Host = host
		config.Redis.Slave.Host = host
	}
	if port := viper.GetInt("STREAM_DEMO_REDIS_PORT"); port != 0 {
		config.Redis.Master.Port = port
		config.Redis.Slave.Port = port
	}
	if password := viper.GetString("STREAM_DEMO_REDIS_PASSWORD"); password != "" {
		config.Redis.Master.Password = password
	}

	// JWT 配置覆蓋
	if secret := viper.GetString("STREAM_DEMO_JWT_SECRET"); secret != "" {
		config.JWT.Secret = secret
	}

	// S3 配置覆蓋
	if region := viper.GetString("STREAM_DEMO_S3_REGION"); region != "" {
		if config.Storage.S3.Region == "" {
			config.Storage.S3.Region = region
		}
	}
	if bucket := viper.GetString("STREAM_DEMO_S3_BUCKET"); bucket != "" {
		if config.Storage.S3.Bucket == "" {
			config.Storage.S3.Bucket = bucket
		}
	}
	if accessKey := viper.GetString("STREAM_DEMO_S3_ACCESS_KEY"); accessKey != "" {
		config.Storage.S3.AccessKey = accessKey
	}
	if secretKey := viper.GetString("STREAM_DEMO_S3_SECRET_KEY"); secretKey != "" {
		config.Storage.S3.SecretKey = secretKey
	}

	// 直播配置覆蓋
	if apiKey := viper.GetString("STREAM_DEMO_LIVE_API_KEY"); apiKey != "" {
		config.Live.Cloud.APIKey = apiKey
	}
	if apiSecret := viper.GetString("STREAM_DEMO_LIVE_API_SECRET"); apiSecret != "" {
		config.Live.Cloud.APISecret = apiSecret
	}

	// 伺服器配置覆蓋
	if host := viper.GetString("STREAM_DEMO_HOST"); host != "" {
		if config.Gin.Host == "" {
			config.Gin.Host = host
		}
	}
	if port := viper.GetInt("STREAM_DEMO_PORT"); port != 0 {
		if config.Gin.Port == 0 {
			config.Gin.Port = port
		}
	}
	if mode := viper.GetString("STREAM_DEMO_MODE"); mode != "" {
		if config.Gin.Mode == "" {
			config.Gin.Mode = mode
		}
	}
}

// determineDatabase 決定使用哪個資料庫
func determineDatabase(dbType string, databases map[string]DatabaseConfiguration) string {
	// 1. 優先使用命令行參數
	if dbType != "" {
		utils.LogInfo("使用命令行指定的資料庫類型: %s", dbType)
		return dbType
	}

	// 2. 檢查環境變數
	if envDB := viper.GetString("DATABASE_TYPE"); envDB != "" {
		utils.LogInfo("使用環境變數指定的資料庫類型: %s", envDB)
		return envDB
	}

	// 3. 使用默認順序：postgresql > mysql
	if _, exists := databases["postgresql"]; exists {
		utils.LogInfo("使用默認資料庫類型: postgresql")
		return "postgresql"
	}
	if _, exists := databases["mysql"]; exists {
		utils.LogInfo("使用默認資料庫類型: mysql")
		return "mysql"
	}

	utils.LogFatal("No valid database configuration found")
	return ""
}

// InitRedis 初始化Redis連接
func InitRedis(config RedisConfiguration) error {
	masterConfig := utils.RedisConfig{
		Host:           config.Master.Host,
		Port:           config.Master.Port,
		Password:       config.Master.Password,
		DB:             config.Master.DB,
		MaxActive:      config.Pool.MaxActive,
		MaxIdle:        config.Pool.MaxIdle,
		IdleTimeout:    config.Pool.IdleTimeout,
		ConnectTimeout: config.Pool.ConnectTimeout,
		ReadTimeout:    config.Pool.ReadTimeout,
		WriteTimeout:   config.Pool.WriteTimeout,
	}

	slaveConfig := utils.RedisConfig{
		Host:           config.Slave.Host,
		Port:           config.Slave.Port,
		Password:       config.Slave.Password,
		DB:             config.Slave.DB,
		MaxActive:      config.Pool.MaxActive,
		MaxIdle:        config.Pool.MaxIdle,
		IdleTimeout:    config.Pool.IdleTimeout,
		ConnectTimeout: config.Pool.ConnectTimeout,
		ReadTimeout:    config.Pool.ReadTimeout,
		WriteTimeout:   config.Pool.WriteTimeout,
	}

	return utils.InitRedisClient(masterConfig, slaveConfig)
}

// ReconnectDatabases 重新連接資料庫
func (c *Config) ReconnectDatabases() {
	for name, db := range c.DB {
		sqlDB, err := db.DB()
		if err != nil {
			utils.LogError("Failed to get underlying sql.DB for %s: %v", name, err)
			continue
		}

		if err := sqlDB.Ping(); err != nil {
			utils.LogInfo("Reconnecting database: %s", name)

			isSlave := name == "slave"
			newDB, err := c.DatabaseFactory.CreateDatabase(isSlave)
			if err != nil {
				utils.LogError("Failed to reconnect database %s: %v", name, err)
			} else {
				c.DB[name] = newDB
				utils.LogInfo("Database reconnected: %s", name)
			}
		}
	}
}

// CheckDatabaseConnections 檢查資料庫連接狀態
func (c *Config) CheckDatabaseConnections() bool {
	allConnected := true

	for name, db := range c.DB {
		sqlDB, err := db.DB()
		if err != nil {
			utils.LogError("Failed to get underlying sql.DB for %s: %v", name, err)
			allConnected = false
			continue
		}

		if err := sqlDB.Ping(); err != nil {
			utils.LogError("Database connection failed for %s: %v", name, err)
			allConnected = false
		} else {
			utils.LogInfo("Database connection OK for %s", name)
		}
	}

	return allConnected
}

// GetAvailableDatabases 獲取可用的資料庫配置
func (c *Config) GetAvailableDatabases() []string {
	var available []string
	for name := range c.Databases {
		available = append(available, name)
	}
	return available
}

// SwitchDatabase 動態切換資料庫（運行時切換）
func (c *Config) SwitchDatabase(dbType string) error {
	dbConfig, exists := c.Databases[dbType]
	if !exists {
		return fmt.Errorf("database configuration not found for type: %s", dbType)
	}

	// 驗證資料庫類型
	if err := ValidateDatabaseType(dbConfig.Type); err != nil {
		return fmt.Errorf("database configuration error: %w", err)
	}

	// 關閉現有連接
	for name, db := range c.DB {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		utils.LogInfo("Closed database connection: %s", name)
	}

	// 創建新的資料庫工廠
	c.DatabaseFactory = NewDatabaseFactory(dbConfig)
	c.ActiveDatabase = dbType

	// 重新初始化連接
	utils.LogInfo("切換到資料庫配置: %s (類型: %s)", dbType, dbConfig.Type)

	// 主資料庫
	masterDB, err := c.DatabaseFactory.CreateDatabase(false)
	if err != nil {
		return fmt.Errorf("failed to create master database connection: %w", err)
	}
	c.DB["master"] = masterDB

	// 從資料庫
	if dbConfig.Slave.Host != dbConfig.Master.Host ||
		dbConfig.Slave.DBName != dbConfig.Master.DBName {
		slaveDB, err := c.DatabaseFactory.CreateDatabase(true)
		if err != nil {
			utils.LogWarn("Failed to create slave database connection, using master: %v", err)
			c.DB["slave"] = masterDB
		} else {
			c.DB["slave"] = slaveDB
		}
	} else {
		c.DB["slave"] = masterDB
	}

	utils.LogInfo("資料庫切換完成: %s", dbType)
	return nil
}

// GetDatabaseInfo 獲取資料庫信息
func (c *Config) GetDatabaseInfo() map[string]interface{} {
	info := make(map[string]interface{})

	info["active"] = c.ActiveDatabase
	info["available"] = c.GetAvailableDatabases()

	if c.DatabaseFactory != nil {
		info["master"] = c.DatabaseFactory.GetDatabaseInfo(false)
		info["slave"] = c.DatabaseFactory.GetDatabaseInfo(true)
	}

	return info
}
