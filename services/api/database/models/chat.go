package models

import "time"

// ChatMessage 聊天訊息模型
type ChatMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	LiveID    uint      `json:"live_id" gorm:"not null;index:idx_chat_live_created,priority:1"`
	UserID    uint      `json:"user_id" gorm:"not null;index:idx_chat_user_created,priority:1"`
	Username  string    `json:"username" gorm:"size:50;not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Type      string    `json:"type" gorm:"size:20;default:text"` // text, system, gift, etc.
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_chat_live_created,priority:2;index:idx_chat_user_created,priority:2"`

	// 關聯關係
	Live *Live `json:"live,omitempty" gorm:"foreignKey:LiveID;constraint:OnDelete:CASCADE"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
