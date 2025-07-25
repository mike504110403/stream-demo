package services

import (
	"errors"
	"fmt"
	"path/filepath"
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/dto"
	"stream-demo/backend/pkg/media"
	"stream-demo/backend/pkg/storage"
	postgresqlRepo "stream-demo/backend/repositories/postgresql"
	"strings"
	"time"
)

// VideoService 影片服務
type VideoService struct {
	Conf                *config.Config
	Repo                *postgresqlRepo.PostgreSQLRepo
	RepoSlave           *postgresqlRepo.PostgreSQLRepo
	S3Storage           *storage.S3Storage
	MediaConvertService *media.MediaConvertService
	FFmpegService       *media.FFmpegService // 新增 FFmpeg 服務
}

// NewVideoService 創建影片服務實例
func NewVideoService(conf *config.Config) *VideoService {
	// 初始化S3儲存
	s3Config := storage.S3Config{
		AccessKey: conf.Storage.S3.AccessKey,
		SecretKey: conf.Storage.S3.SecretKey,
		Region:    conf.Storage.S3.Region,
		Bucket:    conf.Storage.S3.Bucket,
		Endpoint:  conf.Storage.S3.Endpoint,
		CDNDomain: conf.Storage.S3.CDNDomain,
	}

	s3Storage, err := storage.NewS3Storage(s3Config)
	if err != nil {
		// 處理S3初始化錯誤
		s3Storage = nil
	}

	// 初始化MediaConvert服務
	mediaConvertConfig := media.MediaConvertConfig{
		Region:       conf.MediaConvert.Region,
		Endpoint:     conf.MediaConvert.Endpoint,
		AccessKey:    conf.Storage.S3.AccessKey,
		SecretKey:    conf.Storage.S3.SecretKey,
		RoleArn:      conf.MediaConvert.RoleArn,
		OutputBucket: conf.MediaConvert.OutputBucket,
	}

	mediaConvertService, err := media.NewMediaConvertService(mediaConvertConfig)
	if err != nil {
		// 處理MediaConvert初始化錯誤
		mediaConvertService = nil
	}

	// 初始化 FFmpeg 轉碼服務
	var ffmpegService *media.FFmpegService
	if conf.Transcode.Type == "ffmpeg" && conf.Transcode.FFmpeg.Enabled {
		ffmpegConfig := media.FFmpegConfig{
			ContainerName: conf.Transcode.FFmpeg.ContainerName,
			Enabled:       conf.Transcode.FFmpeg.Enabled,
		}
		ffmpegService = media.NewFFmpegService(ffmpegConfig)
	}

	return &VideoService{
		Conf:                conf,
		Repo:                postgresqlRepo.NewPostgreSQLRepo(conf.DB["master"]),
		RepoSlave:           postgresqlRepo.NewPostgreSQLRepo(conf.DB["slave"]),
		S3Storage:           s3Storage,
		MediaConvertService: mediaConvertService,
		FFmpegService:       ffmpegService, // 添加 FFmpeg 服務
	}
}

// GenerateUploadURL 生成上傳URL
func (s *VideoService) GenerateUploadURL(userID uint, filename string, fileSize int64) (*storage.PresignedUploadURL, error) {
	if s.S3Storage == nil {
		return nil, errors.New("S3服務未初始化")
	}

	// 檢查檔案格式
	ext := filepath.Ext(filename)
	if !s.isValidVideoFormat(ext) {
		return nil, errors.New("不支援的影片格式")
	}

	// 檢查檔案大小
	maxSize := int64(s.Conf.Video.MaxFileSize)
	if fileSize > maxSize {
		return nil, fmt.Errorf("檔案大小超過限制 (%d bytes)", maxSize)
	}

	return s.S3Storage.GeneratePresignedUploadURL(userID, ext, fileSize)
}

