package services

import (
	"stream-demo/backend/dto"
	"stream-demo/backend/pkg/storage"
	"time"
)

// UserServiceInterface 用戶服務接口
type UserServiceInterface interface {
	Register(username, email, password string) (*dto.UserDTO, error)
	Login(username, password string) (string, *dto.UserDTO, time.Time, error)
	GetUserByID(userID uint) (*dto.UserDTO, error)
	UpdateUser(userID uint, username, email, avatar, bio string) (*dto.UserDTO, error)
	DeleteUser(userID uint) error
}

// VideoServiceInterface 影片服務接口
type VideoServiceInterface interface {
	GenerateUploadURL(userID uint, filename string, fileSize int64) (*storage.PresignedUploadURL, error)
	CreateVideoRecord(userID uint, title, description, s3Key string) (*dto.VideoDTO, error)
	ConfirmUploadOnly(videoID uint, s3Key string) error
	GetVideos(offset, limit int) ([]*dto.VideoDTO, int64, error)
	GetVideoByID(videoID uint) (*dto.VideoDTO, error)
	GetVideosByUserID(userID uint) ([]*dto.VideoDTO, int64, error)
	UpdateVideo(id uint, title string, description string, videoData *dto.VideoDTO) error
	DeleteVideo(id uint) error
	SearchVideos(query string, offset, limit int) ([]*dto.VideoDTO, int64, error)
	LikeVideo(id uint) error
	IncrementViews(id uint) error
	IncrementLikes(id uint) error
}

// LiveServiceInterface 直播服務接口
type LiveServiceInterface interface {
	ListLives(offset, limit int) ([]*dto.LiveDTO, int64, error)
	CreateLive(userID uint, title, description string, startTime time.Time) (*dto.LiveDTO, error)
	GetLiveByID(id uint) (*dto.LiveDTO, error)
	GetLivesByUserID(userID uint) ([]*dto.LiveDTO, int64, error)
	UpdateLive(id uint, title, description string, startTime time.Time) (*dto.LiveDTO, error)
	DeleteLive(id uint) error
	StartLive(id uint) error
	EndLive(id uint) error
	GetStreamKey(id uint) (string, error)
	UpdateViewerCount(id uint, count int64) error
	ToggleChat(id uint, enabled bool) error
	GetActiveLives() ([]*dto.LiveDTO, error)
} 