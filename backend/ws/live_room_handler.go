package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"stream-demo/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// LiveRoomClient 直播間客戶端
type LiveRoomClient struct {
	conn     *websocket.Conn
	send     chan []byte
	roomID   string
	userID   int
	username string
	role     string // creator, viewer
	handler  *LiveRoomHandler
	mu       sync.Mutex
}

// LiveRoomHandler 直播間 WebSocket 處理器
type LiveRoomHandler struct {
	// 房間映射：roomID -> room
	rooms map[string]*LiveRoom
	mu    sync.RWMutex
	// JWT 工具
	jwtUtil *utils.JWTUtil
}

// LiveRoom 直播間
type LiveRoom struct {
	roomID  string
	clients map[*LiveRoomClient]bool
	mu      sync.RWMutex
	// 房間狀態
	status      string
	viewerCount int
	creatorID   int
	title       string
	lastUpdate  time.Time
}

// LiveRoomMessage 直播間消息
type LiveRoomMessage struct {
	Type      string      `json:"type"`
	RoomID    string      `json:"room_id,omitempty"`
	UserID    int         `json:"user_id,omitempty"`
	Username  string      `json:"username,omitempty"`
	Role      string      `json:"role,omitempty"`
	Content   string      `json:"content,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// NewLiveRoomHandler 創建直播間處理器
func NewLiveRoomHandler(jwtUtil *utils.JWTUtil) *LiveRoomHandler {
	return &LiveRoomHandler{
		rooms:   make(map[string]*LiveRoom),
		jwtUtil: jwtUtil,
	}
}

// ServeWS WebSocket 連接處理
func (h *LiveRoomHandler) ServeWS(c *gin.Context) {
	roomID := c.Param("roomID")
	if roomID == "" {
		c.JSON(400, gin.H{"error": "房間ID不能為空"})
		return
	}

	// 從 URL 參數或 header 獲取 JWT token
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if token != "" && len(token) > 7 {
			token = token[7:] // 移除 "Bearer " 前綴
		}
	}

	if token == "" {
		c.JSON(401, gin.H{"error": "未提供認證 token"})
		return
	}

	// 驗證 JWT token
	claims, err := h.jwtUtil.ValidateToken(token)
	if err != nil {
		c.JSON(401, gin.H{"error": "無效的 token"})
		return
	}

	userID := int(claims.UserID)

	// 檢查用戶是否在房間中
	ctx := context.Background()
	isMember, err := utils.GetRedisClient().SIsMember(ctx, fmt.Sprintf("live:room:%s:users", roomID), userID).Result()
	if err != nil || !isMember {
		c.JSON(403, gin.H{"error": "用戶不在房間中"})
		return
	}

	// 獲取用戶角色
	role, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s:roles", roomID), strconv.Itoa(userID)).Result()
	if err != nil {
		c.JSON(500, gin.H{"error": "獲取用戶角色失敗"})
		return
	}

	// 獲取用戶名（這裡簡化處理，實際應該從資料庫獲取）
	username := fmt.Sprintf("user_%d", userID)

	// 升級 HTTP 連接為 WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允許所有來源
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升級失敗: %v", err)
		return
	}

	// 創建客戶端
	client := &LiveRoomClient{
		conn:     conn,
		send:     make(chan []byte, 256),
		roomID:   roomID,
		userID:   userID,
		username: username,
		role:     role,
		handler:  h,
	}

	// 加入房間
	h.joinRoom(roomID, client)

	// 發送歡迎消息
	welcomeMsg := LiveRoomMessage{
		Type:     "welcome",
		RoomID:   roomID,
		UserID:   userID,
		Username: username,
		Role:     role,
		Data: map[string]interface{}{
			"message":      "歡迎加入直播間",
			"viewer_count": h.getRoomViewerCount(roomID),
			"room_status":  h.getRoomStatus(roomID),
		},
		Timestamp: time.Now().Unix(),
	}
	client.sendMessage(welcomeMsg)

	// 廣播用戶加入消息（只給主播）
	h.broadcastToCreator(roomID, LiveRoomMessage{
		Type:     "user_joined",
		RoomID:   roomID,
		UserID:   userID,
		Username: username,
		Role:     role,
		Data: map[string]interface{}{
			"viewer_count": h.getRoomViewerCount(roomID),
		},
		Timestamp: time.Now().Unix(),
	})

	// 啟動客戶端處理
	go client.writePump()
	go client.readPump()

	// 如果是主播，啟動定期廣播觀眾數量
	if role == "creator" {
		go h.startViewerCountBroadcast(roomID)
	}
}

// sendMessage 發送消息
func (c *LiveRoomClient) sendMessage(msg LiveRoomMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("消息序列化失敗: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		c.handler.leaveRoom(c.roomID, c)
	}
}

// writePump 寫入泵
func (c *LiveRoomClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump 讀取泵
func (c *LiveRoomClient) readPump() {
	defer func() {
		c.handler.leaveRoom(c.roomID, c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 讀取錯誤: %v", err)
			}
			break
		}

		c.handler.handleClientMessage(c, message)
	}
}

// joinRoom 加入房間
func (h *LiveRoomHandler) joinRoom(roomID string, client *LiveRoomClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[roomID]
	if !exists {
		room = &LiveRoom{
			roomID:      roomID,
			clients:     make(map[*LiveRoomClient]bool),
			status:      "created",
			viewerCount: 0,
			lastUpdate:  time.Now(),
		}
		h.rooms[roomID] = room
	}

	room.mu.Lock()
	room.clients[client] = true
	room.viewerCount = len(room.clients)
	room.lastUpdate = time.Now()
	room.mu.Unlock()

	// 更新 Redis 中的觀眾數量
	ctx := context.Background()
	utils.GetRedisClient().HSet(ctx, fmt.Sprintf("live:room:%s", roomID), "viewer_count", room.viewerCount)
}

// leaveRoom 離開房間
func (h *LiveRoomHandler) leaveRoom(roomID string, client *LiveRoomClient) {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.Lock()
	delete(room.clients, client)
	room.viewerCount = len(room.clients)
	room.lastUpdate = time.Now()
	room.mu.Unlock()

	// 更新 Redis 中的觀眾數量
	ctx := context.Background()
	utils.GetRedisClient().HSet(ctx, fmt.Sprintf("live:room:%s", roomID), "viewer_count", room.viewerCount)

	// 廣播用戶離開消息（只給主播）
	h.broadcastToCreator(roomID, LiveRoomMessage{
		Type:     "user_left",
		RoomID:   roomID,
		UserID:   client.userID,
		Username: client.username,
		Role:     client.role,
		Data: map[string]interface{}{
			"viewer_count": room.viewerCount,
		},
		Timestamp: time.Now().Unix(),
	})

	// 如果房間沒有客戶端了，清理房間
	if room.viewerCount == 0 {
		h.mu.Lock()
		delete(h.rooms, roomID)
		h.mu.Unlock()
	}
}

// handleClientMessage 處理客戶端消息
func (h *LiveRoomHandler) handleClientMessage(client *LiveRoomClient, message []byte) {
	var msg LiveRoomMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("消息解析失敗: %v", err)
		return
	}

	switch msg.Type {
	case "chat":
		// 處理聊天消息
		h.handleChatMessage(client, msg)
	case "ping":
		// 回應 ping
		client.sendMessage(LiveRoomMessage{
			Type:      "pong",
			Timestamp: time.Now().Unix(),
		})
	default:
		log.Printf("未知消息類型: %s", msg.Type)
	}
}

// handleChatMessage 處理聊天消息
func (h *LiveRoomHandler) handleChatMessage(client *LiveRoomClient, msg LiveRoomMessage) {
	// 廣播聊天消息
	h.broadcastToRoom(client.roomID, LiveRoomMessage{
		Type:      "chat",
		RoomID:    client.roomID,
		UserID:    client.userID,
		Username:  client.username,
		Role:      client.role,
		Content:   msg.Content,
		Timestamp: time.Now().Unix(),
	})

	// 記錄聊天消息到 Redis（可選）
	ctx := context.Background()
	chatKey := fmt.Sprintf("live:room:%s:chat", client.roomID)
	chatMsg := map[string]interface{}{
		"user_id":   client.userID,
		"username":  client.username,
		"role":      client.role,
		"content":   msg.Content,
		"timestamp": time.Now().Unix(),
	}

	chatData, _ := json.Marshal(chatMsg)
	utils.GetRedisClient().LPush(ctx, chatKey, chatData)
	utils.GetRedisClient().LTrim(ctx, chatKey, 0, 99) // 只保留最近100條消息
}

// broadcastToRoom 廣播到房間
func (h *LiveRoomHandler) broadcastToRoom(roomID string, message LiveRoomMessage) {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	clients := make([]*LiveRoomClient, 0, len(room.clients))
	for client := range room.clients {
		clients = append(clients, client)
	}
	room.mu.RUnlock()

	for _, client := range clients {
		client.sendMessage(message)
	}
}

// broadcastToCreator 只廣播給主播
func (h *LiveRoomHandler) broadcastToCreator(roomID string, message LiveRoomMessage) {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	clients := make([]*LiveRoomClient, 0, len(room.clients))
	for client := range room.clients {
		if client.role == "creator" {
			clients = append(clients, client)
		}
	}
	room.mu.RUnlock()

	for _, client := range clients {
		client.sendMessage(message)
	}
}

// getRoomViewerCount 獲取房間觀眾數量
func (h *LiveRoomHandler) getRoomViewerCount(roomID string) int {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if !exists {
		return 0
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	return room.viewerCount
}

// getRoomStatus 獲取房間狀態
func (h *LiveRoomHandler) getRoomStatus(roomID string) string {
	ctx := context.Background()
	status, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s", roomID), "status").Result()
	if err != nil {
		return "unknown"
	}
	return status
}

// BroadcastRoomUpdate 廣播房間狀態更新
func (h *LiveRoomHandler) BroadcastRoomUpdate(roomID string, updateType string, data interface{}) {
	h.broadcastToRoom(roomID, LiveRoomMessage{
		Type:      updateType,
		RoomID:    roomID,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// GetRoomStats 獲取房間統計
func (h *LiveRoomHandler) GetRoomStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]interface{})
	for roomID, room := range h.rooms {
		room.mu.RLock()
		stats[roomID] = map[string]interface{}{
			"viewer_count": room.viewerCount,
			"status":       room.status,
			"last_update":  room.lastUpdate,
		}
		room.mu.RUnlock()
	}
	return stats
}

// startViewerCountBroadcast 定期廣播觀眾數量
func (h *LiveRoomHandler) startViewerCountBroadcast(roomID string) {
	ticker := time.NewTicker(10 * time.Second) // 每10秒更新一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 檢查房間是否還存在
			h.mu.RLock()
			_, exists := h.rooms[roomID]
			h.mu.RUnlock()

			if !exists {
				return // 房間不存在，停止廣播
			}

			// 廣播觀眾數量更新
			h.broadcastToRoom(roomID, LiveRoomMessage{
				Type:   "viewer_count_update",
				RoomID: roomID,
				Data: map[string]interface{}{
					"viewer_count": h.getRoomViewerCount(roomID),
				},
				Timestamp: time.Now().Unix(),
			})
		}
	}
}
