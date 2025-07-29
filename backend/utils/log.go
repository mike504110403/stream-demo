package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Writer 實現 gorm logger 的 Writer 接口
type Writer struct{}

// Printf 實現 gorm logger 的輸出方法
func (w *Writer) Printf(format string, args ...interface{}) {
	LogInfo(format, args...)
}

// InitLogger 初始化日誌工具（簡化版本，只輸出到控制台）
func InitLogger() {
	// 設置標準日誌輸出格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("📝 日誌系統初始化完成（控制台輸出模式）")
}

// LogError 記錄錯誤日誌
func LogError(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	message := fmt.Sprintf(format, v...)
	fmt.Printf("❌ [ERROR] %s:%d - %s\n", filepath.Base(file), line, message)
}

// LogInfo 記錄一般日誌
func LogInfo(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	message := fmt.Sprintf(format, v...)
	fmt.Printf("ℹ️  [INFO] %s:%d - %s\n", filepath.Base(file), line, message)
}

// LogDebug 記錄除錯日誌
func LogDebug(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	message := fmt.Sprintf(format, v...)
	fmt.Printf("🔍 [DEBUG] %s:%d - %s\n", filepath.Base(file), line, message)
}

// LogFatal 記錄致命錯誤日誌並退出程序
func LogFatal(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	message := fmt.Sprintf(format, v...)
	fmt.Printf("💥 [FATAL] %s:%d - %s\n", filepath.Base(file), line, message)
	os.Exit(1)
}

// LogWarn 記錄警告日誌
func LogWarn(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	message := fmt.Sprintf(format, v...)
	fmt.Printf("⚠️  [WARN] %s:%d - %s\n", filepath.Base(file), line, message)
}
