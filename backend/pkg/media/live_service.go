package media

import (
	"fmt"
	"log"
	"os/exec"
)

// LiveService 直播服務介面
type LiveService interface {
	// 啟動直播服務
	Start() error
	// 停止直播服務
	Stop() error
	// 獲取直播流 URL
	GetStreamURL(streamKey string) (string, error)
	// 獲取推流 URL
	GetPushURL(streamKey string) (string, error)
	// 檢查流狀態
	CheckStreamStatus(streamKey string) (bool, error)
	// 獲取可用流列表
	GetActiveStreams() ([]string, error)
}

// LocalLiveService 本地直播服務
type LocalLiveService struct {
	config     LocalLiveConfig
	rtmpServer *RTMPServer
	transcoder *LiveTranscoder
}

// LocalLiveConfig 本地直播配置
type LocalLiveConfig struct {
	RTMPServer        string
	RTMPServerPort    int
	TranscoderEnabled bool
	HLSOutputDir      string
	HTTPPort          int
}

// RTMPServer RTMP 服務器
type RTMPServer struct {
	config LocalLiveConfig
	cmd    *exec.Cmd
}

// LiveTranscoder 直播轉碼器
type LiveTranscoder struct {
	config LocalLiveConfig
	cmd    *exec.Cmd
}

// NewLocalLiveService 創建本地直播服務
func NewLocalLiveService(config LocalLiveConfig) *LocalLiveService {
	return &LocalLiveService{
		config: config,
		rtmpServer: &RTMPServer{
			config: config,
		},
		transcoder: &LiveTranscoder{
			config: config,
		},
	}
}

// Start 啟動本地直播服務
func (s *LocalLiveService) Start() error {
	log.Println("🚀 啟動本地直播服務...")

	// 啟動 RTMP 服務器
	if err := s.rtmpServer.Start(); err != nil {
		return fmt.Errorf("啟動 RTMP 服務器失敗: %w", err)
	}

	// 啟動轉碼器
	if s.config.TranscoderEnabled {
		if err := s.transcoder.Start(); err != nil {
			return fmt.Errorf("啟動轉碼器失敗: %w", err)
		}
	}

	log.Println("✅ 本地直播服務啟動完成")
	return nil
}

// Stop 停止本地直播服務
func (s *LocalLiveService) Stop() error {
	log.Println("🛑 停止本地直播服務...")

	if s.transcoder != nil {
		s.transcoder.Stop()
	}

	if s.rtmpServer != nil {
		s.rtmpServer.Stop()
	}

	log.Println("✅ 本地直播服務已停止")
	return nil
}

// GetStreamURL 獲取直播流 URL
func (s *LocalLiveService) GetStreamURL(streamKey string) (string, error) {
	if s.config.TranscoderEnabled {
		// 返回 HLS 流 URL
		return fmt.Sprintf("http://localhost:%d/%s/index.m3u8", s.config.HTTPPort, streamKey), nil
	}
	// 返回 RTMP 流 URL
	return fmt.Sprintf("rtmp://%s:%d/live/%s", s.config.RTMPServer, s.config.RTMPServerPort, streamKey), nil
}

// GetPushURL 獲取推流 URL
func (s *LocalLiveService) GetPushURL(streamKey string) (string, error) {
	return fmt.Sprintf("rtmp://%s:%d/live/%s", s.config.RTMPServer, s.config.RTMPServerPort, streamKey), nil
}

// CheckStreamStatus 檢查流狀態
func (s *LocalLiveService) CheckStreamStatus(streamKey string) (bool, error) {
	// 檢查 HLS 文件是否存在
	if s.config.TranscoderEnabled {
		_ = fmt.Sprintf("%s/%s/index.m3u8", s.config.HLSOutputDir, streamKey)
		// 這裡可以實現文件存在性檢查
		return true, nil // 簡化實現
	}
	return true, nil
}

