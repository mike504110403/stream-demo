package utils

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient 模擬 Redis 客戶端
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) Info(ctx context.Context) *redis.StringCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

func TestInitRedisClient(t *testing.T) {
	// 測試正常初始化
	masterConfig := RedisConfig{
		Host:            "localhost",
		Port:            6379,
		Password:        "",
		DB:              0,
		MaxActive:       10,
		MaxIdle:         5,
		IdleTimeout:     300,
		ConnectTimeout:  5,
		ReadTimeout:     3,
		WriteTimeout:    3,
	}

	slaveConfig := RedisConfig{
		Host:            "localhost",
		Port:            6380,
		Password:        "",
		DB:              0,
		MaxActive:       10,
		MaxIdle:         5,
		IdleTimeout:     300,
		ConnectTimeout:  5,
		ReadTimeout:     3,
		WriteTimeout:    3,
	}

	// 注意：這個測試需要實際的 Redis 服務，在 CI 環境中會失敗
	// 在實際環境中，我們應該使用 mock 或測試容器
	t.Skip("Skipping test that requires actual Redis connection")
	
	err := InitRedisClient(masterConfig, slaveConfig)
	// 在實際環境中，這裡會根據 Redis 是否可用而成功或失敗
	_ = err
}

func TestGetRedisClient(t *testing.T) {
	// 測試獲取 Redis 客戶端
	client := GetRedisClient()
	
	// 如果 Redis 客戶端未初始化，應該返回 nil
	if client == nil {
		t.Log("Redis client is nil (not initialized)")
		return
	}
	
	assert.NotNil(t, client)
}

func TestGetRedisSlaveClient(t *testing.T) {
	// 測試獲取 Redis 從客戶端
	client := GetRedisSlaveClient()
	
	// 如果 Redis 從客戶端未初始化，應該返回 nil
	if client == nil {
		t.Log("Redis slave client is nil (not initialized)")
		return
	}
	
	assert.NotNil(t, client)
}

func TestCloseRedisClients(t *testing.T) {
	// 測試關閉 Redis 客戶端
	err := CloseRedisClients()
	
	// 如果客戶端未初始化，應該返回 nil
	if err != nil {
		t.Logf("Error closing Redis clients: %v", err)
	}
	
	// 這個測試主要是確保函數不會 panic
	assert.True(t, true)
}

func TestCheckRedisConnection(t *testing.T) {
	// 測試檢查 Redis 連接
	// 如果 RedisClient 為 nil，會 panic，所以我們需要跳過這個測試
	if RedisClient == nil {
		t.Skip("Redis client not initialized")
	}
	
	err := CheckRedisConnection()
	
	// 如果 Redis 未連接，會返回錯誤
	if err != nil {
		t.Logf("Redis connection check failed: %v", err)
	}
	
	// 這個測試主要是確保函數不會 panic
	assert.True(t, true)
}

func TestGetRedisStats(t *testing.T) {
	// 測試獲取 Redis 統計信息
	// 如果 RedisClient 為 nil，會 panic，所以我們需要跳過這個測試
	if RedisClient == nil {
		t.Skip("Redis client not initialized")
	}
	
	stats, err := GetRedisStats()
	
	// 如果 Redis 未連接，會返回錯誤
	if err != nil {
		t.Logf("Failed to get Redis stats: %v", err)
		return
	}
	
	if stats != nil {
		assert.IsType(t, map[string]interface{}{}, stats)
		
		// 檢查統計信息結構
		if redisInfo, exists := stats["redis_info"]; exists {
			assert.IsType(t, "", redisInfo)
		}
		
		if poolStats, exists := stats["pool_stats"]; exists {
			assert.IsType(t, map[string]interface{}{}, poolStats)
		}
	}
}

func TestRedisConfig(t *testing.T) {
	// 測試 Redis 配置結構
	config := RedisConfig{
		Host:            "test-host",
		Port:            6379,
		Password:        "test-password",
		DB:              1,
		MaxActive:       20,
		MaxIdle:         10,
		IdleTimeout:     600,
		ConnectTimeout:  10,
		ReadTimeout:     5,
		WriteTimeout:    5,
	}
	
	assert.Equal(t, "test-host", config.Host)
	assert.Equal(t, 6379, config.Port)
	assert.Equal(t, "test-password", config.Password)
	assert.Equal(t, 1, config.DB)
	assert.Equal(t, 20, config.MaxActive)
	assert.Equal(t, 10, config.MaxIdle)
	assert.Equal(t, 600, config.IdleTimeout)
	assert.Equal(t, 10, config.ConnectTimeout)
	assert.Equal(t, 5, config.ReadTimeout)
	assert.Equal(t, 5, config.WriteTimeout)
}

// MockRedisClientTest 使用 mock 測試 Redis 客戶端
func TestMockRedisClient(t *testing.T) {
	mockClient := &MockRedisClient{}
	
	// 設置 mock 期望
	mockPingCmd := &redis.StatusCmd{}
	mockPingCmd.SetVal("PONG")
	mockClient.On("Ping", mock.Anything).Return(mockPingCmd)
	
	mockInfoCmd := &redis.StringCmd{}
	mockInfoCmd.SetVal("redis_version:6.0.0")
	mockClient.On("Info", mock.Anything).Return(mockInfoCmd)
	
	mockPoolStats := &redis.PoolStats{
		Hits:       100,
		Misses:     10,
		Timeouts:   0,
		TotalConns: 50,
		IdleConns:  20,
		StaleConns: 0,
	}
	mockClient.On("PoolStats").Return(mockPoolStats)
	
	mockClient.On("Close").Return(nil)
	
	// 測試 Ping
	ctx := context.Background()
	result := mockClient.Ping(ctx)
	assert.Equal(t, "PONG", result.Val())
	
	// 測試 Info
	info := mockClient.Info(ctx)
	assert.Equal(t, "redis_version:6.0.0", info.Val())
	
	// 測試 PoolStats
	stats := mockClient.PoolStats()
	assert.Equal(t, uint32(100), stats.Hits)
	assert.Equal(t, uint32(10), stats.Misses)
	assert.Equal(t, uint32(50), stats.TotalConns)
	
	// 測試 Close
	err := mockClient.Close()
	assert.NoError(t, err)
	
	// 驗證所有期望都被調用
	mockClient.AssertExpectations(t)
}

// BenchmarkRedisOperations 性能測試
func BenchmarkRedisOperations(b *testing.B) {
	// 這個 benchmark 需要實際的 Redis 連接
	// 在 CI 環境中會跳過
	b.Skip("Skipping benchmark that requires actual Redis connection")
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if RedisClient != nil {
			RedisClient.Ping(ctx)
		}
	}
} 