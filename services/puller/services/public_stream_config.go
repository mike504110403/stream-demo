package services

import (
	"stream-demo/backend/database/models"
	"stream-demo/backend/repositories/postgresql"
)

// PublicStreamConfigService 公開流配置服務
type PublicStreamConfigService struct {
	repo *postgresql.PublicStreamRepository
}

// NewPublicStreamConfigService 創建公開流配置服務
func NewPublicStreamConfigService(repo *postgresql.PublicStreamRepository) *PublicStreamConfigService {
	return &PublicStreamConfigService{repo: repo}
}

// GetAllStreams 獲取所有公開流配置
func (s *PublicStreamConfigService) GetAllStreams() ([]models.PublicStream, error) {
	return s.repo.GetAll()
}

// GetEnabledStreams 獲取所有啟用的公開流配置
func (s *PublicStreamConfigService) GetEnabledStreams() ([]models.PublicStream, error) {
	return s.repo.GetEnabled()
}

// GetStreamByName 根據名稱獲取公開流配置
func (s *PublicStreamConfigService) GetStreamByName(name string) (*models.PublicStream, error) {
	return s.repo.GetByName(name)
}

// CreateStream 創建公開流配置
func (s *PublicStreamConfigService) CreateStream(stream *models.PublicStream) error {
	return s.repo.Create(stream)
}

// UpdateStream 更新公開流配置
func (s *PublicStreamConfigService) UpdateStream(stream *models.PublicStream) error {
	return s.repo.Update(stream)
}

// DeleteStream 刪除公開流配置
func (s *PublicStreamConfigService) DeleteStream(id uint) error {
	return s.repo.Delete(id)
}

// ToggleStreamEnabled 切換流啟用狀態
func (s *PublicStreamConfigService) ToggleStreamEnabled(name string, enabled bool) error {
	return s.repo.ToggleEnabled(name, enabled)
}