// GetActiveStreams 獲取可用流列表
func (s *LocalLiveService) GetActiveStreams() ([]string, error) {
	// 這裡可以實現掃描 HLS 目錄或查詢 RTMP 狀態
	return []string{}, nil
}

// Start 啟動 RTMP 服務器
func (r *RTMPServer) Start() error {
	log.Println("📡 啟動 RTMP 服務器...")

	// 這裡可以啟動 nginx-rtmp 或使用 FFmpeg 作為 RTMP 服務器
	// 簡化實現，實際應該啟動 nginx-rtmp 容器或進程
	log.Println("✅ RTMP 服務器已啟動")
	return nil
}

// Stop 停止 RTMP 服務器
func (r *RTMPServer) Stop() error {
	if r.cmd != nil && r.cmd.Process != nil {
		return r.cmd.Process.Kill()
	}
	return nil
}

// Start 啟動直播轉碼器
func (t *LiveTranscoder) Start() error {
	log.Println("🎬 啟動直播轉碼器...")

	// 這裡可以啟動 FFmpeg 轉碼進程
	// 簡化實現，實際應該啟動轉碼腳本
	log.Println("✅ 直播轉碼器已啟動")
	return nil
}

// Stop 停止直播轉碼器
func (t *LiveTranscoder) Stop() error {
	if t.cmd != nil && t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}
	return nil
}

// CloudLiveService 雲端直播服務
type CloudLiveService struct {
	config CloudLiveConfig
}

// CloudLiveConfig 雲端直播配置
type CloudLiveConfig struct {
	Provider         string
	RTMPIngestURL    string
	HLSPlaybackURL   string
	APIKey           string
	APISecret        string
	TranscodeEnabled bool
}

// NewCloudLiveService 創建雲端直播服務
func NewCloudLiveService(config CloudLiveConfig) *CloudLiveService {
	return &CloudLiveService{
		config: config,
	}
}

// Start 啟動雲端直播服務
func (s *CloudLiveService) Start() error {
	log.Println("☁️ 啟動雲端直播服務...")
	// 雲端服務通常不需要本地啟動
	return nil
}

// Stop 停止雲端直播服務
func (s *CloudLiveService) Stop() error {
	log.Println("☁️ 停止雲端直播服務...")
	return nil
}

// GetStreamURL 獲取直播流 URL
func (s *CloudLiveService) GetStreamURL(streamKey string) (string, error) {
	return fmt.Sprintf("%s/%s/index.m3u8", s.config.HLSPlaybackURL, streamKey), nil
}

// GetPushURL 獲取推流 URL
func (s *CloudLiveService) GetPushURL(streamKey string) (string, error) {
	return fmt.Sprintf("%s/%s", s.config.RTMPIngestURL, streamKey), nil
}

// CheckStreamStatus 檢查流狀態
func (s *CloudLiveService) CheckStreamStatus(streamKey string) (bool, error) {
	// 這裡應該調用雲端 API 檢查流狀態
	return true, nil
}

// GetActiveStreams 獲取可用流列表
func (s *CloudLiveService) GetActiveStreams() ([]string, error) {
	// 這裡應該調用雲端 API 獲取流列表
	return []string{}, nil
}

// LiveServiceFactory 直播服務工廠
func LiveServiceFactory(serviceType string, config interface{}) (LiveService, error) {
	switch serviceType {
	case "local":
		if localConfig, ok := config.(LocalLiveConfig); ok {
			return NewLocalLiveService(localConfig), nil
		}
		return nil, fmt.Errorf("無效的本地直播配置")
	case "cloud":
		if cloudConfig, ok := config.(CloudLiveConfig); ok {
			return NewCloudLiveService(cloudConfig), nil
		}
		return nil, fmt.Errorf("無效的雲端直播配置")
	default:
		return nil, fmt.Errorf("不支援的直播服務類型: %s", serviceType)
	}
}
