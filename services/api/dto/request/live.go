package request

import "time"

// CreateLiveRequest 創建直播請求
type CreateLiveRequest struct {
	Title       string    `json:"title" binding:"required,min=1,max=100"`
	Description string    `json:"description" binding:"max=500"`
	StartTime   time.Time `json:"start_time" binding:"required"`
}

// UpdateLiveRequest 更新直播資訊請求
type UpdateLiveRequest struct {
	Title       string    `json:"title" binding:"omitempty,min=1,max=100"`
	Description string    `json:"description" binding:"omitempty,max=500"`
	StartTime   time.Time `json:"start_time" binding:"omitempty"`
}

// ToggleChatRequest 切換聊天功能請求
type ToggleChatRequest struct {
	Enabled bool `form:"enabled" binding:"required"`
}
