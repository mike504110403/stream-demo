package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// LiveRoomService Redis 驅動的直播間服務
type LiveRoomService struct {
	conf      *config.Config
	db        *gorm.DB
	wsHandler interface{} // WebSocket 處理器接口
}

// LiveRoomInfo 直播間信息
type LiveRoomInfo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatorID   int       `json:"creator_id"`
	Status      string    `json:"status"`
	StreamKey   string    `json:"stream_key"`
	ViewerCount int       `json:"viewer_count"`
	MaxViewers  int       `json:"max_viewers"`
	StartedAt   time.Time `json:"started_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// NewLiveRoomService 創建直播間服務
func NewLiveRoomService(conf *config.Config, db *gorm.DB) *LiveRoomService {
	return &LiveRoomService{
		conf:      conf,
		db:        db,
		wsHandler: nil,
	}
}

// SetWSHandler 設置 WebSocket 處理器
func (s *LiveRoomService) SetWSHandler(handler interface{}) {
	s.wsHandler = handler
}

// CreateRoom 創建直播間
func (s *LiveRoomService) CreateRoom(userID int, title, description string) (*LiveRoomInfo, error) {
	ctx := context.Background()

	// 檢查用戶是否已經有活躍的直播間
	existingRoomID, err := utils.GetRedisClient().Get(ctx, fmt.Sprintf("user:%d:current_room", userID)).Result()
	if err == nil && existingRoomID != "" {
		// 檢查現有房間是否還活躍
		roomStatus, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s", existingRoomID), "status").Result()
		if err == nil && (roomStatus == "created" || roomStatus == "live") {
			return nil, fmt.Errorf("用戶已有活躍的直播間，請先結束現有直播間")
		}
	}

	// 生成唯一房間ID和推流密鑰
	roomID := fmt.Sprintf("room_%s", uuid.New().String()[:8])
	streamKey := fmt.Sprintf("stream_%s", uuid.New().String()[:12])

	now := time.Now()
	roomInfo := &LiveRoomInfo{
		ID:          roomID,
		Title:       title,
		Description: description,
		CreatorID:   userID,
		Status:      "created",
		StreamKey:   streamKey,
		ViewerCount: 0,
		MaxViewers:  1000,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 保存到 Redis
	if err := s.saveRoomToRedis(ctx, roomInfo); err != nil {
		return nil, fmt.Errorf("save room to redis failed: %v", err)
	}

	// 添加用戶到房間
	if err := s.addUserToRoom(ctx, roomID, userID, "creator"); err != nil {
		return nil, fmt.Errorf("add user to room failed: %v", err)
	}

	// 設置用戶當前房間
	if err := s.setUserCurrentRoom(ctx, userID, roomID); err != nil {
		return nil, fmt.Errorf("set user current room failed: %v", err)
	}

	// 添加到活躍房間列表
	if err := s.addToActiveRooms(ctx, roomID); err != nil {
		return nil, fmt.Errorf("add to active rooms failed: %v", err)
	}

	// 保存到 PostgreSQL (異步)
	go s.saveRoomToDatabase(roomInfo)

	utils.LogInfo("直播間創建成功: %s, 用戶: %d", roomID, userID)
	return roomInfo, nil
}

// GetActiveRooms 獲取活躍直播間列表
func (s *LiveRoomService) GetActiveRooms(limit int) ([]*LiveRoomInfo, error) {
	ctx := context.Background()

	if limit <= 0 {
		limit = 20
	}

	// 從 Redis 獲取活躍房間ID列表
	roomIDs, err := utils.GetRedisClient().ZRevRange(ctx, "live:active_rooms", 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("get active room ids failed: %v", err)
	}

	var rooms []*LiveRoomInfo
	for _, roomID := range roomIDs {
		room, err := s.GetRoomByID(roomID)
		if err != nil {
			utils.LogError("獲取房間信息失敗: %s, %v", roomID, err)
			continue
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// GetAllRooms 獲取所有直播間（包括已結束的）
func (s *LiveRoomService) GetAllRooms(limit int) ([]*LiveRoomInfo, error) {
	ctx := context.Background()

	if limit <= 0 {
		limit = 50
	}

	// 從 Redis 獲取所有房間ID（使用 pattern 匹配，只匹配房間主數據）
	pattern := "live:room:room_*"
	keys, err := utils.GetRedisClient().Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("get all room keys failed: %v", err)
	}

	var rooms []*LiveRoomInfo
	count := 0

	// 過濾出真正的房間主數據（排除 :roles, :users 等子鍵）
	var roomKeys []string
	for _, key := range keys {
		// 只保留房間主數據，排除包含冒號的子鍵
		if !strings.Contains(strings.TrimPrefix(key, "live:room:"), ":") {
			roomKeys = append(roomKeys, key)
		}
	}

	// 按創建時間排序（從新到舊）
	for i := len(roomKeys) - 1; i >= 0 && count < limit; i-- {
		key := roomKeys[i]
		roomID := strings.TrimPrefix(key, "live:room:")

		room, err := s.GetRoomByID(roomID)
		if err != nil {
			utils.LogError("獲取房間信息失敗: %s, %v", roomID, err)
			continue
		}

		rooms = append(rooms, room)
		count++
	}

	return rooms, nil
}

// GetRoomByID 根據ID獲取直播間信息
func (s *LiveRoomService) GetRoomByID(roomID string) (*LiveRoomInfo, error) {
	ctx := context.Background()

	// 從 Redis 獲取房間信息
	roomData, err := utils.GetRedisClient().HGetAll(ctx, fmt.Sprintf("live:room:%s", roomID)).Result()
	if err != nil {
		return nil, fmt.Errorf("get room from redis failed: %v", err)
	}

	if len(roomData) == 0 {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	// 解析房間信息
	room := &LiveRoomInfo{}
	if err := s.parseRoomFromRedis(roomData, room); err != nil {
		return nil, fmt.Errorf("parse room data failed: %v", err)
	}

	return room, nil
}

// JoinRoom 加入直播間
func (s *LiveRoomService) JoinRoom(roomID string, userID int) error {
	ctx := context.Background()

	// 檢查房間是否存在
	exists, err := utils.GetRedisClient().Exists(ctx, fmt.Sprintf("live:room:%s", roomID)).Result()
	if err != nil {
		return fmt.Errorf("check room exists failed: %v", err)
	}

	if exists == 0 {
		return fmt.Errorf("room not found: %s", roomID)
	}

	// 檢查房間狀態
	status, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s", roomID), "status").Result()
	if err != nil {
		return fmt.Errorf("get room status failed: %v", err)
	}

	if status == "cancelled" {
		return fmt.Errorf("room is not active: %s", status)
	}

	// 檢查用戶是否已在房間中
	isMember, err := utils.GetRedisClient().SIsMember(ctx, fmt.Sprintf("live:room:%s:users", roomID), userID).Result()
	if err != nil {
		return fmt.Errorf("check user in room failed: %v", err)
	}

	if isMember {
		return nil // 用戶已在房間中
	}

	// 後踢前：先離開當前所在的房間
	if err := s.leaveCurrentRoom(userID); err != nil {
		utils.LogError("離開當前房間失敗: %v", err)
		// 不返回錯誤，繼續加入新房間
	}

	// 檢查用戶是否為房間創建者
	creatorID, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s", roomID), "creator_id").Result()
	if err != nil {
		return fmt.Errorf("get room creator failed: %v", err)
	}

	// 根據用戶身份設置角色
	role := "viewer"
	if creatorID == strconv.Itoa(userID) {
		role = "creator"
	}

	// 添加用戶到房間
	if err := s.addUserToRoom(ctx, roomID, userID, role); err != nil {
		return fmt.Errorf("add user to room failed: %v", err)
	}

	// 增加觀眾數量
	if err := utils.GetRedisClient().HIncrBy(ctx, fmt.Sprintf("live:room:%s", roomID), "viewer_count", 1).Err(); err != nil {
		return fmt.Errorf("increment viewer count failed: %v", err)
	}

	// 設置用戶當前房間
	if err := s.setUserCurrentRoom(ctx, userID, roomID); err != nil {
		return fmt.Errorf("set user current room failed: %v", err)
	}

	utils.LogInfo("用戶 %d 加入直播間 %s", userID, roomID)
	return nil
}

// LeaveRoom 離開直播間
func (s *LiveRoomService) LeaveRoom(roomID string, userID int) error {
	ctx := context.Background()

	// 檢查用戶是否在房間中
	isMember, err := utils.GetRedisClient().SIsMember(ctx, fmt.Sprintf("live:room:%s:users", roomID), userID).Result()
	if err != nil {
		return fmt.Errorf("check user in room failed: %v", err)
	}

	if !isMember {
		return nil // 用戶不在房間中
	}

	// 從房間移除用戶
	if err := utils.GetRedisClient().SRem(ctx, fmt.Sprintf("live:room:%s:users", roomID), userID).Err(); err != nil {
		return fmt.Errorf("remove user from room failed: %v", err)
	}

	// 移除用戶角色
	if err := utils.GetRedisClient().HDel(ctx, fmt.Sprintf("live:room:%s:roles", roomID), strconv.Itoa(userID)).Err(); err != nil {
		return fmt.Errorf("remove user role failed: %v", err)
	}

	// 減少觀眾數量
	if err := utils.GetRedisClient().HIncrBy(ctx, fmt.Sprintf("live:room:%s", roomID), "viewer_count", -1).Err(); err != nil {
		return fmt.Errorf("decrement viewer count failed: %v", err)
	}

	// 清除用戶當前房間
	if err := utils.GetRedisClient().Del(ctx, fmt.Sprintf("user:%d:current_room", userID)).Err(); err != nil {
		return fmt.Errorf("clear user current room failed: %v", err)
	}

	utils.LogInfo("用戶 %d 離開直播間 %s", userID, roomID)
	return nil
}

// StartLive 開始直播
func (s *LiveRoomService) StartLive(roomID string, userID int) error {
	ctx := context.Background()

	// 檢查用戶是否為房間創建者
	role, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s:roles", roomID), strconv.Itoa(userID)).Result()
	if err != nil || role != "creator" {
		return fmt.Errorf("user is not room creator")
	}

	// 檢查房間狀態，允許從 created 或 ended 狀態開始直播
	status, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s", roomID), "status").Result()
	if err != nil {
		return fmt.Errorf("get room status failed: %v", err)
	}

	if status != "created" && status != "ended" {
		return fmt.Errorf("cannot start live from status: %s", status)
	}

	// 更新房間狀態
	now := time.Now()
	updates := map[string]interface{}{
		"status":     "live",
		"started_at": now.Format(time.RFC3339),
		"updated_at": now.Format(time.RFC3339),
	}

	// 如果是重新開始直播，清除結束時間
	if status == "ended" {
		updates["ended_at"] = ""
	}

	if err := utils.GetRedisClient().HMSet(ctx, fmt.Sprintf("live:room:%s", roomID), updates).Err(); err != nil {
		return fmt.Errorf("update room status failed: %v", err)
	}

	// 重新加入活躍房間列表
	if err := s.addToActiveRooms(ctx, roomID); err != nil {
		utils.LogError("重新加入活躍房間列表失敗: %v", err)
	}

	// 通過 WebSocket 通知所有用戶直播已開始
	if s.wsHandler != nil {
		if handler, ok := s.wsHandler.(interface {
			BroadcastRoomUpdate(roomID string, updateType string, data interface{})
		}); ok {
			handler.BroadcastRoomUpdate(roomID, "live_started", map[string]interface{}{
				"message": "直播已開始",
				"room_id": roomID,
				"status":  "live",
			})
		}
	}

	// 同步到資料庫
	go s.syncRoomToDatabase(roomID)

	utils.LogInfo("直播間 %s 開始直播", roomID)
	return nil
}

// EndLive 結束直播
func (s *LiveRoomService) EndLive(roomID string, userID int) error {
	ctx := context.Background()

	// 檢查用戶是否為房間創建者
	role, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s:roles", roomID), strconv.Itoa(userID)).Result()
	if err != nil || role != "creator" {
		return fmt.Errorf("user is not room creator")
	}

	// 更新房間狀態
	now := time.Now()
	updates := map[string]interface{}{
		"status":     "ended",
		"ended_at":   now.Format(time.RFC3339),
		"updated_at": now.Format(time.RFC3339),
	}

	if err := utils.GetRedisClient().HMSet(ctx, fmt.Sprintf("live:room:%s", roomID), updates).Err(); err != nil {
		return fmt.Errorf("update room status failed: %v", err)
	}

	// 通過 WebSocket 通知所有用戶直播已結束
	if s.wsHandler != nil {
		if handler, ok := s.wsHandler.(interface {
			BroadcastRoomUpdate(roomID string, updateType string, data interface{})
		}); ok {
			handler.BroadcastRoomUpdate(roomID, "live_ended", map[string]interface{}{
				"message": "直播已結束",
				"room_id": roomID,
				"status":  "ended",
			})
		}
	}

	// 從活躍房間列表移除
	if err := utils.GetRedisClient().ZRem(ctx, "live:active_rooms", roomID).Err(); err != nil {
		return fmt.Errorf("remove from active rooms failed: %v", err)
	}

	// 異步保存到資料庫
	go s.syncRoomToDatabase(roomID)

	utils.LogInfo("直播間 %s 結束直播", roomID)
	return nil
}

// CloseRoom 關閉直播間（完全刪除）
func (s *LiveRoomService) CloseRoom(roomID string, userID int) error {
	ctx := context.Background()

	// 檢查用戶是否為房間創建者
	role, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s:roles", roomID), strconv.Itoa(userID)).Result()
	if err != nil || role != "creator" {
		return fmt.Errorf("只有直播間創建者可以關閉直播間")
	}

	// 檢查房間狀態
	roomStatus, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s", roomID), "status").Result()
	if err != nil {
		return fmt.Errorf("獲取房間狀態失敗: %v", err)
	}

	// 如果房間正在直播中，先結束直播
	if roomStatus == "live" {
		if err := s.EndLive(roomID, userID); err != nil {
			return fmt.Errorf("結束直播失敗: %v", err)
		}
	}

	// 通過 WebSocket 通知所有用戶直播間已關閉
	if s.wsHandler != nil {
		if handler, ok := s.wsHandler.(interface {
			BroadcastRoomUpdate(roomID string, updateType string, data interface{})
		}); ok {
			handler.BroadcastRoomUpdate(roomID, "room_closed", map[string]interface{}{
				"message": "直播間已關閉",
				"room_id": roomID,
			})
		}
	}

	// 清除用戶當前房間記錄
	creatorID, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s", roomID), "creator_id").Result()
	if err == nil {
		if creatorIDInt, err := strconv.Atoi(creatorID); err == nil {
			// 清除創建者的當前房間記錄
			if err := utils.GetRedisClient().Del(ctx, fmt.Sprintf("user:%d:current_room", creatorIDInt)).Err(); err != nil {
				utils.LogError("清除創建者當前房間記錄失敗: %v", err)
			} else {
				utils.LogInfo("成功清除創建者當前房間記錄: user:%d:current_room", creatorIDInt)
			}
		}
	}
	
	// 清除房間內所有用戶的當前房間記錄
	roomUsers, err := utils.GetRedisClient().SMembers(ctx, fmt.Sprintf("live:room:%s:users", roomID)).Result()
	if err == nil {
		for _, userIDStr := range roomUsers {
			if userIDInt, err := strconv.Atoi(userIDStr); err == nil {
				// 檢查該用戶的當前房間是否為此房間
				currentRoom, err := utils.GetRedisClient().Get(ctx, fmt.Sprintf("user:%d:current_room", userIDInt)).Result()
				if err == nil && currentRoom == roomID {
					// 清除該用戶的當前房間記錄
					if err := utils.GetRedisClient().Del(ctx, fmt.Sprintf("user:%d:current_room", userIDInt)).Err(); err != nil {
						utils.LogError("清除用戶 %d 當前房間記錄失敗: %v", userIDInt, err)
					} else {
						utils.LogInfo("成功清除用戶 %d 當前房間記錄", userIDInt)
					}
				}
			}
		}
	}

	// 刪除房間相關的所有 Redis 數據
	keys := []string{
		fmt.Sprintf("live:room:%s", roomID),
		fmt.Sprintf("live:room:%s:users", roomID),
		fmt.Sprintf("live:room:%s:roles", roomID),
		fmt.Sprintf("live:room:%s:chat", roomID), // 清除聊天記錄
	}
	
	// 動態查找所有相關的鍵（使用 pattern 匹配）
	pattern := fmt.Sprintf("live:room:%s*", roomID)
	relatedKeys, err := utils.GetRedisClient().Keys(ctx, pattern).Result()
	if err == nil {
		// 將動態找到的鍵添加到刪除列表
		for _, key := range relatedKeys {
			// 避免重複添加
			found := false
			for _, existingKey := range keys {
				if existingKey == key {
					found = true
					break
				}
			}
			if !found {
				keys = append(keys, key)
			}
		}
	}

	// 使用 pipeline 批量刪除，提高效率
	pipe := utils.GetRedisClient().Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
	}

	// 從活躍房間列表移除
	pipe.ZRem(ctx, "live:active_rooms", roomID)

	// 執行所有操作
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		utils.LogError("清除房間 Redis 數據失敗: %v", err)
		return fmt.Errorf("清除房間數據失敗: %v", err)
	}

	// 檢查刪除結果
	for i, cmd := range cmds {
		if cmd.Err() != nil {
			utils.LogError("刪除鍵 %s 失敗: %v", keys[i], cmd.Err())
		} else {
			utils.LogInfo("成功刪除鍵: %s", keys[i])
		}
	}

	// 同步最終狀態到資料庫
	go s.syncRoomToDatabase(roomID)

	utils.LogInfo("直播間 %s 已關閉", roomID)
	return nil
}

// GetUserRole 獲取用戶在房間中的角色
func (s *LiveRoomService) GetUserRole(roomID string, userID int) (string, error) {
	ctx := context.Background()

	// 檢查房間是否存在
	exists, err := utils.GetRedisClient().Exists(ctx, fmt.Sprintf("live:room:%s", roomID)).Result()
	if err != nil {
		return "", fmt.Errorf("check room exists failed: %v", err)
	}

	if exists == 0 {
		return "", fmt.Errorf("room not found: %s", roomID)
	}

	// 檢查用戶是否在房間中
	isMember, err := utils.GetRedisClient().SIsMember(ctx, fmt.Sprintf("live:room:%s:users", roomID), userID).Result()
	if err != nil {
		return "", fmt.Errorf("check user in room failed: %v", err)
	}

	if !isMember {
		return "", fmt.Errorf("user not in room")
	}

	// 獲取用戶角色
	role, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s:roles", roomID), strconv.Itoa(userID)).Result()
	if err != nil {
		// 如果沒有角色記錄，檢查是否為創建者
		creatorID, err := utils.GetRedisClient().HGet(ctx, fmt.Sprintf("live:room:%s", roomID), "creator_id").Result()
		if err != nil {
			return "viewer", nil // 默認為觀眾
		}

		if creatorID == strconv.Itoa(userID) {
			return "creator", nil
		}
		return "viewer", nil
	}

	return role, nil
}

// 私有方法

// saveRoomToRedis 保存房間信息到 Redis
func (s *LiveRoomService) saveRoomToRedis(ctx context.Context, room *LiveRoomInfo) error {
	key := fmt.Sprintf("live:room:%s", room.ID)
	data := map[string]interface{}{
		"id":           room.ID,
		"title":        room.Title,
		"description":  room.Description,
		"creator_id":   room.CreatorID,
		"status":       room.Status,
		"stream_key":   room.StreamKey,
		"viewer_count": room.ViewerCount,
		"max_viewers":  room.MaxViewers,
	}

	// 安全地格式化時間字段
	if !room.StartedAt.IsZero() {
		data["started_at"] = room.StartedAt.Format(time.RFC3339)
	}
	if !room.CreatedAt.IsZero() {
		data["created_at"] = room.CreatedAt.Format(time.RFC3339)
	}
	if !room.UpdatedAt.IsZero() {
		data["updated_at"] = room.UpdatedAt.Format(time.RFC3339)
	}

	return utils.GetRedisClient().HMSet(ctx, key, data).Err()
}

// addUserToRoom 添加用戶到房間
func (s *LiveRoomService) addUserToRoom(ctx context.Context, roomID string, userID int, role string) error {
	// 添加到用戶列表
	if err := utils.GetRedisClient().SAdd(ctx, fmt.Sprintf("live:room:%s:users", roomID), userID).Err(); err != nil {
		return err
	}

	// 設置用戶角色
	return utils.GetRedisClient().HSet(ctx, fmt.Sprintf("live:room:%s:roles", roomID), userID, role).Err()
}

// setUserCurrentRoom 設置用戶當前房間
func (s *LiveRoomService) setUserCurrentRoom(ctx context.Context, userID int, roomID string) error {
	return utils.GetRedisClient().Set(ctx, fmt.Sprintf("user:%d:current_room", userID), roomID, 0).Err()
}

// addToActiveRooms 添加到活躍房間列表
func (s *LiveRoomService) addToActiveRooms(ctx context.Context, roomID string) error {
	return utils.GetRedisClient().ZAdd(ctx, "live:active_rooms", redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: roomID,
	}).Err()
}

// parseRoomFromRedis 從 Redis 數據解析房間信息
func (s *LiveRoomService) parseRoomFromRedis(data map[string]string, room *LiveRoomInfo) error {
	room.ID = data["id"]
	room.Title = data["title"]
	room.Description = data["description"]

	if creatorID, err := strconv.Atoi(data["creator_id"]); err == nil {
		room.CreatorID = creatorID
	}

	room.Status = data["status"]
	room.StreamKey = data["stream_key"]

	if viewerCount, err := strconv.Atoi(data["viewer_count"]); err == nil {
		room.ViewerCount = viewerCount
	}

	if maxViewers, err := strconv.Atoi(data["max_viewers"]); err == nil {
		room.MaxViewers = maxViewers
	}

	// 解析時間字段，如果為空或格式錯誤則使用零值
	if startedAtStr := data["started_at"]; startedAtStr != "" {
		if startedAt, err := time.Parse(time.RFC3339, startedAtStr); err == nil {
			room.StartedAt = startedAt
		}
	}

	if createdAtStr := data["created_at"]; createdAtStr != "" {
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			room.CreatedAt = createdAt
		}
	}

	if updatedAtStr := data["updated_at"]; updatedAtStr != "" {
		if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			room.UpdatedAt = updatedAt
		}
	}

	return nil
}

// saveRoomToDatabase 保存房間到資料庫
func (s *LiveRoomService) saveRoomToDatabase(room *LiveRoomInfo) error {
	session := &models.UserLiveSession{
		UserID:      room.CreatorID,
		RoomID:      room.ID,
		Title:       room.Title,
		Description: room.Description,
		StreamKey:   room.StreamKey,
		Status:      room.Status,
	}

	return s.db.Create(session).Error
}

// syncRoomToDatabase 同步房間數據到資料庫
func (s *LiveRoomService) syncRoomToDatabase(roomID string) error {
	// 獲取房間完整信息
	room, err := s.GetRoomByID(roomID)
	if err != nil {
		utils.LogError("同步房間數據失敗，房間不存在: %s, %v", roomID, err)
		return err
	}

	// 查找或創建資料庫記錄
	var session models.UserLiveSession
	if err := s.db.Where("room_id = ?", roomID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果記錄不存在，創建新記錄
			session = models.UserLiveSession{
				UserID:      room.CreatorID,
				RoomID:      room.ID,
				Title:       room.Title,
				Description: room.Description,
				StreamKey:   room.StreamKey,
				Status:      room.Status,
			}
			if err := s.db.Create(&session).Error; err != nil {
				return fmt.Errorf("創建直播記錄失敗: %v", err)
			}
		} else {
			return fmt.Errorf("查詢直播記錄失敗: %v", err)
		}
	}

	// 計算直播時長
	var duration int
	if !room.StartedAt.IsZero() {
		endTime := room.UpdatedAt
		if room.Status == "live" {
			endTime = time.Now()
		}
		duration = int(endTime.Sub(room.StartedAt).Seconds())
	}

	// 統計聊天消息數量
	totalMessages := s.getChatMessageCount(roomID)

	// 計算峰值觀眾數
	peakViewers := room.ViewerCount
	if session.PeakViewers > peakViewers {
		peakViewers = session.PeakViewers
	}

	updates := map[string]interface{}{
		"status":         room.Status,
		"peak_viewers":   peakViewers,
		"total_viewers":  room.ViewerCount,
		"total_messages": totalMessages,
	}

	// 只有在有開始時間時才更新時間相關字段
	if !room.StartedAt.IsZero() {
		updates["started_at"] = &room.StartedAt
		if room.Status == "ended" || room.Status == "cancelled" {
			updates["ended_at"] = &room.UpdatedAt
		}
		updates["duration"] = duration
	}

	return s.db.Model(&session).Updates(updates).Error
}

// getChatMessageCount 獲取聊天消息數量
func (s *LiveRoomService) getChatMessageCount(roomID string) int {
	ctx := context.Background()
	chatKey := fmt.Sprintf("live:room:%s:chat", roomID)

	count, err := utils.GetRedisClient().LLen(ctx, chatKey).Result()
	if err != nil {
		utils.LogError("獲取聊天消息數量失敗: %s, %v", roomID, err)
		return 0
	}

	return int(count)
}

// leaveCurrentRoom 離開當前房間（後踢前功能）
func (s *LiveRoomService) leaveCurrentRoom(userID int) error {
	ctx := context.Background()

	// 獲取用戶當前所在的房間
	currentRoomID, err := utils.GetRedisClient().Get(ctx, fmt.Sprintf("user:%d:current_room", userID)).Result()
	if err != nil {
		if err == redis.Nil {
			// 用戶沒有在房間中，這是正常情況
			return nil
		}
		return fmt.Errorf("get current room failed: %v", err)
	}

	if currentRoomID == "" {
		return nil // 用戶沒有在房間中
	}

	// 離開當前房間
	return s.LeaveRoom(currentRoomID, userID)
}
