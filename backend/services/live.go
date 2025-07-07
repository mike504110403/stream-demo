package services

import (
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/dto"
	postgresqlRepo "stream-demo/backend/repositories/postgresql"
	"time"
)

type LiveService struct {
	Conf      *config.Config
	Repo      *postgresqlRepo.PostgreSQLRepo
	RepoSlave *postgresqlRepo.PostgreSQLRepo
}

func NewLiveService(conf *config.Config) *LiveService {
	return &LiveService{
		Conf:      conf,
		Repo:      postgresqlRepo.NewPostgreSQLRepo(conf.DB["master"]),
		RepoSlave: postgresqlRepo.NewPostgreSQLRepo(conf.DB["slave"]),
	}
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
	live := &models.Live{
		Title:       title,
		Description: description,
		UserID:      userID,
		Status:      "scheduled",
		StartTime:   startTime,
		ChatEnabled: true,
	}

	if err := s.Repo.CreateLive(live); err != nil {
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
