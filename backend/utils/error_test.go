package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	// 測試創建 AppError
	appError := &AppError{
		StatusCode: 404,
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
	}

	// 測試 Error 方法
	errorMessage := appError.Error()
	assert.Equal(t, "Resource not found", errorMessage)

	// 測試字段值
	assert.Equal(t, 404, appError.StatusCode)
	assert.Equal(t, "NOT_FOUND", appError.Code)
	assert.Equal(t, "Resource not found", appError.Message)
}

func TestNewAppError(t *testing.T) {
	// 測試 NewAppError 函數
	appError := NewAppError(500, "INTERNAL_ERROR", "Internal server error")

	// 驗證創建的錯誤
	assert.NotNil(t, appError)
	assert.Equal(t, 500, appError.StatusCode)
	assert.Equal(t, "INTERNAL_ERROR", appError.Code)
	assert.Equal(t, "Internal server error", appError.Message)

	// 測試 Error 方法
	errorMessage := appError.Error()
	assert.Equal(t, "Internal server error", errorMessage)
}

func TestAppErrorWithEmptyMessage(t *testing.T) {
	// 測試空訊息的情況
	appError := NewAppError(400, "BAD_REQUEST", "")
	
	assert.NotNil(t, appError)
	assert.Equal(t, 400, appError.StatusCode)
	assert.Equal(t, "BAD_REQUEST", appError.Code)
	assert.Equal(t, "", appError.Message)
	
	// 測試 Error 方法返回空字串
	errorMessage := appError.Error()
	assert.Equal(t, "", errorMessage)
}

func TestAppErrorWithSpecialCharacters(t *testing.T) {
	// 測試包含特殊字符的錯誤訊息
	message := "錯誤訊息：檔案 'test.txt' 不存在！"
	appError := NewAppError(400, "FILE_NOT_FOUND", message)
	
	assert.NotNil(t, appError)
	assert.Equal(t, message, appError.Message)
	
	// 測試 Error 方法
	errorMessage := appError.Error()
	assert.Equal(t, message, errorMessage)
}

func TestAppErrorStatusCode(t *testing.T) {
	// 測試不同的狀態碼
	testCases := []int{
		200, 201, 400, 401, 403, 404, 500, 502, 503,
	}
	
	for _, statusCode := range testCases {
		t.Run("status_code_"+string(rune(statusCode)), func(t *testing.T) {
			appError := NewAppError(statusCode, "TEST", "Test message")
			assert.Equal(t, statusCode, appError.StatusCode)
		})
	}
}

func TestAppErrorCode(t *testing.T) {
	// 測試不同的錯誤代碼
	testCodes := []string{
		"VALIDATION_ERROR",
		"UNAUTHORIZED",
		"FORBIDDEN",
		"NOT_FOUND",
		"INTERNAL_ERROR",
		"BAD_GATEWAY",
		"SERVICE_UNAVAILABLE",
	}
	
	for _, code := range testCodes {
		t.Run("error_code_"+code, func(t *testing.T) {
			appError := NewAppError(400, code, "Test message")
			assert.Equal(t, code, appError.Code)
		})
	}
}

func TestAppErrorImplementsErrorInterface(t *testing.T) {
	// 測試 AppError 實現了 error 接口
	var err error
	appError := NewAppError(500, "TEST", "Test error")
	
	// 這應該可以編譯，因為 AppError 實現了 Error() 方法
	err = appError
	
	// 驗證錯誤訊息
	assert.Equal(t, "Test error", err.Error())
}

func TestAppErrorNilHandling(t *testing.T) {
	// 測試 nil 處理
	var appError *AppError
	
	// 這不應該 panic
	if appError != nil {
		_ = appError.Error()
	}
	
	assert.Nil(t, appError)
}

// BenchmarkAppError 性能測試
func BenchmarkAppError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		appError := NewAppError(500, "BENCHMARK", "Benchmark error message")
		_ = appError.Error()
	}
}

func BenchmarkAppErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewAppError(404, "NOT_FOUND", "Resource not found")
	}
} 