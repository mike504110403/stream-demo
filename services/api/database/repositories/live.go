package repositories

import (
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

// LiveRepository 直播資料庫操作介面
type LiveRepository interface {
	Create(live *models.Live) error
	FindByID(id uint) (*models.Live, error)
	FindByUserID(userID uint) ([]*models.Live, int64, error)
	FindByStreamKey(streamKey string) (*models.Live, error)
	FindActive() ([]*models.Live, error)
	Update(live *models.Live) error
	Delete(id uint) error
	IncrementViewerCount(id uint) error
	DecrementViewerCount(id uint) error
}

// liveRepository 直播資料庫操作實現
type liveRepository struct {
	db *gorm.DB
}

// NewLiveRepository 創建直播資料庫操作實例
func NewLiveRepository(db *gorm.DB) LiveRepository {
	return &liveRepository{db: db}
}

// Create 創建直播
func (r *liveRepository) Create(live *models.Live) error {
	return r.db.Create(live).Error
}

// FindByID 根據 ID 查找直播
func (r *liveRepository) FindByID(id uint) (*models.Live, error) {
	var live models.Live
	err := r.db.First(&live, id).Error
	if err != nil {
		return nil, err
	}
	return &live, nil
}

// FindByUserID 根據用戶 ID 查找直播（支持分頁）
func (r *liveRepository) FindByUserID(userID uint) ([]*models.Live, int64, error) {
	var lives []*models.Live
	var total int64

	query := r.db.Where("user_id = ?", userID)

	// 計算總數
	if err := query.Model(&models.Live{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查找記錄
	if err := query.Find(&lives).Error; err != nil {
		return nil, 0, err
	}

	return lives, total, nil
}

// FindByStreamKey 根據串流金鑰查找直播
func (r *liveRepository) FindByStreamKey(streamKey string) (*models.Live, error) {
	var live models.Live
	err := r.db.Where("stream_key = ?", streamKey).First(&live).Error
	if err != nil {
		return nil, err
	}
	return &live, nil
}

// FindActive 查找所有活躍的直播
func (r *liveRepository) FindActive() ([]*models.Live, error) {
	var lives []*models.Live
	err := r.db.Where("status = ?", "live").Find(&lives).Error
	if err != nil {
		return nil, err
	}
	return lives, nil
}

// Update 更新直播
func (r *liveRepository) Update(live *models.Live) error {
	return r.db.Save(live).Error
}

// Delete 刪除直播
func (r *liveRepository) Delete(id uint) error {
	return r.db.Delete(&models.Live{}, id).Error
}

// IncrementViewerCount 增加觀看人數
func (r *liveRepository) IncrementViewerCount(id uint) error {
	return r.db.Model(&models.Live{}).Where("id = ?", id).UpdateColumn("viewer_count", gorm.Expr("viewer_count + ?", 1)).Error
}

// DecrementViewerCount 減少觀看人數
func (r *liveRepository) DecrementViewerCount(id uint) error {
	return r.db.Model(&models.Live{}).Where("id = ?", id).UpdateColumn("viewer_count", gorm.Expr("viewer_count - ?", 1)).Error
}