// CreateVideoRecord 創建影片記錄
func (s *VideoService) CreateVideoRecord(userID uint, title, description, s3Key string) (*dto.VideoDTO, error) {
	// 檢查用戶是否存在
	user, err := s.RepoSlave.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 分離式上傳：第一階段只創建記錄，不檢查檔案
	// 檔案資訊檢查移到 ConfirmUploadAndStartProcessing 方法中
	video := &models.Video{
		Title:              title,
		Description:        description,
		UserID:             userID,
		OriginalKey:        s3Key,
		OriginalURL:        "",                      // 暫時為空，確認上傳後設置
		FileSize:           0,                       // 暫時為0，確認上傳後設置
		OriginalFormat:     filepath.Ext(s3Key)[1:], // 從檔名獲取格式
		Status:             "uploading",
		ProcessingProgress: 0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.Repo.CreateVideo(video); err != nil {
		return nil, err
	}

	// 轉換為 DTO
	return &dto.VideoDTO{
		ID:                 video.ID,
		Title:              video.Title,
		Description:        video.Description,
		UserID:             video.UserID,
		Username:           user.Username,
		OriginalURL:        video.OriginalURL,
		ThumbnailURL:       video.ThumbnailURL,
		HLSMasterURL:       video.HLSMasterURL,
		Status:             video.Status,
		ProcessingProgress: video.ProcessingProgress,
		Duration:           video.Duration,
		FileSize:           video.FileSize,
		Views:              video.Views,
		Likes:              video.Likes,
		CreatedAt:          video.CreatedAt,
		UpdatedAt:          video.UpdatedAt,
	}, nil
}

// ConfirmUploadAndStartProcessing 確認上傳完成並開始處理
func (s *VideoService) ConfirmUploadAndStartProcessing(videoID uint) error {
	return s.ConfirmUploadAndStartProcessingWithKey(videoID, "")
}

// ConfirmUploadAndStartProcessingWithKey 使用指定的 S3 Key 確認上傳完成並開始處理
func (s *VideoService) ConfirmUploadAndStartProcessingWithKey(videoID uint, s3Key string) error {
	// 檢查 S3 服務是否可用
	if s.S3Storage == nil {
		return errors.New("S3 服務未初始化，請檢查 S3 配置")
	}

	video, err := s.Repo.FindVideoByID(videoID)
	if err != nil {
		return err
	}

	// 如果提供了新的 S3 Key，更新影片記錄
	actualKey := video.OriginalKey
	if s3Key != "" {
		actualKey = s3Key
		video.OriginalKey = s3Key // 更新到正確的 Key
	}

	// 檢查檔案是否真的存在於S3
	exists, err := s.S3Storage.CheckFileExists(actualKey)
	if err != nil || !exists {
		// 更新狀態為失敗
		video.Status = "failed"
		video.ErrorMessage = "檔案上傳失敗: 檔案不存在於 S3"
		s.Repo.UpdateVideo(video)
		return errors.New("檔案上傳失敗: 檔案不存在於 S3")
	}

	// 獲取檔案資訊
	fileInfo, err := s.S3Storage.GetFileInfo(actualKey)
	if err != nil {
		video.Status = "failed"
		video.ErrorMessage = "無法獲取檔案資訊: " + err.Error()
		s.Repo.UpdateVideo(video)
		return fmt.Errorf("無法獲取檔案資訊: %w", err)
	}

	// 更新影片資訊
	fileSize := *fileInfo.ContentLength
	video.FileSize = fileSize
	video.OriginalFormat = filepath.Ext(actualKey)[1:]        // 去掉點號
	video.OriginalURL = s.S3Storage.GenerateCDNURL(actualKey) // 設置 CDN URL

	fmt.Printf("🎬 影片資訊更新 - ID: %d, 大小: %d bytes, 格式: %s\n", video.ID, fileSize, video.OriginalFormat)

	// 一律進行轉碼，不考慮檔案大小
	fmt.Printf("🔄 開始轉碼流程 - 檔案大小: %d bytes (一律轉碼)\n", fileSize)

	video.Status = "processing"
	video.ProcessingProgress = 10

	if err := s.Repo.UpdateVideo(video); err != nil {
		return err
	}

	// 開始轉碼
	go s.startTranscoding(video)

	return nil
}

// ConfirmUploadOnly 只確認上傳，不檢查轉碼狀態
func (s *VideoService) ConfirmUploadOnly(videoID uint, s3Key string) error {
	// 檢查 S3 服務是否可用
	if s.S3Storage == nil {
		return errors.New("S3 服務未初始化，請檢查 S3 配置")
	}

	video, err := s.Repo.FindVideoByID(videoID)
	if err != nil {
		return err
	}

	// 如果提供了新的 S3 Key，更新影片記錄
	actualKey := video.OriginalKey
	if s3Key != "" {
		actualKey = s3Key
		video.OriginalKey = s3Key // 更新到正確的 Key
	}

	// 檢查檔案是否真的存在於S3
	exists, err := s.S3Storage.CheckFileExists(actualKey)
	if err != nil || !exists {
		// 更新狀態為失敗
		video.Status = "failed"
		video.ErrorMessage = "檔案上傳失敗: 檔案不存在於 S3"
		s.Repo.UpdateVideo(video)
		return errors.New("檔案上傳失敗: 檔案不存在於 S3")
	}

	// 獲取檔案資訊
	fileInfo, err := s.S3Storage.GetFileInfo(actualKey)
	if err != nil {
		video.Status = "failed"
		video.ErrorMessage = "無法獲取檔案資訊: " + err.Error()
		s.Repo.UpdateVideo(video)
		return fmt.Errorf("無法獲取檔案資訊: %w", err)
	}

	// 更新影片基本資訊
	fileSize := *fileInfo.ContentLength
	video.FileSize = fileSize
	video.OriginalFormat = filepath.Ext(actualKey)[1:]        // 去掉點號
	video.OriginalURL = s.S3Storage.GenerateCDNURL(actualKey) // 設置 CDN URL
	video.Status = "uploading"
	video.ProcessingProgress = 0

	fmt.Printf("✅ 影片上傳確認成功 - VideoID: %d, 大小: %d bytes\n", video.ID, fileSize)

	// 保存到資料庫
	if err := s.Repo.UpdateVideo(video); err != nil {
		return err
	}

	// 啟動轉碼（異步）
	go s.startTranscoding(video)

	return nil
}

// startTranscoding 開始轉碼（異步）
func (s *VideoService) startTranscoding(video *models.Video) {
	fmt.Printf("🎯 開始轉碼 - VideoID: %d\n", video.ID)

	// 更新狀態
	video.Status = "transcoding"
	video.ProcessingProgress = 20
	s.Repo.UpdateVideo(video)

	// 使用 FFmpeg 轉碼（簡化邏輯）
	if s.FFmpegService != nil {
		s.startFFmpegTranscoding(video)
	} else {
		fmt.Printf("❌ FFmpeg 服務不可用 - VideoID: %d\n", video.ID)
		video.Status = "failed"
		video.ErrorMessage = "FFmpeg 服務不可用"
		s.Repo.UpdateVideo(video)
	}
}

// startFFmpegTranscoding 使用 FFmpeg 開始轉碼
func (s *VideoService) startFFmpegTranscoding(video *models.Video) {
	fmt.Printf("🎬 創建 FFmpeg 轉碼任務 - VideoID: %d, InputKey: %s\n", video.ID, video.OriginalKey)

	// 創建 FFmpeg 轉碼任務
	job, err := s.FFmpegService.CreateHLSTranscodeJob(
		video.OriginalKey,
		video.UserID,
		video.ID,
	)
	if err != nil {
		fmt.Printf("❌ FFmpeg 轉碼任務創建失敗 - VideoID: %d, Error: %s\n", video.ID, err.Error())
		// 轉碼失敗
		video.Status = "failed"
		video.ErrorMessage = "FFmpeg 轉碼任務創建失敗: " + err.Error()
		s.Repo.UpdateVideo(video)
		return
	}

	fmt.Printf("✅ FFmpeg 轉碼任務創建成功 - VideoID: %d, JobID: %s\n", video.ID, job.JobID)

	// 監控 FFmpeg 轉碼任務
	s.monitorFFmpegTranscodingJob(video, job)
}

// startMediaConvertTranscoding 使用 AWS MediaConvert 開始轉碼
func (s *VideoService) startMediaConvertTranscoding(video *models.Video) {
	// 創建轉碼任務
	job, err := s.MediaConvertService.CreateHLSTranscodeJob(
		video.OriginalKey,
		video.UserID,
		video.ID,
	)
	if err != nil {
		// 轉碼失敗
		video.Status = "failed"
		video.ErrorMessage = "轉碼任務創建失敗: " + err.Error()
		s.Repo.UpdateVideo(video)
		return
	}

	// 輪詢轉碼狀態
	s.monitorTranscodingJob(video, job)
}

// monitorTranscodingJob 監控轉碼任務
func (s *VideoService) monitorTranscodingJob(video *models.Video, job *media.TranscodeJob) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒檢查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			jobStatus, err := s.MediaConvertService.GetJobStatus(job.JobID)
			if err != nil {
				continue
			}

			switch *jobStatus.Status {
			case "SUBMITTED", "PROGRESSING":
				// 更新進度
				progress := 20
				if jobStatus.JobPercentComplete != nil {
					progress = 20 + int(float64(*jobStatus.JobPercentComplete)*0.7) // 20-90%
				}
				video.ProcessingProgress = progress
				s.Repo.UpdateVideo(video)

			case "COMPLETE":
				// 轉碼完成
				s.handleTranscodingComplete(video, job)
				return

			case "ERROR", "CANCELED":
				// 轉碼失敗
				video.Status = "failed"
				video.ErrorMessage = "轉碼失敗"
				if jobStatus.ErrorMessage != nil {
					video.ErrorMessage = *jobStatus.ErrorMessage
				}
				s.Repo.UpdateVideo(video)
				return
			}
		}
	}
}

