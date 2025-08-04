package services

import (
	"fmt"
	"log"
	"stream-demo/backend/database/models"
	"stream-demo/backend/pkg/media"
	"time"

	"gorm.io/gorm/clause"
)

// TranscodeWorker 背景轉碼工作服務
type TranscodeWorker struct {
	videoService *VideoService
	interval     time.Duration
	stopChan     chan bool
	isRunning    bool
}

// NewTranscodeWorker 創建轉碼工作服務
func NewTranscodeWorker(videoService *VideoService) *TranscodeWorker {
	return &TranscodeWorker{
		videoService: videoService,
		interval:     30 * time.Second, // 每30秒檢查一次
		stopChan:     make(chan bool),
		isRunning:    false,
	}
}

// Start 啟動背景轉碼工作
func (w *TranscodeWorker) Start() {
	if w.isRunning {
		log.Println("轉碼工作服務已在運行中")
		return
	}

	w.isRunning = true
	log.Println("🚀 啟動背景轉碼工作服務")

	// 服務啟動時立即處理一次待轉碼影片
	go w.processPendingVideosOnStartup()

	// 啟動定期檢查
	go w.run()
}

// processPendingVideosOnStartup 服務啟動時處理待轉碼影片
func (w *TranscodeWorker) processPendingVideosOnStartup() {
	log.Println("🔍 服務啟動時檢查待轉碼影片...")

	// 等待一下讓資料庫連接穩定
	time.Sleep(2 * time.Second)

	// 處理所有待轉碼的影片
	w.processPendingVideos()
}

// Stop 停止背景轉碼工作
func (w *TranscodeWorker) Stop() {
	if !w.isRunning {
		return
	}

	w.isRunning = false
	w.stopChan <- true
	log.Println("🛑 停止背景轉碼工作服務")
}

// run 運行轉碼工作循環
func (w *TranscodeWorker) run() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processPendingVideos()
		case <-w.stopChan:
			return
		}
	}
}

// processPendingVideos 處理待轉碼的影片
func (w *TranscodeWorker) processPendingVideos() {
	log.Println("🔍 檢查待轉碼影片...")

	// 使用事務和鎖來避免並發問題
	tx := w.videoService.Repo.GetDB().Begin()
	if tx.Error != nil {
		log.Printf("❌ 開始事務失敗: %v", tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ 事務回滾: %v", r)
		}
	}()

	// 查詢待轉碼的影片（包括重新轉檔的測試）
	var videos []models.Video
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("status IN ?", []string{"uploading", "processing", "transcoding"}).
		Or("status = ? AND updated_at < ?", "ready", time.Now().Add(-24*time.Hour)). // 重新轉檔24小時內的影片（測試用）
		Limit(5).                                                                    // 每次最多處理5個影片
		Find(&videos).Error

	if err != nil {
		tx.Rollback()
		log.Printf("❌ 查詢待轉碼影片失敗: %v", err)
		return
	}

	if len(videos) == 0 {
		tx.Rollback()
		log.Println("📭 沒有待轉碼的影片")
		return
	}

	log.Printf("📋 找到 %d 個待轉碼影片", len(videos))

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ 提交事務失敗: %v", err)
		return
	}

	// 並發處理影片
	for _, video := range videos {
		go w.processVideo(&video)
	}
}

// processVideo 處理單個影片的轉碼
func (w *TranscodeWorker) processVideo(video *models.Video) {
	log.Printf("🎬 開始處理影片 ID: %d, 標題: %s, 狀態: %s", video.ID, video.Title, video.Status)

	// 如果是重新轉檔已完成的影片，跳過驗證
	if video.Status == "ready" {
		log.Printf("🔄 重新轉檔已完成的影片 ID: %d", video.ID)
	} else {
		// 1. 驗證影片檔案
		if err := w.validateVideo(video); err != nil {
			w.markVideoAsFailed(video, err.Error())
			return
		}
	}

	// 2. 更新狀態為轉碼中
	if err := w.updateVideoStatus(video.ID, "transcoding", 20); err != nil {
		log.Printf("❌ 更新影片狀態失敗 - VideoID: %d, Error: %v", video.ID, err)
		return
	}

	// 3. 執行轉碼
	if err := w.executeTranscoding(video); err != nil {
		w.markVideoAsFailed(video, err.Error())
		return
	}

	log.Printf("✅ 影片 ID: %d 轉碼完成", video.ID)
}

