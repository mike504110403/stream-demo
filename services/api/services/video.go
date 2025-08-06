package services

import (
	"errors"
	"fmt"
	"path/filepath"
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/dto"
	"stream-demo/backend/pkg/storage"
	postgresqlRepo "stream-demo/backend/repositories/postgresql"
	"strings"
	"time"
)

// VideoService 影片服務
type VideoService struct {
	Conf      *config.Config
	Repo      *postgresqlRepo.PostgreSQLRepo
	RepoSlave *postgresqlRepo.PostgreSQLRepo
	S3Storage *storage.S3Storage
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

	return &VideoService{
		Conf:      conf,
		Repo:      postgresqlRepo.NewPostgreSQLRepo(conf.DB["master"]),
		RepoSlave: postgresqlRepo.NewPostgreSQLRepo(conf.DB["slave"]),
		S3Storage: s3Storage,
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
	// 創建影片記錄
	video := &models.Video{
		Title:       title,
		Description: description,
		UserID:      userID,
		OriginalKey: s3Key,
		Status:      "uploading",
	}

	if err := s.Repo.CreateVideo(video); err != nil {
		return nil, fmt.Errorf("創建影片記錄失敗: %v", err)
	}

	// 轉換為 DTO
	videoDTO := &dto.VideoDTO{
		ID:          video.ID,
		Title:       video.Title,
		Description: video.Description,
		UserID:      video.UserID,
		Status:      video.Status,
		CreatedAt:   video.CreatedAt,
		UpdatedAt:   video.UpdatedAt,
	}

	return videoDTO, nil
}

// ConfirmUploadAndStartProcessing 確認上傳並開始處理
func (s *VideoService) ConfirmUploadAndStartProcessing(videoID uint) error {
	// 更新影片狀態為 processing，讓 converter 服務處理轉碼
	return s.Repo.UpdateVideoStatus(videoID, "processing", 0, "")
}

// ConfirmUploadAndStartProcessingWithKey 確認上傳並開始處理（指定 S3 Key）
func (s *VideoService) ConfirmUploadAndStartProcessingWithKey(videoID uint, s3Key string) error {
	// 檢查影片是否存在
	_, err := s.Repo.FindVideoByID(videoID)
	if err != nil {
		return fmt.Errorf("找不到影片記錄: %v", err)
	}

	// 更新影片資訊
	updates := map[string]interface{}{
		"original_key": s3Key,
		"status":       "processing",
		"updated_at":   time.Now(),
	}

	// 如果有 S3 儲存，生成原始 URL
	if s.S3Storage != nil {
		originalURL := s.S3Storage.GenerateCDNURL(s3Key)
		updates["original_url"] = originalURL

		// 獲取檔案資訊
		fileInfo, err := s.S3Storage.GetFileInfo(s3Key)
		if err == nil && fileInfo.ContentLength != nil {
			updates["file_size"] = *fileInfo.ContentLength
			updates["original_format"] = strings.ToLower(strings.TrimPrefix(filepath.Ext(s3Key), "."))
		}
	}

	// 更新資料庫
	if err := s.Repo.UpdateVideoFields(videoID, updates); err != nil {
		return fmt.Errorf("更新影片資訊失敗: %v", err)
	}

	return nil
}

// ConfirmUploadOnly 僅確認上傳（不開始轉碼）
func (s *VideoService) ConfirmUploadOnly(videoID uint, s3Key string) error {
	// 檢查影片是否存在
	_, err := s.Repo.FindVideoByID(videoID)
	if err != nil {
		return fmt.Errorf("找不到影片記錄: %v", err)
	}

	// 更新影片資訊
	updates := map[string]interface{}{
		"original_key": s3Key,
		"status":       "ready", // 直接設為 ready，不進行轉碼
		"updated_at":   time.Now(),
	}

	// 如果有 S3 儲存，生成原始 URL
	if s.S3Storage != nil {
		originalURL := s.S3Storage.GenerateCDNURL(s3Key)
		updates["original_url"] = originalURL

		// 獲取檔案資訊
		fileInfo, err := s.S3Storage.GetFileInfo(s3Key)
		if err == nil && fileInfo.ContentLength != nil {
			updates["file_size"] = *fileInfo.ContentLength
			updates["original_format"] = strings.ToLower(strings.TrimPrefix(filepath.Ext(s3Key), "."))
		}
	}

	// 更新資料庫
	if err := s.Repo.UpdateVideoFields(videoID, updates); err != nil {
		return fmt.Errorf("更新影片資訊失敗: %v", err)
	}

	return nil
}

// isValidVideoFormat 檢查是否為有效的影片格式
func (s *VideoService) isValidVideoFormat(ext string) bool {
	validFormats := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm", ".m4v"}
	ext = strings.ToLower(ext)
	for _, format := range validFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// GetVideoByID 根據ID獲取影片
func (s *VideoService) GetVideoByID(id uint) (*dto.VideoDTO, error) {
	video, err := s.Repo.FindVideoByID(id)
	if err != nil {
		return nil, fmt.Errorf("找不到影片: %v", err)
	}

	// 轉換為 DTO
	videoDTO := &dto.VideoDTO{
		ID:                 video.ID,
		Title:              video.Title,
		Description:        video.Description,
		UserID:             video.UserID,
		OriginalURL:        video.OriginalURL,
		ThumbnailURL:       video.ThumbnailURL,
		HLSMasterURL:       video.HLSMasterURL,
		MP4URL:             video.MP4URL,
		Duration:           video.Duration,
		FileSize:           video.FileSize,
		OriginalFormat:     video.OriginalFormat,
		Status:             video.Status,
		ProcessingProgress: video.ProcessingProgress,
		ErrorMessage:       video.ErrorMessage,
		Views:              video.Views,
		Likes:              video.Likes,
		CreatedAt:          video.CreatedAt,
		UpdatedAt:          video.UpdatedAt,
	}

	// 如果有用戶資訊，添加用戶名
	if video.User != nil {
		videoDTO.Username = video.User.Username
	}

	// 獲取影片品質資訊
	qualities, err := s.Repo.FindVideoQualitiesByVideoID(id)
	if err == nil && len(qualities) > 0 {
		qualityDTOs := make([]dto.VideoQualityDTO, len(qualities))
		for i, quality := range qualities {
			qualityDTOs[i] = dto.VideoQualityDTO{
				ID:       quality.ID,
				Quality:  quality.Quality,
				Width:    quality.Width,
				Height:   quality.Height,
				Bitrate:  quality.Bitrate,
				FileURL:  quality.FileURL,
				FileSize: quality.FileSize,
				Status:   quality.Status,
			}
		}
		videoDTO.Qualities = qualityDTOs
	}

	return videoDTO, nil
}

// GetVideos 獲取影片列表
func (s *VideoService) GetVideos(offset, limit int) ([]*dto.VideoDTO, int64, error) {
	videos, total, err := s.Repo.FindVideosWithPagination(offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("獲取影片列表失敗: %v", err)
	}

	// 轉換為 DTO
	videoDTOs := make([]*dto.VideoDTO, len(videos))
	for i, video := range videos {
		videoDTO := &dto.VideoDTO{
			ID:                 video.ID,
			Title:              video.Title,
			Description:        video.Description,
			UserID:             video.UserID,
			OriginalURL:        video.OriginalURL,
			ThumbnailURL:       video.ThumbnailURL,
			HLSMasterURL:       video.HLSMasterURL,
			MP4URL:             video.MP4URL,
			Duration:           video.Duration,
			FileSize:           video.FileSize,
			OriginalFormat:     video.OriginalFormat,
			Status:             video.Status,
			ProcessingProgress: video.ProcessingProgress,
			Views:              video.Views,
			Likes:              video.Likes,
			CreatedAt:          video.CreatedAt,
			UpdatedAt:          video.UpdatedAt,
		}

		// 如果有用戶資訊，添加用戶名
		if video.User != nil {
			videoDTO.Username = video.User.Username
		}

		videoDTOs[i] = videoDTO
	}

	return videoDTOs, total, nil
}

// GetVideosByUserID 根據用戶ID獲取影片列表
func (s *VideoService) GetVideosByUserID(userID uint) ([]*dto.VideoDTO, int64, error) {
	videos, err := s.Repo.FindVideoByUserID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("獲取用戶影片列表失敗: %v", err)
	}

	// 轉換為 DTO
	videoDTOs := make([]*dto.VideoDTO, len(videos))
	for i, video := range videos {
		videoDTO := &dto.VideoDTO{
			ID:                 video.ID,
			Title:              video.Title,
			Description:        video.Description,
			UserID:             video.UserID,
			OriginalURL:        video.OriginalURL,
			ThumbnailURL:       video.ThumbnailURL,
			HLSMasterURL:       video.HLSMasterURL,
			MP4URL:             video.MP4URL,
			Duration:           video.Duration,
			FileSize:           video.FileSize,
			OriginalFormat:     video.OriginalFormat,
			Status:             video.Status,
			ProcessingProgress: video.ProcessingProgress,
			Views:              video.Views,
			Likes:              video.Likes,
			CreatedAt:          video.CreatedAt,
			UpdatedAt:          video.UpdatedAt,
		}

		videoDTOs[i] = videoDTO
	}

	return videoDTOs, int64(len(videos)), nil
}

// SearchVideos 搜尋影片
func (s *VideoService) SearchVideos(query string, offset, limit int) ([]*dto.VideoDTO, int64, error) {
	videos, err := s.Repo.SearchVideo(query)
	if err != nil {
		return nil, 0, fmt.Errorf("搜尋影片失敗: %v", err)
	}

	// 手動實現分頁
	total := int64(len(videos))
	start := offset
	end := offset + limit
	if end > len(videos) {
		end = len(videos)
	}
	if start > len(videos) {
		start = len(videos)
	}

	pagedVideos := videos[start:end]

	// 轉換為 DTO
	videoDTOs := make([]*dto.VideoDTO, len(pagedVideos))
	for i, video := range pagedVideos {
		videoDTO := &dto.VideoDTO{
			ID:                 video.ID,
			Title:              video.Title,
			Description:        video.Description,
			UserID:             video.UserID,
			OriginalURL:        video.OriginalURL,
			ThumbnailURL:       video.ThumbnailURL,
			HLSMasterURL:       video.HLSMasterURL,
			MP4URL:             video.MP4URL,
			Duration:           video.Duration,
			FileSize:           video.FileSize,
			OriginalFormat:     video.OriginalFormat,
			Status:             video.Status,
			ProcessingProgress: video.ProcessingProgress,
			Views:              video.Views,
			Likes:              video.Likes,
			CreatedAt:          video.CreatedAt,
			UpdatedAt:          video.UpdatedAt,
		}

		// 如果有用戶資訊，添加用戶名
		if video.User != nil {
			videoDTO.Username = video.User.Username
		}

		videoDTOs[i] = videoDTO
	}

	return videoDTOs, total, nil
}

// UpdateVideo 更新影片資訊
func (s *VideoService) UpdateVideo(id uint, title string, description string, videoData *dto.VideoDTO) error {
	// 檢查影片是否存在
	_, err := s.Repo.FindVideoByID(id)
	if err != nil {
		return fmt.Errorf("找不到影片: %v", err)
	}

	// 更新影片資訊
	updates := map[string]interface{}{
		"title":       title,
		"description": description,
		"updated_at":  time.Now(),
	}

	// 更新資料庫
	if err := s.Repo.UpdateVideoFields(id, updates); err != nil {
		return fmt.Errorf("更新影片失敗: %v", err)
	}

	return nil
}

// DeleteVideo 刪除影片
func (s *VideoService) DeleteVideo(id uint) error {
	// 獲取影片資訊
	video, err := s.Repo.FindVideoByID(id)
	if err != nil {
		return fmt.Errorf("找不到影片: %v", err)
	}

	// 從 S3 刪除檔案（如果存在）
	if s.S3Storage != nil {
		// 刪除原始檔案
		if video.OriginalKey != "" {
			s.S3Storage.DeleteFile(video.OriginalKey)
		}

		// 刪除轉碼後的檔案
		if video.HLSKey != "" {
			s.S3Storage.DeleteFile(video.HLSKey)
		}
		if video.MP4Key != "" {
			s.S3Storage.DeleteFile(video.MP4Key)
		}
	}

	// 從資料庫刪除
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

// CheckS3Configuration 檢查 S3 配置
func (s *VideoService) CheckS3Configuration() error {
	if s.S3Storage == nil {
		return errors.New("S3 儲存未初始化")
	}
	return nil
}
