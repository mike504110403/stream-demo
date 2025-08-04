package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/EDDYCJY/go-gin-example/pkg/file"
)

type Level int

var (
	F *os.File

	DefaultPrefix      = ""
	DefaultCallerDepth = 2

	logger     *log.Logger
	logPrefix  = ""
	levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
)

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

// Setup initialize the log instance
func Setup() {
	var err error
	var filePath string
	var fileName string
	filePath = "logs/"
	fileName = fmt.Sprintf("%s%s.%s", "log", time.Now().Format("20060102"), "log")

	F, err = file.MustOpen(fileName, filePath)
	if err != nil {
		log.Fatalf("[%s] logging.Setup err: %v", time.Now().Format("2006-01-02 15:04:05"), err)
	}

	// 設置標準日誌格式，包含時間戳
	logger = log.New(F, DefaultPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)
}

// formatLogMessage formats a log message with timestamp
func formatLogMessage(level Level, message string) string {
	return fmt.Sprintf("[%s][%s] %s",
		time.Now().Format("2006-01-02 15:04:05.000"),
		levelFlags[level],
		message,
	)
}

// Debug output logs at debug level
func Debug(v ...interface{}) {
	setPrefix(DEBUG)
	logger.Println(v...)
	fmt.Println(formatLogMessage(DEBUG, fmt.Sprint(v...)))
}

// Info output logs at info level
func Info(v ...interface{}) {
	setPrefix(INFO)
	logger.Println(v...)
	fmt.Println(formatLogMessage(INFO, fmt.Sprint(v...)))
}

// Warn output logs at warn level
func Warn(v ...interface{}) {
	setPrefix(WARNING)
	logger.Println(v...)
	fmt.Println(formatLogMessage(WARNING, fmt.Sprint(v...)))
}

// Error output logs at error level
func Error(v ...interface{}) {
	setPrefix(ERROR)
	logger.Println(v...)
	fmt.Println(formatLogMessage(ERROR, fmt.Sprint(v...)))
}

// Fatal output logs at fatal level
func Fatal(v ...interface{}) {
	setPrefix(FATAL)
	logger.Fatalln(v...)
	fmt.Println(formatLogMessage(FATAL, fmt.Sprint(v...)))
}

// setPrefix set the prefix of the log output
func setPrefix(level Level) {
	_, file, line, ok := runtime.Caller(DefaultCallerDepth)
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s][%s:%d]", time.Now().Format("2006-01-02 15:04:05"), levelFlags[level], filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s][%s]", time.Now().Format("2006-01-02 15:04:05"), levelFlags[level])
	}

	logger.SetPrefix(logPrefix)
}

type Writer struct{}

// Printf 實現 gorm logger 的輸出方法
func (w *Writer) Printf(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

// SystemLog 用於記錄系統級別的日誌
func SystemLog(component string, message string) {
	logMessage := fmt.Sprintf("[%s][SYSTEM][%s] %s",
		time.Now().Format("2006-01-02 15:04:05.000"),
		component,
		message,
	)
	logger.Println(logMessage)
	fmt.Println(logMessage)
}

// MonitorLog 用於記錄監控資訊
func MonitorLog(metric string, value interface{}) {
	logMessage := fmt.Sprintf("[%s][MONITOR][%s] %v",
		time.Now().Format("2006-01-02 15:04:05.000"),
		metric,
		value,
	)
	logger.Println(logMessage)
	fmt.Println(logMessage)
}
