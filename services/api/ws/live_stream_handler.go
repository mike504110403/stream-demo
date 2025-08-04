package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// LiveStreamClient 直播流客戶端
type LiveStreamClient struct {
	conn       *websocket.Conn
	send       chan []byte
	streamName string
	handler    *LiveStreamHandler
}

// LiveStreamHandler 直播流 WebSocket 處理器
type LiveStreamHandler struct {
	// 流房間映射：streamName -> room
	rooms map[string]*LiveStreamRoom
	mu    sync.RWMutex
}

// LiveStreamRoom 直播流房間
type LiveStreamRoom struct {
	streamName string
	clients    map[*LiveStreamClient]bool
	mu         sync.RWMutex
	// 流狀態
	status      string
	viewerCount int
	lastUpdate  time.Time
}

// LiveStreamMessage 直播流消息
type LiveStreamMessage struct {
	Type       string      `json:"type"`
	StreamName string      `json:"stream_name,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Timestamp  int64       `json:"timestamp"`
}

// NewLiveStreamHandler 創建直播流處理器
func NewLiveStreamHandler() *LiveStreamHandler {
	return &LiveStreamHandler{
		rooms: make(map[string]*LiveStreamRoom),
	}
}

// HandleLiveStream WebSocket 連接處理
func (h *LiveStreamHandler) HandleLiveStream(c *gin.Context) {
	streamName := c.Param("streamName")
	if streamName == "" {
		c.JSON(400, gin.H{"error": "流名稱不能為空"})
		return
	}

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
	client := &LiveStreamClient{
		conn:       conn,
		send:       make(chan []byte, 256),
		streamName: streamName,
		handler:    h,
	}

	// 加入流房間
	h.joinStreamRoom(streamName, client)

	// 發送歡迎消息
	welcomeMsg := LiveStreamMessage{
		Type:       "welcome",
		StreamName: streamName,
		Data: map[string]interface{}{
			"message":      "歡迎加入直播流",
			"viewer_count": h.getRoomViewerCount(streamName),
		},
		Timestamp: time.Now().Unix(),
	}
	client.sendMessage(welcomeMsg)

	// 啟動客戶端處理
	go client.writePump()
	go client.readPump()
}

// sendMessage 發送消息給客戶端
func (c *LiveStreamClient) sendMessage(msg LiveStreamMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化消息失敗: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		// 客戶端緩衝區滿，關閉連接
		close(c.send)
		c.conn.Close()
	}
}

// writePump 寫入泵
func (c *LiveStreamClient) writePump() {
	ticker := time.NewTicker(30 * time.Second) // 30秒 ping
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.handler.leaveStreamRoom(c.streamName, c)
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
func (c *LiveStreamClient) readPump() {
	defer func() {
		c.handler.leaveStreamRoom(c.streamName, c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg LiveStreamMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 錯誤: %v", err)
			}
			break
		}

		c.handler.handleClientMessage(c, msg)
	}
}

// joinStreamRoom 加入流房間
func (h *LiveStreamHandler) joinStreamRoom(streamName string, client *LiveStreamClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 創建房間（如果不存在）
	if _, exists := h.rooms[streamName]; !exists {
		h.rooms[streamName] = &LiveStreamRoom{
			streamName:  streamName,
			clients:     make(map[*LiveStreamClient]bool),
			status:      "active",
			viewerCount: 0,
			lastUpdate:  time.Now(),
		}
	}

	room := h.rooms[streamName]
	room.mu.Lock()
	room.clients[client] = true
	room.viewerCount++
	room.lastUpdate = time.Now()
	room.mu.Unlock()

	// 廣播新觀眾加入
	h.broadcastToRoom(streamName, LiveStreamMessage{
		Type:       "viewer_joined",
		StreamName: streamName,
		Data: map[string]interface{}{
			"viewer_count": room.viewerCount,
		},
		Timestamp: time.Now().Unix(),
	})

	log.Printf("觀眾加入流 %s，當前觀眾數: %d", streamName, room.viewerCount)
}

// leaveStreamRoom 離開流房間
func (h *LiveStreamHandler) leaveStreamRoom(streamName string, client *LiveStreamClient) {
	h.mu.RLock()
	room, exists := h.rooms[streamName]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.Lock()
	if _, exists := room.clients[client]; exists {
		delete(room.clients, client)
		room.viewerCount--
		room.lastUpdate = time.Now()
	}
	room.mu.Unlock()

	// 廣播觀眾離開
	h.broadcastToRoom(streamName, LiveStreamMessage{
		Type:       "viewer_left",
		StreamName: streamName,
		Data: map[string]interface{}{
			"viewer_count": room.viewerCount,
		},
		Timestamp: time.Now().Unix(),
	})

	log.Printf("觀眾離開流 %s，當前觀眾數: %d", streamName, room.viewerCount)

	// 如果房間空了，清理房間
	if room.viewerCount == 0 {
		h.mu.Lock()
		delete(h.rooms, streamName)
		h.mu.Unlock()
		log.Printf("流房間 %s 已清理", streamName)
	}
}

// handleClientMessage 處理客戶端消息
func (h *LiveStreamHandler) handleClientMessage(client *LiveStreamClient, msg LiveStreamMessage) {
	switch msg.Type {
	case "ping":
		// 回應 ping
		client.sendMessage(LiveStreamMessage{
			Type:       "pong",
			StreamName: client.streamName,
			Timestamp:  time.Now().Unix(),
		})

	case "chat":
		// 廣播聊天消息
		h.broadcastToRoom(client.streamName, LiveStreamMessage{
			Type:       "chat",
			StreamName: client.streamName,
			Data:       msg.Data,
			Timestamp:  time.Now().Unix(),
		})

	case "heartbeat":
		// 心跳回應
		client.sendMessage(LiveStreamMessage{
			Type:       "heartbeat_ack",
			StreamName: client.streamName,
			Timestamp:  time.Now().Unix(),
		})

	default:
		log.Printf("未知消息類型: %s", msg.Type)
	}
}

// broadcastToRoom 廣播消息到房間
func (h *LiveStreamHandler) broadcastToRoom(streamName string, message LiveStreamMessage) {
	h.mu.RLock()
	room, exists := h.rooms[streamName]
	h.mu.RUnlock()

	if !exists {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化消息失敗: %v", err)
		return
	}

	room.mu.RLock()
	for client := range room.clients {
		select {
		case client.send <- data:
		default:
			// 客戶端緩衝區滿，關閉連接
			close(client.send)
			delete(room.clients, client)
		}
	}
	room.mu.RUnlock()
}

// getRoomViewerCount 獲取房間觀眾數
func (h *LiveStreamHandler) getRoomViewerCount(streamName string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, exists := h.rooms[streamName]; exists {
		room.mu.RLock()
		defer room.mu.RUnlock()
		return room.viewerCount
	}
	return 0
}

// GetStreamStats 獲取流統計信息
func (h *LiveStreamHandler) GetStreamStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]interface{})
	for streamName, room := range h.rooms {
		room.mu.RLock()
		stats[streamName] = map[string]interface{}{
			"viewer_count": room.viewerCount,
			"status":       room.status,
			"last_update":  room.lastUpdate,
		}
		room.mu.RUnlock()
	}
	return stats
}