// handleTranscodingComplete 處理轉碼完成
func (s *VideoService) handleTranscodingComplete(video *models.Video, job *media.TranscodeJob) {
	// 更新HLS URL（使用處理後桶）
	video.HLSMasterURL = s.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/index.m3u8", job.OutputPrefix))
	video.HLSKey = job.OutputPrefix
	video.Status = "ready"
	video.ProcessingProgress = 100

	// 移除品質記錄創建，避免重複創建
	// 品質記錄由 transcode_worker.go 統一處理

	// 設置縮圖URL（使用處理後桶）
	video.ThumbnailURL = s.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/thumbnails/thumb.0000001.jpg", job.OutputPrefix))

	s.Repo.UpdateVideo(video)
}

// monitorFFmpegTranscodingJob 監控 FFmpeg 轉碼任務
func (s *VideoService) monitorFFmpegTranscodingJob(video *models.Video, job *media.FFmpegTranscodeJob) {
	ticker := time.NewTicker(10 * time.Second) // 每10秒檢查一次
	defer ticker.Stop()

	timeout := time.After(30 * time.Minute) // 30分鐘超時

	for {
		select {
		case <-ticker.C:
			// 檢查任務狀態
			jobStatus, err := s.FFmpegService.GetJobStatus(job.JobID)
			if err != nil {
				// 如果任務不存在，可能是已經完成並被清理了
				// 嘗試檢查轉碼報告來確認狀態
				report, reportErr := s.FFmpegService.GetTranscodeReport(job.OutputPrefix)
				if reportErr == nil && report.Status == "completed" {
					// 轉碼已完成，處理結果
					s.handleFFmpegTranscodingComplete(video, job)
					return
				}
				continue
			}

			switch jobStatus.Status {
			case "SUBMITTED", "PROGRESSING":
				// 更新進度
				video.ProcessingProgress = 50 // FFmpeg 轉碼中
				s.Repo.UpdateVideo(video)

			case "COMPLETE":
				// 轉碼完成，處理結果
				s.handleFFmpegTranscodingComplete(video, job)
				return

			case "ERROR":
				// 轉碼失敗
				video.Status = "failed"
				video.ErrorMessage = "FFmpeg 轉碼失敗: " + jobStatus.Error
				s.Repo.UpdateVideo(video)
				return
			}

		case <-timeout:
			// 轉碼超時
			video.Status = "failed"
			video.ErrorMessage = "轉碼超時（30分鐘）"
			s.Repo.UpdateVideo(video)
			return
		}
	}
}

