package services

import (
	"fmt"
	"log"
	"time"

	"stream-demo/backend/utils"
)

// StreamPersistenceService 流持久化服務
type StreamPersistenceService struct {
	cache *utils.RedisCache
}

// NewStreamPersistenceService 創建流持久化服務
func NewStreamPersistenceService(cache *utils.RedisCache) *StreamPersistenceService {
	return &StreamPersistenceService{
		cache: cache,
	}
}

// SaveStreamConfig 保存流配置到持久化存儲
func (s *StreamPersistenceService) SaveStreamConfig(streams map[string]*PublicStreamInfo) error {
	key := "persistent:stream_configs"

	// 轉換為可序列化的格式
	configs := make(map[string]map[string]interface{})
	for name, info := range streams {
		configs[name] = map[string]interface{}{
			"name":        info.Name,
			"title":       info.Title,
			"description": info.Description,
			"url":         info.URL,
			"status":      info.Status,
			"category":    info.Category,
		}
	}

	// 保存到 Redis，永不過期
	return s.cache.Set(key, configs, 0) // 0 表示永不過期
}

// LoadStreamConfig 從持久化存儲加載流配置
func (s *StreamPersistenceService) LoadStreamConfig() (map[string]*PublicStreamInfo, error) {
	key := "persistent:stream_configs"

	var configs map[string]map[string]interface{}
	exists, err := s.cache.Get(key, &configs)
	if err != nil {
		return nil, fmt.Errorf("加載流配置失敗: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("流配置不存在")
	}

	// 轉換回 PublicStreamInfo 格式
	streams := make(map[string]*PublicStreamInfo)
	for name, config := range configs {
		streams[name] = &PublicStreamInfo{
			Name:        config["name"].(string),
			Title:       config["title"].(string),
			Description: config["description"].(string),
			URL:         config["url"].(string),
			Status:      config["status"].(string),
			Category:    config["category"].(string),
			LastUpdate:  time.Now(),
		}
	}

	return streams, nil
}

// SaveStreamStatus 保存流狀態
func (s *StreamPersistenceService) SaveStreamStatus(streamName string, status string, viewerCount int) error {
	key := fmt.Sprintf("persistent:stream_status:%s", streamName)

	statusData := map[string]interface{}{
		"status":       status,
		"viewer_count": viewerCount,
		"last_update":  time.Now().Unix(),
	}

	// 保存狀態，過期時間 1 小時
	return s.cache.Set(key, statusData, time.Hour)
}

// LoadStreamStatus 加載流狀態
func (s *StreamPersistenceService) LoadStreamStatus(streamName string) (string, int, error) {
	key := fmt.Sprintf("persistent:stream_status:%s", streamName)

	var statusData map[string]interface{}
	exists, err := s.cache.Get(key, &statusData)
	if err != nil {
		return "", 0, fmt.Errorf("加載流狀態失敗: %w", err)
	}

	if !exists {
		return "active", 0, nil // 默認狀態
	}

	status := statusData["status"].(string)
	viewerCount := int(statusData["viewer_count"].(float64))

	return status, viewerCount, nil
}

// BackupStreamData 備份流數據
func (s *StreamPersistenceService) BackupStreamData(streams map[string]*PublicStreamInfo) error {
	backupKey := fmt.Sprintf("backup:streams:%d", time.Now().Unix())

	// 創建備份數據
	backupData := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"streams":   streams,
	}

	// 保存備份，保留 7 天
	return s.cache.Set(backupKey, backupData, 7*24*time.Hour)
}

// RestoreFromBackup 從備份恢復
func (s *StreamPersistenceService) RestoreFromBackup() (map[string]*PublicStreamInfo, error) {
	// 查找最新的備份
	// 這裡需要 Redis 的 SCAN 命令，簡化處理
	// 實際實現中可以使用 Redis 的 SCAN 命令查找最新的備份

	log.Printf("嘗試從備份恢復流數據...")
	return nil, fmt.Errorf("備份恢復功能待實現")
}

// GetStreamStats 獲取流統計信息
func (s *StreamPersistenceService) GetStreamStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 獲取所有流配置
	streams, err := s.LoadStreamConfig()
	if err != nil {
		return nil, err
	}

	for name := range streams {
		status, viewerCount, err := s.LoadStreamStatus(name)
		if err != nil {
			log.Printf("獲取流 %s 狀態失敗: %v", name, err)
			continue
		}

		stats[name] = map[string]interface{}{
			"status":       status,
			"viewer_count": viewerCount,
		}
	}

	return stats, nil
}
