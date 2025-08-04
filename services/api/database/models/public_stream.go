package models

import (
	"time"
)

// PublicStream 公開流配置
type PublicStream struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	URL         string    `json:"url" gorm:"not null"`
	Category    string    `json:"category"`
	Type        string    `json:"type" gorm:"default:'hls'"` // "rtmp" 或 "hls"
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (PublicStream) TableName() string {
	return "public_streams"
} 