package test

import (
	"testing"

	"stream-demo/backend/dto"
	"stream-demo/backend/ws"

	"github.com/stretchr/testify/assert"
)

func TestNewHub(t *testing.T) {
	// 創建一個nil的messaging來測試
	hub := ws.NewHub(nil)
	assert.NotNil(t, hub)
}

func TestHub_GetRoom(t *testing.T) {
	hub := ws.NewHub(nil)

	// 測試獲取房間
	room := hub.GetRoom(1)
	assert.NotNil(t, room)

	// 測試獲取已存在的房間
	existingRoom := hub.GetRoom(1)
	assert.Equal(t, room, existingRoom)
}

func TestRoom_GetClientCount(t *testing.T) {
	hub := ws.NewHub(nil)
	room := hub.GetRoom(1)

	// 初始客戶端數量應該為0
	assert.Equal(t, 0, room.GetClientCount())
}

func TestHub_GetRoomStats(t *testing.T) {
	hub := ws.NewHub(nil)

	// 創建一個房間
	room := hub.GetRoom(1)
	assert.NotNil(t, room)

	// 獲取房間統計
	stats := hub.GetRoomStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, uint(1))
}

func TestRoom_BroadcastMessage(t *testing.T) {
	hub := ws.NewHub(nil)
	room := hub.GetRoom(1)

	// 測試廣播消息到空房間
	message := &dto.ChatMessageDTO{
		Type:     "chat",
		UserID:   1,
		Username: "testuser",
		Content:  "test message",
	}
	room.BroadcastMessage(message)
	// 不應該panic
}

func TestRoom_AddRemoveClient(t *testing.T) {
	hub := ws.NewHub(nil)
	room := hub.GetRoom(1)

	// 測試添加客戶端（需要一個有效的websocket連接，這裡我們跳過）
	// 在實際測試中，我們會創建一個mock websocket連接
	assert.Equal(t, 0, room.GetClientCount())
}

func TestHub_PublishChatMessage(t *testing.T) {
	hub := ws.NewHub(nil)

	// 測試發布聊天消息
	err := hub.PublishChatMessage(1, 1, "testuser", "Hello world", "chat")
	// 因為messaging是nil，所以會返回nil（沒有錯誤）
	assert.NoError(t, err)
}

func TestHub_Close(t *testing.T) {
	hub := ws.NewHub(nil)

	// 創建一個房間
	room := hub.GetRoom(1)
	assert.NotNil(t, room)

	// 關閉hub
	err := hub.Close()
	assert.NoError(t, err)
}

// 測試基本功能
func TestBasicFunctionality(t *testing.T) {
	hub := ws.NewHub(nil)
	assert.NotNil(t, hub)

	// 測試創建多個房間
	room1 := hub.GetRoom(1)
	room2 := hub.GetRoom(2)
	assert.NotEqual(t, room1, room2)

	// 測試房間統計
	stats := hub.GetRoomStats()
	assert.Len(t, stats, 2)
}
