package config

import (
	"fmt"
	"time"

	log "stream-demo/backend/pkg/logging"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// InitPostgreSQL 初始化PostgreSQL連接
func InitPostgreSQL(dbConfig DatabaseConnectionConfig, isSlave bool) *gorm.DB {
	log.Info("=== PostgreSQL Connection Info ===")

	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.New(
			&log.Writer{},
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)
	}

	// 構建PostgreSQL DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Taipei",
		dbConfig.Host,
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.DBName,
		dbConfig.Port,
		dbConfig.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: ormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: false,
		},
	})

	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	// 設置連接池參數
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)

	// 測試連接
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping PostgreSQL database:", err)
	}

	log.Info("PostgreSQL connected successfully")

	if isSlave {
		log.Info("Connected as slave database")
	} else {
		log.Info("Connected as master database")
	}

	return db
}

// InitCacheTable 初始化緩存表
func InitCacheTable(db *gorm.DB, tableName string) {
	if tableName == "" {
		tableName = "cache_data"
	}

	// 創建緩存表的SQL
	sql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			key VARCHAR(255) PRIMARY KEY,
			value JSONB NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		
		-- 創建過期時間索引以便清理
		CREATE INDEX IF NOT EXISTS idx_%s_expires_at ON %s (expires_at);
		
		-- 創建更新時間觸發器
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
		
		DROP TRIGGER IF EXISTS update_%s_updated_at ON %s;
		CREATE TRIGGER update_%s_updated_at
			BEFORE UPDATE ON %s
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
	`, tableName, tableName, tableName, tableName, tableName, tableName, tableName)

	if err := db.Exec(sql).Error; err != nil {
		log.Error("Failed to create cache table:", err)
	} else {
		log.Info("Cache table initialized successfully")
	}
}

// PostgreSQLCache PostgreSQL緩存實現
type PostgreSQLCache struct {
	db        *gorm.DB
	tableName string
}

// NewPostgreSQLCache 創建PostgreSQL緩存實例
func NewPostgreSQLCache(db *gorm.DB, tableName string) *PostgreSQLCache {
	if tableName == "" {
		tableName = "cache_data"
	}

	return &PostgreSQLCache{
		db:        db,
		tableName: tableName,
	}
}

// CacheItem 緩存項目結構
type CacheItem struct {
	Key       string     `gorm:"primaryKey;column:key"`
	Value     string     `gorm:"type:jsonb;column:value"`
	ExpiresAt *time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

// Set 設置緩存
func (c *PostgreSQLCache) Set(key string, value interface{}, expiration time.Duration) error {
	var expiresAt *time.Time
	if expiration > 0 {
		exp := time.Now().Add(expiration)
		expiresAt = &exp
	}

	// 將value序列化為JSON字符串
	valueJSON := fmt.Sprintf("%v", value)

	item := CacheItem{
		Key:       key,
		Value:     valueJSON,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 使用UPSERT操作
	sql := fmt.Sprintf(`
		INSERT INTO %s (key, value, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (key) 
		DO UPDATE SET 
			value = EXCLUDED.value,
			expires_at = EXCLUDED.expires_at,
			updated_at = CURRENT_TIMESTAMP
	`, c.tableName)

	return c.db.Exec(sql, item.Key, item.Value, item.ExpiresAt, item.CreatedAt, item.UpdatedAt).Error
}

// Get 獲取緩存
func (c *PostgreSQLCache) Get(key string) (interface{}, bool) {
	var item CacheItem

	sql := fmt.Sprintf(`
		SELECT * FROM %s 
		WHERE key = ? 
		AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`, c.tableName)

	if err := c.db.Raw(sql, key).Scan(&item).Error; err != nil {
		return nil, false
	}

	if item.Key == "" {
		return nil, false
	}

	return item.Value, true
}

// Delete 刪除緩存
func (c *PostgreSQLCache) Delete(key string) error {
	sql := fmt.Sprintf("DELETE FROM %s WHERE key = ?", c.tableName)
	return c.db.Exec(sql, key).Error
}

// CleanExpired 清理過期的緩存項目
func (c *PostgreSQLCache) CleanExpired() error {
	sql := fmt.Sprintf("DELETE FROM %s WHERE expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP", c.tableName)
	return c.db.Exec(sql).Error
}

// PostgreSQLMessaging PostgreSQL訊息佇列實現
type PostgreSQLMessaging struct {
	db *gorm.DB
}

// NewPostgreSQLMessaging 創建PostgreSQL訊息佇列實例
func NewPostgreSQLMessaging(db *gorm.DB) *PostgreSQLMessaging {
	return &PostgreSQLMessaging{db: db}
}

// Publish 發布訊息到指定頻道
func (m *PostgreSQLMessaging) Publish(channel string, payload interface{}) error {
	payloadJSON := fmt.Sprintf("%v", payload)
	sql := fmt.Sprintf("SELECT pg_notify('%s', ?)", channel)
	return m.db.Exec(sql, payloadJSON).Error
}

// Listen 監聽指定頻道的訊息
func (m *PostgreSQLMessaging) Listen(channel string, callback func(string)) error {
	// 這需要使用database/sql的原生連接來實現LISTEN
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}

	// PostgreSQL的LISTEN需要專用連接
	sql := fmt.Sprintf("LISTEN %s", channel)
	_, err = sqlDB.Exec(sql)
	return err
}
