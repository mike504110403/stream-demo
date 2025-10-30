package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"stream-demo/backend/database/models"
	"stream-demo/backend/repositories/postgresql"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// StreamConfig 外部流配置 (使用資料庫模型)
type StreamConfig = models.PublicStream

// StreamProcess 流進程管理
type StreamProcess struct {
	Config   *StreamConfig
	Process  *exec.Cmd
	StopChan chan bool
	Running  bool
	mu       sync.Mutex
}

// StreamPuller 優化的流拉取器
type StreamPuller struct {
	streams       map[string]*StreamProcess
	outputDir     string
	httpPort      int
	mu            sync.RWMutex
	db            *gorm.DB
	configService *services.PublicStreamConfigService
	maxConcurrent int // 最大同時轉檔數
}

// 簡單的日誌函數
func logInfo(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func logError(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

func logWarning(format string, args ...interface{}) {
	fmt.Printf("[WARNING] "+format+"\n", args...)
}

// getEnv 從環境變數讀取字串，如果不存在則返回預設值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 從環境變數讀取整數，如果不存在或解析失敗則返回預設值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// NewStreamPuller 創建優化的流拉取器
func NewStreamPuller(outputDir string, httpPort int, db *gorm.DB) *StreamPuller {
	repo := postgresql.NewPublicStreamRepository(db)
	configService := services.NewPublicStreamConfigService(repo)

	return &StreamPuller{
		streams:       make(map[string]*StreamProcess),
		outputDir:     outputDir,
		httpPort:      httpPort,
		db:            db,
		configService: configService,
		maxConcurrent: 2, // 限制最多同時轉檔 2 個流
	}
}

// AddStream 添加外部流配置
func (sp *StreamPuller) AddStream(config StreamConfig) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	streamProcess := &StreamProcess{
		Config:   &config,
		StopChan: make(chan bool, 1),
		Running:  false,
	}

	sp.streams[config.Name] = streamProcess
	logInfo("📺 添加外部流: %s (%s) - 類型: %s", config.Name, config.Title, config.Type)
}

// Start 啟動優化的拉流服務
func (sp *StreamPuller) Start() error {
	logInfo("🎬 啟動優化流拉取服務...")
	logInfo("📁 輸出目錄: %s", sp.outputDir)
	logInfo("🌐 HTTP 端口: %d", sp.httpPort)

	// 創建輸出目錄
	if err := os.MkdirAll(sp.outputDir, 0755); err != nil {
		return fmt.Errorf("創建輸出目錄失敗: %w", err)
	}

	// 從資料庫載入啟用的流配置
	if err := sp.loadStreamsFromDatabase(); err != nil {
		logError("載入資料庫配置失敗: %v", err)
		return err
	}

	// 啟動所有啟用的外部流
	sp.mu.RLock()
	for name, streamProcess := range sp.streams {
		if streamProcess.Config.Enabled {
			go sp.startExternalStream(name, streamProcess)
		}
	}
	sp.mu.RUnlock()

	// 啟動配置重載器
	go sp.startConfigReloader()

	// 啟動 HTTP 服務器
	go sp.startHTTPServer()

	// 等待信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logInfo("🛑 收到停止信號，正在關閉服務...")
	sp.Stop()
	return nil
}

// loadStreamsFromDatabase 從資料庫載入流配置
func (sp *StreamPuller) loadStreamsFromDatabase() error {
	if sp.db == nil {
		logWarning("資料庫未連接，跳過載入配置")
		return nil
	}

	// 先查詢所有記錄，看看資料庫中有什麼
	var allStreams []StreamConfig
	if err := sp.db.Find(&allStreams).Error; err != nil {
		return fmt.Errorf("查詢所有流配置失敗: %w", err)
	}
	logInfo("📊 資料庫中共有 %d 個流配置", len(allStreams))

	// 查詢啟用的流配置
	var streams []StreamConfig
	if err := sp.db.Where("enabled = ?", true).Find(&streams).Error; err != nil {
		return fmt.Errorf("查詢啟用的流配置失敗: %w", err)
	}

	for _, stream := range streams {
		sp.AddStream(stream)
		logInfo("📺 載入流配置: %s (%s) - 啟用: %t", stream.Name, stream.Title, stream.Enabled)
	}

	logInfo("📊 從資料庫載入了 %d 個啟用的流配置", len(streams))
	return nil
}

// stopStreamProcess 停止流進程
func (sp *StreamPuller) stopStreamProcess(name string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	streamProcess, exists := sp.streams[name]
	if !exists {
		return
	}

	streamProcess.mu.Lock()
	defer streamProcess.mu.Unlock()

	if streamProcess.Running && streamProcess.Process != nil {
		logInfo("🛑 停止流進程: %s", name)
		streamProcess.StopChan <- true
		streamProcess.Process.Process.Kill()
		streamProcess.Running = false
	}
}

// startConfigReloader 啟動配置重載器
func (sp *StreamPuller) startConfigReloader() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		logInfo("🔄 重新載入配置...")
		if err := sp.loadStreamsFromDatabase(); err != nil {
			logError("重新載入配置失敗: %v", err)
		}
	}
}