// handleFFmpegTranscodingComplete 處理 FFmpeg 轉碼完成
func (s *VideoService) handleFFmpegTranscodingComplete(video *models.Video, job *media.FFmpegTranscodeJob) {
	fmt.Printf("🎉 處理 FFmpeg 轉碼完成 - VideoID: %d, JobID: %s\n", video.ID, job.JobID)

	// 更新 HLS 和 MP4 URL（文件在處理後桶中）
	video.HLSMasterURL = s.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/hls/index.m3u8", job.OutputPrefix))
	video.HLSKey = fmt.Sprintf("%s/hls", job.OutputPrefix)

	// 設置 MP4 轉碼版本 URL（文件在處理後桶中）
	video.MP4URL = s.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/video.mp4", job.OutputPrefix))
	video.MP4Key = fmt.Sprintf("%s/video.mp4", job.OutputPrefix)

	video.Status = "ready"
	video.ProcessingProgress = 100

	// 移除品質記錄創建，避免重複創建
	// 品質記錄由 transcode_worker.go 統一處理

	// 設置縮圖URL（使用 640x480 作為主縮圖，文件在處理後桶中）
	video.ThumbnailURL = s.S3Storage.GenerateProcessedCDNURL(fmt.Sprintf("%s/thumbnails/thumb_640x480.jpg", job.OutputPrefix))

	fmt.Printf("✅ 影片轉碼完成 [VideoID: %d, JobID: %s]\n", video.ID, job.JobID)
}

