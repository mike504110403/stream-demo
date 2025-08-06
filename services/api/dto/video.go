package dto

import "time"

// VideoDTO 影片資料傳輸物件
type VideoDTO struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`

	// 影片URL相關
	OriginalURL  string `json:"original_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	HLSMasterURL string `json:"hls_master_url"`
	MP4URL       string `json:"mp4_url"`

	// 影片屬性
	Duration       int    `json:"duration"`
	FileSize       int64  `json:"file_size"`
	OriginalFormat string `json:"original_format"`

	// 狀態相關
	Status             string `json:"status"`
	ProcessingProgress int    `json:"processing_progress"`
	ErrorMessage       string `json:"error_message,omitempty"`

	// 統計資料
	Views int64 `json:"views"`
	Likes int64 `json:"likes"`

	// 品質資訊
	Qualities []VideoQualityDTO `json:"qualities,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VideoQualityDTO 影片品質資料傳輸物件
type VideoQualityDTO struct {
	ID       uint   `json:"id"`
	Quality  string `json:"quality"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Bitrate  int    `json:"bitrate"`
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
	Status   string `json:"status"`
}

// VideoCreateDTO 建立影片請求
type VideoCreateDTO struct {
	Title       string `json:"title" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
	Filename    string `json:"filename" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required,min=1"`
}

// VideoUploadURLDTO 上傳URL響應
type VideoUploadURLDTO struct {
	UploadURL string            `json:"upload_url"`
	FormData  map[string]string `json:"form_data"`
	Key       string            `json:"key"`
	CDNUrl    string            `json:"cdn_url"`
	VideoID   uint              `json:"video_id"`
}

// VideoConfirmUploadDTO 確認上傳請求
type VideoConfirmUploadDTO struct {
	VideoID uint   `json:"video_id" binding:"required"`
	S3Key   string `json:"s3_key" binding:"required"`
}

// VideoUpdateDTO 更新影片請求
type VideoUpdateDTO struct {
	Title       string `json:"title" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

// VideoListDTO 影片列表回應
type VideoListDTO struct {
	Total  int64      `json:"total"`
	Videos []VideoDTO `json:"videos"`
}

// VideoSearchDTO 搜尋影片請求
type VideoSearchDTO struct {
	Query  string `form:"q" binding:"required"`
	Offset int    `form:"offset" binding:"min=0"`
	Limit  int    `form:"limit" binding:"min=1,max=50"`
	Status string `form:"status"`
}
