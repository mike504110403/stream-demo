package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// RedisClient Redis主客戶端
	RedisClient *redis.Client
	// RedisSlaveClient Redis從客戶端（只讀）
	RedisSlaveClient *redis.Client
)

// RedisConfig Redis配置結構
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	// 連接池配置
	MaxActive      int
	MaxIdle        int
	IdleTimeout    int
	ConnectTimeout int
	ReadTimeout    int
	WriteTimeout   int
}

// InitRedisClient 初始化Redis客戶端
func InitRedisClient(masterConfig, slaveConfig RedisConfig) error {
	// 初始化主Redis客戶端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", masterConfig.Host, masterConfig.Port),
		Password:        masterConfig.Password,
		DB:              masterConfig.DB,
		PoolSize:        masterConfig.MaxActive,
		MinIdleConns:    masterConfig.MaxIdle,
		DialTimeout:     time.Duration(masterConfig.ConnectTimeout) * time.Second,
		ReadTimeout:     time.Duration(masterConfig.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(masterConfig.WriteTimeout) * time.Second,
		ConnMaxIdleTime: time.Duration(masterConfig.IdleTimeout) * time.Second,
	})

	// 測試主Redis連接
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis master connection failed: %w", err)
	}

	// 初始化從Redis客戶端（如果配置不同）
	if slaveConfig.Host != masterConfig.Host || slaveConfig.Port != masterConfig.Port {
		RedisSlaveClient = redis.NewClient(&redis.Options{
			Addr:            fmt.Sprintf("%s:%d", slaveConfig.Host, slaveConfig.Port),
			Password:        slaveConfig.Password,
			DB:              slaveConfig.DB,
			PoolSize:        slaveConfig.MaxActive,
			MinIdleConns:    slaveConfig.MaxIdle,
			DialTimeout:     time.Duration(slaveConfig.ConnectTimeout) * time.Second,
			ReadTimeout:     time.Duration(slaveConfig.ReadTimeout) * time.Second,
			WriteTimeout:    time.Duration(slaveConfig.WriteTimeout) * time.Second,
			ConnMaxIdleTime: time.Duration(slaveConfig.IdleTimeout) * time.Second,
		})

		// 測試從Redis連接
		if err := RedisSlaveClient.Ping(ctx).Err(); err != nil {
			fmt.Printf("WARNING: Redis slave connection failed, using master for reads: %v\n", err)
			RedisSlaveClient = RedisClient
		}
	} else {
		// 如果配置相同，使用同一個客戶端
		RedisSlaveClient = RedisClient
	}

	fmt.Println("INFO: Redis clients initialized successfully")
	return nil
}

// GetRedisClient 獲取Redis主客戶端（寫操作）
func GetRedisClient() *redis.Client {
	return RedisClient
}

// GetRedisSlaveClient 獲取Redis從客戶端（讀操作）
func GetRedisSlaveClient() *redis.Client {
	return RedisSlaveClient
}

// CloseRedisClients 關閉Redis連接
func CloseRedisClients() error {
	var errs []error

	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis master client: %w", err))
		}
	}

	if RedisSlaveClient != nil && RedisSlaveClient != RedisClient {
		if err := RedisSlaveClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis slave client: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Redis clients: %v", errs)
	}

	return nil
}

// CheckRedisConnection 檢查Redis連接狀態
func CheckRedisConnection() error {
	ctx := context.Background()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis master ping failed: %w", err)
	}

	if RedisSlaveClient != RedisClient {
		if err := RedisSlaveClient.Ping(ctx).Err(); err != nil {
			fmt.Printf("WARNING: Redis slave ping failed: %v\n", err)
		}
	}

	return nil
}

// GetRedisStats 獲取Redis狀態信息
func GetRedisStats() (map[string]interface{}, error) {
	ctx := context.Background()

	info, err := RedisClient.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	poolStats := RedisClient.PoolStats()

	return map[string]interface{}{
		"redis_info": info,
		"pool_stats": map[string]interface{}{
			"hits":        poolStats.Hits,
			"misses":      poolStats.Misses,
			"timeouts":    poolStats.Timeouts,
			"total_conns": poolStats.TotalConns,
			"idle_conns":  poolStats.IdleConns,
			"stale_conns": poolStats.StaleConns,
		},
	}, nil
}