// startExternalStream 啟動外部流
func (sp *StreamPuller) startExternalStream(name string, streamProcess *StreamProcess) {
	streamProcess.mu.Lock()
	if streamProcess.Running {
		streamProcess.mu.Unlock()
		return
	}
	streamProcess.Running = true
	streamProcess.mu.Unlock()

	defer func() {
		streamProcess.mu.Lock()
		streamProcess.Running = false
		streamProcess.mu.Unlock()
	}()

	config := streamProcess.Config
	outputPath := fmt.Sprintf("%s/%s", sp.outputDir, name)

	logInfo("🎬 啟動外部流: %s -> %s", config.URL, outputPath)

	// 根據流類型選擇 FFmpeg 命令
	var cmd *exec.Cmd
	switch strings.ToLower(config.Type) {
	case "hls":
		cmd = exec.Command("ffmpeg",
			"-i", config.URL,
			"-c", "copy",
			"-f", "hls",
			"-hls_time", "2",
			"-hls_list_size", "5",
			"-hls_flags", "delete_segments",
			"-hls_segment_filename", fmt.Sprintf("%s/%%03d.ts", outputPath),
			fmt.Sprintf("%s/index.m3u8", outputPath))
	case "rtmp":
		cmd = exec.Command("ffmpeg",
			"-i", config.URL,
			"-c", "copy",
			"-f", "hls",
			"-hls_time", "2",
			"-hls_list_size", "5",
			"-hls_flags", "delete_segments",
			"-hls_segment_filename", fmt.Sprintf("%s/%%03d.ts", outputPath),
			fmt.Sprintf("%s/index.m3u8", outputPath))
	case "rtsp":
		cmd = exec.Command("ffmpeg",
			"-i", config.URL,
			"-c", "copy",
			"-f", "hls",
			"-hls_time", "2",
			"-hls_list_size", "5",
			"-hls_flags", "delete_segments",
			"-hls_segment_filename", fmt.Sprintf("%s/%%03d.ts", outputPath),
			fmt.Sprintf("%s/index.m3u8", outputPath))
	default:
		logError("不支援的流類型: %s", config.Type)
		return
	}

	// 設置輸出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	streamProcess.Process = cmd

	// 啟動進程
	if err := cmd.Start(); err != nil {
		logError("啟動 FFmpeg 失敗: %v", err)
		return
	}

	logInfo("✅ 外部流啟動成功: %s", name)

	// 等待進程結束或收到停止信號
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			logError("外部流進程結束: %s, 錯誤: %v", name, err)
		} else {
			logInfo("外部流進程正常結束: %s", name)
		}
	case <-streamProcess.StopChan:
		logInfo("收到停止信號，結束外部流: %s", name)
	}
}

// StopStream 停止指定流
func (sp *StreamPuller) StopStream(name string) {
	logInfo("🛑 停止流: %s", name)
	sp.stopStreamProcess(name)
}

