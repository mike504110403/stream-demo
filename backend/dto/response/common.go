package response

// Response 通用響應結構
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResponse 分頁回應結構
type PageResponse struct {
	Total  int64       `json:"total"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
	Data   interface{} `json:"data"`
}

// ErrorResponse 錯誤響應結構
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ListResponse 列表響應結構
type ListResponse struct {
	Total int64         `json:"total"`
	Items []interface{} `json:"items"`
}

// NewSuccessResponse 創建成功響應
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// NewErrorResponse 創建錯誤響應
func NewErrorResponse(code int, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// NewPageResponse 創建分頁回應
func NewPageResponse(total int64, offset, limit int, data interface{}) *PageResponse {
	return &PageResponse{
		Total:  total,
		Offset: offset,
		Limit:  limit,
		Data:   data,
	}
}

// NewListResponse 創建列表響應
func NewListResponse[T any](total int64, items []T) *ListResponse {
	interfaceItems := make([]interface{}, len(items))
	for i, item := range items {
		interfaceItems[i] = item
	}
	return &ListResponse{
		Total: total,
		Items: interfaceItems,
	}
}
