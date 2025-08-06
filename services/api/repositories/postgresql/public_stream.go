package postgresql

import (
	"gorm.io/gorm"
	"stream-demo/backend/database/models"
)

// PublicStreamRepository 公開流倉庫
type PublicStreamRepository struct {
	DB *gorm.DB
}

// NewPublicStreamRepository 創建公開流倉庫
func NewPublicStreamRepository(db *gorm.DB) *PublicStreamRepository {
	return &PublicStreamRepository{DB: db}
}

// GetAll 獲取所有公開流
func (r *PublicStreamRepository) GetAll() ([]models.PublicStream, error) {
	var streams []models.PublicStream
	err := r.DB.Find(&streams).Error
	return streams, err
}

// GetEnabled 獲取所有啟用的公開流
func (r *PublicStreamRepository) GetEnabled() ([]models.PublicStream, error) {
	var streams []models.PublicStream
	err := r.DB.Where("enabled = ?", true).Find(&streams).Error
	return streams, err
}

// GetByName 根據名稱獲取公開流
func (r *PublicStreamRepository) GetByName(name string) (*models.PublicStream, error) {
	var stream models.PublicStream
	err := r.DB.Where("name = ?", name).First(&stream).Error
	if err != nil {
		return nil, err
	}
	return &stream, nil
}

// Create 創建公開流
func (r *PublicStreamRepository) Create(stream *models.PublicStream) error {
	return r.DB.Create(stream).Error
}

// Update 更新公開流
func (r *PublicStreamRepository) Update(stream *models.PublicStream) error {
	return r.DB.Save(stream).Error
}

// Delete 刪除公開流
func (r *PublicStreamRepository) Delete(id uint) error {
	return r.DB.Delete(&models.PublicStream{}, id).Error
}

// ToggleEnabled 切換啟用狀態
func (r *PublicStreamRepository) ToggleEnabled(name string, enabled bool) error {
	return r.DB.Model(&models.PublicStream{}).Where("name = ?", name).Update("enabled", enabled).Error
}
