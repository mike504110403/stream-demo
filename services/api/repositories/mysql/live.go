package repo

import (
	"errors"
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

func (r *MysqlRepo) CreateLive(live *models.Live) error {
	return r.MysqlDB.Create(live).Error
}

func (r *MysqlRepo) FindLiveByID(id uint) (*models.Live, error) {
	var live models.Live
	if err := r.MysqlDB.Preload("User").First(&live, id).Error; err != nil {
		return nil, err
	}
	return &live, nil
}

func (r *MysqlRepo) FindLiveByUserID(userID uint) ([]*models.Live, int64, error) {
	var lives []*models.Live
	var total int64

	query := r.MysqlDB.Where("user_id = ?", userID)

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

func (r *MysqlRepo) FindLiveByStreamKey(streamKey string) (*models.Live, error) {
	var live models.Live
	if err := r.MysqlDB.Where("stream_key = ?", streamKey).First(&live).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &live, nil
}

func (r *MysqlRepo) UpdateLive(live *models.Live) error {
	return r.MysqlDB.Save(live).Error
}

func (r *MysqlRepo) DeleteLive(id uint) error {
	return r.MysqlDB.Delete(&models.Live{}, id).Error
}

func (r *MysqlRepo) FindActiveLive() ([]*models.Live, error) {
	var lives []*models.Live
	if err := r.MysqlDB.Where("status = ?", "live").Find(&lives).Error; err != nil {
		return nil, err
	}
	return lives, nil
}

func (r *MysqlRepo) IncrementLiveViewerCount(id uint) error {
	return r.MysqlDB.Model(&models.Live{}).Where("id = ?", id).UpdateColumn("viewer_count", gorm.Expr("viewer_count + ?", 1)).Error
}

func (r *MysqlRepo) DecrementLiveViewerCount(id uint) error {
	return r.MysqlDB.Model(&models.Live{}).Where("id = ?", id).UpdateColumn("viewer_count", gorm.Expr("viewer_count - ?", 1)).Error
}
