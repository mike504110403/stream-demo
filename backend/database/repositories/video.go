package repositories

import (
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

// VideoRepository 影片資料庫操作介面
type VideoRepository interface {
	Create(video *models.Video) error
	FindByID(id uint) (*models.Video, error)
	FindByUserID(userID uint) ([]models.Video, error)
	FindAll() ([]models.Video, error)
	Search(query string) ([]models.Video, error)
	Update(video *models.Video) error
	Delete(id uint) error
	IncrementViews(id uint) error
	IncrementLikes(id uint) error
}

// videoRepository 影片資料庫操作實現
type videoRepository struct {
	db *gorm.DB
}

// NewVideoRepository 創建影片資料庫操作實例
func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepository{db: db}
}

// Create 創建影片
func (r *videoRepository) Create(video *models.Video) error {
	return r.db.Create(video).Error
}

// FindByID 根據 ID 查找影片
func (r *videoRepository) FindByID(id uint) (*models.Video, error) {
	var video models.Video
	err := r.db.First(&video, id).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// FindByUserID 根據用戶 ID 查找影片
func (r *videoRepository) FindByUserID(userID uint) ([]models.Video, error) {
	var videos []models.Video
	err := r.db.Where("user_id = ?", userID).Find(&videos).Error
	if err != nil {
		return nil, err
	}
	return videos, nil
}

// FindAll 查找所有影片
func (r *videoRepository) FindAll() ([]models.Video, error) {
	var videos []models.Video
	err := r.db.Where("status = ?", "published").Order("created_at DESC").Find(&videos).Error
	if err != nil {
		return nil, err
	}
	return videos, nil
}

// Search 搜尋影片
func (r *videoRepository) Search(query string) ([]models.Video, error) {
	var videos []models.Video
	searchQuery := "%" + query + "%"
	err := r.db.Where("(title LIKE ? OR description LIKE ?) AND status = ?", searchQuery, searchQuery, "published").Order("created_at DESC").Find(&videos).Error
	if err != nil {
		return nil, err
	}
	return videos, nil
}

// Update 更新影片
func (r *videoRepository) Update(video *models.Video) error {
	return r.db.Save(video).Error
}

// Delete 刪除影片
func (r *videoRepository) Delete(id uint) error {
	return r.db.Delete(&models.Video{}, id).Error
}

// IncrementViews 增加觀看次數
func (r *videoRepository) IncrementViews(id uint) error {
	return r.db.Model(&models.Video{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}

// IncrementLikes 增加喜歡次數
func (r *videoRepository) IncrementLikes(id uint) error {
	return r.db.Model(&models.Video{}).Where("id = ?", id).UpdateColumn("likes", gorm.Expr("likes + ?", 1)).Error
}
