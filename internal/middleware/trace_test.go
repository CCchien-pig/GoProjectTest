package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestTraceID(t *testing.T) {
	r := gin.New()
	r.Use(TraceID())

	var capturedTraceID string
	r.GET("/test", func(c *gin.Context) {
		val, exists := c.Get("request_id")
		if exists {
			capturedTraceID = val.(string)
		}
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	headerTraceID := w.Header().Get("X-Request-ID")
	if headerTraceID == "" {
		t.Error("expected X-Request-ID header to be set, but got empty")
	}

	if _, err := uuid.Parse(headerTraceID); err != nil {
		t.Errorf("X-Request-ID header is not a valid UUID: %s, error: %v", headerTraceID, err)
	}

	if capturedTraceID != headerTraceID {
		t.Errorf("context trace_id (%s) does not match header X-Request-ID (%s)", capturedTraceID, headerTraceID)
	}
}
