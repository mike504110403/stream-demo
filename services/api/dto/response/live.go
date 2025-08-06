package response

import "time"

// LiveResponse 直播資訊回應
type LiveResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UserID      uint      `json:"user_id"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
	ViewerCount int64     `json:"viewer_count"`
	ChatEnabled bool      `json:"chat_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LiveListResponse 直播列表回應
type LiveListResponse struct {
	Total int64          `json:"total"`
	Lives []LiveResponse `json:"lives"`
}

// StreamKeyResponse 串流金鑰回應
type StreamKeyResponse struct {
	StreamKey string `json:"stream_key"`
}

// NewLiveResponse 從模型創建直播回應
func NewLiveResponse(live interface{}) *LiveResponse {
	// 這裡需要實現從模型到 DTO 的轉換邏輯
	return &LiveResponse{}
}

// NewLiveListResponse 創建直播列表回應
func NewLiveListResponse(total int64, lives []interface{}) *LiveListResponse {
	liveResponses := make([]LiveResponse, len(lives))
	for i, live := range lives {
		liveResponses[i] = *NewLiveResponse(live)
	}
	return &LiveListResponse{
		Total: total,
		Lives: liveResponses,
	}
}

// NewStreamKeyResponse 創建串流金鑰回應
func NewStreamKeyResponse(streamKey string) *StreamKeyResponse {
	return &StreamKeyResponse{
		StreamKey: streamKey,
	}
}
