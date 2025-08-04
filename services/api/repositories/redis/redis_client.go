package redisclient

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient 封裝了 Redis 客戶端
type RedisClient struct {
	Client *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// Close 關閉 Redis 客戶端連接
func (r *RedisClient) Close() error {
	r.cancel() // 先取消 context
	return r.Client.Close()
}

// NewRedisClient 初始化一個新的 Redis 客戶端
func NewRedisClient(addr, username, password string, db int) *RedisClient {
	ctx, cancel := context.WithCancel(context.Background())

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Username:     username,
		Password:     password,
		DB:           db,
		MaxRetries:   3,
		PoolSize:     50,              // 增加連接池大小
		MinIdleConns: 10,              // 調整最小空閒連接
		PoolTimeout:  time.Second * 3, // 設置池超時
	})

	return &RedisClient{
		Client: rdb,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Publish 發佈消息到指定頻道
func (r *RedisClient) Publish(channel string, message []byte) error {
	err := r.Client.Publish(r.ctx, channel, message).Err()
	if err != nil {
		log.Println("publish error:", err)
		return err
	}
	return nil
}

// Subscribe 訂閱指定頻道
func (r *RedisClient) Subscribe(channel string) *redis.PubSub {
	return r.Client.Subscribe(r.ctx, channel)
}

func (r *RedisClient) Set(key string, val string, ttl time.Duration) error {
	return r.Client.Set(r.ctx, key, val, ttl).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	result, err := r.Client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		errMsg := "key does not exist"
		return "", fmt.Errorf("%s", errMsg)
	}

	if err != nil {
		log.Printf("[Redis] get error: %v", err)
	}
	return result, nil
}

func (r *RedisClient) Del(key string) error {
	return r.Client.Del(r.ctx, key).Err()
}

func (r *RedisClient) HGet(key, field string) (string, error) {
	result, err := r.Client.HGet(r.ctx, key, field).Result()
	if err == redis.Nil {
		errMsg := "key does not exist"
		log.Println(errMsg + ": " + key)
		return "", fmt.Errorf("%s", errMsg)
	}
	return result, nil
}

func (r *RedisClient) HSet(key, field string, increment int64, liveTime time.Duration) error {
	err := r.Client.HSet(r.ctx, key, field, increment).Err()
	if err != nil {
		return err
	}
	return r.Client.Expire(r.ctx, key, liveTime).Err()
}

// 增加计数
func (r *RedisClient) HIncrBy(key, field string, increment int64) error {
	return r.Client.HIncrBy(r.ctx, key, field, increment).Err()
}

// Get All
func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	return r.Client.HGetAll(r.ctx, key).Result()
}

// ScanKeys 扫描匹配指定模式的键
// func (r *RedisClient) ScanKeys(pattern string, count int) ([]string, error) {
// 	var cursor uint64
// 	var keys []string
// 	for {
// 		foundKeys, newCursor, err := r.Client.Scan(r.ctx, cursor, pattern, int64(count)).Result()
// 		if err != nil {
// 			return nil, err
// 		}
// 		keys = append(keys, foundKeys...)
// 		cursor = newCursor
// 		if cursor == 0 {
// 			break
// 		}
// 	}
// 	return keys, nil
// }

func (r *RedisClient) ScanKeys(pattern string, count int64) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error
		batch, cursor, err = r.Client.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func (r *RedisClient) LPushToList(key string, userId int64) error {
	_, err := r.Client.LPush(r.ctx, key, userId).Result()
	return err
}

func (r *RedisClient) LRemFromList(key string, userId int64) error {
	_, err := r.Client.LRem(r.ctx, key, 0, userId).Result()
	return err
}

func (r *RedisClient) GetListFromRedis(key string) ([]string, error) {
	return r.Client.LRange(r.ctx, key, 0, -1).Result()
}

// 檢查集合是否包含指定的值
func (c *RedisClient) SExistsInSet(key string, value int64) (bool, error) {
	return c.Client.SIsMember(c.ctx, key, value).Result()
}

// 從集合中刪除指定值
func (c *RedisClient) SRemFromSet(key string, value int64) error {
	_, err := c.Client.SRem(c.ctx, key, value).Result()
	return err
}

// 將值添加到集合
func (c *RedisClient) SAddToSet(key string, value int64) error {
	_, err := c.Client.SAdd(c.ctx, key, value).Result()
	return err
}

// 將值加一
func (r *RedisClient) IncrementValueAddOne(key string) (int64, error) {
	return r.Client.Incr(r.ctx, key).Result()
}

// 添加 Ping 方法
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}
