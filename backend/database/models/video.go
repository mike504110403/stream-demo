package models

import "time"

// Video 影片模型
type Video struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	UserID      uint   `json:"user_id" gorm:"not null;index:idx_videos_user_status,priority:1;index:idx_videos_user_created,priority:1"`

	// 原始影片資訊
	OriginalURL  string `json:"original_url" gorm:"size:500;not null"`
	OriginalKey  string `json:"original_key" gorm:"size:500"`
	ThumbnailURL string `json:"thumbnail_url" gorm:"size:500"`

	// HLS串流資訊
	HLSMasterURL string `json:"hls_master_url" gorm:"size:500"`
	HLSKey       string `json:"hls_key" gorm:"size:500"`

	// 影片屬性
	Duration       int    `json:"duration" gorm:"default:0"`      // 秒數
	FileSize       int64  `json:"file_size" gorm:"default:0"`     // 位元組
	OriginalFormat string `json:"original_format" gorm:"size:10"` // mp4, avi等

	// 狀態管理
	Status string `json:"status" gorm:"size:20;not null;index:idx_videos_user_status,priority:2;index:idx_videos_status_created,priority:1"`
	// 狀態: uploading, processing, transcoding, ready, failed
	ProcessingProgress int    `json:"processing_progress" gorm:"default:0"` // 0-100
	ErrorMessage       string `json:"error_message" gorm:"size:500"`

	// 統計資料
	Views     int64     `json:"views" gorm:"default:0"`
	Likes     int64     `json:"likes" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_videos_user_created,priority:2;index:idx_videos_status_created,priority:2"`
	UpdatedAt time.Time `json:"updated_at"`

	// 關聯關係
	User           *User          `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	VideoQualities []VideoQuality `json:"video_qualities,omitempty" gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`
}

// VideoQuality 影片品質資訊模型
type VideoQuality struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	VideoID  uint   `json:"video_id" gorm:"not null;index"`
	Quality  string `json:"quality" gorm:"size:10;not null"` // 360p, 480p, 720p, 1080p
	Width    int    `json:"width" gorm:"not null"`
	Height   int    `json:"height" gorm:"not null"`
	Bitrate  int    `json:"bitrate" gorm:"not null"`
	FileURL  string `json:"file_url" gorm:"size:500;not null"`
	FileKey  string `json:"file_key" gorm:"size:500"`
	FileSize int64  `json:"file_size" gorm:"default:0"`
	Status   string `json:"status" gorm:"size:20;default:pending"` // pending, processing, ready, failed

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 關聯關係
	Video *Video `json:"video,omitempty" gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`
}
