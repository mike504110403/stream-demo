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

	"stream-demo/backend/database/models"
	"stream-demo/backend/repositories/postgresql"
	"stream-demo/backend/services"
	"stream-demo/backend/utils"

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
	utils.LogInfo("📺 添加外部流: %s (%s) - 類型: %s", config.Name, config.Title, config.Type)
}

// Start 啟動優化的拉流服務
func (sp *StreamPuller) Start() error {
	utils.LogInfo("🎬 啟動優化流拉取服務...")
	utils.LogInfo("📁 輸出目錄: %s", sp.outputDir)
	utils.LogInfo("🌐 HTTP 端口: %d", sp.httpPort)

	// 創建輸出目錄
	if err := os.MkdirAll(sp.outputDir, 0755); err != nil {
		return fmt.Errorf("創建輸出目錄失敗: %w", err)
	}

	// 從資料庫載入啟用的流配置
	if err := sp.loadStreamsFromDatabase(); err != nil {
		utils.LogError("載入資料庫配置失敗: %v", err)
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

	// 啟動 HTTP 服務器
	go sp.startHTTPServer()

	// 啟動定期重新載入配置
	go sp.startConfigReloader()

	utils.LogInfo("✅ 優化流拉取服務啟動成功")
	return nil
}

// loadStreamsFromDatabase 從資料庫載入流配置
func (sp *StreamPuller) loadStreamsFromDatabase() error {
	streams, err := sp.configService.GetEnabledStreams()
	if err != nil {
		return fmt.Errorf("獲取啟用流失敗: %w", err)
	}

	sp.mu.Lock()
	defer sp.mu.Unlock()

	// 創建新的流映射
	newStreams := make(map[string]*StreamProcess)

	// 載入資料庫配置
	for _, stream := range streams {
		// 檢查是否已存在
		if existingProcess, exists := sp.streams[stream.Name]; exists {
			// 如果已存在且配置相同，保持現有進程
			if existingProcess.Config.Enabled == stream.Enabled &&
				existingProcess.Config.URL == stream.URL &&
				existingProcess.Config.Type == stream.Type {
				newStreams[stream.Name] = existingProcess
				continue
			}
			// 如果配置改變，停止現有進程
			utils.LogInfo("🔄 配置改變，停止現有流: %s", stream.Name)
			sp.stopStreamProcess(stream.Name)
		}

		// 創建新的流進程
		streamProcess := &StreamProcess{
			Config:   &stream,
			StopChan: make(chan bool, 1),
			Running:  false,
		}
		newStreams[stream.Name] = streamProcess
		utils.LogInfo("📺 載入資料庫流配置: %s (%s) - 類型: %s", stream.Name, stream.Title, stream.Type)

		// 如果啟用，檢查併發數限制
		if stream.Enabled {
			// 計算當前運行的流數量
			runningCount := 0
			for _, existingProcess := range sp.streams {
				if existingProcess.Running {
					runningCount++
				}
			}

			if runningCount < sp.maxConcurrent {
				go sp.startExternalStream(stream.Name, streamProcess)
			} else {
				utils.LogInfo("⚠️ 達到最大併發數限制 (%d)，延遲啟動流: %s", sp.maxConcurrent, stream.Name)
				// 可以考慮加入排隊機制
			}
		}
	}

	// 停止不再存在的流
	for name := range sp.streams {
		if _, exists := newStreams[name]; !exists {
			utils.LogInfo("🛑 停止不再存在的流: %s", name)
			sp.stopStreamProcess(name)
		}
	}

	// 更新流映射
	sp.streams = newStreams

	return nil
}

// stopStreamProcess 停止流進程的輔助函數
func (sp *StreamPuller) stopStreamProcess(name string) {
	if process, exists := sp.streams[name]; exists {
		process.mu.Lock()
		if process.Running {
			select {
			case process.StopChan <- true:
			default:
			}
			if process.Process != nil && process.Process.Process != nil {
				process.Process.Process.Kill()
			}
		}
		process.Running = false
		process.mu.Unlock()
	}
}

// startConfigReloader 啟動配置重新載入器
func (sp *StreamPuller) startConfigReloader() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒檢查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sp.loadStreamsFromDatabase(); err != nil {
				utils.LogError("重新載入配置失敗: %v", err)
			}
		}
	}
}

