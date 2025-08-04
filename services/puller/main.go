package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// StreamConfig 外部流配置
type StreamConfig struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"uniqueIndex"`
	Title       string `json:"title"`
	Type        string `json:"type"` // "hls", "rtmp", "rtsp"
	URL         string `json:"url"`
	Enabled     bool   `json:"enabled"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

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

// NewStreamPuller 創建優化的流拉取器
func NewStreamPuller(outputDir string, httpPort int, db *gorm.DB) *StreamPuller {
	return &StreamPuller{
		streams:       make(map[string]*StreamProcess),
		outputDir:     outputDir,
		httpPort:      httpPort,
		db:            db,
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

	var streams []StreamConfig
	if err := sp.db.Table("public_streams").Where("enabled = ?", true).Find(&streams).Error; err != nil {
		return fmt.Errorf("查詢流配置失敗: %w", err)
	}

	for _, stream := range streams {
		sp.AddStream(stream)
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
	mux := http.NewServeMux()

	// 健康檢查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "stream-puller",
		})
	})

	// 流狀態
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		sp.mu.RLock()
		defer sp.mu.RUnlock()

		status := make(map[string]interface{})
		for name, streamProcess := range sp.streams {
			streamProcess.mu.Lock()
			status[name] = map[string]interface{}{
				"running": streamProcess.Running,
				"config":  streamProcess.Config,
			}
			streamProcess.mu.Unlock()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// 靜態文件服務 (HLS 文件)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 設置 CORS 頭
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 提供靜態文件
		http.FileServer(http.Dir(sp.outputDir)).ServeHTTP(w, r)
	})

	// 啟動服務器
	addr := fmt.Sprintf(":%d", sp.httpPort)
	logInfo("🌐 HTTP 服務器啟動在端口 %d", sp.httpPort)
	if err := http.ListenAndServe(addr, mux); err != nil {
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
	// 解析命令行參數
	outputDir := flag.String("output", "/tmp/public_streams", "輸出目錄")
	port := flag.Int("port", 8081, "HTTP 端口")
	dbHost := flag.String("db-host", "localhost", "資料庫主機")
	dbPort := flag.Int("db-port", 5432, "資料庫端口")
	dbUser := flag.String("db-user", "stream_user", "資料庫用戶")
	dbPass := flag.String("db-pass", "stream_password", "資料庫密碼")
	dbName := flag.String("db-name", "stream_demo", "資料庫名稱")
	flag.Parse()

	// 連接資料庫
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPass, *dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logError("連接資料庫失敗: %v", err)
		// 繼續運行，但不載入資料庫配置
		db = nil
	} else {
		logInfo("✅ 資料庫連接成功")
	}

	// 創建並啟動 StreamPuller
	sp := NewStreamPuller(*outputDir, *port, db)
	if err := sp.Start(); err != nil {
		logError("啟動 StreamPuller 失敗: %v", err)
		os.Exit(1)
	}
}
