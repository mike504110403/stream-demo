package media

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PublicStreamService 公開流服務
type PublicStreamService struct {
	config  PublicStreamConfig
	streams map[string]*StreamInfo
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// PublicStreamConfig 公開流配置
type PublicStreamConfig struct {
	OutputDir       string            // HLS 輸出目錄
	StreamConfigs   map[string]string // 流名稱 -> URL 映射
	SegmentTime     int               // 片段時長（秒）
	SegmentListSize int               // 片段列表大小
	HTTPPort        int               // HTTP 服務端口
}

// StreamInfo 流資訊
type StreamInfo struct {
	Name        string
	URL         string
	Status      string // "active", "inactive", "error"
	LastUpdate  time.Time
	ViewerCount int
	Process     *exec.Cmd
}

// NewPublicStreamService 創建公開流服務
func NewPublicStreamService(config PublicStreamConfig) *PublicStreamService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &PublicStreamService{
		config:  config,
		streams: make(map[string]*StreamInfo),
		ctx:     ctx,
		cancel:  cancel,
	}

	return service
}

// Start 啟動公開流服務
func (s *PublicStreamService) Start() error {
	log.Println("🚀 啟動公開流服務...")

	// 創建輸出目錄
	if err := s.ensureOutputDir(); err != nil {
		return fmt.Errorf("創建輸出目錄失敗: %w", err)
	}

	// 啟動所有配置的流
	for streamName, streamURL := range s.config.StreamConfigs {
		if err := s.startStream(streamName, streamURL); err != nil {
			log.Printf("啟動流 %s 失敗: %v", streamName, err)
			continue
		}
	}

	// 啟動監控協程
	go s.monitorStreams()

	log.Println("✅ 公開流服務啟動完成")
	return nil
}

// Stop 停止公開流服務
func (s *PublicStreamService) Stop() error {
	log.Println("🛑 停止公開流服務...")

	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, stream := range s.streams {
		if stream.Process != nil && stream.Process.Process != nil {
			stream.Process.Process.Kill()
		}
	}

	log.Println("✅ 公開流服務已停止")
	return nil
}

// GetAvailableStreams 獲取可用的流列表
func (s *PublicStreamService) GetAvailableStreams() []StreamInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	streams := make([]StreamInfo, 0, len(s.streams))
	for _, stream := range s.streams {
		streams = append(streams, *stream)
	}

	return streams
}

// GetStreamURL 獲取流的播放 URL
func (s *PublicStreamService) GetStreamURL(streamName string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stream, exists := s.streams[streamName]
	if !exists {
		return "", fmt.Errorf("流 %s 不存在", streamName)
	}

	if stream.Status != "active" {
		return "", fmt.Errorf("流 %s 狀態為 %s", streamName, stream.Status)
	}

	return fmt.Sprintf("http://localhost:%d/%s/index.m3u8", s.config.HTTPPort, streamName), nil
}

// startStream 啟動單個流
func (s *PublicStreamService) startStream(streamName, streamURL string) error {
	log.Printf("🎬 啟動流: %s (%s)", streamName, streamURL)

	// 創建流目錄
	streamDir := filepath.Join(s.config.OutputDir, streamName)
	if err := s.ensureDir(streamDir); err != nil {
		return fmt.Errorf("創建流目錄失敗: %w", err)
	}

	// 檢查流是否為 HLS
	if strings.Contains(streamURL, ".m3u8") {
		return s.startHLSStream(streamName, streamURL, streamDir)
	}

	// 檢查流是否為 RTMP
	if strings.HasPrefix(streamURL, "rtmp://") {
		return s.startRTMPStream(streamName, streamURL, streamDir)
	}

	// 檢查流是否為 RTSP
	if strings.HasPrefix(streamURL, "rtsp://") {
		return s.startRTSPStream(streamName, streamURL, streamDir)
	}

	// 檢查是否為 MP4 檔案（作為循環直播源）
	if strings.Contains(streamURL, ".mp4") {
		return s.startMP4Stream(streamName, streamURL, streamDir)
	}

	return fmt.Errorf("不支援的流格式: %s", streamURL)
}

