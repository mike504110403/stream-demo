package api

import (
	"net/http"
	"strconv"
	"stream-demo/backend/dto"
	"stream-demo/backend/dto/request"
	"stream-demo/backend/dto/response"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	videoService services.VideoServiceInterface
}

func NewVideoHandler(videoService services.VideoServiceInterface) *VideoHandler {
	return &VideoHandler{videoService: videoService}
}

// ListVideos 列出所有影片
func (h *VideoHandler) ListVideos(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	videos, total, err := h.videoService.GetVideos(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewListResponse(total, videos)))
}

// GenerateUploadURL 生成S3上傳URL
func (h *VideoHandler) GenerateUploadURL(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, "未登入"))
		return
	}

	var req request.GenerateUploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	// 生成預簽名上傳URL
	uploadURL, err := h.videoService.GenerateUploadURL(userID.(uint), req.Filename, req.FileSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	// 創建影片記錄
	video, err := h.videoService.CreateVideoRecord(userID.(uint), req.Title, req.Description, uploadURL.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	// 返回上傳URL和影片資訊
	c.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"upload_url": uploadURL.UploadURL,
		"form_data":  uploadURL.FormData,
		"key":        uploadURL.Key, // 添加 key 字段供前端使用
		"video":      video,
	}))
}

// ConfirmUpload 確認上傳完成並開始處理
func (h *VideoHandler) ConfirmUpload(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, "未登入"))
		return
	}

	var req request.ConfirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	// 只確認上傳，不檢查轉碼狀態
	if err := h.videoService.ConfirmUploadOnly(req.VideoID, req.S3Key); err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	// 返回簡單的成功回應，不包含轉碼狀態
	c.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"message":  "影片上傳確認成功，轉碼處理已開始",
		"video_id": req.VideoID,
		"status":   "uploading",
	}))
}

// GetVideoTranscodeStatus 獲取影片轉碼狀態
func (h *VideoHandler) GetVideoTranscodeStatus(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(http.StatusBadRequest, "無效的影片ID"))
		return
	}

	video, err := h.videoService.GetVideoByID(uint(videoID))
	if err != nil {
		c.JSON(http.StatusNotFound, response.NewErrorResponse(http.StatusNotFound, "影片不存在"))
		return
	}

	// 檢查轉碼狀態
	status := map[string]interface{}{
		"video_id":            video.ID,
		"status":              video.Status,
		"processing_progress": video.ProcessingProgress,
		"original_url":        video.OriginalURL,
		"mp4_url":             video.MP4URL,
		"hls_master_url":      video.HLSMasterURL,
		"thumbnail_url":       video.ThumbnailURL,
		"file_size":           video.FileSize,
		"original_format":     video.OriginalFormat,
		"created_at":          video.CreatedAt,
		"updated_at":          video.UpdatedAt,
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(status))
}

// UploadVideo 傳統表單上傳方式（保留相容性）
func (h *VideoHandler) UploadVideo(c *gin.Context) {
	// 從 context 獲取使用者 ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, "未登入"))
		return
	}

	// 獲取表單數據
	title := c.PostForm("title")
	description := c.PostForm("description")

	if title == "" {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "標題不能為空"))
		return
	}

	// 獲取影片檔案
	videoFile, videoHeader, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "請上傳影片檔案"))
		return
	}
	defer videoFile.Close()

	// 生成上傳URL
	uploadURL, err := h.videoService.GenerateUploadURL(userID.(uint), videoHeader.Filename, videoHeader.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	// 創建影片記錄
	video, err := h.videoService.CreateVideoRecord(userID.(uint), title, description, uploadURL.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	// TODO: 在實際應用中，將檔案上傳到S3
	// 這裡可以實現直接上傳邏輯，或者返回預簽名 URL 讓前端處理

	c.JSON(http.StatusCreated, response.NewSuccessResponse(gin.H{
		"video":      video,
		"upload_url": uploadURL.UploadURL,
		"form_data":  uploadURL.FormData,
		"message":    "檔案已接收，請使用返回的upload_url完成S3上傳，或調用確認上傳API",
	}))
}

// GetVideo 獲取影片詳情
func (h *VideoHandler) GetVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的影片 ID"))
		return
	}

	video, err := h.videoService.GetVideoByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, response.NewErrorResponse(404, "影片不存在"))
		return
	}

	// 直接返回 DTO，無需額外包裝
	c.JSON(http.StatusOK, response.NewSuccessResponse(video))
}

// UpdateVideo 更新影片
func (h *VideoHandler) UpdateVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的影片 ID"))
		return
	}

	var req request.UpdateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	err = h.videoService.UpdateVideo(uint(id), req.Title, req.Description, &dto.VideoDTO{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	// 獲取更新後的影片
	video, err := h.videoService.GetVideoByID(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{"message": "更新成功"}))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(video))
}

// DeleteVideo 刪除影片
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的影片 ID"))
		return
	}

	if err := h.videoService.DeleteVideo(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{"message": "刪除成功"}))
}

// SearchVideos 搜尋影片
func (h *VideoHandler) SearchVideos(c *gin.Context) {
	var req request.SearchVideoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	videos, total, err := h.videoService.SearchVideos(req.Query, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewListResponse(total, videos)))
}

// LikeVideo 點讚影片
func (h *VideoHandler) LikeVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的影片 ID"))
		return
	}

	if err := h.videoService.LikeVideo(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{"message": "按讚成功"}))
}

// GetUserVideos 獲取用戶影片
func (h *VideoHandler) GetUserVideos(c *gin.Context) {
	// 獲取用戶 ID 參數
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的用戶 ID"))
		return
	}

	videos, total, err := h.videoService.GetVideosByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewListResponse(total, videos)))
}
