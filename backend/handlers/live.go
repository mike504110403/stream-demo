package handlers

import (
	"net/http"
	"strconv"
	"stream-demo/backend/dto"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

// LiveHandler 直播處理器
type LiveHandler struct {
	liveService *services.LiveService
}

// NewLiveHandler 創建直播處理器實例
func NewLiveHandler(liveService *services.LiveService) *LiveHandler {
	return &LiveHandler{
		liveService: liveService,
	}
}

// CreateLive 創建直播
func (h *LiveHandler) CreateLive(c *gin.Context) {
	userID := c.GetUint("user_id")

	var createDTO dto.LiveCreateDTO
	if err := c.ShouldBindJSON(&createDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	live, err := h.liveService.CreateLive(userID, createDTO.Title, createDTO.Description, createDTO.StartTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, live)
}

// GetLive 獲取直播資訊
func (h *LiveHandler) GetLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的直播 ID"})
		return
	}

	live, err := h.liveService.GetLiveByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "直播不存在"})
		return
	}

	c.JSON(http.StatusOK, live)
}

// GetUserLives 獲取用戶的直播列表
func (h *LiveHandler) GetUserLives(c *gin.Context) {
	userID := c.GetUint("user_id")

	lives, _, err := h.liveService.GetLivesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lives)
}

// UpdateLive 更新直播資訊
func (h *LiveHandler) UpdateLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的直播 ID"})
		return
	}

	var updateDTO dto.LiveUpdateDTO
	if err := c.ShouldBindJSON(&updateDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	live, err := h.liveService.UpdateLive(uint(id), updateDTO.Title, updateDTO.Description, updateDTO.StartTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, live)
}

// DeleteLive 刪除直播
func (h *LiveHandler) DeleteLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的直播 ID"})
		return
	}

	if err := h.liveService.DeleteLive(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// StartLive 開始直播
func (h *LiveHandler) StartLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的直播 ID"})
		return
	}

	err = h.liveService.StartLive(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// EndLive 結束直播
func (h *LiveHandler) EndLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的直播 ID"})
		return
	}

	err = h.liveService.EndLive(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
