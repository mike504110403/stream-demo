package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PostgreSQLCache PostgreSQL緩存管理器
type PostgreSQLCache struct {
	db                *gorm.DB
	tableName         string
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
}

// CacheItem 緩存項目結構
type CacheItem struct {
	Key       string     `gorm:"primaryKey;column:key"`
	Value     string     `gorm:"type:jsonb;column:value"`
	ExpiresAt *time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

// NewPostgreSQLCache 創建PostgreSQL緩存實例
func NewPostgreSQLCache(db *gorm.DB, tableName string, defaultExpiration, cleanupInterval time.Duration) *PostgreSQLCache {
	if tableName == "" {
		tableName = "cache_data"
	}

	cache := &PostgreSQLCache{
		db:                db,
		tableName:         tableName,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}

	// 啟動後台清理任務
	go cache.startCleanupTask()

	return cache
}

// Set 設置緩存
func (c *PostgreSQLCache) Set(key string, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = c.defaultExpiration
	}

	var expiresAt *time.Time
	if expiration > 0 {
		exp := time.Now().Add(expiration)
		expiresAt = &exp
	}

	// 將value序列化為JSON
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化值失敗: %w", err)
	}

	item := CacheItem{
		Key:       key,
		Value:     string(valueJSON),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 使用UPSERT操作（PostgreSQL特有語法）
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
func (c *PostgreSQLCache) Get(key string, dest interface{}) (bool, error) {
	var item CacheItem

	sql := fmt.Sprintf(`
		SELECT * FROM %s 
		WHERE key = ? 
		AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`, c.tableName)

	if err := c.db.Raw(sql, key).Scan(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	if item.Key == "" {
		return false, nil
	}

	// 反序列化JSON
	if err := json.Unmarshal([]byte(item.Value), dest); err != nil {
		return false, fmt.Errorf("反序列化值失敗: %w", err)
	}

	return true, nil
}

// GetString 獲取字符串類型的緩存
func (c *PostgreSQLCache) GetString(key string) (string, bool, error) {
	var value string
	exists, err := c.Get(key, &value)
	return value, exists, err
}

// GetInt 獲取整數類型的緩存
func (c *PostgreSQLCache) GetInt(key string) (int, bool, error) {
	var value int
	exists, err := c.Get(key, &value)
	return value, exists, err
}

// Delete 刪除緩存
func (c *PostgreSQLCache) Delete(key string) error {
	sql := fmt.Sprintf("DELETE FROM %s WHERE key = ?", c.tableName)
	return c.db.Exec(sql, key).Error
}

// Exists 檢查鍵是否存在
func (c *PostgreSQLCache) Exists(key string) (bool, error) {
	var count int64
	sql := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		WHERE key = ? 
		AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`, c.tableName)

	if err := c.db.Raw(sql, key).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// Clear 清空所有緩存
func (c *PostgreSQLCache) Clear() error {
	sql := fmt.Sprintf("DELETE FROM %s", c.tableName)
	return c.db.Exec(sql).Error
}

// Expire 設置鍵的過期時間
func (c *PostgreSQLCache) Expire(key string, expiration time.Duration) error {
	expiresAt := time.Now().Add(expiration)
	sql := fmt.Sprintf("UPDATE %s SET expires_at = ? WHERE key = ?", c.tableName)
	return c.db.Exec(sql, expiresAt, key).Error
}

// TTL 獲取鍵的剩餘生存時間
func (c *PostgreSQLCache) TTL(key string) (time.Duration, error) {
	var expiresAt *time.Time
	sql := fmt.Sprintf("SELECT expires_at FROM %s WHERE key = ?", c.tableName)

	if err := c.db.Raw(sql, key).Scan(&expiresAt).Error; err != nil {
		return 0, err
	}

	if expiresAt == nil {
		return -1, nil // 永不過期
	}

	if time.Now().After(*expiresAt) {
		return 0, nil // 已過期
	}

	return expiresAt.Sub(time.Now()), nil
}

// cleanExpired 清理過期的緩存項目
func (c *PostgreSQLCache) cleanExpired() error {
	sql := fmt.Sprintf("DELETE FROM %s WHERE expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP", c.tableName)
	result := c.db.Exec(sql)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		LogInfo("清理了 %d 個過期的緩存項目", result.RowsAffected)
	}

	return nil
}

// startCleanupTask 啟動後台清理任務
func (c *PostgreSQLCache) startCleanupTask() {
	if c.cleanupInterval <= 0 {
		return
	}

	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.cleanExpired(); err != nil {
			LogError("清理過期緩存失敗: %v", err)
		}
	}
}

// Stats 獲取緩存統計資訊
func (c *PostgreSQLCache) Stats() (map[string]interface{}, error) {
	var total, expired int64

	// 總數
	totalSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s", c.tableName)
	if err := c.db.Raw(totalSQL).Count(&total).Error; err != nil {
		return nil, err
	}

	// 過期數
	expiredSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP", c.tableName)
	if err := c.db.Raw(expiredSQL).Count(&expired).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_keys":   total,
		"expired_keys": expired,
		"active_keys":  total - expired,
		"table_name":   c.tableName,
	}, nil
}
