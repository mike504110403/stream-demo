package models

import "time"

// Live 直播模型
type Live struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:500"`
	UserID      uint      `json:"user_id" gorm:"not null;index:idx_lives_user_status,priority:1"`
	Status      string    `json:"status" gorm:"size:20;not null;index:idx_lives_user_status,priority:2;index:idx_lives_status_start,priority:1"` // scheduled, live, ended
	StartTime   time.Time `json:"start_time" gorm:"index:idx_lives_status_start,priority:2"`
	EndTime     time.Time `json:"end_time"`
	StreamKey   string    `json:"stream_key" gorm:"size:100;uniqueIndex"`
	ViewerCount int64     `json:"viewer_count" gorm:"default:0"`
	ChatEnabled bool      `json:"chat_enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 關聯關係
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ChatMessages []ChatMessage `json:"chat_messages,omitempty" gorm:"foreignKey:LiveID;constraint:OnDelete:CASCADE"`
}
