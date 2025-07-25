package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisMessaging Redis 訊息佇列管理器
type RedisMessaging struct {
	client      *redis.Client
	pubsub      *redis.PubSub
	channels    map[string][]MessageHandler
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	isListening bool
	db          int
}

// MessageHandler 訊息處理函數類型
type MessageHandler func(channel string, payload []byte) error

// Message 訊息結構
type Message struct {
	ID        string                 `json:"id"`
	Channel   string                 `json:"channel"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewRedisMessaging 創建Redis訊息佇列實例
func NewRedisMessaging(db int) (*RedisMessaging, error) {
	// 創建指定DB的Redis客戶端
	messagingClient := redis.NewClient(&redis.Options{
		Addr:     RedisClient.Options().Addr,
		Password: RedisClient.Options().Password,
		DB:       db,
	})

	// 測試連接
	ctx := context.Background()
	if err := messagingClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis messaging connection failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	messaging := &RedisMessaging{
		client:   messagingClient,
		channels: make(map[string][]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
		db:       db,
	}

	return messaging, nil
}

// Subscribe 訂閱頻道
func (m *RedisMessaging) Subscribe(channel string, handler MessageHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 添加處理器
	m.channels[channel] = append(m.channels[channel], handler)

	// 如果是新頻道，開始監聽
	if len(m.channels[channel]) == 1 {
		if m.pubsub == nil {
			m.pubsub = m.client.Subscribe(m.ctx, channel)
		} else {
			if err := m.pubsub.Subscribe(m.ctx, channel); err != nil {
				return fmt.Errorf("訂閱頻道 %s 失敗: %w", channel, err)
			}
		}
		fmt.Printf("INFO: 開始監聽頻道: %s\n", channel)
	}

	// 啟動監聽器（如果還沒啟動）
	if !m.isListening {
		go m.startListening()
		m.isListening = true
	}

	return nil
}

// Unsubscribe 取消訂閱頻道
func (m *RedisMessaging) Unsubscribe(channel string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.channels, channel)

	if m.pubsub != nil {
		if err := m.pubsub.Unsubscribe(m.ctx, channel); err != nil {
			return fmt.Errorf("停止監聽頻道 %s 失敗: %w", channel, err)
		}
	}

	fmt.Printf("INFO: 停止監聽頻道: %s\n", channel)
	return nil
}

// Publish 發布訊息到指定頻道
func (m *RedisMessaging) Publish(channel string, messageType string, payload map[string]interface{}) error {
	message := Message{
		ID:        generateMessageID(),
		Channel:   channel,
		Type:      messageType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化訊息失敗: %w", err)
	}

	if err := m.client.Publish(m.ctx, channel, messageJSON).Err(); err != nil {
		return fmt.Errorf("發送訊息到頻道 %s 失敗: %w", channel, err)
	}

	fmt.Printf("INFO: 訊息已發送到頻道 %s: %s\n", channel, messageType)
	return nil
}

// PublishVideoProcessing 發布影片處理訊息
func (m *RedisMessaging) PublishVideoProcessing(videoID uint, status string, progress int) error {
	return m.Publish("video_processing", "status_update", map[string]interface{}{
		"video_id": videoID,
		"status":   status,
		"progress": progress,
	})
}

// PublishLiveUpdate 發布直播更新訊息
func (m *RedisMessaging) PublishLiveUpdate(liveID uint, eventType string, data map[string]interface{}) error {
	payload := map[string]interface{}{
		"live_id": liveID,
		"event":   eventType,
	}

	// 合併額外數據
	for k, v := range data {
		payload[k] = v
	}

	return m.Publish("live_updates", eventType, payload)
}

// PublishUserNotification 發布用戶通知訊息
func (m *RedisMessaging) PublishUserNotification(userID uint, notificationType string, title, content string) error {
	return m.Publish("user_notifications", notificationType, map[string]interface{}{
		"user_id": userID,
		"title":   title,
		"content": content,
	})
}

// PublishChatMessage 發布聊天訊息
func (m *RedisMessaging) PublishChatMessage(liveID, userID uint, username, content, messageType string) error {
	return m.Publish("chat_messages", "new_message", map[string]interface{}{
		"live_id":  liveID,
		"user_id":  userID,
		"username": username,
		"content":  content,
		"type":     messageType,
	})
}

// startListening 開始監聽訊息
func (m *RedisMessaging) startListening() {
	fmt.Println("INFO: Redis訊息監聽器已啟動")

	if m.pubsub == nil {
		fmt.Println("ERROR: PubSub未初始化")
		return
	}

	// 獲取訊息通道
	ch := m.pubsub.Channel()

	for {
		select {
		case <-m.ctx.Done():
			fmt.Println("INFO: Redis訊息監聽器已停止")
			return

		case msg := <-ch:
			if msg != nil {
				m.handleMessage(msg)
			}

		case <-time.After(30 * time.Second):
			// 定期ping以保持連接
			if err := m.pubsub.Ping(m.ctx); err != nil {
				fmt.Printf("ERROR: Redis PubSub ping失敗: %v\n", err)
			}
		}
	}
}

// handleMessage 處理收到的訊息
func (m *RedisMessaging) handleMessage(msg *redis.Message) {
	m.mu.RLock()
	handlers, exists := m.channels[msg.Channel]
	m.mu.RUnlock()

	if !exists {
		return
	}

	payload := []byte(msg.Payload)

	// 並行處理所有處理器
	for _, handler := range handlers {
		go func(h MessageHandler) {
			if err := h(msg.Channel, payload); err != nil {
				fmt.Printf("ERROR: 處理訊息失敗 (頻道: %s): %v\n", msg.Channel, err)
			}
		}(handler)
	}
}

// Close 關閉訊息系統
func (m *RedisMessaging) Close() error {
	m.cancel()

	var errs []error

	if m.pubsub != nil {
		if err := m.pubsub.Close(); err != nil {
			errs = append(errs, fmt.Errorf("關閉PubSub失敗: %w", err))
		}
	}

	if m.client != nil {
		if err := m.client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("關閉Redis客戶端失敗: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("關閉訊息系統時發生錯誤: %v", errs)
	}

	return nil
}

// GetChannels 獲取所有訂閱的頻道
func (m *RedisMessaging) GetChannels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channels := make([]string, 0, len(m.channels))
	for channel := range m.channels {
		channels = append(channels, channel)
	}
	return channels
}

// GetStats 獲取訊息系統統計信息
func (m *RedisMessaging) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"db":             m.db,
		"channels_count": len(m.channels),
		"channels":       m.GetChannels(),
		"is_listening":   m.isListening,
	}

	return stats
}

// generateMessageID 生成訊息ID
func generateMessageID() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// VideoProcessingHandler 影片處理訊息處理器
func VideoProcessingHandler(channel string, payload []byte) error {
	var message Message
	if err := UnmarshalMessage(payload, &message); err != nil {
		return err
	}

	videoID, ok := message.Payload["video_id"].(float64)
	if !ok {
		return fmt.Errorf("無效的video_id")
	}

	status, ok := message.Payload["status"].(string)
	if !ok {
		return fmt.Errorf("無效的status")
	}

	progress, ok := message.Payload["progress"].(float64)
	if !ok {
		progress = 0
	}

	fmt.Printf("INFO: 處理影片訊息 - VideoID: %d, Status: %s, Progress: %.0f%%\n",
		uint(videoID), status, progress)

	// 這裡可以添加具體的業務邏輯
	// 例如更新資料庫、發送通知等

	return nil
}

// LiveUpdateHandler 直播更新訊息處理器
func LiveUpdateHandler(channel string, payload []byte) error {
	var message Message
	if err := UnmarshalMessage(payload, &message); err != nil {
		return err
	}

	liveID, ok := message.Payload["live_id"].(float64)
	if !ok {
		return fmt.Errorf("無效的live_id")
	}

	event, ok := message.Payload["event"].(string)
	if !ok {
		return fmt.Errorf("無效的event")
	}

	fmt.Printf("INFO: 處理直播訊息 - LiveID: %d, Event: %s\n", uint(liveID), event)

	// 這裡可以添加具體的業務邏輯
	// 例如更新觀看人數、發送系統訊息等

	return nil
}

// ChatMessageHandler 聊天訊息處理器
func ChatMessageHandler(channel string, payload []byte) error {
	var message Message
	if err := UnmarshalMessage(payload, &message); err != nil {
		return err
	}

	liveID, ok := message.Payload["live_id"].(float64)
	if !ok {
		return fmt.Errorf("無效的live_id")
	}

	username, ok := message.Payload["username"].(string)
	if !ok {
		return fmt.Errorf("無效的username")
	}

	content, ok := message.Payload["content"].(string)
	if !ok {
		return fmt.Errorf("無效的content")
	}

	fmt.Printf("INFO: 處理聊天訊息 - LiveID: %d, User: %s, Content: %s\n",
		uint(liveID), username, content)

	// 這裡可以添加具體的業務邏輯
	// 例如保存聊天記錄、廣播給WebSocket等

	return nil
}

// UnmarshalMessage 反序列化訊息
func UnmarshalMessage(payload []byte, dest *Message) error {
	if err := json.Unmarshal(payload, dest); err != nil {
		return fmt.Errorf("反序列化訊息失敗: %w", err)
	}
	return nil
}
