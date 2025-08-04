package utils

// AppError 應用程式錯誤
type AppError struct {
	StatusCode int    // HTTP 狀態碼
	Code       string // 錯誤代碼
	Message    string // 錯誤訊息
}

// Error 實現 error 介面
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError 創建新的應用程式錯誤
func NewAppError(statusCode int, code string, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}
