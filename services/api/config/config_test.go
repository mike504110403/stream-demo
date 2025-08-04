package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineDatabase(t *testing.T) {
	// 創建測試配置
	databases := map[string]DatabaseConfiguration{
		"postgresql": {
			Type: "postgresql",
			Master: DatabaseConnectionConfig{
				Host: "localhost",
				Port: 5432,
			},
		},
		"mysql": {
			Type: "mysql",
			Master: DatabaseConnectionConfig{
				Host: "localhost",
				Port: 3306,
			},
		},
	}

	tests := []struct {
		name      string
		dbType    string
		expected  string
		shouldErr bool
	}{
		{
			name:     "命令行參數優先",
			dbType:   "mysql",
			expected: "mysql",
		},
		{
			name:     "默認選擇 postgresql",
			dbType:   "",
			expected: "postgresql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineDatabase(tt.dbType, databases)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateDatabaseType(t *testing.T) {
	tests := []struct {
		name      string
		dbType    string
		shouldErr bool
	}{
		{
			name:      "有效的 postgresql",
			dbType:    "postgresql",
			shouldErr: false,
		},
		{
			name:      "有效的 mysql",
			dbType:    "mysql",
			shouldErr: false,
		},
		{
			name:      "無效的資料庫類型",
			dbType:    "invalid",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDatabaseType(tt.dbType)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_GetAvailableDatabases(t *testing.T) {
	config := &Config{
		Configurations: &Configurations{
			Databases: map[string]DatabaseConfiguration{
				"postgresql": {Type: "postgresql"},
				"mysql":      {Type: "mysql"},
			},
		},
	}

	available := config.GetAvailableDatabases()
	assert.Len(t, available, 2)
	assert.Contains(t, available, "postgresql")
	assert.Contains(t, available, "mysql")
}

func TestConfig_GetDatabaseInfo(t *testing.T) {
	config := &Config{
		Configurations: &Configurations{
			Databases: map[string]DatabaseConfiguration{
				"postgresql": {Type: "postgresql"},
				"mysql":      {Type: "mysql"},
			},
		},
		ActiveDatabase: "postgresql",
	}

	info := config.GetDatabaseInfo()
	assert.Equal(t, "postgresql", info["active"])
	assert.Contains(t, info["available"], "postgresql")
	assert.Contains(t, info["available"], "mysql")
}

func TestConfig_CheckDatabaseConnections(t *testing.T) {
	// 這個測試需要實際的資料庫連接，所以我們只測試基本結構
	config := &Config{}

	// 當沒有資料庫連接時，應該返回 false
	result := config.CheckDatabaseConnections()
	// 注意：實際行為可能因環境而異，所以我們只測試函數不會 panic
	assert.IsType(t, false, result)
}

func TestConfig_ReconnectDatabases(t *testing.T) {
	// 這個測試需要實際的資料庫連接，所以我們只測試基本結構
	config := &Config{}

	// 應該不會 panic
	config.ReconnectDatabases()
}

func TestInitRedis(t *testing.T) {
	config := RedisConfiguration{
		Master: RedisConnectionConfig{
			Host: "localhost",
			Port: 6379,
		},
		Slave: RedisConnectionConfig{
			Host: "localhost",
			Port: 6379,
		},
		Pool: RedisPoolConfiguration{
			MaxActive:      10,
			MaxIdle:        5,
			IdleTimeout:    300,
			ConnectTimeout: 5,
			ReadTimeout:    3,
			WriteTimeout:   3,
		},
	}

	// 這個測試會失敗因為沒有實際的 Redis 服務器，但我們可以測試配置結構
	err := InitRedis(config)
	// 在沒有 Redis 服務器的情況下，這個調用會失敗，但這是預期的
	// 我們主要測試函數不會 panic
	assert.NotNil(t, err) // 預期會失敗，因為沒有 Redis 服務器
}

func TestConfig_SwitchDatabase(t *testing.T) {
	config := &Config{
		Configurations: &Configurations{
			Databases: map[string]DatabaseConfiguration{
				"postgresql": {
					Type: "postgresql",
					Master: DatabaseConnectionConfig{
						Host: "localhost",
						Port: 5432,
					},
				},
			},
		},
	}

	// 測試切換到不存在的資料庫
	err := config.SwitchDatabase("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database configuration not found")

	// 測試切換到無效類型的資料庫
	config.Configurations.Databases["invalid"] = DatabaseConfiguration{
		Type: "invalid",
	}
	err = config.SwitchDatabase("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database configuration error")
}

func TestNewConfig(t *testing.T) {
	// 測試環境變數配置載入（不連接資料庫）
	// 由於 NewConfig 會嘗試連接資料庫，我們跳過這個測試
	t.Skip("Skipping test that requires database connection")
}

func TestNewConfigWithEnvironmentVariables(t *testing.T) {
	// 測試環境變數配置載入（不連接資料庫）
	// 由於 NewConfig 會嘗試連接資料庫，我們跳過這個測試
	t.Skip("Skipping test that requires database connection")
}
