package dto

import "time"

// PaymentDTO 支付資料傳輸物件
type PaymentDTO struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	Username      string    `json:"username"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	TransactionID string    `json:"transaction_id"`
	Description   string    `json:"description"`
	RefundReason  string    `json:"refund_reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PaymentCreateDTO 建立支付請求
type PaymentCreateDTO struct {
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,len=3"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	Description   string  `json:"description" binding:"max=500"`
}

// PaymentRefundDTO 退款請求
type PaymentRefundDTO struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// PaymentListDTO 支付列表回應
type PaymentListDTO struct {
	Total    int64        `json:"total"`
	Payments []PaymentDTO `json:"payments"`
}
