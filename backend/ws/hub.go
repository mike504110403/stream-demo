package ws

import (
	"sync"

	"stream-demo/backend/dto"
	"stream-demo/backend/utils"
)

// Hub 管理所有聊天室連線
type Hub struct {
	// 直播間 ID 對應的聊天室
	rooms map[uint]*Room
	// 互斥鎖
	mu sync.RWMutex
	// Redis訊息系統
	messaging *utils.RedisMessaging
}

// Room 單一聊天室
type Room struct {
	// 直播間 ID
	liveID uint
	// 所有連線的客戶端
	clients map[*Client]bool
	// 廣播頻道
	broadcast chan *dto.ChatMessageDTO
	// 註冊頻道
	register chan *Client
	// 註銷頻道
	unregister chan *Client
	// 互斥鎖
	mu sync.RWMutex
	// Hub引用（用於訊息發布）
	hub *Hub
}

// NewHub 建立新的 Hub
func NewHub(messaging *utils.RedisMessaging) *Hub {
	hub := &Hub{
		rooms:     make(map[uint]*Room),
		messaging: messaging,
	}

	// 訂閱聊天訊息頻道
	if messaging != nil {
		messaging.Subscribe("chat_messages", hub.handleChatMessage)
		messaging.Subscribe("live_updates", hub.handleLiveUpdate)
	}

	return hub
}

// GetRoom 取得或建立聊天室
func (h *Hub) GetRoom(liveID uint) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[liveID]
	if !exists {
		room = &Room{
			liveID:     liveID,
			clients:    make(map[*Client]bool),
			broadcast:  make(chan *dto.ChatMessageDTO),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			hub:        h,
		}
		h.rooms[liveID] = room
		go room.run()
	}
	return room
}

// handleChatMessage 處理聊天訊息
func (h *Hub) handleChatMessage(channel string, payload []byte) error {
	var message utils.Message
	if err := utils.UnmarshalMessage(payload, &message); err != nil {
		return err
	}

	// 解析聊天訊息
	liveID, ok := message.Payload["live_id"].(float64)
	if !ok {
		return nil
	}

	userID, ok := message.Payload["user_id"].(float64)
	if !ok {
		return nil
	}

	username, ok := message.Payload["username"].(string)
	if !ok {
		return nil
	}

	content, ok := message.Payload["content"].(string)
	if !ok {
		return nil
	}

	messageType, ok := message.Payload["type"].(string)
	if !ok {
		messageType = "text"
	}

	chatMsg := &dto.ChatMessageDTO{
		Type:      messageType,
		LiveID:    uint(liveID),
		UserID:    uint(userID),
		Username:  username,
		Content:   content,
		CreatedAt: message.Timestamp,
	}

	// 廣播到對應聊天室
	h.broadcastToRoom(uint(liveID), chatMsg)
	return nil
}

// handleLiveUpdate 處理直播更新
func (h *Hub) handleLiveUpdate(channel string, payload []byte) error {
	var message utils.Message
	if err := utils.UnmarshalMessage(payload, &message); err != nil {
		return err
	}

	liveID, ok := message.Payload["live_id"].(float64)
	if !ok {
		return nil
	}

	event, ok := message.Payload["event"].(string)
	if !ok {
		return nil
	}

	// 創建系統訊息
	systemMsg := &dto.ChatMessageDTO{
		Type:      "system",
		LiveID:    uint(liveID),
		UserID:    0,
		Username:  "系統",
		Content:   getSystemMessage(event),
		CreatedAt: message.Timestamp,
	}

	// 廣播系統訊息
	h.broadcastToRoom(uint(liveID), systemMsg)
	return nil
}

// broadcastToRoom 廣播訊息到指定聊天室
func (h *Hub) broadcastToRoom(liveID uint, message *dto.ChatMessageDTO) {
	h.mu.RLock()
	room, exists := h.rooms[liveID]
	h.mu.RUnlock()

	if exists {
		select {
		case room.broadcast <- message:
		default:
			// 頻道滿了，可能需要記錄錯誤
		}
	}
}

// PublishChatMessage 發布聊天訊息到Redis
func (h *Hub) PublishChatMessage(liveID, userID uint, username, content, messageType string) error {
	if h.messaging == nil {
		return nil
	}

	return h.messaging.PublishChatMessage(liveID, userID, username, content, messageType)
}

// getSystemMessage 根據事件類型獲取系統訊息
func getSystemMessage(event string) string {
	switch event {
	case "live_started":
		return "直播開始了！"
	case "live_ended":
		return "直播結束了"
	case "viewer_joined":
		return "有新觀眾加入"
	case "viewer_count_update":
		return "觀看人數更新"
	default:
		return "直播狀態更新"
	}
}

// run 執行聊天室
func (r *Room) run() {
	for {
		select {
		case client := <-r.register:
			r.mu.Lock()
			r.clients[client] = true
			r.mu.Unlock()

			// 發送歡迎訊息
			welcomeMsg := &dto.ChatMessageDTO{
				Type:     "system",
				LiveID:   r.liveID,
				UserID:   0,
				Username: "系統",
				Content:  "歡迎進入直播間！",
			}

			select {
			case client.send <- welcomeMsg:
			default:
				close(client.send)
				r.mu.Lock()
				delete(r.clients, client)
				r.mu.Unlock()
			}

		case client := <-r.unregister:
			r.mu.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
			r.mu.Unlock()

		case message := <-r.broadcast:
			r.mu.RLock()
			clients := make([]*Client, 0, len(r.clients))
			for client := range r.clients {
				clients = append(clients, client)
			}
			r.mu.RUnlock()

			// 廣播給所有客戶端
			for _, client := range clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					r.mu.Lock()
					delete(r.clients, client)
					r.mu.Unlock()
				}
			}
		}
	}
}

// AddClient 添加客戶端到聊天室
func (r *Room) AddClient(client *Client) {
	r.register <- client
}

// RemoveClient 從聊天室移除客戶端
func (r *Room) RemoveClient(client *Client) {
	r.unregister <- client
}

// BroadcastMessage 廣播訊息到聊天室
func (r *Room) BroadcastMessage(message *dto.ChatMessageDTO) {
	r.broadcast <- message
}

// GetClientCount 獲取聊天室客戶端數量
func (r *Room) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// GetRoomStats 獲取所有聊天室統計
func (h *Hub) GetRoomStats() map[uint]int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[uint]int)
	for liveID, room := range h.rooms {
		stats[liveID] = room.GetClientCount()
	}
	return stats
}

// Close 關閉Hub
func (h *Hub) Close() error {
	if h.messaging != nil {
		return h.messaging.Close()
	}
	return nil
}
