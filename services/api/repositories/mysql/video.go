package repo

import (
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

func (r *MysqlRepo) CreateVideo(video *models.Video) error {
	return r.MysqlDB.Create(video).Error
}

func (r *MysqlRepo) FindVideoByID(id uint) (*models.Video, error) {
	var video models.Video
	if err := r.MysqlDB.Preload("User").First(&video, id).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *MysqlRepo) FindVideoByUserID(userID uint) ([]models.Video, error) {
	var videos []models.Video
	if err := r.MysqlDB.Where("user_id = ?", userID).Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

func (r *MysqlRepo) FindAllVideo() ([]models.Video, error) {
	var videos []models.Video
	if err := r.MysqlDB.Preload("User").Where("status = ?", "published").Order("created_at DESC").Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

func (r *MysqlRepo) SearchVideo(query string) ([]models.Video, error) {
	var videos []models.Video
	searchQuery := "%" + query + "%"
	if err := r.MysqlDB.Where("(title LIKE ? OR description LIKE ?) AND status = ?", searchQuery, searchQuery, "published").Order("created_at DESC").Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

func (r *MysqlRepo) UpdateVideo(video *models.Video) error {
	return r.MysqlDB.Save(video).Error
}

func (r *MysqlRepo) DeleteVideo(id uint) error {
	return r.MysqlDB.Delete(&models.Video{}, id).Error
}

func (r *MysqlRepo) IncrementVideoViews(id uint) error {
	return r.MysqlDB.Model(&models.Video{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}

func (r *MysqlRepo) IncrementVideoLikes(id uint) error {
	return r.MysqlDB.Model(&models.Video{}).Where("id = ?", id).UpdateColumn("likes", gorm.Expr("likes + ?", 1)).Error
}

func (r *MysqlRepo) FindVideosWithPagination(offset, limit int) ([]models.Video, int64, error) {
	var videos []models.Video
	var total int64

	if err := r.MysqlDB.Model(&models.Video{}).Where("status IN ?", []string{"ready", "processing"}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.MysqlDB.Where("status IN ?", []string{"ready", "processing"}).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&videos).Error

	if err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

func (r *MysqlRepo) CreateVideoQuality(quality *models.VideoQuality) error {
	return r.MysqlDB.Create(quality).Error
}

func (r *MysqlRepo) FindVideoQualitiesByVideoID(videoID uint) ([]models.VideoQuality, error) {
	var qualities []models.VideoQuality
	if err := r.MysqlDB.Where("video_id = ?", videoID).Find(&qualities).Error; err != nil {
		return nil, err
	}
	return qualities, nil
}

func (r *MysqlRepo) UpdateVideoQuality(quality *models.VideoQuality) error {
	return r.MysqlDB.Save(quality).Error
}

func (r *MysqlRepo) DeleteVideoQuality(id uint) error {
	return r.MysqlDB.Delete(&models.VideoQuality{}, id).Error
}