// startExternalStream 啟動外部流處理
func (sp *StreamPuller) startExternalStream(name string, streamProcess *StreamProcess) {
	streamProcess.mu.Lock()
	if streamProcess.Running {
		streamProcess.mu.Unlock()
		return
	}
	streamProcess.Running = true
	streamProcess.mu.Unlock()

	utils.LogInfo("📺 啟動外部流: %s (%s)", name, streamProcess.Config.Title)

	streamDir := fmt.Sprintf("%s/%s", sp.outputDir, name)
	if err := os.MkdirAll(streamDir, 0755); err != nil {
		utils.LogError("創建流目錄失敗: %v", err)
		return
	}

	// 根據流類型選擇不同的 FFmpeg 參數
	var args []string

	if streamProcess.Config.Type == "rtmp" {
		// RTMP 輸入參數
		args = []string{
			"-i", streamProcess.Config.URL,
			"-c:v", "libx264",
			"-preset", "fast", // 改為 fast，平衡性能
			"-crf", "23", // 控制品質
			"-c:a", "aac",
			"-b:a", "128k",
			"-maxrate", "2M", // 限制最大比特率
			"-bufsize", "4M",
		}
	} else if streamProcess.Config.Type == "mp4" {
		// MP4 文件輸入參數 - 優化效能
		args = []string{
			"-i", streamProcess.Config.URL,
			"-c:v", "libx264",
			"-preset", "ultrafast", // 改為 ultrafast，大幅降低 CPU 使用
			"-crf", "28", // 稍微降低品質以節省 CPU
			"-vf", "scale=1280:720", // 限制解析度為 720p
			"-c:a", "aac",
			"-b:a", "96k", // 降低音頻比特率
			"-maxrate", "1M", // 降低最大比特率
			"-bufsize", "2M",
			"-loop", "1", // 循環播放 MP4 文件
		}
	} else {
		// HLS 輸入參數 (通常已經編碼過)
		args = []string{
			"-i", streamProcess.Config.URL,
			"-c", "copy", // 直接複製，不重新編碼
		}
	}

	// 統一的 HLS 輸出參數 (標準延遲，不是 LL-HLS)
	args = append(args, []string{
		"-f", "hls",
		"-hls_time", "2", // 2秒片段，平衡延遲和性能
		"-hls_list_size", "10", // 保留10個片段
		"-hls_flags", "delete_segments+independent_segments",
		"-hls_segment_type", "mpegts",
		"-hls_segment_filename", fmt.Sprintf("%s/segment_%%03d.ts", streamDir),
		"-hls_playlist_type", "vod", // 改為 vod，適合公開流
		fmt.Sprintf("%s/index.m3u8", streamDir),
	}...)

	// 啟動 FFmpeg 進程
	cmd := exec.Command("ffmpeg", args...)
	cmd.Dir = streamDir

	// 設置進程組，便於管理
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	streamProcess.Process = cmd

	if err := cmd.Start(); err != nil {
		utils.LogError("啟動外部流 FFmpeg 失敗: %v", err)
		streamProcess.mu.Lock()
		streamProcess.Running = false
		streamProcess.mu.Unlock()
		return
	}

	utils.LogInfo("✅ 外部流啟動成功: %s", name)

	// 監控進程
	go func() {
		cmd.Wait()
		utils.LogInfo("外部流 %s 已停止", name)

		streamProcess.mu.Lock()
		streamProcess.Running = false
		streamProcess.mu.Unlock()

		// 如果沒有手動停止，檢查是否應該重啟
		select {
		case <-streamProcess.StopChan:
			utils.LogInfo("外部流 %s 手動停止", name)
		default:
			// 檢查流是否仍然啟用
			streamProcess.mu.Lock()
			shouldRestart := streamProcess.Config.Enabled
			streamProcess.mu.Unlock()

			if shouldRestart {
				utils.LogInfo("外部流 %s 意外停止，5秒後重啟...", name)
				time.Sleep(5 * time.Second)
				go sp.startExternalStream(name, streamProcess)
			} else {
				utils.LogInfo("外部流 %s 已停用，不重啟", name)
			}
		}
	}()
}

