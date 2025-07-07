package request

// UploadVideoRequest 上傳影片請求
type UploadVideoRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateVideoRequest 更新影片資訊請求
type UpdateVideoRequest struct {
	Title       string `json:"title" binding:"omitempty,min=1,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

// SearchVideoRequest 搜尋影片請求
type SearchVideoRequest struct {
	Query  string `form:"q" binding:"required"`
	Offset int    `form:"offset" binding:"min=0"`
	Limit  int    `form:"limit" binding:"min=1,max=50"`
}

// GenerateUploadURLRequest 生成上傳URL請求
type GenerateUploadURLRequest struct {
	Filename    string `json:"filename" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required,min=1"`
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// ConfirmUploadRequest 確認上傳完成請求
type ConfirmUploadRequest struct {
	VideoID uint   `json:"video_id" binding:"required"`
	S3Key   string `json:"s3_key" binding:"required"`
}
