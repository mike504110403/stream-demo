package ws

import (
	"fmt"
	"net/http"
	"strconv"
	"stream-demo/backend/dto"
	"stream-demo/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允許所有來源的 WebSocket 連線
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler 處理 WebSocket 連線
type Handler struct {
	hub *Hub
}

// NewHandler 建立新的 Handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// ServeWS 處理 WebSocket 連線請求
func (h *Handler) ServeWS(c *gin.Context) {
	// 從 URL 參數取得直播間 ID
	liveID, err := strconv.ParseUint(c.Param("liveID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的直播間 ID"})
		return
	}

	// 從 JWT 取得用戶資訊
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登入"})
		return
	}

	// 從 claims 獲取用戶名（如果有的話）
	username := "用戶" // 預設用戶名
	if claims, exists := c.Get("claims"); exists {
		if jwtClaims, ok := claims.(*utils.JWTClaims); ok {
			username = fmt.Sprintf("用戶%d", jwtClaims.UserID)
		}
	}

	// 升級 HTTP 連線為 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "無法建立 WebSocket 連線"})
		return
	}

	// 取得或建立聊天室
	room := h.hub.GetRoom(uint(liveID))

	// 建立新的客戶端
	client := &Client{
		hub:      h.hub,
		room:     room,
		conn:     conn,
		send:     make(chan *dto.ChatMessageDTO, 256),
		userID:   userID.(uint),
		username: username,
	}

	// 註冊客戶端
	client.room.register <- client

	// 啟動讀寫 goroutine
	go client.writePump()
	go client.readPump()
}