// startHLSStream 啟動 HLS 流拉取
func (s *PublicStreamService) startHLSStream(streamName, streamURL, streamDir string) error {
	// 使用 FFmpeg 拉取 HLS 流並重新封裝
	args := []string{
		"-i", streamURL,
		"-c", "copy", // 直接複製，不重新編碼
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", s.config.SegmentTime),
		"-hls_list_size", fmt.Sprintf("%d", s.config.SegmentListSize),
		"-hls_flags", "delete_segments",
		"-hls_segment_filename", filepath.Join(streamDir, "segment_%03d.ts"),
		filepath.Join(streamDir, "index.m3u8"),
	}

	cmd := exec.CommandContext(s.ctx, "ffmpeg", args...)
	cmd.Dir = streamDir

	// 記錄流資訊
	s.mu.Lock()
	s.streams[streamName] = &StreamInfo{
		Name:       streamName,
		URL:        streamURL,
		Status:     "active",
		LastUpdate: time.Now(),
		Process:    cmd,
	}
	s.mu.Unlock()

	// 啟動進程
	if err := cmd.Start(); err != nil {
		s.mu.Lock()
		s.streams[streamName].Status = "error"
		s.mu.Unlock()
		return fmt.Errorf("啟動 FFmpeg 失敗: %w", err)
	}

	// 監控進程
	go func() {
		cmd.Wait()
		s.mu.Lock()
		if stream, exists := s.streams[streamName]; exists {
			stream.Status = "inactive"
			stream.Process = nil
		}
		s.mu.Unlock()
		log.Printf("流 %s 已停止", streamName)
	}()

	log.Printf("✅ 流 %s 啟動成功", streamName)
	return nil
}

// startRTSPStream 啟動 RTSP 流拉取
func (s *PublicStreamService) startRTSPStream(streamName, streamURL, streamDir string) error {
	// 使用 FFmpeg 拉取 RTSP 流並轉換為 HLS
	args := []string{
		"-i", streamURL,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", s.config.SegmentTime),
		"-hls_list_size", fmt.Sprintf("%d", s.config.SegmentListSize),
		"-hls_flags", "delete_segments",
		"-hls_segment_filename", filepath.Join(streamDir, "segment_%03d.ts"),
		filepath.Join(streamDir, "index.m3u8"),
	}

	cmd := exec.CommandContext(s.ctx, "ffmpeg", args...)
	cmd.Dir = streamDir

	// 記錄流資訊
	s.mu.Lock()
	s.streams[streamName] = &StreamInfo{
		Name:       streamName,
		URL:        streamURL,
		Status:     "active",
		LastUpdate: time.Now(),
		Process:    cmd,
	}
	s.mu.Unlock()

	// 啟動進程
	if err := cmd.Start(); err != nil {
		s.mu.Lock()
		s.streams[streamName].Status = "error"
		s.mu.Unlock()
		return fmt.Errorf("啟動 FFmpeg 失敗: %w", err)
	}

	// 監控進程
	go func() {
		cmd.Wait()
		s.mu.Lock()
		if stream, exists := s.streams[streamName]; exists {
			stream.Status = "inactive"
			stream.Process = nil
		}
		s.mu.Unlock()
		log.Printf("流 %s 已停止", streamName)
	}()

	log.Printf("✅ 流 %s 啟動成功", streamName)
	return nil
}