// StopStream 停止特定流
func (sp *StreamPuller) StopStream(name string) {
	sp.mu.RLock()
	streamProcess, exists := sp.streams[name]
	sp.mu.RUnlock()

	if !exists {
		return
	}

	streamProcess.mu.Lock()
	if !streamProcess.Running {
		streamProcess.mu.Unlock()
		return
	}
	streamProcess.mu.Unlock()

	utils.LogInfo("🛑 停止外部流: %s", name)

	// 發送停止信號
	select {
	case streamProcess.StopChan <- true:
	default:
	}

	// 終止進程
	if streamProcess.Process != nil && streamProcess.Process.Process != nil {
		// 終止整個進程組
		syscall.Kill(-streamProcess.Process.Process.Pid, syscall.SIGTERM)

		// 等待進程結束
		done := make(chan error, 1)
		go func() {
			done <- streamProcess.Process.Wait()
		}()

		select {
		case <-done:
			utils.LogInfo("外部流 %s 已停止", name)
		case <-time.After(5 * time.Second):
			// 強制終止
			syscall.Kill(-streamProcess.Process.Process.Pid, syscall.SIGKILL)
			utils.LogInfo("外部流 %s 強制終止", name)
		}
	}
}

// startHTTPServer 啟動 HTTP 服務器
func (sp *StreamPuller) startHTTPServer() {
	mux := http.NewServeMux()

	// 健康檢查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

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

		fmt.Fprintf(w, `{"status":"healthy","streams":%v}`, status)
	})

	// 流控制 API
	mux.HandleFunc("/api/streams", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "GET":
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

			fmt.Fprintf(w, `{"streams":%v}`, streams)

		case "POST":
			// 啟動流
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			streamName := r.FormValue("name")
			sp.mu.RLock()
			streamProcess, exists := sp.streams[streamName]
			sp.mu.RUnlock()

			if !exists {
				http.Error(w, "Stream not found", http.StatusNotFound)
				return
			}

			go sp.startExternalStream(streamName, streamProcess)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"started"}`))

		case "DELETE":
			// 停止流
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			streamName := r.FormValue("name")
			sp.StopStream(streamName)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"stopped"}`))
		}
	})

	// 靜態文件服務 - 提供 HLS 文件
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 處理 API 路由
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// 讓其他處理器處理 API 請求
			return
		}

		// 處理健康檢查
		if r.URL.Path == "/health" {
			// 讓健康檢查處理器處理
			return
		}

		// 處理靜態文件
		if strings.HasSuffix(r.URL.Path, ".m3u8") || strings.HasSuffix(r.URL.Path, ".ts") {
			// 設置 CORS 頭
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Range")

			// 設置正確的 MIME 類型
			if strings.HasSuffix(r.URL.Path, ".m3u8") {
				w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			} else if strings.HasSuffix(r.URL.Path, ".ts") {
				w.Header().Set("Content-Type", "video/mp2t")
			}

			// 構建文件路徑
			filePath := sp.outputDir + r.URL.Path

			// 檢查文件是否存在
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}

			// 提供文件
			http.ServeFile(w, r, filePath)
			return
		}

		// 其他請求返回 404
		http.NotFound(w, r)
	})

	// 新增：公開流配置管理 API
	mux.HandleFunc("/api/public-streams", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "GET":
			// 獲取所有公開流配置
			streams, err := sp.configService.GetAllStreams()
			if err != nil {
				http.Error(w, "Failed to get streams", http.StatusInternalServerError)
				return
			}

			// 轉換為 JSON 格式
			streamList := make([]map[string]interface{}, 0)
			for _, stream := range streams {
				streamList = append(streamList, map[string]interface{}{
					"id":          stream.ID,
					"name":        stream.Name,
					"title":       stream.Title,
					"description": stream.Description,
					"url":         stream.URL,
					"category":    stream.Category,
					"type":        stream.Type,
					"enabled":     stream.Enabled,
					"created_at":  stream.CreatedAt,
					"updated_at":  stream.UpdatedAt,
				})
			}

			response := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"streams": streamList,
					"total":   len(streamList),
				},
			}

			jsonResponse, err := json.Marshal(response)
			if err != nil {
				http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
				return
			}

			w.Write(jsonResponse)

		case "POST":
			// 創建新的公開流配置
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			// 解析表單數據
			stream := &models.PublicStream{
				Name:        r.FormValue("name"),
				Title:       r.FormValue("title"),
				Description: r.FormValue("description"),
				URL:         r.FormValue("url"),
				Category:    r.FormValue("category"),
				Type:        r.FormValue("type"),
				Enabled:     r.FormValue("enabled") == "true",
			}

			// 驗證必填欄位
			if stream.Name == "" || stream.Title == "" || stream.URL == "" {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
				return
			}

			// 創建流配置
			if err := sp.configService.CreateStream(stream); err != nil {
				http.Error(w, "Failed to create stream", http.StatusInternalServerError)
				return
			}

			// 重新載入配置
			if err := sp.loadStreamsFromDatabase(); err != nil {
				utils.LogError("重新載入配置失敗: %v", err)
			}

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"success":true,"message":"Stream created successfully"}`))

		case "PUT":
			// 更新公開流配置
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			streamName := r.FormValue("name")
			if streamName == "" {
				http.Error(w, "Stream name is required", http.StatusBadRequest)
				return
			}

			// 獲取現有配置
			existingStream, err := sp.configService.GetStreamByName(streamName)
			if err != nil {
				http.Error(w, "Stream not found", http.StatusNotFound)
				return
			}

			// 更新欄位
			if title := r.FormValue("title"); title != "" {
				existingStream.Title = title
			}
			if description := r.FormValue("description"); description != "" {
				existingStream.Description = description
			}
			if url := r.FormValue("url"); url != "" {
				existingStream.URL = url
			}
			if category := r.FormValue("category"); category != "" {
				existingStream.Category = category
			}
			if streamType := r.FormValue("type"); streamType != "" {
				existingStream.Type = streamType
			}
			if enabled := r.FormValue("enabled"); enabled != "" {
				existingStream.Enabled = enabled == "true"
			}

			// 更新配置
			if err := sp.configService.UpdateStream(existingStream); err != nil {
				http.Error(w, "Failed to update stream", http.StatusInternalServerError)
				return
			}

			// 重新載入配置
			if err := sp.loadStreamsFromDatabase(); err != nil {
				utils.LogError("重新載入配置失敗: %v", err)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true,"message":"Stream updated successfully"}`))

		case "DELETE":
			// 刪除公開流配置
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			streamName := r.FormValue("name")
			if streamName == "" {
				http.Error(w, "Stream name is required", http.StatusBadRequest)
				return
			}

			// 先停止流
			sp.StopStream(streamName)

			// 獲取流配置
			stream, err := sp.configService.GetStreamByName(streamName)
			if err != nil {
				http.Error(w, "Stream not found", http.StatusNotFound)
				return
			}

			// 刪除配置
			if err := sp.configService.DeleteStream(stream.ID); err != nil {
				http.Error(w, "Failed to delete stream", http.StatusInternalServerError)
				return
			}

			// 重新載入配置
			if err := sp.loadStreamsFromDatabase(); err != nil {
				utils.LogError("重新載入配置失敗: %v", err)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true,"message":"Stream deleted successfully"}`))
		}
	})

	// 靜態文件服務 (HLS 播放)
	mux.Handle("/streams/", http.StripPrefix("/streams/", http.FileServer(http.Dir(sp.outputDir))))

	addr := fmt.Sprintf(":%d", sp.httpPort)
	utils.LogInfo("🌐 HTTP 服務器啟動在 %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		utils.LogError("HTTP 服務器啟動失敗: %v", err)
	}
}

// Stop 停止所有服務
func (sp *StreamPuller) Stop() {
	utils.LogInfo("🛑 停止優化流拉取服務...")

	sp.mu.RLock()
	for name := range sp.streams {
		sp.StopStream(name)
	}
	sp.mu.RUnlock()

	utils.LogInfo("✅ 服務已停止")
}

func main() {
	var (
		outputDir = flag.String("output", "/tmp/public_streams", "HLS 輸出目錄")
		httpPort  = flag.Int("port", 8081, "HTTP 服務端口")
		dbHost    = flag.String("db-host", "localhost", "資料庫主機")
		dbPort    = flag.Int("db-port", 5432, "資料庫端口")
		dbUser    = flag.String("db-user", "stream_user", "資料庫用戶")
		dbPass    = flag.String("db-pass", "stream_password", "資料庫密碼")
		dbName    = flag.String("db-name", "stream_demo", "資料庫名稱")
		showHelp  = flag.Bool("help", false, "顯示幫助信息")
	)
	flag.Parse()

	if *showHelp {
		fmt.Println("🎬 Optimized Stream Puller - 外部流拉取服務")
		fmt.Println("")
		fmt.Println("用法:")
		fmt.Printf("  %s [選項]\n", os.Args[0])
		fmt.Println("")
		fmt.Println("選項:")
		fmt.Println("  -output string")
		fmt.Println("        HLS 輸出目錄 (默認: /tmp/public_streams)")
		fmt.Println("  -port int")
		fmt.Println("        HTTP 服務端口 (默認: 8081)")
		fmt.Println("  -db-host string")
		fmt.Println("        資料庫主機 (默認: localhost)")
		fmt.Println("  -db-port int")
		fmt.Println("        資料庫端口 (默認: 5432)")
		fmt.Println("  -db-user string")
		fmt.Println("        資料庫用戶 (默認: stream_user)")
		fmt.Println("  -db-pass string")
		fmt.Println("        資料庫密碼 (默認: stream_password)")
		fmt.Println("  -db-name string")
		fmt.Println("        資料庫名稱 (默認: stream_demo)")
		fmt.Println("  -help")
		fmt.Println("        顯示幫助信息")
		fmt.Println("")
		return
	}

	// 初始化日誌
	utils.InitLogger()

	// 連接資料庫
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPass, *dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		utils.LogFatal("連接資料庫失敗: %v", err)
	}

	// 自動遷移資料表
	if err := db.AutoMigrate(&models.PublicStream{}); err != nil {
		utils.LogFatal("資料庫遷移失敗: %v", err)
	}

	// 創建優化拉流器
	puller := NewStreamPuller(*outputDir, *httpPort, db)

	// 啟動服務
	if err := puller.Start(); err != nil {
		utils.LogFatal("啟動服務失敗: %v", err)
	}

	// 設置信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 監控服務狀態
	go func() {
		ticker := time.NewTicker(60 * time.Second) // 改為60秒檢查一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				utils.LogInfo("📊 優化服務運行中...")
			}
		}
	}()

	// 等待信號
	<-sigChan
	utils.LogInfo("🛑 收到停止信號...")
	puller.Stop()
}