// validateVideo 驗證影片檔案
func (w *TranscodeWorker) validateVideo(video *models.Video) error {
	// 檢查檔案 Key
	if video.OriginalKey == "" {
		return fmt.Errorf("沒有原始檔案 Key")
	}

	// 檢查檔案是否存在於 S3
	exists, err := w.videoService.S3Storage.CheckFileExists(video.OriginalKey)
	if err != nil {
		return fmt.Errorf("檢查 S3 檔案失敗: %v", err)
	}
	if !exists {
		return fmt.Errorf("原始檔案不存在於 S3")
	}

	// 獲取檔案資訊
	fileInfo, err := w.videoService.S3Storage.GetFileInfo(video.OriginalKey)
	if err != nil {
		return fmt.Errorf("無法獲取檔案資訊: %v", err)
	}

	// 更新影片基本資訊
	fileSize := *fileInfo.ContentLength
	updates := map[string]interface{}{
		"file_size":       fileSize,
		"original_format": getFileExtension(video.OriginalKey),
		"original_url":    w.videoService.S3Storage.GenerateCDNURL(video.OriginalKey),
		"updated_at":      time.Now(),
	}

	if err := w.videoService.Repo.UpdateVideoFields(video.ID, updates); err != nil {
		return fmt.Errorf("更新影片資訊失敗: %v", err)
	}

	log.Printf("📊 影片 ID: %d 驗證完成 - 大小: %d bytes", video.ID, fileSize)
	return nil
}

// updateVideoStatus 更新影片狀態
func (w *TranscodeWorker) updateVideoStatus(videoID uint, status string, progress int) error {
	updates := map[string]interface{}{
		"status":              status,
		"processing_progress": progress,
		"updated_at":          time.Now(),
	}
	return w.videoService.Repo.UpdateVideoFields(videoID, updates)
}

// executeTranscoding 執行轉碼
func (w *TranscodeWorker) executeTranscoding(video *models.Video) error {
	// 根據配置選擇轉碼方式
	if w.videoService.FFmpegService != nil {
		return w.executeFFmpegTranscoding(video)
	}
	if w.videoService.MediaConvertService != nil {
		return w.executeMediaConvertTranscoding(video)
	}
	return fmt.Errorf("沒有可用的轉碼服務")
}

// executeFFmpegTranscoding 執行 FFmpeg 轉碼
func (w *TranscodeWorker) executeFFmpegTranscoding(video *models.Video) error {
	log.Printf("🎬 開始 FFmpeg 轉碼 - VideoID: %d", video.ID)

	// 使用現有的轉碼邏輯
	w.startFFmpegTranscoding(video)
	return nil
}

// executeMediaConvertTranscoding 執行 MediaConvert 轉碼
func (w *TranscodeWorker) executeMediaConvertTranscoding(video *models.Video) error {
	log.Printf("🎬 開始 MediaConvert 轉碼 - VideoID: %d", video.ID)

	// 使用現有的轉碼邏輯
	w.startMediaConvertTranscoding(video)
	return nil
}