// startHTTPServer 啟動 HTTP 服務器
func (sp *StreamPuller) startHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 設置 CORS 中間件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Range")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 健康檢查
	r.GET("/health", func(c *gin.Context) {
		// 返回流狀態
		sp.mu.RLock()
		status := make(map[string]interface{})
		for name, streamProcess := range sp.streams {
			streamProcess.mu.Lock()
			status[name] = map[string]interface{}{
				"running": streamProcess.Running,
				"title":   streamProcess.Config.Title,
				"type":    streamProcess.Config.Type,
			}
			streamProcess.mu.Unlock()
		}
		sp.mu.RUnlock()

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"streams": status,
		})
	})

	// 流控制 API
	api := r.Group("/api")
	{
		api.GET("/streams", func(c *gin.Context) {
			// 獲取所有流狀態
			sp.mu.RLock()
			streams := make([]map[string]interface{}, 0)
			for name, streamProcess := range sp.streams {
				streamProcess.mu.Lock()
				streams = append(streams, map[string]interface{}{
					"name":    name,
					"title":   streamProcess.Config.Title,
					"running": streamProcess.Running,
					"type":    streamProcess.Config.Type,
					"enabled": streamProcess.Config.Enabled,
				})
				streamProcess.mu.Unlock()
			}
			sp.mu.RUnlock()

			c.JSON(http.StatusOK, gin.H{"streams": streams})
		})

		api.POST("/streams", func(c *gin.Context) {
			// 啟動流
			streamName := c.PostForm("name")
			if streamName == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Stream name is required"})
				return
			}

			sp.mu.RLock()
			streamProcess, exists := sp.streams[streamName]
			sp.mu.RUnlock()

			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
				return
			}

			go sp.startExternalStream(streamName, streamProcess)
			c.JSON(http.StatusOK, gin.H{"status": "started"})
		})

		api.DELETE("/streams", func(c *gin.Context) {
			// 停止流
			streamName := c.PostForm("name")
			if streamName == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Stream name is required"})
				return
			}

			sp.StopStream(streamName)
			c.JSON(http.StatusOK, gin.H{"status": "stopped"})
		})
	}

	// RTMP 事件處理
	r.GET("/rtmp/publish", func(c *gin.Context) {
		// RTMP 推流開始事件
		streamName := c.Query("name")
		logInfo("RTMP 推流開始: %s", streamName)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/rtmp/publish_done", func(c *gin.Context) {
		// RTMP 推流結束事件
		streamName := c.Query("name")
		logInfo("RTMP 推流結束: %s", streamName)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/rtmp/error", func(c *gin.Context) {
		// RTMP 錯誤事件
		streamName := c.Query("name")
		errorMsg := c.Query("error")
		logError("RTMP 錯誤: %s - %s", streamName, errorMsg)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 靜態文件服務 - 提供 HLS 文件 (最後註冊)
	r.GET("/hls/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")

		// 處理靜態文件
		if strings.HasSuffix(filepath, ".m3u8") || strings.HasSuffix(filepath, ".ts") {
			// 設置正確的 MIME 類型
			if strings.HasSuffix(filepath, ".m3u8") {
				c.Header("Content-Type", "application/vnd.apple.mpegurl")
			} else if strings.HasSuffix(filepath, ".ts") {
				c.Header("Content-Type", "video/mp2t")
			}

			// 構建文件路徑
			filePath := sp.outputDir + "/" + filepath

			// 檢查文件是否存在
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				c.Status(http.StatusNotFound)
				return
			}

			// 提供文件
			c.File(filePath)
			return
		}

		// 其他請求返回 404
		c.Status(http.StatusNotFound)
	})

	// 啟動服務器
	addr := fmt.Sprintf(":%d", sp.httpPort)
	logInfo("🌐 HTTP 服務器啟動在端口 %d", sp.httpPort)
	if err := r.Run(addr); err != nil {
		logError("HTTP 服務器啟動失敗: %v", err)
	}
}

// Stop 停止所有服務
func (sp *StreamPuller) Stop() {
	logInfo("🛑 停止所有流...")
	sp.mu.RLock()
	for name := range sp.streams {
		sp.StopStream(name)
	}
	sp.mu.RUnlock()
	logInfo("✅ 所有流已停止")
}

func main() {
	// 從環境變數讀取配置
	outputDir := getEnv("OUTPUT_DIR", "/tmp/public_streams")
	port := getEnvAsInt("HTTP_PORT", 8081)
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvAsInt("DB_PORT", 5432)
	dbUser := getEnv("DB_USER", "stream_user")
	dbPass := getEnv("DB_PASS", "stream_password")
	dbName := getEnv("DB_NAME", "stream_demo")

	// 連接資料庫
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logError("連接資料庫失敗: %v", err)
		// 繼續運行，但不載入資料庫配置
		db = nil
	} else {
		logInfo("✅ 資料庫連接成功")
	}

	// 創建並啟動 StreamPuller
	sp := NewStreamPuller(outputDir, port, db)
	if err := sp.Start(); err != nil {
		logError("啟動 StreamPuller 失敗: %v", err)
		os.Exit(1)
	}
}
