package dto

import "time"

// LiveDTO 直播資料傳輸物件
type LiveDTO struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	ViewerCount int64     `json:"viewer_count"`
	StreamKey   string    `json:"stream_key"`
	PushURL     string    `json:"push_url"`
	StreamURL   string    `json:"stream_url"`
	ChatEnabled bool      `json:"chat_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LiveCreateDTO 建立直播請求
type LiveCreateDTO struct {
	Title       string    `json:"title" binding:"required,max=100"`
	Description string    `json:"description" binding:"max=500"`
	StartTime   time.Time `json:"start_time" binding:"required"`
}

// LiveUpdateDTO 更新直播請求
type LiveUpdateDTO struct {
	Title       string    `json:"title" binding:"omitempty,max=100"`
	Description string    `json:"description" binding:"omitempty,max=500"`
	StartTime   time.Time `json:"start_time" binding:"omitempty"`
	ChatEnabled bool      `json:"chat_enabled"`
}

// LiveListDTO 直播列表回應
type LiveListDTO struct {
	Total int64     `json:"total"`
	Lives []LiveDTO `json:"lives"`
}
