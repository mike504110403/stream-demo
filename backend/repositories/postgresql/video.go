package postgresql

import (
	"stream-demo/backend/database/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateVideo 創建影片
func (r *PostgreSQLRepo) CreateVideo(video *models.Video) error {
	return r.PostgreSQLDB.Create(video).Error
}

// FindVideoByID 根據ID查找影片
func (r *PostgreSQLRepo) FindVideoByID(id uint) (*models.Video, error) {
	var video models.Video
	if err := r.PostgreSQLDB.Preload("User").First(&video, id).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

// FindVideoByUserID 根據用戶ID查找影片列表
func (r *PostgreSQLRepo) FindVideoByUserID(userID uint) ([]models.Video, error) {
	var videos []models.Video
	if err := r.PostgreSQLDB.Where("user_id = ?", userID).Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// FindVideosByStatus 根據狀態查找影片列表
func (r *PostgreSQLRepo) FindVideosByStatus(statuses []string) ([]models.Video, error) {
	var videos []models.Video
	if err := r.PostgreSQLDB.Where("status IN ?", statuses).Order("created_at ASC").Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// FindVideosByStatusForUpdate 根據狀態查找影片列表並加鎖（用於轉碼）
func (r *PostgreSQLRepo) FindVideosByStatusForUpdate(statuses []string, limit int) ([]models.Video, error) {
	var videos []models.Video
	if err := r.PostgreSQLDB.Where("status IN ?", statuses).
		Order("created_at ASC").
		Limit(limit).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// UpdateVideoStatus 更新影片狀態（用於轉碼進度）
func (r *PostgreSQLRepo) UpdateVideoStatus(videoID uint, status string, progress int, errorMsg string) error {
	updates := map[string]interface{}{
		"status":              status,
		"processing_progress": progress,
		"updated_at":          time.Now(),
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	return r.PostgreSQLDB.Model(&models.Video{}).Where("id = ?", videoID).Updates(updates).Error
}

// GetDB 獲取資料庫連接（用於事務）
func (r *PostgreSQLRepo) GetDB() *gorm.DB {
	return r.PostgreSQLDB
}

// UpdateVideoFields 更新影片欄位
func (r *PostgreSQLRepo) UpdateVideoFields(videoID uint, updates map[string]interface{}) error {
	return r.PostgreSQLDB.Model(&models.Video{}).Where("id = ?", videoID).Updates(updates).Error
}

// FindAllVideo 查找所有已發布的影片
func (r *PostgreSQLRepo) FindAllVideo() ([]models.Video, error) {
	var videos []models.Video
	if err := r.PostgreSQLDB.Preload("User").Where("status = ?", "completed").Order("created_at DESC").Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// SearchVideo 搜尋影片
func (r *PostgreSQLRepo) SearchVideo(query string) ([]models.Video, error) {
	var videos []models.Video
	searchQuery := "%" + query + "%"
	// PostgreSQL使用ILIKE進行不區分大小寫的搜尋
	if err := r.PostgreSQLDB.Where("(title ILIKE ? OR description ILIKE ?) AND status = ?", searchQuery, searchQuery, "completed").Order("created_at DESC").Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// UpdateVideo 更新影片
func (r *PostgreSQLRepo) UpdateVideo(video *models.Video) error {
	return r.PostgreSQLDB.Save(video).Error
}

// DeleteVideo 刪除影片
func (r *PostgreSQLRepo) DeleteVideo(id uint) error {
	return r.PostgreSQLDB.Delete(&models.Video{}, id).Error
}

// IncrementVideoViews 增加影片觀看次數
func (r *PostgreSQLRepo) IncrementVideoViews(id uint) error {
	return r.PostgreSQLDB.Model(&models.Video{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}

// IncrementVideoLikes 增加影片喜歡次數
func (r *PostgreSQLRepo) IncrementVideoLikes(id uint) error {
	return r.PostgreSQLDB.Model(&models.Video{}).Where("id = ?", id).UpdateColumn("likes", gorm.Expr("likes + ?", 1)).Error
}

// FindVideosWithPagination 分頁查找影片
func (r *PostgreSQLRepo) FindVideosWithPagination(offset, limit int) ([]models.Video, int64, error) {
	var videos []models.Video
	var total int64

	// 使用 IN 語法替代 ANY，更簡潔且兼容性更好
	statuses := []string{"ready", "processing", "transcoding"}
	if err := r.PostgreSQLDB.Model(&models.Video{}).Where("status IN ?", statuses).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.PostgreSQLDB.Where("status IN ?", statuses).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&videos).Error

	if err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// CreateVideoQuality 創建影片品質記錄
func (r *PostgreSQLRepo) CreateVideoQuality(quality *models.VideoQuality) error {
	return r.PostgreSQLDB.Create(quality).Error
}

// FindVideoQualitiesByVideoID 根據影片ID查找品質列表
func (r *PostgreSQLRepo) FindVideoQualitiesByVideoID(videoID uint) ([]models.VideoQuality, error) {
	var qualities []models.VideoQuality
	if err := r.PostgreSQLDB.Where("video_id = ?", videoID).Find(&qualities).Error; err != nil {
		return nil, err
	}
	return qualities, nil
}

// UpdateVideoQuality 更新影片品質
func (r *PostgreSQLRepo) UpdateVideoQuality(quality *models.VideoQuality) error {
	return r.PostgreSQLDB.Save(quality).Error
}

// DeleteVideoQuality 刪除影片品質
func (r *PostgreSQLRepo) DeleteVideoQuality(id uint) error {
	return r.PostgreSQLDB.Delete(&models.VideoQuality{}, id).Error
}