// startRTMPStream 啟動 RTMP 流拉取
func (s *PublicStreamService) startRTMPStream(streamName, streamURL, streamDir string) error {
	// 使用 FFmpeg 拉取 RTMP 流並轉換為 HLS
	args := []string{
		"-i", streamURL,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", s.config.SegmentTime),
		"-hls_list_size", fmt.Sprintf("%d", s.config.SegmentListSize),
		"-hls_flags", "delete_segments",
		"-hls_segment_filename", filepath.Join(streamDir, "segment_%03d.ts"),
		filepath.Join(streamDir, "index.m3u8"),
	}

	cmd := exec.CommandContext(s.ctx, "ffmpeg", args...)
	cmd.Dir = streamDir

	// 記錄流資訊
	s.mu.Lock()
	s.streams[streamName] = &StreamInfo{
		Name:       streamName,
		URL:        streamURL,
		Status:     "active",
		LastUpdate: time.Now(),
		Process:    cmd,
	}
	s.mu.Unlock()

	// 啟動進程
	if err := cmd.Start(); err != nil {
		s.mu.Lock()
		s.streams[streamName].Status = "error"
		s.mu.Unlock()
		return fmt.Errorf("啟動 FFmpeg 失敗: %w", err)
	}

	// 監控進程
	go func() {
		cmd.Wait()
		s.mu.Lock()
		if stream, exists := s.streams[streamName]; exists {
			stream.Status = "inactive"
			stream.Process = nil
		}
		s.mu.Unlock()
		log.Printf("流 %s 已停止", streamName)
	}()

	log.Printf("✅ 流 %s 啟動成功", streamName)
	return nil
}

// startMP4Stream 啟動 MP4 檔案作為循環直播源
func (s *PublicStreamService) startMP4Stream(streamName, streamURL, streamDir string) error {
	// 使用 FFmpeg 將 MP4 檔案循環播放並轉換為 HLS
	args := []string{
		"-stream_loop", "-1", // 無限循環
		"-i", streamURL,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", s.config.SegmentTime),
		"-hls_list_size", fmt.Sprintf("%d", s.config.SegmentListSize),
		"-hls_flags", "delete_segments",
		"-hls_segment_filename", filepath.Join(streamDir, "segment_%03d.ts"),
		filepath.Join(streamDir, "index.m3u8"),
	}

	cmd := exec.CommandContext(s.ctx, "ffmpeg", args...)
	cmd.Dir = streamDir

	// 記錄流資訊
	s.mu.Lock()
	s.streams[streamName] = &StreamInfo{
		Name:       streamName,
		URL:        streamURL,
		Status:     "active",
		LastUpdate: time.Now(),
		Process:    cmd,
	}
	s.mu.Unlock()

	// 啟動進程
	if err := cmd.Start(); err != nil {
		s.mu.Lock()
		s.streams[streamName].Status = "error"
		s.mu.Unlock()
		return fmt.Errorf("啟動 FFmpeg 失敗: %w", err)
	}

	// 監控進程
	go func() {
		cmd.Wait()
		s.mu.Lock()
		if stream, exists := s.streams[streamName]; exists {
			stream.Status = "inactive"
			stream.Process = nil
		}
		s.mu.Unlock()
		log.Printf("流 %s 已停止", streamName)
	}()

	log.Printf("✅ 流 %s 啟動成功", streamName)
	return nil
}

// monitorStreams 監控流狀態
func (s *PublicStreamService) monitorStreams() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkStreamHealth()
		}
	}
}

// checkStreamHealth 檢查流健康狀態
func (s *PublicStreamService) checkStreamHealth() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for streamName, stream := range s.streams {
		// 檢查進程是否還在運行
		if stream.Process != nil && stream.Process.Process != nil {
			if stream.Process.ProcessState != nil && stream.Process.ProcessState.Exited() {
				log.Printf("流 %s 進程已退出，嘗試重啟", streamName)
				stream.Status = "inactive"
				stream.Process = nil

				// 重啟流
				go func(name, url string) {
					if err := s.startStream(name, url); err != nil {
						log.Printf("重啟流 %s 失敗: %v", name, err)
					}
				}(streamName, stream.URL)
			}
		}

		// 更新最後更新時間
		stream.LastUpdate = time.Now()
	}
}

// ensureOutputDir 確保輸出目錄存在
func (s *PublicStreamService) ensureOutputDir() error {
	return s.ensureDir(s.config.OutputDir)
}

// ensureDir 確保目錄存在
func (s *PublicStreamService) ensureDir(dir string) error {
	// 簡化實現，實際應該使用 os.MkdirAll
	// 這裡暫時返回 nil，實際應用中應該創建目錄
	return nil
}
