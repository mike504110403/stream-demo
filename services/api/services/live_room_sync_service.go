package services

import (
	"context"
	"time"

	"stream-demo/backend/utils"
)

// LiveRoomSyncService 直播間數據同步服務
type LiveRoomSyncService struct {
	liveRoomService *LiveRoomService
	stopChan        chan bool
	ticker          *time.Ticker
}

// NewLiveRoomSyncService 創建同步服務
func NewLiveRoomSyncService(liveRoomService *LiveRoomService) *LiveRoomSyncService {
	return &LiveRoomSyncService{
		liveRoomService: liveRoomService,
		stopChan:        make(chan bool),
	}
}

// Start 啟動同步服務
func (s *LiveRoomSyncService) Start() {
	// 每5分鐘同步一次活躍房間數據
	s.ticker = time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.syncActiveRooms()
			case <-s.stopChan:
				s.ticker.Stop()
				return
			}
		}
	}()

	utils.LogInfo("直播間數據同步服務已啟動")
}

// Stop 停止同步服務
func (s *LiveRoomSyncService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
	utils.LogInfo("直播間數據同步服務已停止")
}

// syncActiveRooms 同步所有活躍房間數據
func (s *LiveRoomSyncService) syncActiveRooms() {
	ctx := context.Background()

	// 獲取所有活躍房間ID
	roomIDs, err := utils.GetRedisClient().ZRange(ctx, "live:active_rooms", 0, -1).Result()
	if err != nil {
		utils.LogError("獲取活躍房間列表失敗: %v", err)
		return
	}

	if len(roomIDs) == 0 {
		utils.LogInfo("沒有活躍房間需要同步")
		return
	}

	utils.LogInfo("開始同步 %d 個活躍房間的數據", len(roomIDs))

	// 同步每個房間的數據
	for _, roomID := range roomIDs {
		go func(id string) {
			if err := s.liveRoomService.syncRoomToDatabase(id); err != nil {
				utils.LogError("同步房間 %s 數據失敗: %v", id, err)
			} else {
				utils.LogInfo("房間 %s 數據同步成功", id)
			}
		}(roomID)
	}
}

// SyncRoom 立即同步指定房間數據
func (s *LiveRoomSyncService) SyncRoom(roomID string) error {
	return s.liveRoomService.syncRoomToDatabase(roomID)
}
