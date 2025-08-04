package ws

import (
	"log"
	"time"

	"stream-demo/backend/dto"

	"github.com/gorilla/websocket"
)

const (
	// 寫入逾時
	writeWait = 10 * time.Second
	// 讀取逾時
	pongWait = 60 * time.Second
	// 發送 pong 間隔
	pingPeriod = (pongWait * 9) / 10
	// 最大訊息大小
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client 代表一個 WebSocket 連線
type Client struct {
	hub      *Hub
	room     *Room
	conn     *websocket.Conn
	send     chan *dto.ChatMessageDTO
	userID   uint
	username string
}

// readPump 從 WebSocket 連線讀取訊息
func (c *Client) readPump() {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg dto.ChatMessageDTO
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// 設置訊息屬性
		msg.UserID = c.userID
		msg.Username = c.username
		msg.LiveID = c.room.liveID
		msg.Type = "text" // 用戶訊息類型
		msg.CreatedAt = time.Now()

		// 發布訊息到Redis（會廣播到所有實例）
		if c.hub.messaging != nil {
			c.hub.PublishChatMessage(c.room.liveID, c.userID, c.username, msg.Content, "text")
		} else {
			// 如果沒有Redis訊息系統，直接廣播到本地
			c.room.broadcast <- &msg
		}
	}
}

// writePump 將訊息寫入 WebSocket 連線
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			json, err := message.MarshalJSON()
			if err != nil {
				return
			}
			w.Write(json)

			// 加入換行
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				message = <-c.send
				json, err := message.MarshalJSON()
				if err != nil {
					return
				}
				w.Write(json)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
