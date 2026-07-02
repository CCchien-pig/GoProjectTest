package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

// OK 回傳 200 成功
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "ok",
		Data:    data,
	})
}

// OKWithPagination 回傳 200 成功且帶有分頁資訊
func OKWithPagination(c *gin.Context, data interface{}, p *Pagination) {
	c.JSON(http.StatusOK, Response{
		Code:       http.StatusOK,
		Message:    "ok",
		Data:       data,
		Pagination: p,
	})
}

// BadRequest 回傳 400 客戶端請求錯誤
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    http.StatusBadRequest,
		Message: message,
	})
}

// NotFound 回傳 404 資源不存在
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    http.StatusNotFound,
		Message: message,
	})
}

// InternalError 回傳 500 伺服器內部錯誤
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: message,
	})
}

// ServiceUnavailable 回傳 503 服務不可用（例如資料庫斷線降級）
func ServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, Response{
		Code:    http.StatusServiceUnavailable,
		Message: message,
	})
}
