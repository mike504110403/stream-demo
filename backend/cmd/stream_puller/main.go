package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"stream-demo/backend/utils"
)

// StreamConfig 流配置
type StreamConfig struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Category    string `json:"category"`
}

// StreamPuller 流拉取器
type StreamPuller struct {
	streams   map[string]*StreamConfig
	outputDir string
	httpPort  int
}

// NewStreamPuller 創建流拉取器
func NewStreamPuller(outputDir string, httpPort int) *StreamPuller {
	return &StreamPuller{
		streams:   make(map[string]*StreamConfig),
		outputDir: outputDir,
		httpPort:  httpPort,
	}
}

// AddStream 添加流配置
func (sp *StreamPuller) AddStream(config StreamConfig) {
	sp.streams[config.Name] = &config
}

// Start 啟動拉流服務
func (sp *StreamPuller) Start() error {
	utils.LogInfo("🎬 啟動流拉取服務...")
	utils.LogInfo("📁 輸出目錄: %s", sp.outputDir)
	utils.LogInfo("🌐 HTTP 端口: %d", sp.httpPort)

	// 創建輸出目錄
	if err := os.MkdirAll(sp.outputDir, 0755); err != nil {
		return fmt.Errorf("創建輸出目錄失敗: %w", err)
	}

	// 啟動所有流
	for name, config := range sp.streams {
		go sp.startStream(name, config)
	}

	// 啟動 HTTP 服務器
	go sp.startHTTPServer()

	utils.LogInfo("✅ 流拉取服務啟動成功")
	return nil
}

// startStream 啟動單個流
func (sp *StreamPuller) startStream(name string, config *StreamConfig) {
	utils.LogInfo("📺 啟動流: %s (%s)", name, config.Title)

	streamDir := fmt.Sprintf("%s/%s", sp.outputDir, name)
	if err := os.MkdirAll(streamDir, 0755); err != nil {
		utils.LogError("創建流目錄失敗: %v", err)
		return
	}

	// 同時生成 HLS 和 RTMP 流
	args := []string{
		"-i", config.URL,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-c:a", "aac",
		"-b:a", "128k",
		// HLS 輸出
		"-f", "hls",
		"-hls_time", "2", // 減少到 2 秒，降低延遲
		"-hls_list_size", "5", // 減少片段數量
		"-hls_flags", "delete_segments",
		"-hls_segment_filename", fmt.Sprintf("%s/segment_%%03d.ts", streamDir),
		fmt.Sprintf("%s/index.m3u8", streamDir),
		// RTMP 輸出
		"-f", "flv",
		fmt.Sprintf("rtmp://localhost:1935/live/%s", name),
	}

	// 啟動 FFmpeg 進程
	cmd := exec.Command("ffmpeg", args...)
	cmd.Dir = streamDir

	if err := cmd.Start(); err != nil {
		utils.LogError("啟動 FFmpeg 失敗: %v", err)
		return
	}

	utils.LogInfo("✅ 流 %s 啟動成功 (HLS + RTMP)", name)

	// 監控進程
	go func() {
		cmd.Wait()
		utils.LogInfo("流 %s 已停止，嘗試重啟...", name)
		// 重啟流
		time.Sleep(5 * time.Second)
		sp.startStream(name, config)
	}()
}

// startHTTPServer 啟動 HTTP 服務器
func (sp *StreamPuller) startHTTPServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 設置 CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 提供 HLS 文件
		http.FileServer(http.Dir(sp.outputDir)).ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", sp.httpPort)
	utils.LogInfo("🌐 HTTP 服務器啟動在 %s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		utils.LogError("HTTP 服務器啟動失敗: %v", err)
	}
}

// Stop 停止服務
func (sp *StreamPuller) Stop() {
	utils.LogInfo("🛑 停止流拉取服務...")

	// 停止所有 FFmpeg 進程
	cmd := exec.Command("pkill", "-f", "ffmpeg.*"+sp.outputDir)
	cmd.Run()

	utils.LogInfo("✅ 服務已停止")
}

func main() {
	// 命令行參數
	var (
		outputDir = flag.String("output", "/tmp/public_streams", "HLS 輸出目錄")
		httpPort  = flag.Int("port", 8081, "HTTP 服務端口")
		showHelp  = flag.Bool("help", false, "顯示幫助信息")
	)
	flag.Parse()

	if *showHelp {
		fmt.Println("🎬 Stream Puller - 獨立流拉取服務")
		fmt.Println("")
		fmt.Println("用法:")
		fmt.Printf("  %s [選項]\n", os.Args[0])
		fmt.Println("")
		fmt.Println("選項:")
		fmt.Println("  -output string")
		fmt.Println("        HLS 輸出目錄 (默認: /tmp/public_streams)")
		fmt.Println("  -port int")
		fmt.Println("        HTTP 服務端口 (默認: 8081)")
		fmt.Println("  -help")
		fmt.Println("        顯示幫助信息")
		fmt.Println("")
		return
	}

	// 初始化日誌
	utils.InitLogger()

	// 創建拉流器
	puller := NewStreamPuller(*outputDir, *httpPort)

	// 配置流
	puller.AddStream(StreamConfig{
		Name:        "tears_of_steel",
		Title:       "Tears of Steel",
		Description: "Unified Streaming 測試影片 - 科幻短片",
		URL:         "https://demo.unified-streaming.com/k8s/features/stable/video/tears-of-steel/tears-of-steel.ism/.m3u8",
		Category:    "demo",
	})

	puller.AddStream(StreamConfig{
		Name:        "mux_test",
		Title:       "Mux 測試流",
		Description: "Mux 提供的測試 HLS 流",
		URL:         "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8",
		Category:    "demo",
	})

	// 啟動服務
	if err := puller.Start(); err != nil {
		utils.LogFatal("啟動服務失敗: %v", err)
	}

	// 設置信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 監控服務狀態
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				utils.LogInfo("📊 服務運行中...")
			}
		}
	}()

	// 等待信號
	<-sigChan
	utils.LogInfo("🛑 收到停止信號...")
	puller.Stop()
}
