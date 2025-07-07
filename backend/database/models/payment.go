package models

import "time"

// Payment 支付模型
type Payment struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"not null;index:idx_payments_user_status,priority:1"`
	Amount        float64   `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency      string    `json:"currency" gorm:"size:3;not null"`
	Status        string    `json:"status" gorm:"size:20;not null;index:idx_payments_user_status,priority:2;index:idx_payments_status_created,priority:1"` // pending, completed, failed, refunded
	PaymentMethod string    `json:"payment_method" gorm:"size:50"`
	TransactionID string    `json:"transaction_id" gorm:"size:100;uniqueIndex"`
	Description   string    `json:"description" gorm:"size:500"`
	RefundReason  string    `json:"refund_reason" gorm:"size:500"`
	CreatedAt     time.Time `json:"created_at" gorm:"index:idx_payments_status_created,priority:2"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 關聯關係
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
