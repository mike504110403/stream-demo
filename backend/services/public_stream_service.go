package services

import (
	"fmt"
	"log"
	"os"
	"stream-demo/backend/config"
	"stream-demo/backend/pkg/media"
	"stream-demo/backend/utils"
	"time"
)

// PublicStreamService 公開流服務
type PublicStreamService struct {
	config      *config.Config
	media       *media.PublicStreamService
	cache       *utils.RedisCache
	persistence *StreamPersistenceService
	streams     map[string]*PublicStreamInfo
}

// PublicStreamInfo 公開流資訊
type PublicStreamInfo struct {
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Status      string    `json:"status"`
	LastUpdate  time.Time `json:"last_update"`
	ViewerCount int       `json:"viewer_count"`
	Category    string    `json:"category"`
}

// NewPublicStreamService 創建公開流服務
func NewPublicStreamService(cfg *config.Config, cache *utils.RedisCache) (*PublicStreamService, error) {
	service := &PublicStreamService{
		config:      cfg,
		media:       nil, // 不再需要媒體服務
		cache:       cache,
		persistence: NewStreamPersistenceService(cache),
		streams:     make(map[string]*PublicStreamInfo),
	}

	// 嘗試從持久化存儲加載流配置
	if streams, err := service.persistence.LoadStreamConfig(); err == nil {
		service.streams = streams
		log.Printf("從持久化存儲加載了 %d 個流配置", len(streams))
	} else {
		// 如果沒有持久化數據，初始化默認流
		service.initializeStreams()
		// 保存到持久化存儲
		if err := service.persistence.SaveStreamConfig(service.streams); err != nil {
			log.Printf("保存流配置到持久化存儲失敗: %v", err)
		}
	}

	// 確保所有流資訊都持久化到 Redis
	service.ensureStreamsPersisted()

	// 啟動緩存更新協程
	go service.updateCache()

	return service, nil
}

// initializeStreams 初始化流資訊
func (s *PublicStreamService) initializeStreams() {
	// 預設流資訊
	defaultStreams := map[string]*PublicStreamInfo{
		"tears_of_steel": {
			Name:        "tears_of_steel",
			Title:       "Tears of Steel",
			Description: "Unified Streaming 測試影片 - 科幻短片",
			URL:         "https://demo.unified-streaming.com/k8s/features/stable/video/tears-of-steel/tears-of-steel.ism/.m3u8",
			Status:      "active",
			Category:    "demo",
		},
		"mux_test": {
			Name:        "mux_test",
			Title:       "Mux 測試流",
			Description: "Mux 提供的測試 HLS 流",
			URL:         "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8",
			Status:      "active",
			Category:    "demo",
		},
	}

	for name, info := range defaultStreams {
		s.streams[name] = info
		// 存儲到 Redis
		s.cacheStreamInfo(name, info)
	}
}

// GetAvailableStreams 獲取可用的流列表
func (s *PublicStreamService) GetAvailableStreams() ([]*PublicStreamInfo, error) {
	streams := make([]*PublicStreamInfo, 0)

	for name, info := range s.streams {
		// 首先嘗試從緩存獲取
		cachedInfo, err := s.getStreamInfoFromCache(name)
		if err != nil {
			log.Printf("從緩存獲取流 %s 資訊失敗，使用內存數據: %v", name, err)
			// 如果緩存失敗，使用內存中的數據
			if info.Status == "active" {
				streams = append(streams, info)
			}
			// 重新存儲到緩存
			s.cacheStreamInfo(name, info)
			continue
		}

		if cachedInfo.Status == "active" {
			streams = append(streams, cachedInfo)
		}
	}

	return streams, nil
}

// GetStreamURL 獲取流的播放 URL
func (s *PublicStreamService) GetStreamURL(streamName string) (string, error) {
	// 檢查流是否存在
	if _, exists := s.streams[streamName]; !exists {
		return "", fmt.Errorf("流 %s 不存在", streamName)
	}

	// 更新觀看者數量
	s.incrementViewerCount(streamName)

	// 返回本地 HLS 服務器的 URL
	return fmt.Sprintf("http://localhost:8083/%s/index.m3u8", streamName), nil
}

