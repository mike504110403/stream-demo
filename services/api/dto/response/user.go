package response

import (
	"stream-demo/backend/database/models"
	"stream-demo/backend/dto"
	"time"
)

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
	if user == nil {
		return &UserResponse{}
	}

	// 類型斷言 - 處理 *models.User
	if u, ok := user.(*models.User); ok {
		return &UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			Role:      "user",   // 默認角色
			Status:    "active", // 默認狀態
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	// 類型斷言 - 處理 *dto.UserDTO
	if u, ok := user.(*dto.UserDTO); ok {
		return &UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			Role:      "user",   // 默認角色
			Status:    "active", // 默認狀態
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	// 如果不是支持的類型，返回空結構
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
