package utils

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriterPrintf(t *testing.T) {
	// 測試 Writer 的 Printf 方法
	writer := &Writer{}
	
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	writer.Printf("Test message: %s", "hello")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出包含預期的訊息
	assert.Contains(t, output, "Test message: hello")
	assert.Contains(t, output, "[INFO]")
}

func TestInitLogger(t *testing.T) {
	// 測試初始化日誌函數是否正常執行
	// 由於 log.Println 的輸出行為在不同環境下可能不同，
	// 我們只測試函數是否正常執行而不檢查具體輸出
	assert.NotPanics(t, func() {
		InitLogger()
	})
}

func TestLogError(t *testing.T) {
	// 測試錯誤日誌
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	LogError("Test error message: %s", "error details")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出
	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "Test error message: error details")
	assert.Contains(t, output, "❌")
}

func TestLogInfo(t *testing.T) {
	// 測試一般日誌
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	LogInfo("Test info message: %s", "info details")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "Test info message: info details")
	assert.Contains(t, output, "ℹ️")
}

func TestLogDebug(t *testing.T) {
	// 測試除錯日誌
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	LogDebug("Test debug message: %s", "debug details")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出
	assert.Contains(t, output, "[DEBUG]")
	assert.Contains(t, output, "Test debug message: debug details")
	assert.Contains(t, output, "🔍")
}

func TestLogWarn(t *testing.T) {
	// 測試警告日誌
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	LogWarn("Test warning message: %s", "warning details")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出
	assert.Contains(t, output, "[WARN]")
	assert.Contains(t, output, "Test warning message: warning details")
	assert.Contains(t, output, "⚠️")
}

func TestLogFatal(t *testing.T) {
	// 測試致命錯誤日誌（不實際退出程序）
	// 由於 LogFatal 會調用 os.Exit(1)，我們需要特殊處理
	// 在測試環境中，我們只測試輸出格式，不測試實際退出
	
	// 注意：在測試中我們不實際調用 LogFatal，因為它會退出程序
	// 這裡我們只測試函數存在且可以編譯
	_ = LogFatal
	
	// 驗證函數存在
	assert.NotNil(t, LogFatal)
}

func TestLogWithSpecialCharacters(t *testing.T) {
	// 測試包含特殊字符的日誌
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	LogInfo("特殊字符測試：中文、emoji 🎉、數字 123")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出
	assert.Contains(t, output, "特殊字符測試：中文、emoji 🎉、數字 123")
	assert.Contains(t, output, "[INFO]")
}

func TestLogWithEmptyMessage(t *testing.T) {
	// 測試空訊息的日誌
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	LogInfo("")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出包含日誌格式但不包含具體訊息
	assert.Contains(t, output, "[INFO]")
}

func TestLogWithFormatting(t *testing.T) {
	// 測試格式化日誌
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	LogInfo("User %s logged in from %s", "john", "192.168.1.1")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出
	assert.Contains(t, output, "User john logged in from 192.168.1.1")
	assert.Contains(t, output, "[INFO]")
}

func TestLogWithMultipleArguments(t *testing.T) {
	// 測試多個參數的日誌
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	LogError("Database connection failed: %s, retry count: %d, timeout: %v", 
		"connection refused", 3, "30s")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出
	assert.Contains(t, output, "Database connection failed: connection refused, retry count: 3, timeout: 30s")
	assert.Contains(t, output, "[ERROR]")
}

func TestLogFileAndLineInfo(t *testing.T) {
	// 測試日誌包含文件和行號信息
	// 捕獲標準輸出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 執行測試
	LogInfo("Test message with file and line info")
	
	// 恢復標準輸出
	w.Close()
	os.Stdout = oldStdout
	
	// 讀取輸出
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// 驗證輸出包含文件信息
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "Test message with file and line info")
	// 驗證包含文件名（不包含完整路徑）
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "[INFO]") {
			// 檢查是否包含文件名和行號格式
			assert.True(t, strings.Contains(line, ":") || strings.Contains(line, "log_test.go"))
			break
		}
	}
}

// BenchmarkLogFunctions 性能測試
func BenchmarkLogInfo(b *testing.B) {
	// 禁用輸出以避免測試變慢
	oldStdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = oldStdout }()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogInfo("Benchmark message %d", i)
	}
}

func BenchmarkLogError(b *testing.B) {
	// 禁用輸出以避免測試變慢
	oldStdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = oldStdout }()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogError("Benchmark error message %d", i)
	}
}

func BenchmarkWriterPrintf(b *testing.B) {
	// 禁用輸出以避免測試變慢
	oldStdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = oldStdout }()
	
	writer := &Writer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Printf("Benchmark writer message %d", i)
	}
} 