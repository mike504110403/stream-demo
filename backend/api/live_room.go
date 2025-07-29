package api

import (
	"fmt"
	"net/http"
	"strconv"

	"stream-demo/backend/services"
	"stream-demo/backend/utils"

	"github.com/gin-gonic/gin"
)

// LiveRoomHandler 直播間處理器
type LiveRoomHandler struct {
	liveRoomService *services.LiveRoomService
}

// NewLiveRoomHandler 創建直播間處理器
func NewLiveRoomHandler(liveRoomService *services.LiveRoomService) *LiveRoomHandler {
	return &LiveRoomHandler{
		liveRoomService: liveRoomService,
	}
}

// getUserIDFromContext 從上下文獲取用戶ID
func getUserIDFromContext(c *gin.Context) (int, error) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("未授權")
	}

	// 處理不同類型的 user_id
	switch v := userIDInterface.(type) {
	case int:
		return v, nil
	case uint:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		// 添加調試信息
		utils.LogError("用戶ID類型錯誤: %T, 值: %v", userIDInterface, userIDInterface)
		return 0, fmt.Errorf("用戶ID類型錯誤: %T", userIDInterface)
	}
}

// CreateRoom 創建直播間
func (h *LiveRoomHandler) CreateRoom(c *gin.Context) {
	// 從 JWT 獲取用戶ID
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請求參數錯誤", "details": err.Error()})
		return
	}

	// 創建直播間
	room, err := h.liveRoomService.CreateRoom(userID, req.Title, req.Description)
	if err != nil {
		utils.LogError("創建直播間失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "創建直播間失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "直播間創建成功",
		"data":    room,
	})
}

// GetActiveRooms 獲取活躍直播間列表
func (h *LiveRoomHandler) GetActiveRooms(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	rooms, err := h.liveRoomService.GetActiveRooms(limit)
	if err != nil {
		utils.LogError("獲取直播間列表失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取直播間列表失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "獲取成功",
		"data":    rooms,
		"total":   len(rooms),
	})
}

// GetAllRooms 獲取所有直播間列表（包括已結束的）
func (h *LiveRoomHandler) GetAllRooms(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	rooms, err := h.liveRoomService.GetAllRooms(limit)
	if err != nil {
		utils.LogError("獲取所有直播間列表失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取直播間列表失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "獲取成功",
		"data":    rooms,
		"total":   len(rooms),
	})
}

// GetRoomByID 根據ID獲取直播間信息
func (h *LiveRoomHandler) GetRoomByID(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "房間ID不能為空"})
		return
	}

	room, err := h.liveRoomService.GetRoomByID(roomID)
	if err != nil {
		utils.LogError("獲取直播間信息失敗: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "直播間不存在", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "獲取成功",
		"data":    room,
	})
}

// JoinRoom 加入直播間
func (h *LiveRoomHandler) JoinRoom(c *gin.Context) {
	// 從 JWT 獲取用戶ID
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "房間ID不能為空"})
		return
	}

	err = h.liveRoomService.JoinRoom(roomID, userID)
	if err != nil {
		utils.LogError("加入直播間失敗: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "加入直播間失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "加入直播間成功",
		"room_id": roomID,
	})
}

// LeaveRoom 離開直播間
func (h *LiveRoomHandler) LeaveRoom(c *gin.Context) {
	// 從 JWT 獲取用戶ID
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "房間ID不能為空"})
		return
	}

	err = h.liveRoomService.LeaveRoom(roomID, userID)
	if err != nil {
		utils.LogError("離開直播間失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "離開直播間失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "離開直播間成功",
		"room_id": roomID,
	})
}

// StartLive 開始直播
func (h *LiveRoomHandler) StartLive(c *gin.Context) {
	// 從 JWT 獲取用戶ID
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "房間ID不能為空"})
		return
	}

	err = h.liveRoomService.StartLive(roomID, userID)
	if err != nil {
		utils.LogError("開始直播失敗: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "開始直播失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "直播開始成功",
		"room_id": roomID,
	})
}

// EndLive 結束直播
func (h *LiveRoomHandler) EndLive(c *gin.Context) {
	// 從 JWT 獲取用戶ID
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "房間ID不能為空"})
		return
	}

	err = h.liveRoomService.EndLive(roomID, userID)
	if err != nil {
		utils.LogError("結束直播失敗: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "結束直播失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "直播結束成功",
		"room_id": roomID,
	})
}

// CloseRoom 關閉直播間
func (h *LiveRoomHandler) CloseRoom(c *gin.Context) {
	// 從 JWT 獲取用戶ID
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "房間ID不能為空"})
		return
	}

	err = h.liveRoomService.CloseRoom(roomID, userID)
	if err != nil {
		utils.LogError("關閉直播間失敗: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "關閉直播間失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "直播間關閉成功",
		"room_id": roomID,
	})
}

// GetUserRole 獲取用戶在房間中的角色
func (h *LiveRoomHandler) GetUserRole(c *gin.Context) {
	roomID := c.Param("id")
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登入"})
		return
	}

	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "房間ID不能為空"})
		return
	}

	role, err := h.liveRoomService.GetUserRole(roomID, userID)
	if err != nil {
		utils.LogError("獲取用戶角色失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取用戶角色失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "獲取成功",
		"data": gin.H{
			"role": role,
		},
	})
}
