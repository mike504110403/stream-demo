package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis緩存管理器
type RedisCache struct {
	client            *redis.Client
	slaveClient       *redis.Client
	db                int
	keyPrefix         string
	defaultExpiration time.Duration
}

// NewRedisCache 創建Redis緩存實例
func NewRedisCache(db int, keyPrefix string, defaultExpiration time.Duration) *RedisCache {
	// 創建指定DB的Redis客戶端
	cacheClient := redis.NewClient(&redis.Options{
		Addr:     RedisClient.Options().Addr,
		Password: RedisClient.Options().Password,
		DB:       db,
	})

	slaveClient := cacheClient
	if RedisSlaveClient != RedisClient {
		slaveClient = redis.NewClient(&redis.Options{
			Addr:     RedisSlaveClient.Options().Addr,
			Password: RedisSlaveClient.Options().Password,
			DB:       db,
		})
	}

	return &RedisCache{
		client:            cacheClient,
		slaveClient:       slaveClient,
		db:                db,
		keyPrefix:         keyPrefix,
		defaultExpiration: defaultExpiration,
	}
}

// buildKey 構建緩存鍵名
func (c *RedisCache) buildKey(key string) string {
	if c.keyPrefix != "" {
		return c.keyPrefix + key
	}
	return key
}

// Set 設置緩存
func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()

	if expiration == 0 {
		expiration = c.defaultExpiration
	}

	// 將value序列化為JSON
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化值失敗: %w", err)
	}

	fullKey := c.buildKey(key)
	return c.client.Set(ctx, fullKey, valueJSON, expiration).Err()
}

// Get 獲取緩存
func (c *RedisCache) Get(key string, dest interface{}) (bool, error) {
	ctx := context.Background()
	fullKey := c.buildKey(key)

	val, err := c.slaveClient.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // 鍵不存在
		}
		return false, err
	}

	// 反序列化JSON
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("反序列化值失敗: %w", err)
	}

	return true, nil
}

// GetString 獲取字符串類型的緩存
func (c *RedisCache) GetString(key string) (string, bool, error) {
	var value string
	exists, err := c.Get(key, &value)
	return value, exists, err
}

// GetInt 獲取整數類型的緩存
func (c *RedisCache) GetInt(key string) (int, bool, error) {
	var value int
	exists, err := c.Get(key, &value)
	return value, exists, err
}

// Delete 刪除緩存
func (c *RedisCache) Delete(key string) error {
	ctx := context.Background()
	fullKey := c.buildKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

// Exists 檢查鍵是否存在
func (c *RedisCache) Exists(key string) (bool, error) {
	ctx := context.Background()
	fullKey := c.buildKey(key)

	count, err := c.slaveClient.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Clear 清空所有緩存（危險操作，慎用）
func (c *RedisCache) Clear() error {
	ctx := context.Background()

	// 使用SCAN命令安全地獲取所有帶前綴的鍵
	pattern := c.buildKey("*")

	var cursor uint64
	var keys []string

	for {
		var err error
		var scanKeys []string
		scanKeys, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Expire 設置鍵的過期時間
func (c *RedisCache) Expire(key string, expiration time.Duration) error {
	ctx := context.Background()
	fullKey := c.buildKey(key)
	return c.client.Expire(ctx, fullKey, expiration).Err()
}

// TTL 獲取鍵的剩餘生存時間
func (c *RedisCache) TTL(key string) (time.Duration, error) {
	ctx := context.Background()
	fullKey := c.buildKey(key)
	return c.slaveClient.TTL(ctx, fullKey).Result()
}

// Increment 原子遞增
func (c *RedisCache) Increment(key string, delta int64) (int64, error) {
	ctx := context.Background()
	fullKey := c.buildKey(key)
	return c.client.IncrBy(ctx, fullKey, delta).Result()
}

// Decrement 原子遞減
func (c *RedisCache) Decrement(key string, delta int64) (int64, error) {
	ctx := context.Background()
	fullKey := c.buildKey(key)
	return c.client.DecrBy(ctx, fullKey, delta).Result()
}

// MGet 批量獲取
func (c *RedisCache) MGet(keys []string) (map[string]interface{}, error) {
	ctx := context.Background()

	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.buildKey(key)
	}

	values, err := c.slaveClient.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i, val := range values {
		if val != nil {
			var decoded interface{}
			if err := json.Unmarshal([]byte(val.(string)), &decoded); err == nil {
				result[keys[i]] = decoded
			}
		}
	}

	return result, nil
}

// MSet 批量設置
func (c *RedisCache) MSet(items map[string]interface{}, expiration time.Duration) error {
	ctx := context.Background()

	if expiration == 0 {
		expiration = c.defaultExpiration
	}

	// 使用Pipeline提高性能
	pipe := c.client.Pipeline()

	for key, value := range items {
		valueJSON, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("序列化值失敗 (key: %s): %w", key, err)
		}

		fullKey := c.buildKey(key)
		pipe.Set(ctx, fullKey, valueJSON, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Stats 獲取緩存統計信息
func (c *RedisCache) Stats() (map[string]interface{}, error) {
	ctx := context.Background()

	info, err := c.client.Info(ctx, "memory", "keyspace").Result()
	if err != nil {
		return nil, err
	}

	dbSize, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"db":      c.db,
		"db_size": dbSize,
		"prefix":  c.keyPrefix,
		"info":    info,
	}, nil
}

// Close 關閉緩存連接
func (c *RedisCache) Close() error {
	var errs []error

	if c.client != nil {
		if err := c.client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis cache client: %w", err))
		}
	}

	if c.slaveClient != nil && c.slaveClient != c.client {
		if err := c.slaveClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis cache slave client: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Redis cache: %v", errs)
	}

	return nil
}
