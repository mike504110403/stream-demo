package handlers

import (
	"net/http"
	"strconv"
	"stream-demo/backend/dto"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

// UserHandler 用戶處理器
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler 創建用戶處理器實例
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用戶註冊
func (h *UserHandler) Register(c *gin.Context) {
	var registerDTO dto.UserRegisterDTO
	if err := c.ShouldBindJSON(&registerDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Register(registerDTO.Username, registerDTO.Email, registerDTO.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login 用戶登入
func (h *UserHandler) Login(c *gin.Context) {
	var loginDTO dto.UserLoginDTO
	if err := c.ShouldBindJSON(&loginDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err	or": err.Error()})
		return
	}

	token, user, expiresAt, err := h.userService.Login(loginDTO.Email, loginDTO.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token, "expires_at": expiresAt})
}

// GetUser 獲取用戶資訊
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的用戶 ID"})
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用戶不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser 更新用戶資訊
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的用戶 ID"})
		return
	}

	var updateDTO dto.UserUpdateDTO
	if err := c.ShouldBindJSON(&updateDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(uint(id), updateDTO.Username, updateDTO.Email, updateDTO.Avatar, updateDTO.Bio)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser 刪除用戶
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的用戶 ID"})
		return
	}

	if err := h.userService.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
