package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// PostgreSQLMessaging PostgreSQL訊息佇列管理器
type PostgreSQLMessaging struct {
	db          *gorm.DB
	listener    *pq.Listener
	channels    map[string][]MessageHandler
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	isListening bool
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

// NewPostgreSQLMessaging 創建PostgreSQL訊息佇列實例
func NewPostgreSQLMessaging(db *gorm.DB) (*PostgreSQLMessaging, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("獲取底層SQL連接失敗: %w", err)
	}

	// 獲取連接字符串（這裡需要從配置中讀取）
	connStr := buildConnectionString(sqlDB)

	listener := pq.NewListener(connStr, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			LogError("PostgreSQL Listener錯誤: %v", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())

	messaging := &PostgreSQLMessaging{
		db:       db,
		listener: listener,
		channels: make(map[string][]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
	}

	return messaging, nil
}

// buildConnectionString 構建連接字符串（簡化版本）
func buildConnectionString(sqlDB *sql.DB) string {
	// 這裡應該從配置文件讀取，暫時使用預設值
	return "postgres://stream_user:stream_password@localhost:5432/stream_demo?sslmode=disable"
}

// Subscribe 訂閱頻道
func (m *PostgreSQLMessaging) Subscribe(channel string, handler MessageHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 添加處理器
	m.channels[channel] = append(m.channels[channel], handler)

	// 如果是新頻道，開始監聽
	if len(m.channels[channel]) == 1 {
		if err := m.listener.Listen(channel); err != nil {
			return fmt.Errorf("監聽頻道 %s 失敗: %w", channel, err)
		}
		LogInfo("開始監聽頻道: %s", channel)
	}

	// 啟動監聽器（如果還沒啟動）
	if !m.isListening {
		go m.startListening()
		m.isListening = true
	}

	return nil
}

// Unsubscribe 取消訂閱頻道
func (m *PostgreSQLMessaging) Unsubscribe(channel string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.channels, channel)

	if err := m.listener.Unlisten(channel); err != nil {
		return fmt.Errorf("停止監聽頻道 %s 失敗: %w", channel, err)
	}

	LogInfo("停止監聽頻道: %s", channel)
	return nil
}

// Publish 發布訊息到指定頻道
func (m *PostgreSQLMessaging) Publish(channel string, messageType string, payload map[string]interface{}) error {
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

	sql := fmt.Sprintf("SELECT pg_notify('%s', $1)", channel)
	if err := m.db.Exec(sql, string(messageJSON)).Error; err != nil {
		return fmt.Errorf("發送訊息到頻道 %s 失敗: %w", channel, err)
	}

	LogInfo("訊息已發送到頻道 %s: %s", channel, messageType)
	return nil
}

// PublishVideoProcessing 發布影片處理訊息
func (m *PostgreSQLMessaging) PublishVideoProcessing(videoID uint, status string, progress int) error {
	return m.Publish("video_processing", "status_update", map[string]interface{}{
		"video_id": videoID,
		"status":   status,
		"progress": progress,
	})
}

// PublishLiveUpdate 發布直播更新訊息
func (m *PostgreSQLMessaging) PublishLiveUpdate(liveID uint, eventType string, data map[string]interface{}) error {
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
func (m *PostgreSQLMessaging) PublishUserNotification(userID uint, notificationType string, title, content string) error {
	return m.Publish("user_notifications", notificationType, map[string]interface{}{
		"user_id": userID,
		"title":   title,
		"content": content,
	})
}

// startListening 開始監聽訊息
func (m *PostgreSQLMessaging) startListening() {
	LogInfo("PostgreSQL訊息監聽器已啟動")

	for {
		select {
		case <-m.ctx.Done():
			LogInfo("PostgreSQL訊息監聽器已停止")
			return

		case notification := <-m.listener.Notify:
			if notification != nil {
				m.handleNotification(notification)
			}

		case <-time.After(90 * time.Second):
			// 定期ping以保持連接
			go func() {
				if err := m.listener.Ping(); err != nil {
					LogError("PostgreSQL Listener ping失敗: %v", err)
				}
			}()
		}
	}
}

// handleNotification 處理收到的通知
func (m *PostgreSQLMessaging) handleNotification(notification *pq.Notification) {
	m.mu.RLock()
	handlers, exists := m.channels[notification.Channel]
	m.mu.RUnlock()

	if !exists {
		return
	}

	// 解析訊息
	var message Message
	if err := json.Unmarshal([]byte(notification.Extra), &message); err != nil {
		LogError("解析訊息失敗: %v", err)
		return
	}

	// 調用所有處理器
	for _, handler := range handlers {
		go func(h MessageHandler) {
			if err := h(notification.Channel, []byte(notification.Extra)); err != nil {
				LogError("訊息處理失敗 (頻道: %s): %v", notification.Channel, err)
			}
		}(handler)
	}
}

// Close 關閉訊息佇列
func (m *PostgreSQLMessaging) Close() error {
	m.cancel()

	if m.listener != nil {
		if err := m.listener.Close(); err != nil {
			return fmt.Errorf("關閉監聽器失敗: %w", err)
		}
	}

	LogInfo("PostgreSQL訊息佇列已關閉")
	return nil
}

// GetChannels 獲取當前訂閱的頻道列表
func (m *PostgreSQLMessaging) GetChannels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channels := make([]string, 0, len(m.channels))
	for channel := range m.channels {
		channels = append(channels, channel)
	}

	return channels
}

// generateMessageID 生成訊息ID
func generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// VideoProcessingHandler 影片處理訊息處理器
func VideoProcessingHandler(channel string, payload []byte) error {
	var message Message
	if err := json.Unmarshal(payload, &message); err != nil {
		return fmt.Errorf("解析影片處理訊息失敗: %w", err)
	}

	videoID, ok := message.Payload["video_id"].(float64)
	if !ok {
		return fmt.Errorf("無效的video_id")
	}

	status, ok := message.Payload["status"].(string)
	if !ok {
		return fmt.Errorf("無效的status")
	}

	LogInfo("處理影片 %d 狀態更新: %s", uint(videoID), status)

	// 這裡可以添加具體的業務邏輯
	// 例如：更新資料庫、通知前端等

	return nil
}

// LiveUpdateHandler 直播更新訊息處理器
func LiveUpdateHandler(channel string, payload []byte) error {
	var message Message
	if err := json.Unmarshal(payload, &message); err != nil {
		return fmt.Errorf("解析直播更新訊息失敗: %w", err)
	}

	liveID, ok := message.Payload["live_id"].(float64)
	if !ok {
		return fmt.Errorf("無效的live_id")
	}

	event, ok := message.Payload["event"].(string)
	if !ok {
		return fmt.Errorf("無效的event")
	}

	LogInfo("處理直播 %d 事件: %s", uint(liveID), event)

	// 這裡可以添加具體的業務邏輯
	// 例如：更新觀看人數、通知其他觀眾等

	return nil
}

// UnmarshalMessage 解析訊息
func UnmarshalMessage(payload []byte, dest *Message) error {
	return json.Unmarshal(payload, dest)
}
