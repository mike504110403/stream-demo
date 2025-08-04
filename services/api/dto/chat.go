package dto

import (
	"encoding/json"
	"time"
)

// ChatMessageDTO 聊天訊息傳輸物件
type ChatMessageDTO struct {
	Type      string    `json:"type"` // message, join, leave, system
	LiveID    uint      `json:"live_id"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// MarshalJSON 自定義 JSON 序列化
func (m *ChatMessageDTO) MarshalJSON() ([]byte, error) {
	type Alias ChatMessageDTO
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
	}{
		Alias:     (*Alias)(m),
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	})
}
