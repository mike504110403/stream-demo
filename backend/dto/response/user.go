package response

import "time"

// UserResponse 使用者資訊回應
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginResponse 登入回應
type LoginResponse struct {
	Token     string       `json:"token"`
	User      UserResponse `json:"user"`
	ExpiresAt time.Time    `json:"expires_at"`
}

// NewUserResponse 從模型創建使用者回應
func NewUserResponse(user interface{}) *UserResponse {
	// 這裡需要實現從模型到 DTO 的轉換邏輯
	return &UserResponse{}
}

// NewLoginResponse 創建登入回應
func NewLoginResponse(token string, user interface{}, expiresAt time.Time) *LoginResponse {
	return &LoginResponse{
		Token:     token,
		User:      *NewUserResponse(user),
		ExpiresAt: expiresAt,
	}
}
