package response

import "time"

// VideoResponse 影片資訊回應
type VideoResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UserID      uint      `json:"user_id"`
	Status      string    `json:"status"`
	Duration    int       `json:"duration"`
	Thumbnail   string    `json:"thumbnail"`
	VideoURL    string    `json:"video_url"`
	Views       int64     `json:"views"`
	Likes       int64     `json:"likes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VideoListResponse 影片列表回應
type VideoListResponse struct {
	Total  int64           `json:"total"`
	Videos []VideoResponse `json:"videos"`
}

// NewVideoResponse 從模型創建影片回應
func NewVideoResponse(video interface{}) *VideoResponse {
	// 這裡需要實現從模型到 DTO 的轉換邏輯
	return &VideoResponse{}
}

// NewVideoListResponse 創建影片列表回應
func NewVideoListResponse(total int64, videos []interface{}) *VideoListResponse {
	videoResponses := make([]VideoResponse, len(videos))
	for i, video := range videos {
		videoResponses[i] = *NewVideoResponse(video)
	}
	return &VideoListResponse{
		Total:  total,
		Videos: videoResponses,
	}
}
