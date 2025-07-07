package handlers

import (
	"net/http"
	"strconv"
	"stream-demo/backend/dto"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

// VideoHandler 影片處理器
type VideoHandler struct {
	videoService *services.VideoService
}

// NewVideoHandler 創建影片處理器實例
func NewVideoHandler(videoService *services.VideoService) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
	}
}

// UploadVideo 上傳影片
func (h *VideoHandler) UploadVideo(c *gin.Context) {
	userID := c.GetUint("user_id")

	var createDTO dto.VideoCreateDTO
	if err := c.ShouldBindJSON(&createDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 獲取上傳的檔案
	videoFile, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請上傳影片檔案"})
		return
	}

	// 獲取縮圖檔案
	thumbnailFile, err := c.FormFile("thumbnail")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請上傳縮圖檔案"})
		return
	}

	// 儲存檔案
	videoPath := "uploads/videos/" + videoFile.Filename
	thumbnailPath := "uploads/thumbnails/" + thumbnailFile.Filename

	if err := c.SaveUploadedFile(videoFile, videoPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "儲存影片檔案失敗"})
		return
	}

	if err := c.SaveUploadedFile(thumbnailFile, thumbnailPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "儲存縮圖檔案失敗"})
		return
	}

	err = h.videoService.UpdateVideo(userID, createDTO.Title, createDTO.Description, &dto.VideoDTO{
		Title:        createDTO.Title,
		Description:  createDTO.Description,
		OriginalURL:  videoPath,
		ThumbnailURL: thumbnailPath,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// GetVideo 獲取影片資訊
func (h *VideoHandler) GetVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的影片 ID"})
		return
	}

	video, err := h.videoService.GetVideoByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "影片不存在"})
		return
	}

	// 增加觀看次數
	if err := h.videoService.IncrementViews(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新觀看次數失敗"})
		return
	}

	c.JSON(http.StatusOK, video)
}

// GetUserVideos 獲取用戶的影片列表
func (h *VideoHandler) GetUserVideos(c *gin.Context) {
	userID := c.GetUint("user_id")

	videos, total, err := h.videoService.GetVideosByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"videos": videos, "total": total})
}

// UpdateVideo 更新影片資訊
func (h *VideoHandler) UpdateVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的影片 ID"})
		return
	}

	var updateDTO dto.VideoUpdateDTO
	if err := c.ShouldBindJSON(&updateDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.videoService.UpdateVideo(uint(id), updateDTO.Title, updateDTO.Description, &dto.VideoDTO{
		Title:       updateDTO.Title,
		Description: updateDTO.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteVideo 刪除影片
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的影片 ID"})
		return
	}

	if err := h.videoService.DeleteVideo(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// LikeVideo 喜歡影片
func (h *VideoHandler) LikeVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的影片 ID"})
		return
	}

	if err := h.videoService.IncrementLikes(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新喜歡次數失敗"})
		return
	}

	c.Status(http.StatusOK)
}