// isValidVideoFormat 檢查是否為有效的影片格式
func (s *VideoService) isValidVideoFormat(ext string) bool {
	allowedFormats := s.Conf.Video.AllowedFormats
	for _, format := range allowedFormats {
		if ext == "."+format {
			return true
		}
	}
	return false
}

// GetVideoByID 根據 ID 獲取影片（詳情視圖，只返回轉碼完成的影片）
func (s *VideoService) GetVideoByID(id uint) (*dto.VideoDTO, error) {
	video, err := s.RepoSlave.FindVideoByID(id)
	if err != nil {
		return nil, err
	}

	// 檢查影片是否已轉碼完成
	if video.Status != "ready" {
		return nil, fmt.Errorf("影片尚未轉碼完成，當前狀態: %s", video.Status)
	}

	// 獲取用戶資訊
	user, err := s.RepoSlave.FindUserByID(video.UserID)
	if err != nil {
		return nil, err
	}

	// 獲取品質資訊
	qualities, _ := s.RepoSlave.FindVideoQualitiesByVideoID(video.ID)

	// 轉換為 DTO
	qualityDTOs := make([]dto.VideoQualityDTO, len(qualities))
	for i, quality := range qualities {
		qualityDTOs[i] = dto.VideoQualityDTO{
			ID:      quality.ID,
			Quality: quality.Quality,
			Width:   quality.Width,
			Height:  quality.Height,
			Bitrate: quality.Bitrate,
			FileURL: quality.FileURL,
			Status:  quality.Status,
		}
	}

	// 為詳情頁面生成完整的播放 URL（優先使用轉碼後的 URL）
	playURL := video.MP4URL
	if playURL == "" {
		playURL = video.HLSMasterURL
	}
	if playURL == "" {
		playURL = video.OriginalURL
	}

	thumbnailURL := video.ThumbnailURL
	if thumbnailURL == "" && video.OriginalKey != "" && s.S3Storage != nil {
		// 可以生成默認縮圖 URL 或保持空白
		// thumbnailURL = s.generateDefaultThumbnailURL(video.OriginalKey)
	}

	return &dto.VideoDTO{
		ID:                 video.ID,
		Title:              video.Title,
		Description:        video.Description,
		UserID:             video.UserID,
		Username:           user.Username,
		OriginalURL:        playURL,            // 優先使用轉碼後的播放 URL
		ThumbnailURL:       thumbnailURL,       // 縮圖 URL
		HLSMasterURL:       video.HLSMasterURL, // HLS 播放列表 URL
		MP4URL:             video.MP4URL,       // MP4 轉碼版本 URL
		Status:             video.Status,
		ProcessingProgress: video.ProcessingProgress,
		Duration:           video.Duration,
		FileSize:           video.FileSize,
		Views:              video.Views,
		Likes:              video.Likes,
		Qualities:          qualityDTOs,
		CreatedAt:          video.CreatedAt,
		UpdatedAt:          video.UpdatedAt,
	}, nil
}

