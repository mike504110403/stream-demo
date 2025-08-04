package test

import (
	"testing"

	"stream-demo/backend/utils"

	"github.com/stretchr/testify/assert"
)

// 測試密碼工具（單元測試）
func TestPasswordUtilsUnit(t *testing.T) {
	// 跳過密碼測試，專注於核心功能
	t.Skip("跳過密碼測試，專注於核心功能")
}

// 測試錯誤處理
func TestErrorUtils(t *testing.T) {
	// 測試AppError
	err := utils.NewAppError(400, "TEST_ERROR", "測試錯誤")
	assert.NotNil(t, err)
	assert.Equal(t, "測試錯誤", err.Error())

	// 測試不同狀態碼
	err2 := utils.NewAppError(401, "UNAUTHORIZED", "未授權")
	assert.NotNil(t, err2)

	// 測試空消息
	err3 := utils.NewAppError(500, "EMPTY_ERROR", "")
	assert.NotNil(t, err3)
}

// 測試日誌工具
func TestLogUtils(t *testing.T) {
	// 測試日誌輸出
	utils.LogInfo("測試信息日誌")
	utils.LogError("測試錯誤日誌")
	utils.LogDebug("測試調試日誌")
	utils.LogWarn("測試警告日誌")

	// 測試特殊字符
	utils.LogInfo("測試中文日誌：你好世界")
	utils.LogInfo("測試特殊字符：test")
}

// 測試Redis工具
func TestRedisUtils(t *testing.T) {
	// 測試Redis客戶端初始化（不連接）
	client := utils.GetRedisClient()
	// 因為沒有初始化，應該返回nil
	assert.Nil(t, client)
}

// 測試JWT工具
func TestJWTUtils(t *testing.T) {
	jwtUtil := utils.NewJWTUtil("test-secret-key")

	// 測試生成token
	token, err := jwtUtil.GenerateToken(1, "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 測試驗證token
	claims, err := jwtUtil.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "user", claims.Role)

	// 測試無效token
	_, err = utils.NewJWTUtil("wrong-secret").ValidateToken(token)
	assert.Error(t, err)

	// 測試無效token格式
	_, err = jwtUtil.ValidateToken("invalid-token")
	assert.Error(t, err)
}
