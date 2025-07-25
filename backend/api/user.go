package api

import (
	"net/http"
	"stream-demo/backend/dto/request"
	"stream-demo/backend/dto/response"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

// UserHandler 用戶處理器
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler 創建用戶處理器
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register 用戶註冊
func (h *UserHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "參數格式錯誤"))
		return
	}
	user, err := h.userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewUserResponse(user)))
}

// Login 用戶登入
func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "參數格式錯誤"))
		return
	}
	token, user, expiresAt, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"token":      token,
		"user":       response.NewUserResponse(user),
		"expires_at": expiresAt,
	}))
}

// GetUser 獲取用戶資訊
func (h *UserHandler) GetUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, "未登入"))
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, response.NewErrorResponse(404, "使用者不存在"))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewUserResponse(user)))
}

// UpdateUser 更新用戶資訊
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, "未登入"))
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	user, err := h.userService.UpdateUser(userID.(uint), req.Username, req.Email, req.Avatar, req.Bio)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewUserResponse(user)))
}

// DeleteUser 刪除用戶
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, "未登入"))
		return
	}

	if err := h.userService.DeleteUser(userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
