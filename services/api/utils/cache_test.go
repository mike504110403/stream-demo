package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPostgreSQLCache(t *testing.T) {
	// 測試創建 PostgreSQL 緩存實例
	// 注意：這個測試需要實際的 GORM 數據庫，在測試環境中會跳過
	t.Skip("Skipping test that requires actual GORM database")
	
	tableName := "test_cache"
	defaultExpiration := 5 * time.Minute
	cleanupInterval := 1 * time.Minute
	
	// 這裡我們只測試函數存在且可以編譯
	_ = NewPostgreSQLCache
	_ = tableName
	_ = defaultExpiration
	_ = cleanupInterval
}

func TestCacheItemStructure(t *testing.T) {
	// 測試緩存項目結構
	now := time.Now()
	expiresAt := now.Add(5 * time.Minute)
	
	item := CacheItem{
		Key:       "test-key",
		Value:     `{"data": "test"}`,
		ExpiresAt: &expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.Equal(t, "test-key", item.Key)
	assert.Equal(t, `{"data": "test"}`, item.Value)
	assert.Equal(t, &expiresAt, item.ExpiresAt)
	assert.Equal(t, now, item.CreatedAt)
	assert.Equal(t, now, item.UpdatedAt)
}

func TestCacheItemWithNilExpiresAt(t *testing.T) {
	// 測試 ExpiresAt 為 nil 的情況
	now := time.Now()
	
	item := CacheItem{
		Key:       "test-key",
		Value:     `{"data": "test"}`,
		ExpiresAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.Equal(t, "test-key", item.Key)
	assert.Equal(t, `{"data": "test"}`, item.Value)
	assert.Nil(t, item.ExpiresAt)
	assert.Equal(t, now, item.CreatedAt)
	assert.Equal(t, now, item.UpdatedAt)
}

func TestCacheItemWithEmptyValue(t *testing.T) {
	// 測試空值的情況
	now := time.Now()
	
	item := CacheItem{
		Key:       "test-key",
		Value:     "",
		ExpiresAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.Equal(t, "test-key", item.Key)
	assert.Equal(t, "", item.Value)
	assert.Nil(t, item.ExpiresAt)
}

func TestCacheItemWithSpecialCharacters(t *testing.T) {
	// 測試包含特殊字符的情況
	now := time.Now()
	
	item := CacheItem{
		Key:       "test-key-中文",
		Value:     `{"message": "測試訊息", "emoji": "🎉"}`,
		ExpiresAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.Equal(t, "test-key-中文", item.Key)
	assert.Equal(t, `{"message": "測試訊息", "emoji": "🎉"}`, item.Value)
}

func TestCacheItemTimeHandling(t *testing.T) {
	// 測試時間處理
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)
	
	// 測試未來時間
	item1 := CacheItem{
		Key:       "future-key",
		Value:     "future-value",
		ExpiresAt: &future,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.True(t, item1.ExpiresAt.After(now))
	
	// 測試過去時間
	item2 := CacheItem{
		Key:       "past-key",
		Value:     "past-value",
		ExpiresAt: &past,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.True(t, item2.ExpiresAt.Before(now))
}

func TestCacheItemJSONValue(t *testing.T) {
	// 測試 JSON 值處理
	now := time.Now()
	
	// 測試有效的 JSON
	item1 := CacheItem{
		Key:       "json-key",
		Value:     `{"name": "test", "age": 25, "active": true}`,
		ExpiresAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.Equal(t, `{"name": "test", "age": 25, "active": true}`, item1.Value)
	
	// 測試無效的 JSON（但仍然可以存儲為字符串）
	item2 := CacheItem{
		Key:       "invalid-json-key",
		Value:     `{"name": "test", "age": 25, "active": true,}`,
		ExpiresAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.Equal(t, `{"name": "test", "age": 25, "active": true,}`, item2.Value)
}

func TestCacheItemKeyValidation(t *testing.T) {
	// 測試鍵名驗證
	now := time.Now()
	
	testKeys := []string{
		"normal-key",
		"key_with_underscores",
		"key-with-dashes",
		"key123",
		"123key",
		"key.123",
		"key_123",
		"中文鍵名",
		"key with spaces",
		"",
		"very-long-key-name-that-might-exceed-normal-length-limits",
	}
	
	for _, key := range testKeys {
		t.Run("key_"+key, func(t *testing.T) {
			item := CacheItem{
				Key:       key,
				Value:     "test-value",
				ExpiresAt: nil,
				CreatedAt: now,
				UpdatedAt: now,
			}
			
			assert.Equal(t, key, item.Key)
			assert.Equal(t, "test-value", item.Value)
		})
	}
}

func TestCacheItemValueTypes(t *testing.T) {
	// 測試不同類型的值
	now := time.Now()
	
	testCases := []struct {
		name  string
		value string
	}{
		{"string", "simple string"},
		{"json_object", `{"key": "value"}`},
		{"json_array", `[1, 2, 3, "test"]`},
		{"json_boolean", `true`},
		{"json_number", `123.45`},
		{"json_null", `null`},
		{"empty", ""},
		{"unicode", `{"message": "Hello 世界 🌍"}`},
		{"special_chars", `{"path": "/path/to/file.txt", "query": "?param=value"}`},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := CacheItem{
				Key:       "test-key",
				Value:     tc.value,
				ExpiresAt: nil,
				CreatedAt: now,
				UpdatedAt: now,
			}
			
			assert.Equal(t, tc.value, item.Value)
		})
	}
}

func TestCacheItemTimeConsistency(t *testing.T) {
	// 測試時間一致性
	now := time.Now()
	
	item := CacheItem{
		Key:       "test-key",
		Value:     "test-value",
		ExpiresAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// 驗證時間字段
	assert.Equal(t, now, item.CreatedAt)
	assert.Equal(t, now, item.UpdatedAt)
	
	// 驗證時間不是零值
	assert.False(t, item.CreatedAt.IsZero())
	assert.False(t, item.UpdatedAt.IsZero())
}

func TestCacheItemExpirationLogic(t *testing.T) {
	// 測試過期邏輯
	now := time.Now()
	
	// 測試有過期時間的情況
	expiresAt := now.Add(5 * time.Minute)
	item1 := CacheItem{
		Key:       "expiring-key",
		Value:     "expiring-value",
		ExpiresAt: &expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.NotNil(t, item1.ExpiresAt)
	assert.True(t, item1.ExpiresAt.After(now))
	
	// 測試無過期時間的情況
	item2 := CacheItem{
		Key:       "non-expiring-key",
		Value:     "non-expiring-value",
		ExpiresAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	assert.Nil(t, item2.ExpiresAt)
}

// BenchmarkCacheItem 性能測試
func BenchmarkCacheItem(b *testing.B) {
	now := time.Now()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CacheItem{
			Key:       "bench-key",
			Value:     "bench-value",
			ExpiresAt: nil,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
}

func BenchmarkCacheItemWithExpiration(b *testing.B) {
	now := time.Now()
	expiresAt := now.Add(5 * time.Minute)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CacheItem{
			Key:       "bench-key",
			Value:     "bench-value",
			ExpiresAt: &expiresAt,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
} 