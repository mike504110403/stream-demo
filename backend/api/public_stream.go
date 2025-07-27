package api

import (
	"net/http"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

// PublicStreamHandler 公開流處理器
type PublicStreamHandler struct {
	service *services.PublicStreamService
}

// NewPublicStreamHandler 創建公開流處理器
func NewPublicStreamHandler(service *services.PublicStreamService) *PublicStreamHandler {
	return &PublicStreamHandler{
		service: service,
	}
}

// GetAvailableStreams 獲取可用的公開流列表
func (h *PublicStreamHandler) GetAvailableStreams(c *gin.Context) {
	streams, err := h.service.GetAvailableStreams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "獲取流列表失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"streams": streams,
			"total":   len(streams),
		},
	})
}

// GetStreamInfo 獲取流詳細資訊
func (h *PublicStreamHandler) GetStreamInfo(c *gin.Context) {
	streamName := c.Param("name")
	if streamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "流名稱不能為空",
		})
		return
	}

	info, err := h.service.GetStreamInfo(streamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "流不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// GetStreamURL 獲取流的播放 URL
func (h *PublicStreamHandler) GetStreamURL(c *gin.Context) {
	streamName := c.Param("name")
	if streamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "流名稱不能為空",
		})
		return
	}

	url, err := h.service.GetStreamURL(streamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "獲取流 URL 失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stream_name": streamName,
			"url":         url,
		},
	})
}

// GetStreamURLs 獲取流的所有播放 URL
func (h *PublicStreamHandler) GetStreamURLs(c *gin.Context) {
	streamName := c.Param("name")
	if streamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "流名稱不能為空",
		})
		return
	}

	urls, err := h.service.GetStreamURLs(streamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "獲取流 URL 失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stream_name": streamName,
			"urls":        urls,
		},
	})
}

// GetStreamByCategory 根據分類獲取流列表
func (h *PublicStreamHandler) GetStreamByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "分類不能為空",
		})
		return
	}

	streams, err := h.service.GetAvailableStreams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "獲取流列表失敗",
			"details": err.Error(),
		})
		return
	}

	// 過濾分類
	var filteredStreams []*services.PublicStreamInfo
	for _, stream := range streams {
		if stream.Category == category {
			filteredStreams = append(filteredStreams, stream)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"category": category,
			"streams":  filteredStreams,
			"total":    len(filteredStreams),
		},
	})
}

// GetStreamStats 獲取流統計資訊
func (h *PublicStreamHandler) GetStreamStats(c *gin.Context) {
	streamName := c.Param("name")
	if streamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "流名稱不能為空",
		})
		return
	}

	info, err := h.service.GetStreamInfo(streamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "流不存在",
			"details": err.Error(),
		})
		return
	}

	// 這裡可以添加更多統計資訊
	stats := gin.H{
		"stream_name":  info.Name,
		"title":        info.Title,
		"status":       info.Status,
		"viewer_count": info.ViewerCount,
		"last_update":  info.LastUpdate,
		"category":     info.Category,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// RegisterPublicStreamRoutes 註冊公開流路由
func RegisterPublicStreamRoutes(r *gin.RouterGroup, handler *PublicStreamHandler) {
	// 公開流相關路由
	publicStreams := r.Group("/public-streams")
	{
		publicStreams.GET("", handler.GetAvailableStreams)                    // 獲取所有可用流
		publicStreams.GET("/:name", handler.GetStreamInfo)                    // 獲取流資訊
		publicStreams.GET("/:name/url", handler.GetStreamURL)                 // 獲取播放 URL
		publicStreams.GET("/:name/urls", handler.GetStreamURLs)               // 獲取所有播放 URL (HLS + RTMP)
		publicStreams.GET("/:name/stats", handler.GetStreamStats)             // 獲取統計資訊
		publicStreams.GET("/category/:category", handler.GetStreamByCategory) // 按分類獲取流
	}
}