// startTranscoding 啟動轉碼
func (w *TranscodeWorker) startTranscoding(video *models.Video) {
	log.Printf("🎯 啟動轉碼 - 影片 ID: %d", video.ID)

	// 更新狀態
	video.Status = "transcoding"
	video.ProcessingProgress = 20
	w.videoService.Repo.UpdateVideo(video)

	// 根據配置選擇轉碼服務
	if w.videoService.FFmpegService != nil && w.videoService.FFmpegService.IsEnabled() {
		log.Printf("🎬 使用 FFmpeg 轉碼 - 影片 ID: %d", video.ID)
		w.startFFmpegTranscoding(video)
	} else if w.videoService.MediaConvertService != nil {
		log.Printf("☁️ 使用 AWS MediaConvert 轉碼 - 影片 ID: %d", video.ID)
		w.startMediaConvertTranscoding(video)
	} else {
		log.Printf("❌ 沒有可用的轉碼服務 - 影片 ID: %d", video.ID)
		w.markVideoAsFailed(video, "沒有可用的轉碼服務")
	}
}

// startFFmpegTranscoding 使用 FFmpeg 轉碼
func (w *TranscodeWorker) startFFmpegTranscoding(video *models.Video) {
	log.Printf("🎬 創建 FFmpeg 轉碼任務 - 影片 ID: %d, InputKey: %s", video.ID, video.OriginalKey)

	// 創建 FFmpeg 轉碼任務
	job, err := w.videoService.FFmpegService.CreateHLSTranscodeJob(
		video.OriginalKey,
		video.UserID,
		video.ID,
	)
	if err != nil {
		log.Printf("❌ FFmpeg 轉碼任務創建失敗 - 影片 ID: %d, Error: %s", video.ID, err.Error())
		w.markVideoAsFailed(video, "FFmpeg 轉碼任務創建失敗: "+err.Error())
		return
	}

	log.Printf("✅ FFmpeg 轉碼任務創建成功 - 影片 ID: %d, JobID: %s", video.ID, job.JobID)

	// 監控 FFmpeg 轉碼任務
	w.monitorFFmpegTranscodingJob(video, job)
}

// startMediaConvertTranscoding 使用 AWS MediaConvert 轉碼
func (w *TranscodeWorker) startMediaConvertTranscoding(video *models.Video) {
	log.Printf("☁️ 創建 MediaConvert 轉碼任務 - 影片 ID: %d", video.ID)

	// 創建轉碼任務
	job, err := w.videoService.MediaConvertService.CreateHLSTranscodeJob(
		video.OriginalKey,
		video.UserID,
		video.ID,
	)
	if err != nil {
		log.Printf("❌ MediaConvert 轉碼任務創建失敗 - 影片 ID: %d, Error: %s", video.ID, err.Error())
		w.markVideoAsFailed(video, "轉碼任務創建失敗: "+err.Error())
		return
	}

	log.Printf("✅ MediaConvert 轉碼任務創建成功 - 影片 ID: %d, JobID: %s", video.ID, job.JobID)

	// 監控轉碼任務
	w.monitorTranscodingJob(video, job)
}

// monitorFFmpegTranscodingJob 監控 FFmpeg 轉碼任務
func (w *TranscodeWorker) monitorFFmpegTranscodingJob(video *models.Video, job *media.FFmpegTranscodeJob) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Minute)

	for {
		select {
		case <-ticker.C:
			// 檢查任務狀態
			jobStatus, err := w.videoService.FFmpegService.GetJobStatus(job.JobID)
			if err != nil {
				// 如果任務不存在，嘗試檢查轉碼報告
				report, reportErr := w.videoService.FFmpegService.GetTranscodeReport(job.OutputPrefix)
				if reportErr == nil && report.Status == "completed" {
					w.handleFFmpegTranscodingComplete(video, job)
					return
				}
				continue
			}

			switch jobStatus.Status {
			case "SUBMITTED", "PROGRESSING":
				// 更新進度
				video.ProcessingProgress = 50
				w.videoService.Repo.UpdateVideo(video)

			case "COMPLETE":
				// 轉碼完成，處理結果
				w.handleFFmpegTranscodingComplete(video, job)
				return

			case "ERROR":
				// 轉碼失敗
				w.markVideoAsFailed(video, "FFmpeg 轉碼失敗: "+jobStatus.Error)
				return
			}

		case <-timeout:
			// 轉碼超時
			w.markVideoAsFailed(video, "轉碼超時（30分鐘）")
			return
		}
	}
}

