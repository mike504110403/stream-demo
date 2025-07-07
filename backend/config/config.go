package config

import (
	"strings"

	log "stream-demo/backend/pkg/logging"

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
	// PostgreSQL配置
	Database  DatabaseConfiguration  `mapstructure:"database"`
	Cache     CacheConfiguration     `mapstructure:"cache"`
	Messaging MessagingConfiguration `mapstructure:"messaging"`
	JWT       JWTConfiguration       `mapstructure:"jwt"`
	// 新增的配置字段
	Storage      StorageConfiguration      `mapstructure:"storage"`
	MediaConvert MediaConvertConfiguration `mapstructure:"media_convert"`
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

// DatabaseConfiguration 資料庫配置
type DatabaseConfiguration struct {
	Type   string                    `mapstructure:"type"`
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
	SSLMode  string `mapstructure:"sslmode"`
}

// DatabasePoolConfiguration 資料庫連接池配置
type DatabasePoolConfiguration struct {
	MaxOpenConns    int `mapstructure:"max_open_conns"`
	MaxIdleConns    int `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"`
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
	Region       string `mapstructure:"region"`
	Endpoint     string `mapstructure:"endpoint"`
	RoleArn      string `mapstructure:"role_arn"`
	OutputBucket string `mapstructure:"output_bucket"`
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

// CacheConfiguration PostgreSQL緩存配置
type CacheConfiguration struct {
	Type              string `mapstructure:"type"`
	TableName         string `mapstructure:"table_name"`
	DefaultExpiration int    `mapstructure:"default_expiration"`
	CleanupInterval   int    `mapstructure:"cleanup_interval"`
}

// MessagingConfiguration PostgreSQL訊息佇列配置
type MessagingConfiguration struct {
	Type     string   `mapstructure:"type"`
	Channels []string `mapstructure:"channels"`
}

type Config struct {
	*Configurations
	DB map[string]*gorm.DB
}

// NewPostgreSQLConfig 創建PostgreSQL配置
func NewPostgreSQLConfig(configPath string, env string) *Config {
	var config Configurations
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading config file, ", err)
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Unable to decode into struct, ", err)
	}

	var conf Config
	conf.Configurations = &config
	gin.SetMode(config.Gin.Mode)

	// 初始化PostgreSQL連接
	if config.Database.Type == "postgresql" {
		conf.DB = make(map[string]*gorm.DB)

		// 主資料庫
		masterDB := InitPostgreSQL(config.Database.Master, false)
		conf.DB["master"] = masterDB

		// 從資料庫（如果配置不同的話）
		if config.Database.Slave.Host != config.Database.Master.Host ||
			config.Database.Slave.DBName != config.Database.Master.DBName {
			slaveDB := InitPostgreSQL(config.Database.Slave, true)
			conf.DB["slave"] = slaveDB
		} else {
			// 如果配置相同，使用同一個連接
			conf.DB["slave"] = masterDB
		}

		// 創建緩存表
		InitCacheTable(masterDB, config.Cache.TableName)

		log.Info("PostgreSQL configuration completed successfully")
	} else {
		log.Fatal("Database type must be 'postgresql'")
	}

	return &conf
}

// (c *Config) ReconnectPostgreSQLDatabases 重新連接PostgreSQL資料庫
func (c *Config) ReconnectPostgreSQLDatabases() {
	if c.Database.Type == "postgresql" {
		for name, db := range c.DB {
			sqlDB, err := db.DB()
			if err != nil {
				log.Error("Failed to get underlying sql.DB for", name, ":", err)
				continue
			}

			if err := sqlDB.Ping(); err != nil {
				log.Info("Reconnecting PostgreSQL database:", name)
				if name == "master" {
					c.DB[name] = InitPostgreSQL(c.Database.Master, false)
				} else {
					c.DB[name] = InitPostgreSQL(c.Database.Slave, true)
				}
			}
		}
	}
}

// CheckPostgreSQLConnections 檢查PostgreSQL連接狀態
func (c *Config) CheckPostgreSQLConnections() bool {
	allConnected := true

	for name, db := range c.DB {
		sqlDB, err := db.DB()
		if err != nil {
			log.Error("Failed to get underlying sql.DB for", name, ":", err)
			allConnected = false
			continue
		}

		if err := sqlDB.Ping(); err != nil {
			log.Error("PostgreSQL connection failed for", name, ":", err)
			allConnected = false
		} else {
			log.Info("PostgreSQL connection OK for", name)
		}
	}

	return allConnected
}
