package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

// Writer 實現 gorm logger 的 Writer 接口
type Writer struct{}

// Printf 實現 gorm logger 的輸出方法
func (w *Writer) Printf(format string, args ...interface{}) {
	LogInfo(format, args...)
}

// InitLogger 初始化日誌工具
func InitLogger() {
	// 建立日誌目錄
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatal("無法建立日誌目錄:", err)
	}

	// 設定日誌檔案名稱格式：logs/app-2006-01-02.log
	logFile := filepath.Join(logDir, fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("無法開啟日誌檔案:", err)
	}

	// 初始化不同級別的日誌
	InfoLogger = log.New(file, "[INFO] ", log.LstdFlags|log.Lshortfile)
	ErrorLogger = log.New(file, "[ERROR] ", log.LstdFlags|log.Lshortfile)
	DebugLogger = log.New(file, "[DEBUG] ", log.LstdFlags|log.Lshortfile)

	// 同時輸出到控制台
	InfoLogger.SetOutput(os.Stdout)
	ErrorLogger.SetOutput(os.Stderr)
	DebugLogger.SetOutput(os.Stdout)
}

// LogError 記錄錯誤日誌
func LogError(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	ErrorLogger.Printf("%s:%d - %s", filepath.Base(file), line, fmt.Sprintf(format, v...))
}

// LogInfo 記錄一般日誌
func LogInfo(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	InfoLogger.Printf("%s:%d - %s", filepath.Base(file), line, fmt.Sprintf(format, v...))
}

// LogDebug 記錄除錯日誌
func LogDebug(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	DebugLogger.Printf("%s:%d - %s", filepath.Base(file), line, fmt.Sprintf(format, v...))
}

// LogFatal 記錄致命錯誤日誌並退出程序
func LogFatal(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	ErrorLogger.Printf("%s:%d - FATAL: %s", filepath.Base(file), line, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// LogWarn 記錄警告日誌
func LogWarn(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	InfoLogger.Printf("%s:%d - WARN: %s", filepath.Base(file), line, fmt.Sprintf(format, v...))
}