// monitorTranscodingJob 監控 MediaConvert 轉碼任務
func (w *TranscodeWorker) monitorTranscodingJob(video *models.Video, job *media.TranscodeJob) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			jobStatus, err := w.videoService.MediaConvertService.GetJobStatus(job.JobID)
			if err != nil {
				continue
			}

			switch *jobStatus.Status {
			case "SUBMITTED", "PROGRESSING":
				// 更新進度
				progress := 20
				if jobStatus.JobPercentComplete != nil {
					progress = 20 + int(float64(*jobStatus.JobPercentComplete)*0.7)
				}
				video.ProcessingProgress = progress
				w.videoService.Repo.UpdateVideo(video)

			case "COMPLETE":
				// 轉碼完成
				w.handleTranscodingComplete(video, job)
				return

			case "ERROR", "CANCELED":
				// 轉碼失敗
				errorMsg := "轉碼失敗"
				if jobStatus.ErrorMessage != nil {
					errorMsg = *jobStatus.ErrorMessage
				}
				w.markVideoAsFailed(video, errorMsg)
				return
			}
		}
	}
}

// handleFFmpegTranscodingComplete 處理 FFmpeg 轉碼完成
func (w *TranscodeWorker) handleFFmpegTranscodingComplete(video *models.Video, job *media.FFmpegTranscodeJob) {
	log.Printf("🎉 處理 FFmpeg 轉碼完成 - 影片 ID: %d, JobID: %s", video.ID, job.JobID)

	// 更新 HLS 和 MP4 URL（使用處理後桶）
	video.HLSMasterURL = w.videoService.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/hls/index.m3u8", job.OutputPrefix))
	video.HLSKey = fmt.Sprintf("%s/hls", job.OutputPrefix)

	// 設置 MP4 轉碼版本 URL（使用處理後桶）
	video.MP4URL = w.videoService.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/video.mp4", job.OutputPrefix))
	video.MP4Key = fmt.Sprintf("%s/video.mp4", job.OutputPrefix)

	video.Status = "ready"
	video.ProcessingProgress = 100

	// 移除事務外的品質記錄創建，避免重複
	// 品質記錄將在事務內統一創建

	// 設置縮圖URL（使用處理後桶）
	video.ThumbnailURL = w.videoService.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/thumbnails/thumb_640x480.jpg", job.OutputPrefix))

	// 使用事務更新影片狀態
	tx := w.videoService.Repo.GetDB().Begin()
	if tx.Error != nil {
		log.Printf("❌ 開始事務失敗 - VideoID: %d, Error: %v", video.ID, tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ 事務回滾 - VideoID: %d", video.ID)
		}
	}()

	// 更新影片狀態
	updates := map[string]interface{}{
		"status":              "ready",
		"processing_progress": 100,
		"hls_master_url":      video.HLSMasterURL,
		"hls_key":             video.HLSKey,
		"mp4_url":             video.MP4URL,
		"mp4_key":             video.MP4Key,
		"thumbnail_url":       video.ThumbnailURL,
		"updated_at":          time.Now(),
	}

	if err := tx.Model(&models.Video{}).Where("id = ?", video.ID).Updates(updates).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ 更新影片狀態失敗 - VideoID: %d, Error: %v", video.ID, err)
		return
	}

	// 創建品質記錄
	qualities := []struct {
		name    string
		width   int
		height  int
		bitrate int
	}{
		{"720p", 1280, 720, 2500000},
		{"480p", 854, 480, 1200000},
		{"360p", 640, 360, 800000},
	}

	for _, quality := range qualities {
		videoQuality := &models.VideoQuality{
			VideoID:   video.ID,
			Quality:   quality.name,
			Width:     quality.width,
			Height:    quality.height,
			Bitrate:   quality.bitrate,
			FileURL:   w.videoService.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/hls/%s/index.m3u8", job.OutputPrefix, quality.name)),
			FileKey:   fmt.Sprintf("%s/hls/%s/index.m3u8", job.OutputPrefix, quality.name),
			Status:    "ready",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := tx.Create(videoQuality).Error; err != nil {
			tx.Rollback()
			log.Printf("❌ 創建品質記錄失敗 - VideoID: %d, Quality: %s, Error: %v", video.ID, quality.name, err)
			return
		}
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ 提交事務失敗 - VideoID: %d, Error: %v", video.ID, err)
		return
	}

	log.Printf("✅ 影片轉碼完成 [VideoID: %d, JobID: %s]", video.ID, job.JobID)
}

