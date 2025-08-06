package models

import "time"

// User 用戶模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password  string    `json:"-" gorm:"size:100;not null"` // 密碼不返回給前端
	Avatar    string    `json:"avatar" gorm:"size:255"`
	Bio       string    `json:"bio" gorm:"size:500"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 關聯關係
	Videos       []Video       `json:"videos,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Lives        []Live        `json:"lives,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Payments     []Payment     `json:"payments,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ChatMessages []ChatMessage `json:"chat_messages,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
