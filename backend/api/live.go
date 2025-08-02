package api

import (
	"net/http"
	"strconv"
	"stream-demo/backend/dto/request"
	"stream-demo/backend/dto/response"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

type LiveHandler struct {
	liveService services.LiveServiceInterface
}

func NewLiveHandler(liveService services.LiveServiceInterface) *LiveHandler {
	return &LiveHandler{liveService: liveService}
}

func (h *LiveHandler) ListLives(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	lives, total, err := h.liveService.ListLives(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewListResponse(total, lives)))
}

func (h *LiveHandler) CreateLive(c *gin.Context) {
	var req request.CreateLiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	// 從 context 獲取使用者 ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, "未登入"))
		return
	}

	live, err := h.liveService.CreateLive(userID.(uint), req.Title, req.Description, req.StartTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.NewSuccessResponse(response.NewLiveResponse(live)))
}

func (h *LiveHandler) GetLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的直播 ID"))
		return
	}

	live, err := h.liveService.GetLiveByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, response.NewErrorResponse(404, "直播不存在"))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewLiveResponse(live)))
}

func (h *LiveHandler) UpdateLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的直播 ID"))
		return
	}

	var req request.UpdateLiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	live, err := h.liveService.UpdateLive(uint(id), req.Title, req.Description, req.StartTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewLiveResponse(live)))
}

func (h *LiveHandler) DeleteLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的直播 ID"))
		return
	}

	if err := h.liveService.DeleteLive(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LiveHandler) StartLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的直播 ID"))
		return
	}

	if err := h.liveService.StartLive(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

func (h *LiveHandler) EndLive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的直播 ID"))
		return
	}

	if err := h.liveService.EndLive(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

func (h *LiveHandler) GetStreamKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的直播 ID"))
		return
	}

	streamKey, err := h.liveService.GetStreamKey(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewStreamKeyResponse(streamKey)))
}

func (h *LiveHandler) ToggleChat(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的直播 ID"))
		return
	}

	var req request.ToggleChatRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	if err := h.liveService.ToggleChat(uint(id), req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

func (h *LiveHandler) GetUserLives(c *gin.Context) {
	// 獲取用戶 ID 參數
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的用戶 ID"))
		return
	}

	lives, total, err := h.liveService.GetLivesByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewListResponse(total, lives)))
}
