package models

import (
	"time"
)

// UserLiveSession 用戶直播記錄表
type UserLiveSession struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        int        `gorm:"not null;index" json:"user_id"`
	RoomID        string     `gorm:"not null;index;size:255" json:"room_id"`
	Title         string     `gorm:"size:255" json:"title"`
	Description   string     `gorm:"type:text" json:"description"`
	StreamKey     string     `gorm:"size:255" json:"stream_key"`
	Status        string     `gorm:"size:50;default:'created'" json:"status"` // created, waiting, live, paused, ended, cancelled
	StartedAt     *time.Time `json:"started_at"`
	EndedAt       *time.Time `json:"ended_at"`
	Duration      int        `gorm:"default:0" json:"duration"` // 直播時長(秒)
	PeakViewers   int        `gorm:"default:0" json:"peak_viewers"`
	TotalViewers  int        `gorm:"default:0" json:"total_viewers"`
	TotalMessages int        `gorm:"default:0" json:"total_messages"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ChatMessageHistory 聊天消息歷史表
type ChatMessageHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	RoomID      string    `gorm:"not null;index;size:255" json:"room_id"`
	UserID      int       `gorm:"not null;index" json:"user_id"`
	Message     string    `gorm:"type:text;not null" json:"message"`
	MessageType string    `gorm:"size:50;default:'text'" json:"message_type"` // text, image, gift, system
	CreatedAt   time.Time `json:"created_at"`
}

// UserLiveStats 用戶直播統計表
type UserLiveStats struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        int       `gorm:"not null;uniqueIndex" json:"user_id"`
	TotalSessions int       `gorm:"default:0" json:"total_sessions"`
	TotalDuration int       `gorm:"default:0" json:"total_duration"` // 總直播時長(秒)
	TotalViewers  int       `gorm:"default:0" json:"total_viewers"`
	TotalMessages int       `gorm:"default:0" json:"total_messages"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (UserLiveSession) TableName() string {
	return "user_live_sessions"
}

func (ChatMessageHistory) TableName() string {
	return "chat_message_history"
}

func (UserLiveStats) TableName() string {
	return "user_live_stats"
}
