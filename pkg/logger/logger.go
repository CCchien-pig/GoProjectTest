package logger

import (
	"context"
	"log/slog"
	"os"
)

// ContextKey represents context key type
type ContextKey string

const RequestIDKey ContextKey = "request_id"

// ContextHandler wraps a slog.Handler to inject request_id from context
type ContextHandler struct {
	slog.Handler
}

// NewContextHandler creates a new ContextHandler
func NewContextHandler(h slog.Handler) slog.Handler {
	return &ContextHandler{Handler: h}
}

// Handle extracts request_id from context and adds it to the slog.Record attributes
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
			r.AddAttrs(slog.String("request_id", reqID))
		}
	}
	return h.Handler.Handle(ctx, r)
}

// InitLogger initializes the global slog logger with JSON handler and ContextHandler
func InitLogger() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	ctxHandler := NewContextHandler(jsonHandler)
	slog.SetDefault(slog.New(ctxHandler))
}
