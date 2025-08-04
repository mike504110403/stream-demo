package response

import "time"

// PaymentResponse 支付資訊回應
type PaymentResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PaymentListResponse 支付列表回應
type PaymentListResponse struct {
	Total    int64             `json:"total"`
	Payments []PaymentResponse `json:"payments"`
}

// PaymentStatusResponse 支付狀態回應
type PaymentStatusResponse struct {
	Status string `json:"status"`
}

// NewPaymentResponse 從模型創建支付回應
func NewPaymentResponse(payment interface{}) *PaymentResponse {
	// 這裡需要實現從模型到 DTO 的轉換邏輯
	return &PaymentResponse{}
}

// NewPaymentListResponse 創建支付列表回應
func NewPaymentListResponse(total int64, payments []interface{}) *PaymentListResponse {
	paymentResponses := make([]PaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = *NewPaymentResponse(payment)
	}
	return &PaymentListResponse{
		Total:    total,
		Payments: paymentResponses,
	}
}

// NewPaymentStatusResponse 創建支付狀態回應
func NewPaymentStatusResponse(status string) *PaymentStatusResponse {
	return &PaymentStatusResponse{
		Status: status,
	}
}