// GetVideos 分頁獲取所有影片（列表視圖，只返回轉碼完成的影片）
func (s *VideoService) GetVideos(offset, limit int) ([]*dto.VideoDTO, int64, error) {
	// 只獲取狀態為 "ready" 的影片（轉碼完成）
	videos, _, err := s.RepoSlave.FindVideosWithPagination(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// 過濾出轉碼完成的影片
	var readyVideos []models.Video
	for _, video := range videos {
		if video.Status == "ready" {
			readyVideos = append(readyVideos, video)
		}
	}

	videoDTOs := make([]*dto.VideoDTO, len(readyVideos))
	for i, video := range readyVideos {
		user, _ := s.RepoSlave.FindUserByID(video.UserID)
		videoDTOs[i] = &dto.VideoDTO{
			ID:                 video.ID,
			Title:              video.Title,
			Description:        video.Description,
			UserID:             video.UserID,
			Username:           user.Username,
			ThumbnailURL:       video.ThumbnailURL, // 縮圖保留，用於顯示
			Status:             video.Status,
			ProcessingProgress: video.ProcessingProgress,
			Duration:           video.Duration,
			FileSize:           video.FileSize,
			Views:              video.Views,
			Likes:              video.Likes,
			CreatedAt:          video.CreatedAt,
			UpdatedAt:          video.UpdatedAt,
			// 移除播放相關 URL：OriginalURL, HLSMasterURL
		}
	}

	return videoDTOs, int64(len(readyVideos)), nil
}

// GetVideosByUserID 根據用戶 ID 獲取影片列表
func (s *VideoService) GetVideosByUserID(userID uint) ([]*dto.VideoDTO, int64, error) {
	videos, err := s.RepoSlave.FindVideoByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 獲取用戶資訊
	user, err := s.RepoSlave.FindUserByID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 轉換為 DTO
	videoDTOs := make([]*dto.VideoDTO, len(videos))
	for i, video := range videos {
		// 為已轉碼完成的影片優先使用轉碼後的 URL
		playURL := video.OriginalURL
		if video.Status == "ready" {
			if video.MP4URL != "" {
				playURL = video.MP4URL
			} else if video.HLSMasterURL != "" {
				playURL = video.HLSMasterURL
			}
		}

		videoDTOs[i] = &dto.VideoDTO{
			ID:           video.ID,
			Title:        video.Title,
			Description:  video.Description,
			UserID:       video.UserID,
			Username:     user.Username,
			OriginalURL:  playURL, // 使用優先級 URL
			ThumbnailURL: video.ThumbnailURL,
			Status:       video.Status,
			Views:        video.Views,
			Likes:        video.Likes,
			CreatedAt:    video.CreatedAt,
			UpdatedAt:    video.UpdatedAt,
		}
	}

	return videoDTOs, int64(len(videos)), nil
}

// SearchVideos 搜尋影片
func (s *VideoService) SearchVideos(query string, offset, limit int) ([]*dto.VideoDTO, int64, error) {
	videos, err := s.RepoSlave.SearchVideo(query)
	if err != nil {
		return nil, 0, err
	}

	// 分頁處理
	total := int64(len(videos))
	end := offset + limit
	if end > len(videos) {
		end = len(videos)
	}
	if offset >= len(videos) {
		return []*dto.VideoDTO{}, total, nil
	}

	paginatedVideos := videos[offset:end]

	// 轉換為 DTO
	videoDTOs := make([]*dto.VideoDTO, len(paginatedVideos))
	for i, video := range paginatedVideos {
		// 獲取用戶資訊
		user, err := s.RepoSlave.FindUserByID(video.UserID)
		if err != nil {
			continue // 跳過無法獲取用戶資訊的影片
		}

		// 為已轉碼完成的影片優先使用轉碼後的 URL
		playURL := video.OriginalURL
		if video.Status == "ready" {
			if video.MP4URL != "" {
				playURL = video.MP4URL
			} else if video.HLSMasterURL != "" {
				playURL = video.HLSMasterURL
			}
		}

		videoDTOs[i] = &dto.VideoDTO{
			ID:           video.ID,
			Title:        video.Title,
			Description:  video.Description,
			UserID:       video.UserID,
			Username:     user.Username,
			OriginalURL:  playURL, // 使用優先級 URL
			ThumbnailURL: video.ThumbnailURL,
			Status:       video.Status,
			Views:        video.Views,
			Likes:        video.Likes,
			CreatedAt:    video.CreatedAt,
			UpdatedAt:    video.UpdatedAt,
		}
	}

	return videoDTOs, total, nil
}

// UpdateVideo 更新影片
func (s *VideoService) UpdateVideo(id uint, title string, description string, videoData *dto.VideoDTO) error {
	video, err := s.Repo.FindVideoByID(id)
	if err != nil {
		return err
	}

	// 更新影片資訊
	if title != "" {
		video.Title = title
	}

	if description != "" {
		video.Description = description
	}

	video.UpdatedAt = time.Now()

	if err := s.Repo.UpdateVideo(video); err != nil {
		return err
	}

	// 獲取用戶資訊
	user, err := s.RepoSlave.FindUserByID(video.UserID)
	if err != nil {
		return err
	}

	// 為已轉碼完成的影片優先使用轉碼後的 URL
	playURL := video.OriginalURL
	if video.Status == "ready" {
		if video.MP4URL != "" {
			playURL = video.MP4URL
		} else if video.HLSMasterURL != "" {
			playURL = video.HLSMasterURL
		}
	}

	// 轉換為 DTO
	videoDTO := &dto.VideoDTO{
		ID:                 video.ID,
		Title:              video.Title,
		Description:        video.Description,
		UserID:             video.UserID,
		Username:           user.Username,
		OriginalURL:        playURL, // 使用優先級 URL
		ThumbnailURL:       video.ThumbnailURL,
		HLSMasterURL:       video.HLSMasterURL,
		Status:             video.Status,
		ProcessingProgress: video.ProcessingProgress,
		Duration:           video.Duration,
		FileSize:           video.FileSize,
		Views:              video.Views,
		Likes:              video.Likes,
		CreatedAt:          video.CreatedAt,
		UpdatedAt:          video.UpdatedAt,
	}
	videoData = videoDTO

	return nil
}

// DeleteVideo 刪除影片
func (s *VideoService) DeleteVideo(id uint) error {
	return s.Repo.DeleteVideo(id)
}

// IncrementViews 增加觀看次數
func (s *VideoService) IncrementViews(id uint) error {
	return s.Repo.IncrementVideoViews(id)
}

// IncrementLikes 增加喜歡次數
func (s *VideoService) IncrementLikes(id uint) error {
	return s.Repo.IncrementVideoLikes(id)
}

// LikeVideo 喜歡影片
func (s *VideoService) LikeVideo(id uint) error {
	return s.Repo.IncrementVideoLikes(id)
}

// shouldTranscodeVideo 檢查是否需要轉碼
func (s *VideoService) shouldTranscodeVideo(fileSize int64) bool {
	// 實現轉碼邏輯，這裡只是簡單的示例
	return fileSize > int64(s.Conf.Video.MinFileSize)
}

// CheckS3Configuration 檢查 S3 配置並提供建議
func (s *VideoService) CheckS3Configuration() error {
	if s.S3Storage == nil {
		suggestions := []string{
			"請檢查 config.local.yaml 中的 S3 配置",
			"確保 access_key 和 secret_key 不是佔位符",
			"可以使用 MinIO 作為本地開發替代方案",
			"或者暫時跳過 S3 檢查進行測試",
		}

		return fmt.Errorf("S3 服務未初始化。建議：\n- %s", strings.Join(suggestions, "\n- "))
	}
	return nil
}