// GetStreamURLs 獲取流的所有播放 URL
func (s *PublicStreamService) GetStreamURLs(streamName string) (map[string]string, error) {
	// 檢查流是否存在
	if _, exists := s.streams[streamName]; !exists {
		return nil, fmt.Errorf("流 %s 不存在", streamName)
	}

	// 更新觀看者數量
	s.incrementViewerCount(streamName)

	// 返回所有播放 URL
	urls := map[string]string{
		"hls":  fmt.Sprintf("http://localhost:8083/%s/index.m3u8", streamName),
		"rtmp": fmt.Sprintf("rtmp://localhost:1935/live/%s", streamName),
	}

	return urls, nil
}

// GetStreamInfo 獲取流詳細資訊
func (s *PublicStreamService) GetStreamInfo(streamName string) (*PublicStreamInfo, error) {
	return s.getStreamInfoFromCache(streamName)
}

// cacheStreamInfo 緩存流資訊
func (s *PublicStreamService) cacheStreamInfo(name string, info *PublicStreamInfo) {
	key := fmt.Sprintf("public_stream:%s", name)

	// 設置較長的過期時間（24小時），確保數據不會過期
	s.cache.Set(key, info, 24*time.Hour)
}

// getStreamInfoFromCache 從緩存獲取流資訊
func (s *PublicStreamService) getStreamInfoFromCache(name string) (*PublicStreamInfo, error) {
	key := fmt.Sprintf("public_stream:%s", name)
	var info PublicStreamInfo
	exists, err := s.cache.Get(key, &info)
	if err != nil {
		return nil, fmt.Errorf("從緩存獲取流資訊失敗: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("流資訊不存在")
	}

	return &info, nil
}

// incrementViewerCount 增加觀看者數量
func (s *PublicStreamService) incrementViewerCount(streamName string) {
	key := fmt.Sprintf("public_stream_viewers:%s", streamName)

	// 使用 Redis 的 INCR 命令
	s.cache.Increment(key, 1)
}

// updateCache 定期更新緩存
func (s *PublicStreamService) updateCache() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.refreshStreamStatus()
			// 確保所有流資訊都持久化到 Redis
			s.ensureStreamsPersisted()
		}
	}
}

// refreshStreamStatus 刷新流狀態
func (s *PublicStreamService) refreshStreamStatus() {
	// 更新所有已配置的流狀態
	for name, info := range s.streams {
		// 檢查本地 HLS 文件是否存在
		hlsPath := fmt.Sprintf("/tmp/public_streams/%s/index.m3u8", name)
		if _, err := os.Stat(hlsPath); err == nil {
			// HLS 文件存在，流為 active
			info.Status = "active"
		} else {
			// HLS 文件不存在，流為 inactive
			info.Status = "inactive"
		}

		info.LastUpdate = time.Now()

		// 更新緩存
		s.cacheStreamInfo(name, info)
	}
}

// ensureStreamsPersisted 確保所有流資訊都持久化到 Redis
func (s *PublicStreamService) ensureStreamsPersisted() {
	for name, info := range s.streams {
		// 檢查緩存中是否存在
		key := fmt.Sprintf("public_stream:%s", name)
		exists, err := s.cache.Exists(key)
		if err != nil {
			log.Printf("檢查流 %s 緩存失敗: %v", name, err)
			continue
		}

		// 如果不存在或過期，重新存儲
		if !exists {
			log.Printf("重新存儲流 %s 到緩存", name)
			s.cacheStreamInfo(name, info)
		}

		// 同時保存到持久化存儲
		if err := s.persistence.SaveStreamStatus(name, info.Status, info.ViewerCount); err != nil {
			log.Printf("保存流 %s 狀態到持久化存儲失敗: %v", name, err)
		}
	}

	// 定期備份流配置
	if err := s.persistence.BackupStreamData(s.streams); err != nil {
		log.Printf("備份流數據失敗: %v", err)
	}
}

// Stop 停止服務
func (s *PublicStreamService) Stop() error {
	return s.media.Stop()
}
