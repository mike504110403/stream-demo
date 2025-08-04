package postgresql

import (
	"errors"
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

// CreateLive 創建直播
func (r *PostgreSQLRepo) CreateLive(live *models.Live) error {
	return r.PostgreSQLDB.Create(live).Error
}

// FindLiveByID 根據ID查找直播
func (r *PostgreSQLRepo) FindLiveByID(id uint) (*models.Live, error) {
	var live models.Live
	if err := r.PostgreSQLDB.Preload("User").First(&live, id).Error; err != nil {
		return nil, err
	}
	return &live, nil
}

// FindLiveByUserID 根據用戶ID查找直播列表
func (r *PostgreSQLRepo) FindLiveByUserID(userID uint) ([]*models.Live, int64, error) {
	var lives []*models.Live
	var total int64

	query := r.PostgreSQLDB.Where("user_id = ?", userID)

	// 計算總數
	if err := query.Model(&models.Live{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查找記錄
	if err := query.Order("created_at DESC").Find(&lives).Error; err != nil {
		return nil, 0, err
	}

	return lives, total, nil
}

// FindLiveByStreamKey 根據串流密鑰查找直播
func (r *PostgreSQLRepo) FindLiveByStreamKey(streamKey string) (*models.Live, error) {
	var live models.Live
	if err := r.PostgreSQLDB.Where("stream_key = ?", streamKey).First(&live).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &live, nil
}

// UpdateLive 更新直播
func (r *PostgreSQLRepo) UpdateLive(live *models.Live) error {
	return r.PostgreSQLDB.Save(live).Error
}

// DeleteLive 刪除直播
func (r *PostgreSQLRepo) DeleteLive(id uint) error {
	return r.PostgreSQLDB.Delete(&models.Live{}, id).Error
}

// FindActiveLive 查找所有進行中的直播
func (r *PostgreSQLRepo) FindActiveLive() ([]*models.Live, error) {
	var lives []*models.Live
	if err := r.PostgreSQLDB.Where("status = ?", "live").Find(&lives).Error; err != nil {
		return nil, err
	}
	return lives, nil
}

// IncrementLiveViewerCount 增加直播觀看人數
func (r *PostgreSQLRepo) IncrementLiveViewerCount(id uint) error {
	return r.PostgreSQLDB.Model(&models.Live{}).Where("id = ?", id).UpdateColumn("viewer_count", gorm.Expr("viewer_count + ?", 1)).Error
}

// DecrementLiveViewerCount 減少直播觀看人數
func (r *PostgreSQLRepo) DecrementLiveViewerCount(id uint) error {
	return r.PostgreSQLDB.Model(&models.Live{}).Where("id = ?", id).UpdateColumn("viewer_count", gorm.Expr("viewer_count - ?", 1)).Error
}
