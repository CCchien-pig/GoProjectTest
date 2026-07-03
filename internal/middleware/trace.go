package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceID 每個 HTTP Request 注入唯一 UUID 作為 trace_id，方便 log 共用此 ID
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
