package services

import (
	"fmt"
	"log"
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/dto"
	"stream-demo/backend/pkg/media"
	postgresqlRepo "stream-demo/backend/repositories/postgresql"
	"time"
)

type LiveService struct {
	Conf      *config.Config
	Repo      *postgresqlRepo.PostgreSQLRepo
	RepoSlave *postgresqlRepo.PostgreSQLRepo
	LiveMedia media.LiveService
}

func NewLiveService(conf *config.Config) (*LiveService, error) {
	// 根據配置創建直播媒體服務
	var liveMedia media.LiveService
	var err error

	if conf.Live.Enabled {
		switch conf.Live.Type {
		case "local":
			localConfig := media.LocalLiveConfig{
				RTMPServer:        conf.Live.Local.RTMPServer,
				RTMPServerPort:    conf.Live.Local.RTMPServerPort,
				TranscoderEnabled: conf.Live.Local.TranscoderEnabled,
				HLSOutputDir:      conf.Live.Local.HLSOutputDir,
				HTTPPort:          conf.Live.Local.HTTPPort,
			}
			liveMedia, err = media.LiveServiceFactory("local", localConfig)
		case "cloud":
			cloudConfig := media.CloudLiveConfig{
				Provider:         conf.Live.Cloud.Provider,
				RTMPIngestURL:    conf.Live.Cloud.RTMPIngestURL,
				HLSPlaybackURL:   conf.Live.Cloud.HLSPlaybackURL,
				APIKey:           conf.Live.Cloud.APIKey,
				APISecret:        conf.Live.Cloud.APISecret,
				TranscodeEnabled: conf.Live.Cloud.TranscodeEnabled,
			}
			liveMedia, err = media.LiveServiceFactory("cloud", cloudConfig)
		default:
			return nil, fmt.Errorf("不支援的直播服務類型: %s", conf.Live.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("創建直播媒體服務失敗: %w", err)
		}
	}

	service := &LiveService{
		Conf:      conf,
		Repo:      postgresqlRepo.NewPostgreSQLRepo(conf.DB["master"]),
		RepoSlave: postgresqlRepo.NewPostgreSQLRepo(conf.DB["slave"]),
		LiveMedia: liveMedia,
	}

	// 啟動直播服務
	if liveMedia != nil {
		if err := service.Start(); err != nil {
			return nil, fmt.Errorf("啟動直播服務失敗: %w", err)
		}
	}

	return service, nil
}

// Start 啟動直播服務
func (s *LiveService) Start() error {
	if s.LiveMedia == nil {
		log.Println("⚠️ 直播媒體服務未配置，跳過啟動")
		return nil
	}

	log.Println("🚀 啟動直播服務...")
	return s.LiveMedia.Start()
}

// Stop 停止直播服務
func (s *LiveService) Stop() error {
	if s.LiveMedia == nil {
		return nil
	}

	log.Println("🛑 停止直播服務...")
	return s.LiveMedia.Stop()
}

// generateStreamKey 生成串流金鑰
func generateStreamKey(userID uint) string {
	// 簡單的串流金鑰生成邏輯
	// 實際應用中應該使用更安全的方式
	return fmt.Sprintf("user_%d_%d", userID, time.Now().Unix())
}

func (s *LiveService) ListLives(offset, limit int) ([]*dto.LiveDTO, int64, error) {
	// 這裡應該查詢所有直播，不是特定用戶的
	lives, total, err := s.RepoSlave.FindLiveByUserID(0)
	if err != nil {
		return nil, 0, err
	}

	// 轉換為 DTO
	liveDTOs := make([]*dto.LiveDTO, len(lives))
	for i, live := range lives {
		liveDTOs[i] = &dto.LiveDTO{
			ID:          live.ID,
			Title:       live.Title,
			Description: live.Description,
			UserID:      live.UserID,
			Status:      live.Status,
			StartTime:   live.StartTime,
			EndTime:     live.EndTime,
			ViewerCount: live.ViewerCount,
			StreamKey:   live.StreamKey,
			ChatEnabled: live.ChatEnabled,
			CreatedAt:   live.CreatedAt,
			UpdatedAt:   live.UpdatedAt,
		}
	}

	return liveDTOs, total, nil
}

func (s *LiveService) CreateLive(userID uint, title, description string, startTime time.Time) (*dto.LiveDTO, error) {
	// 生成唯一的串流金鑰
	streamKey := generateStreamKey(userID)

	live := &models.Live{
		Title:       title,
		Description: description,
		UserID:      userID,
		Status:      "scheduled",
		StartTime:   startTime,
		StreamKey:   streamKey,
		ChatEnabled: true,
	}

	if err := s.Repo.CreateLive(live); err != nil {
		return nil, err
	}

	// 如果有直播媒體服務，獲取推流和播放 URL
	if s.LiveMedia != nil {
		pushURL, err := s.LiveMedia.GetPushURL(streamKey)
		if err != nil {
			log.Printf("獲取推流 URL 失敗: %v", err)
		} else {
			live.PushURL = pushURL
		}

		streamURL, err := s.LiveMedia.GetStreamURL(streamKey)
		if err != nil {
			log.Printf("獲取播放 URL 失敗: %v", err)
		} else {
			live.StreamURL = streamURL
		}

		// 更新資料庫中的 URL
		if live.PushURL != "" || live.StreamURL != "" {
			s.Repo.UpdateLive(live)
		}
	}

	return &dto.LiveDTO{
		ID:          live.ID,
		Title:       live.Title,
		Description: live.Description,
		UserID:      live.UserID,
		Status:      live.Status,
		StartTime:   live.StartTime,
		EndTime:     live.EndTime,
		ViewerCount: live.ViewerCount,
		StreamKey:   live.StreamKey,
		PushURL:     live.PushURL,
		StreamURL:   live.StreamURL,
		ChatEnabled: live.ChatEnabled,
		CreatedAt:   live.CreatedAt,
		UpdatedAt:   live.UpdatedAt,
	}, nil
}

func (s *LiveService) GetLiveByID(id uint) (*dto.LiveDTO, error) {
	live, err := s.RepoSlave.FindLiveByID(id)
	if err != nil {
		return nil, err
	}

	return &dto.LiveDTO{
		ID:          live.ID,
		Title:       live.Title,
		Description: live.Description,
		UserID:      live.UserID,
		Status:      live.Status,
		StartTime:   live.StartTime,
		EndTime:     live.EndTime,
		ViewerCount: live.ViewerCount,
		StreamKey:   live.StreamKey,
		ChatEnabled: live.ChatEnabled,
		CreatedAt:   live.CreatedAt,
		UpdatedAt:   live.UpdatedAt,
	}, nil
}

func (s *LiveService) GetLivesByUserID(userID uint) ([]*dto.LiveDTO, int64, error) {
	lives, total, err := s.RepoSlave.FindLiveByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 轉換為 DTO
	liveDTOs := make([]*dto.LiveDTO, len(lives))
	for i, live := range lives {
		liveDTOs[i] = &dto.LiveDTO{
			ID:          live.ID,
			Title:       live.Title,
			Description: live.Description,
			UserID:      live.UserID,
			Status:      live.Status,
			StartTime:   live.StartTime,
			EndTime:     live.EndTime,
			ViewerCount: live.ViewerCount,
			StreamKey:   live.StreamKey,
			ChatEnabled: live.ChatEnabled,
			CreatedAt:   live.CreatedAt,
			UpdatedAt:   live.UpdatedAt,
		}
	}

	return liveDTOs, total, nil
}

func (s *LiveService) UpdateLive(id uint, title, description string, startTime time.Time) (*dto.LiveDTO, error) {
	live, err := s.Repo.FindLiveByID(id)
	if err != nil {
		return nil, err
	}

	live.Title = title
	live.Description = description
	live.StartTime = startTime

	if err := s.Repo.UpdateLive(live); err != nil {
		return nil, err
	}

	return &dto.LiveDTO{
		ID:          live.ID,
		Title:       live.Title,
		Description: live.Description,
		UserID:      live.UserID,
		Status:      live.Status,
		StartTime:   live.StartTime,
		EndTime:     live.EndTime,
		ViewerCount: live.ViewerCount,
		StreamKey:   live.StreamKey,
		ChatEnabled: live.ChatEnabled,
		CreatedAt:   live.CreatedAt,
		UpdatedAt:   live.UpdatedAt,
	}, nil
}

func (s *LiveService) DeleteLive(id uint) error {
	return s.Repo.DeleteLive(id)
}

func (s *LiveService) StartLive(id uint) error {
	live, err := s.Repo.FindLiveByID(id)
	if err != nil {
		return err
	}

	live.Status = "live"
	live.StartTime = time.Now()

	return s.Repo.UpdateLive(live)
}

func (s *LiveService) EndLive(id uint) error {
	live, err := s.Repo.FindLiveByID(id)
	if err != nil {
		return err
	}

	live.Status = "ended"
	live.EndTime = time.Now()

	return s.Repo.UpdateLive(live)
}

func (s *LiveService) GetStreamKey(id uint) (string, error) {
	live, err := s.RepoSlave.FindLiveByID(id)
	if err != nil {
		return "", err
	}

	return live.StreamKey, nil
}

func (s *LiveService) UpdateViewerCount(id uint, count int64) error {
	live, err := s.Repo.FindLiveByID(id)
	if err != nil {
		return err
	}

	live.ViewerCount = count

	return s.Repo.UpdateLive(live)
}

func (s *LiveService) ToggleChat(id uint, enabled bool) error {
	live, err := s.Repo.FindLiveByID(id)
	if err != nil {
		return err
	}

	live.ChatEnabled = enabled

	return s.Repo.UpdateLive(live)
}

func (s *LiveService) GetActiveLives() ([]*dto.LiveDTO, error) {
	lives, err := s.RepoSlave.FindActiveLive()
	if err != nil {
		return nil, err
	}

	// 轉換為 DTO
	liveDTOs := make([]*dto.LiveDTO, len(lives))
	for i, live := range lives {
		liveDTOs[i] = &dto.LiveDTO{
			ID:          live.ID,
			Title:       live.Title,
			Description: live.Description,
			UserID:      live.UserID,
			Status:      live.Status,
			StartTime:   live.StartTime,
			EndTime:     live.EndTime,
			ViewerCount: live.ViewerCount,
			StreamKey:   live.StreamKey,
			ChatEnabled: live.ChatEnabled,
			CreatedAt:   live.CreatedAt,
			UpdatedAt:   live.UpdatedAt,
		}
	}

	return liveDTOs, nil
}
