package response

// pkg/response/response.go — 統一 JSON 回傳格式
// 所有 API 回傳使用此格式：
//   {"code": 200, "message": "ok", "data": {...}, "pagination": {...}}

// Response 統一回傳格式
type Response struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination cursor-based 分頁資訊
type Pagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}

// TODO: Day 1 實作
// func OK(c *gin.Context, data interface{}) {}
// func OKWithPagination(c *gin.Context, data interface{}, p *Pagination) {}
// func BadRequest(c *gin.Context, message string) {}
// func NotFound(c *gin.Context, message string) {}
// func InternalError(c *gin.Context, message string) {}
// func ServiceUnavailable(c *gin.Context, message string) {}
