package middleware

// internal/middleware/trace.go — Request ID 中介層
// 參考 USCII internal/middleware/trace.go
//
// 每個 HTTP Request 注入唯一 UUID 作為 trace_id，
// 後續 log 全部帶此 ID，方便追蹤

// TODO: Day 1 實作
// func TraceID() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         requestID := uuid.New().String()
//         c.Set("request_id", requestID)
//         c.Header("X-Request-ID", requestID)
//         c.Next()
//     }
// }
