package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Video 影片模型 - 與 API 服務保持一致
type Video struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	UserID      uint   `json:"user_id" gorm:"not null;index:idx_videos_user_status,priority:1;index:idx_videos_user_created,priority:1"`

	// 原始影片資訊
	OriginalURL  string `json:"original_url" gorm:"size:500;not null"`
	OriginalKey  string `json:"original_key" gorm:"size:500"`
	ThumbnailURL string `json:"thumbnail_url" gorm:"size:500"`

	// HLS串流資訊
	HLSMasterURL string `json:"hls_master_url" gorm:"size:500"`
	HLSKey       string `json:"hls_key" gorm:"size:500"`

	// MP4轉碼版本（網頁播放）
	MP4URL string `json:"mp4_url" gorm:"size:500"`
	MP4Key string `json:"mp4_key" gorm:"size:500"`

	// 影片屬性
	Duration       int    `json:"duration" gorm:"default:0"`      // 秒數
	FileSize       int64  `json:"file_size" gorm:"default:0"`     // 位元組
	OriginalFormat string `json:"original_format" gorm:"size:10"` // mp4, avi等

	// 狀態管理
	Status string `json:"status" gorm:"size:20;not null;index:idx_videos_user_status,priority:2;index:idx_videos_status_created,priority:1"`
	// 狀態: uploading, processing, transcoding, ready, failed
	ProcessingProgress int    `json:"processing_progress" gorm:"default:0"` // 0-100
	ErrorMessage       string `json:"error_message" gorm:"size:500"`

	// 統計資料
	Views     int64     `json:"views" gorm:"default:0"`
	Likes     int64     `json:"likes" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_videos_user_created,priority:2;index:idx_videos_status_created,priority:2"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ConverterService 轉碼服務
type ConverterService struct {
	db          *gorm.DB
	workerCount int
	stopChan    chan bool
	isRunning   bool
}

// NewConverterService 創建轉碼服務
func NewConverterService(db *gorm.DB, workerCount int) *ConverterService {
	return &ConverterService{
		db:          db,
		workerCount: workerCount,
		stopChan:    make(chan bool),
		isRunning:   false,
	}
}

// Start 啟動轉碼服務
func (cs *ConverterService) Start() {
	if cs.isRunning {
		log.Println("轉碼服務已在運行中")
		return
	}

	cs.isRunning = true
	log.Println("🚀 啟動轉碼服務")

	// 啟動多個工作協程
	for i := 0; i < cs.workerCount; i++ {
		go cs.worker(i)
	}

	// 啟動任務監控
	go cs.monitorTasks()
}

// Stop 停止轉碼服務
func (cs *ConverterService) Stop() {
	if !cs.isRunning {
		return
	}

	cs.isRunning = false
	close(cs.stopChan)
	log.Println("🛑 停止轉碼服務")
}

// monitorTasks 監控待處理任務
func (cs *ConverterService) monitorTasks() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("📊 開始監控待轉碼任務...")

	for {
		select {
		case <-ticker.C:
			cs.checkPendingVideos()
		case <-cs.stopChan:
			return
		}
	}
}

// checkPendingVideos 檢查待轉碼影片
func (cs *ConverterService) checkPendingVideos() {
	var videos []Video
	err := cs.db.Where("status = ?", "processing").
		Limit(10).
		Find(&videos).Error

	if err != nil {
		log.Printf("❌ 查詢待轉碼影片失敗: %v", err)
		return
	}

	if len(videos) == 0 {
		return
	}

	log.Printf("📋 找到 %d 個待轉碼影片", len(videos))

	// 將任務發送到工作佇列
	for _, video := range videos {
		cs.processVideo(&video)
	}
}

// worker 工作協程
func (cs *ConverterService) worker(id int) {
	log.Printf("👷 啟動工作協程 %d", id)

	for {
		select {
		case <-cs.stopChan:
			log.Printf("👷 工作協程 %d 停止", id)
			return
		default:
			// 這裡可以實現工作佇列邏輯
			time.Sleep(1 * time.Second)
		}
	}
}

// processVideo 處理單個影片
func (cs *ConverterService) processVideo(video *Video) {
	log.Printf("🎬 開始處理影片 ID: %d, 標題: %s", video.ID, video.Title)

	// 更新狀態為轉碼中
	if err := cs.updateVideoStatus(video.ID, "transcoding", 20); err != nil {
		log.Printf("❌ 更新影片狀態失敗: %v", err)
		return
	}

	// 執行轉碼
	if err := cs.executeTranscoding(video); err != nil {
		cs.markVideoAsFailed(video, err.Error())
		return
	}

	log.Printf("✅ 影片 ID: %d 轉碼完成", video.ID)
}

// executeTranscoding 執行轉碼
func (cs *ConverterService) executeTranscoding(video *Video) error {
	// 生成輸出路徑
	outputPrefix := fmt.Sprintf("videos/processed/%d/%d", video.UserID, video.ID)

	log.Printf("🎬 執行轉碼 - VideoID: %d, InputKey: %s, OutputPrefix: %s",
		video.ID, video.OriginalKey, outputPrefix)

	// 執行 FFmpeg 轉碼腳本
	cmd := exec.Command("/scripts/transcode.sh",
		video.OriginalKey,
		outputPrefix,
		fmt.Sprintf("%d", video.UserID),
		fmt.Sprintf("%d", video.ID),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("轉碼失敗: %v, 輸出: %s", err, string(output))
	}

	log.Printf("✅ 轉碼腳本執行成功: %s", string(output))

	// 更新影片狀態和 URL
	return cs.updateVideoAfterTranscoding(video, outputPrefix)
}

// updateVideoStatus 更新影片狀態
func (cs *ConverterService) updateVideoStatus(videoID uint, status string, progress int) error {
	return cs.db.Model(&Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
		"status":              status,
		"processing_progress": progress,
		"updated_at":          time.Now(),
	}).Error
}

// updateVideoAfterTranscoding 轉碼完成後更新影片資訊
func (cs *ConverterService) updateVideoAfterTranscoding(video *Video, outputPrefix string) error {
	// 從環境變數獲取 CDN 基礎 URL
	cdnBaseURL := os.Getenv("CDN_BASE_URL")
	if cdnBaseURL == "" {
		cdnBaseURL = "http://localhost:9000/stream-demo-processed"
	}

	// 生成各種 URL
	hlsMasterURL := fmt.Sprintf("%s/%s/hls/index.m3u8", cdnBaseURL, outputPrefix)
	mp4URL := fmt.Sprintf("%s/%s/video.mp4", cdnBaseURL, outputPrefix)
	thumbnailURL := fmt.Sprintf("%s/%s/thumbnails/thumb_640x480.jpg", cdnBaseURL, outputPrefix)

	updates := map[string]interface{}{
		"status":              "ready",
		"processing_progress": 100,
		"hls_master_url":      hlsMasterURL,
		"hls_key":             fmt.Sprintf("%s/hls", outputPrefix),
		"mp4_url":             mp4URL,
		"mp4_key":             fmt.Sprintf("%s/video.mp4", outputPrefix),
		"thumbnail_url":       thumbnailURL,
		"updated_at":          time.Now(),
	}

	return cs.db.Model(&Video{}).Where("id = ?", video.ID).Updates(updates).Error
}

// markVideoAsFailed 標記影片為失敗狀態
func (cs *ConverterService) markVideoAsFailed(video *Video, errorMessage string) {
	log.Printf("❌ 影片轉碼失敗 - ID: %d, 錯誤: %s", video.ID, errorMessage)

	// 截斷錯誤訊息，避免超過資料庫欄位長度限制
	if len(errorMessage) > 450 {
		errorMessage = errorMessage[:450] + "..."
	}

	updates := map[string]interface{}{
		"status":        "failed",
		"error_message": errorMessage,
		"updated_at":    time.Now(),
	}

	if err := cs.db.Model(&Video{}).Where("id = ?", video.ID).Updates(updates).Error; err != nil {
		log.Printf("❌ 更新失敗狀態失敗: %v", err)
	}
}

// healthCheck 健康檢查
func healthCheck() error {
	// 檢查資料庫連接
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=postgresql user=stream_user password=stream_password dbname=stream_demo port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("資料庫連接失敗: %v", err)
	}

	// 測試資料庫連接
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("獲取資料庫實例失敗: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("資料庫 ping 失敗: %v", err)
	}

	// 檢查 FFmpeg 是否可用
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("FFmpeg 不可用: %v", err)
	}

	return nil
}

func main() {
	// 檢查命令列參數
	if len(os.Args) > 1 && os.Args[1] == "--health-check" {
		if err := healthCheck(); err != nil {
			log.Printf("❌ 健康檢查失敗: %v", err)
			os.Exit(1)
		}
		log.Println("✅ 健康檢查通過")
		os.Exit(0)
	}

	// 配置日誌
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 資料庫連接
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=postgresql user=stream_user password=stream_password dbname=stream_demo port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ 資料庫連接失敗: %v", err)
	}

	// 自動遷移
	if err := db.AutoMigrate(&Video{}); err != nil {
		log.Fatalf("❌ 資料庫遷移失敗: %v", err)
	}

	// 從環境變數獲取工作協程數量
	workerCount := 3 // 預設值
	if workerCountStr := os.Getenv("WORKER_COUNT"); workerCountStr != "" {
		if count, err := fmt.Sscanf(workerCountStr, "%d", &workerCount); err != nil || count == 0 {
			log.Printf("⚠️ 無效的 WORKER_COUNT: %s，使用預設值 %d", workerCountStr, workerCount)
		}
	}

	log.Printf("🔧 配置: 工作協程數量 = %d", workerCount)

	// 創建轉碼服務
	converterService := NewConverterService(db, workerCount)

	// 啟動服務
	converterService.Start()

	// 等待中斷信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("🛑 收到中斷信號，正在關閉服務...")

	// 優雅關閉
	converterService.Stop()
	log.Println("✅ 服務已關閉")
}
