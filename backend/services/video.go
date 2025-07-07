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
	"time"
)

// VideoService 影片服務
type VideoService struct {
	Conf                *config.Config
	Repo                *postgresqlRepo.PostgreSQLRepo
	RepoSlave           *postgresqlRepo.PostgreSQLRepo
	S3Storage           *storage.S3Storage
	MediaConvertService *media.MediaConvertService
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

	return &VideoService{
		Conf:                conf,
		Repo:                postgresqlRepo.NewPostgreSQLRepo(conf.DB["master"]),
		RepoSlave:           postgresqlRepo.NewPostgreSQLRepo(conf.DB["slave"]),
		S3Storage:           s3Storage,
		MediaConvertService: mediaConvertService,
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

	// 獲取檔案資訊以決定處理策略
	fileInfo, err := s.S3Storage.GetFileInfo(s3Key)
	if err != nil {
		return nil, fmt.Errorf("獲取檔案資訊失敗: %v", err)
	}

	// 根據檔案大小決定處理策略
	fileSize := *fileInfo.ContentLength
	shouldTranscode := s.shouldTranscodeVideo(fileSize)

	// 創建影片記錄
	video := &models.Video{
		Title:              title,
		Description:        description,
		UserID:             userID,
		OriginalKey:        s3Key,
		OriginalURL:        s.S3Storage.GenerateCDNURL(s3Key),
		FileSize:           fileSize,
		OriginalFormat:     filepath.Ext(s3Key)[1:], // 去掉點號
		Status:             "uploading",
		ProcessingProgress: 0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// 如果不需要轉檔，直接設為可播放狀態
	if !shouldTranscode {
		video.Status = "completed"
		video.ProcessingProgress = 100
		// 小檔案可以直接使用原始URL作為播放URL
		video.HLSMasterURL = video.OriginalURL
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
	video, err := s.Repo.FindVideoByID(videoID)
	if err != nil {
		return err
	}

	// 檢查檔案是否真的存在於S3
	exists, err := s.S3Storage.CheckFileExists(video.OriginalKey)
	if err != nil || !exists {
		// 更新狀態為失敗
		video.Status = "failed"
		video.ErrorMessage = "檔案上傳失敗"
		s.Repo.UpdateVideo(video)
		return errors.New("檔案上傳失敗")
	}

	// 獲取檔案資訊
	fileInfo, err := s.S3Storage.GetFileInfo(video.OriginalKey)
	if err != nil {
		return err
	}

	// 更新影片資訊
	fileSize := *fileInfo.ContentLength
	video.FileSize = fileSize
	video.OriginalFormat = filepath.Ext(video.OriginalKey)[1:] // 去掉點號

	// 根據檔案大小決定是否需要轉檔
	if s.shouldTranscodeVideo(fileSize) {
		video.Status = "processing"
		video.ProcessingProgress = 10

		if err := s.Repo.UpdateVideo(video); err != nil {
			return err
		}

		// 開始轉碼
		if s.MediaConvertService != nil {
			go s.startTranscoding(video)
		}
	} else {
		// 小檔案直接標記為完成
		video.Status = "completed"
		video.ProcessingProgress = 100
		video.HLSMasterURL = video.OriginalURL // 使用原始URL

		if err := s.Repo.UpdateVideo(video); err != nil {
			return err
		}
	}

	return nil
}

// startTranscoding 開始轉碼（異步）
func (s *VideoService) startTranscoding(video *models.Video) {
	// 更新狀態
	video.Status = "transcoding"
	video.ProcessingProgress = 20
	s.Repo.UpdateVideo(video)

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
	// 更新HLS URL
	video.HLSMasterURL = s.S3Storage.GenerateCDNURL(fmt.Sprintf("%s/index.m3u8", job.OutputPrefix))
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
			FileURL:   s.S3Storage.GenerateCDNURL(fmt.Sprintf("%s_%s.m3u8", job.OutputPrefix, quality.name)),
			FileKey:   fmt.Sprintf("%s_%s.m3u8", job.OutputPrefix, quality.name),
			Status:    "ready",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		s.Repo.CreateVideoQuality(videoQuality)
	}

	// 設置縮圖URL（取第一張）
	video.ThumbnailURL = s.S3Storage.GenerateCDNURL(fmt.Sprintf("%s/thumbnails/thumb.0000001.jpg", job.OutputPrefix))

	s.Repo.UpdateVideo(video)
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

// GetVideoByID 根據 ID 獲取影片
func (s *VideoService) GetVideoByID(id uint) (*dto.VideoDTO, error) {
	video, err := s.RepoSlave.FindVideoByID(id)
	if err != nil {
		return nil, err
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
		Qualities:          qualityDTOs,
		CreatedAt:          video.CreatedAt,
		UpdatedAt:          video.UpdatedAt,
	}, nil
}

// GetVideos 分頁獲取所有影片
func (s *VideoService) GetVideos(offset, limit int) ([]*dto.VideoDTO, int64, error) {
	videos, total, err := s.RepoSlave.FindVideosWithPagination(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	videoDTOs := make([]*dto.VideoDTO, len(videos))
	for i, video := range videos {
		user, _ := s.RepoSlave.FindUserByID(video.UserID)
		videoDTOs[i] = &dto.VideoDTO{
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
		}
	}

	return videoDTOs, total, nil
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
		videoDTOs[i] = &dto.VideoDTO{
			ID:           video.ID,
			Title:        video.Title,
			Description:  video.Description,
			UserID:       video.UserID,
			Username:     user.Username,
			OriginalURL:  video.OriginalURL,
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

		videoDTOs[i] = &dto.VideoDTO{
			ID:           video.ID,
			Title:        video.Title,
			Description:  video.Description,
			UserID:       video.UserID,
			Username:     user.Username,
			OriginalURL:  video.OriginalURL,
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

	// 轉換為 DTO
	videoDTO := &dto.VideoDTO{
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
