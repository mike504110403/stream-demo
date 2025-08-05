package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

// PublicStreamHandler 公開流處理器
type PublicStreamHandler struct {
	service *services.PublicStreamService
}

// getPullerURL 根據運行環境返回正確的 puller 服務 URL
func getPullerURL() string {
	// 檢查是否在 Docker 容器中運行
	if _, err := os.Stat("/.dockerenv"); err == nil {
		// 在 Docker 中運行，使用內部網路
		return "http://puller:8081/api/streams"
	}
	// 在本地運行，使用外部端口
	return "http://localhost:8083/api/streams"
}

// NewPublicStreamHandler 創建公開流處理器
func NewPublicStreamHandler(service *services.PublicStreamService) *PublicStreamHandler {
	return &PublicStreamHandler{
		service: service,
	}
}

// GetAvailableStreams 獲取可用的公開流列表
func (h *PublicStreamHandler) GetAvailableStreams(c *gin.Context) {
	// 從 puller 服務獲取流配置
	// 如果後端 API 在 Docker 中運行，使用內部網路；否則使用外部端口
	pullerURL := getPullerURL()
	
	resp, err := http.Get(pullerURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "獲取流列表失敗",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "puller 服務獲取流列表失敗",
		})
		return
	}

	// 解析 puller 服務的響應
	var pullerResponse struct {
		Streams []map[string]interface{} `json:"streams"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&pullerResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "解析 puller 響應失敗",
			"details": err.Error(),
		})
		return
	}

	// 轉換為前端期望的格式
	streams := make([]map[string]interface{}, 0)
	for _, stream := range pullerResponse.Streams {
		streams = append(streams, map[string]interface{}{
			"name":        stream["name"],
			"title":       stream["title"],
			"description": stream["description"],
			"url":         stream["url"],
			"category":    stream["category"],
			"type":        stream["type"],
			"enabled":     stream["enabled"],
			"running":     stream["running"],
			"status":      func() string { if stream["running"].(bool) { return "active" } else { return "inactive" } }(),
		})
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

// CreateStream 創建新的公開流
func (h *PublicStreamHandler) CreateStream(c *gin.Context) {
	var request struct {
		Name        string `json:"name" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		URL         string `json:"url" binding:"required"`
		Type        string `json:"type"`
		Category    string `json:"category"`
		Enabled     bool   `json:"enabled"`
	}

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 創建流配置
	stream := &services.PublicStreamInfo{
		Name:        request.Name,
		Title:       request.Title,
		Description: request.Description,
		URL:         request.URL,
		Category:    request.Category,
		Status:      "inactive",
	}

	// 調用 puller 服務的 API 來創建流
	pullerURL := getPullerURL()
	streamData := map[string]string{
		"name":        request.Name,
		"title":       request.Title,
		"description": request.Description,
		"url":         request.URL,
		"type":        request.Type,
		"category":    request.Category,
		"enabled":     fmt.Sprintf("%t", request.Enabled),
	}

	// 發送請求到 puller 服務
	resp, err := http.PostForm(pullerURL, mapToFormData(streamData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "創建流失敗",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "puller 服務創建流失敗",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    stream,
	})
}

// UpdateStream 更新公開流
func (h *PublicStreamHandler) UpdateStream(c *gin.Context) {
	streamName := c.Param("name")
	if streamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "流名稱不能為空",
		})
		return
	}

	var request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Type        string `json:"type"`
		Category    string `json:"category"`
		Enabled     bool   `json:"enabled"`
	}

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 更新流配置
	stream := &services.PublicStreamInfo{
		Name:        streamName,
		Title:       request.Title,
		Description: request.Description,
		URL:         request.URL,
		Category:    request.Category,
	}

	// 暫時返回成功，因為服務層還沒有實現更新方法
	// TODO: 實現 UpdateStream 方法
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stream,
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

// mapToFormData 將 map 轉換為 url.Values
func mapToFormData(data map[string]string) url.Values {
	form := url.Values{}
	for key, value := range data {
		form.Set(key, value)
	}
	return form
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
