package request

// CreatePaymentRequest 創建支付請求
type CreatePaymentRequest struct {
	Amount        float64 `json:"amount" binding:"required"`
	Currency      string  `json:"currency" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	Description   string  `json:"description"`
}

// ProcessPaymentRequest 處理支付請求
type ProcessPaymentRequest struct {
	PaymentMethod string `json:"payment_method" binding:"required"`
	TransactionID string `json:"transaction_id" binding:"required"`
}

// RefundPaymentRequest 退款請求
type RefundPaymentRequest struct {
	Reason string `json:"reason" binding:"required"`
}
