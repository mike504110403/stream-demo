// pkg/redis/subscriber.go
package redisclient

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Subscriber 處理 Redis 訂閱的結構
type Subscriber struct {
	client *RedisClient
	ctx    context.Context
}

// NewSubscriber 創建新的訂閱者
func NewSubscriber(client *RedisClient) *Subscriber {
	return &Subscriber{
		client: client,
		ctx:    context.Background(),
	}
}

// Subscribe 訂閱指定的頻道
func (s *Subscriber) Subscribe(channel string, handler func(message []byte) error) error {
	pubsub := s.client.Subscribe(channel)
	defer pubsub.Close()

	// 等待訂閱確認
	_, err := pubsub.Receive(s.ctx)
	if err != nil {
		return fmt.Errorf("訂閱失敗: %v", err)
	}

	// 接收消息的通道
	ch := pubsub.Channel()

	// 持續監聽消息
	for msg := range ch {
		if err := handler([]byte(msg.Payload)); err != nil {
			log.Printf("處理消息時發生錯誤: %v", err)
		}
	}

	return nil
}

// SubscribeWithRetry 帶重試機制的訂閱
func (s *Subscriber) SubscribeWithRetry(channel string, handler func(message []byte) error) {
	for {
		err := s.Subscribe(channel, handler)
		if err != nil {
			log.Printf("訂閱失敗，5秒後重試: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}
