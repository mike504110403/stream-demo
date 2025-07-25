package config

import (
	"fmt"
	"strings"

	"stream-demo/backend/utils"

	"github.com/gin-gonic/gin"
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

type Config struct {
	*Configurations
	DB              map[string]*gorm.DB
	DatabaseFactory *DatabaseFactory
	ActiveDatabase  string // 當前使用的資料庫類型
}

// NewConfig 創建系統配置（支援 MySQL 和 PostgreSQL）
func NewConfig(configPath string, env string, dbType string) *Config {
	var config Configurations
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))

	if err := viper.ReadInConfig(); err != nil {
		utils.LogFatal("Error reading config file, ", err)
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		utils.LogFatal("Unable to decode into struct, ", err)
	}

	var conf Config
	conf.Configurations = &config
	gin.SetMode(config.Gin.Mode)

	// 決定使用哪個資料庫（優先級：參數 > 環境變數 > 默認）
	selectedDB := determineDatabase(dbType, config.Databases)
	conf.ActiveDatabase = selectedDB

	// 驗證選擇的資料庫配置是否存在
	dbConfig, exists := config.Databases[selectedDB]
	if !exists {
		utils.LogFatal("Database configuration not found for type: ", selectedDB)
	}

	// 驗證資料庫類型
	if err := ValidateDatabaseType(dbConfig.Type); err != nil {
		utils.LogFatal("Database configuration error: ", err)
	}

	// 創建資料庫工廠
	conf.DatabaseFactory = NewDatabaseFactory(dbConfig)

	// 初始化資料庫連接
	utils.LogInfo("初始化資料庫連接，使用配置: %s (類型: %s)", selectedDB, dbConfig.Type)
	conf.DB = make(map[string]*gorm.DB)

	// 主資料庫
	masterDB, err := conf.DatabaseFactory.CreateDatabase(false)
	if err != nil {
		utils.LogFatal("Failed to create master database connection: ", err)
	}
	conf.DB["master"] = masterDB

	// 從資料庫（如果配置不同的話）
	if dbConfig.Slave.Host != dbConfig.Master.Host ||
		dbConfig.Slave.DBName != dbConfig.Master.DBName {
		slaveDB, err := conf.DatabaseFactory.CreateDatabase(true)
		if err != nil {
			utils.LogWarn("Failed to create slave database connection, using master: ", err)
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
			utils.LogError("Failed to get underlying sql.DB for", name, ":", err)
			continue
		}

		if err := sqlDB.Ping(); err != nil {
			utils.LogInfo("Reconnecting database:", name)

			isSlave := name == "slave"
			newDB, err := c.DatabaseFactory.CreateDatabase(isSlave)
			if err != nil {
				utils.LogError("Failed to reconnect database", name, ":", err)
			} else {
				c.DB[name] = newDB
				utils.LogInfo("Database reconnected:", name)
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
			utils.LogError("Failed to get underlying sql.DB for", name, ":", err)
			allConnected = false
			continue
		}

		if err := sqlDB.Ping(); err != nil {
			utils.LogError("Database connection failed for", name, ":", err)
			allConnected = false
		} else {
			utils.LogInfo("Database connection OK for", name)
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
			utils.LogWarn("Failed to create slave database connection, using master: ", err)
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