// handleTranscodingComplete 處理 MediaConvert 轉碼完成
func (w *TranscodeWorker) handleTranscodingComplete(video *models.Video, job *media.TranscodeJob) {
	log.Printf("🎉 處理 MediaConvert 轉碼完成 - 影片 ID: %d, JobID: %s", video.ID, job.JobID)

	// 更新HLS URL（使用處理後桶）
	video.HLSMasterURL = w.videoService.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/index.m3u8", job.OutputPrefix))
	video.HLSKey = job.OutputPrefix
	video.Status = "ready"
	video.ProcessingProgress = 100

	// 創建品質記錄
	qualities := []struct {
		name    string
		width   int
		height  int
		bitrate int
	}{
		{"720p", 1280, 720, 2500000},
		{"480p", 854, 480, 1200000},
		{"360p", 640, 360, 800000},
	}

	for _, quality := range qualities {
		videoQuality := &models.VideoQuality{
			VideoID:   video.ID,
			Quality:   quality.name,
			Width:     quality.width,
			Height:    quality.height,
			Bitrate:   quality.bitrate,
			FileURL:   w.videoService.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s_%s.m3u8", job.OutputPrefix, quality.name)),
			FileKey:   fmt.Sprintf("%s_%s.m3u8", job.OutputPrefix, quality.name),
			Status:    "ready",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		w.videoService.Repo.CreateVideoQuality(videoQuality)
	}

	// 設置縮圖URL（使用處理後桶）
	video.ThumbnailURL = w.videoService.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/thumbnails/thumb.0000001.jpg", job.OutputPrefix))

	w.videoService.Repo.UpdateVideo(video)

	log.Printf("✅ 影片轉碼完成 [VideoID: %d, JobID: %s]", video.ID, job.JobID)
}

// markVideoAsFailed 標記影片為失敗狀態
func (w *TranscodeWorker) markVideoAsFailed(video *models.Video, errorMessage string) {
	log.Printf("❌ 影片轉碼失敗 - ID: %d, 錯誤: %s", video.ID, errorMessage)

	// 使用事務更新狀態
	tx := w.videoService.Repo.GetDB().Begin()
	if tx.Error != nil {
		log.Printf("❌ 開始事務失敗 - VideoID: %d, Error: %v", video.ID, tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updates := map[string]interface{}{
		"status":        "failed",
		"error_message": errorMessage,
		"updated_at":    time.Now(),
	}

	if err := tx.Model(&models.Video{}).Where("id = ?", video.ID).Updates(updates).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ 更新失敗狀態失敗 - VideoID: %d, Error: %v", video.ID, err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ 提交事務失敗 - VideoID: %d, Error: %v", video.ID, err)
		return
	}

	log.Printf("✅ 影片失敗狀態已更新 - ID: %d", video.ID)
}

// getFileExtension 獲取檔案擴展名
func getFileExtension(filename string) string {
	if len(filename) == 0 {
		return ""
	}

	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i+1:]
		}
	}
	return ""
}
