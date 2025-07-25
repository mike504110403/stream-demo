package config

import (
	"fmt"
	"time"

	"stream-demo/backend/utils"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DatabaseType 資料庫類型
type DatabaseType string

const (
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
)

// DatabaseFactory 資料庫工廠
type DatabaseFactory struct {
	config DatabaseConfiguration
}

// NewDatabaseFactory 創建資料庫工廠
func NewDatabaseFactory(config DatabaseConfiguration) *DatabaseFactory {
	return &DatabaseFactory{config: config}
}

// CreateDatabase 根據配置創建資料庫連接
func (f *DatabaseFactory) CreateDatabase(isSlave bool) (*gorm.DB, error) {
	var connectionConfig DatabaseConnectionConfig
	if isSlave {
		connectionConfig = f.config.Slave
		utils.LogInfo("初始化從資料庫連接...")
	} else {
		connectionConfig = f.config.Master
		utils.LogInfo("初始化主資料庫連接...")
	}

	switch DatabaseType(f.config.Type) {
	case DatabaseTypeMySQL:
		return f.createMySQLConnection(connectionConfig, isSlave)
	case DatabaseTypePostgreSQL:
		return f.createPostgreSQLConnection(connectionConfig, isSlave)
	default:
		return nil, fmt.Errorf("不支援的資料庫類型: %s", f.config.Type)
	}
}

// createMySQLConnection 創建 MySQL 連接
func (f *DatabaseFactory) createMySQLConnection(config DatabaseConnectionConfig, isSlave bool) (*gorm.DB, error) {
	utils.LogInfo("=== MySQL Connection Info ===")
	utils.LogInfo(fmt.Sprintf("Host: %s:%d", config.Host, config.Port))
	utils.LogInfo(fmt.Sprintf("Database: %s", config.DBName))
	utils.LogInfo(fmt.Sprintf("User: %s", config.Username))

	// 構建 MySQL DSN
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	// 如果指定了 SSL 模式
	if config.SSLMode != "" {
		if config.SSLMode == "disable" {
			dsn += "&tls=false"
		} else {
			dsn += "&tls=" + config.SSLMode
		}
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: f.createLogger(),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: false,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("MySQL 連接失敗: %w", err)
	}

	// 配置連接池
	if err := f.configureConnectionPool(db); err != nil {
		return nil, err
	}

	utils.LogInfo("MySQL 連接成功")
	if isSlave {
		utils.LogInfo("MySQL 從資料庫已連接")
	} else {
		utils.LogInfo("MySQL 主資料庫已連接")
	}

	return db, nil
}

// createPostgreSQLConnection 創建 PostgreSQL 連接
func (f *DatabaseFactory) createPostgreSQLConnection(config DatabaseConnectionConfig, isSlave bool) (*gorm.DB, error) {
	utils.LogInfo("=== PostgreSQL Connection Info ===")
	utils.LogInfo(fmt.Sprintf("Host: %s:%d", config.Host, config.Port))
	utils.LogInfo(fmt.Sprintf("Database: %s", config.DBName))
	utils.LogInfo(fmt.Sprintf("User: %s", config.Username))

	// 構建 PostgreSQL DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Taipei",
		config.Host,
		config.Username,
		config.Password,
		config.DBName,
		config.Port,
		config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: f.createLogger(),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: false,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("PostgreSQL 連接失敗: %w", err)
	}

	// 配置連接池
	if err := f.configureConnectionPool(db); err != nil {
		return nil, err
	}

	utils.LogInfo("PostgreSQL 連接成功")
	if isSlave {
		utils.LogInfo("PostgreSQL 從資料庫已連接")
	} else {
		utils.LogInfo("PostgreSQL 主資料庫已連接")
	}

	return db, nil
}

// createLogger 創建 GORM 日誌記錄器
func (f *DatabaseFactory) createLogger() logger.Interface {
	return logger.New(
		&utils.Writer{},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

// configureConnectionPool 配置連接池
func (f *DatabaseFactory) configureConnectionPool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("獲取底層 sql.DB 失敗: %w", err)
	}

	// 設置連接池參數
	if f.config.Pool.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(f.config.Pool.MaxOpenConns)
	}
	if f.config.Pool.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(f.config.Pool.MaxIdleConns)
	}
	if f.config.Pool.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(f.config.Pool.ConnMaxLifetime) * time.Second)
	}
	if f.config.Pool.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(f.config.Pool.ConnMaxIdleTime) * time.Second)
	}

	// 測試連接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("資料庫連接測試失敗: %w", err)
	}

	utils.LogInfo("資料庫連接池配置完成")
	return nil
}

// GetSupportedDatabases 獲取支援的資料庫類型
func GetSupportedDatabases() []DatabaseType {
	return []DatabaseType{
		DatabaseTypeMySQL,
		DatabaseTypePostgreSQL,
	}
}

// ValidateDatabaseType 驗證資料庫類型是否支援
func ValidateDatabaseType(dbType string) error {
	supportedTypes := GetSupportedDatabases()
	for _, supportedType := range supportedTypes {
		if string(supportedType) == dbType {
			return nil
		}
	}
	return fmt.Errorf("不支援的資料庫類型: %s，支援的類型: %v", dbType, supportedTypes)
}

// DatabaseInfo 資料庫信息
type DatabaseInfo struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
}

// GetDatabaseInfo 獲取資料庫信息（不包含敏感信息）
func (f *DatabaseFactory) GetDatabaseInfo(isSlave bool) DatabaseInfo {
	var config DatabaseConnectionConfig
	if isSlave {
		config = f.config.Slave
	} else {
		config = f.config.Master
	}

	return DatabaseInfo{
		Type:     f.config.Type,
		Host:     config.Host,
		Port:     config.Port,
		Database: config.DBName,
		Username: config.Username,
	}
}
