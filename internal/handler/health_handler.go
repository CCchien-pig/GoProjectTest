package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"GoProject/udm/internal/keydb"
	"GoProject/udm/internal/scylla"
)

// HealthHandler 處理健康檢查請求
type HealthHandler struct {
	db           *gorm.DB
	scyllaClient *scylla.Client
	keydbClient  *keydb.Client
}

// NewHealthHandler 建立 HealthHandler 實體
func NewHealthHandler(db *gorm.DB, scyllaClient *scylla.Client, keydbClient *keydb.Client) *HealthHandler {
	return &HealthHandler{
		db:           db,
		scyllaClient: scyllaClient,
		keydbClient:  keydbClient,
	}
}

// Health 取得各資料庫連線健康狀態
// GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := "healthy"
	postgresStatus := "healthy"
	scyllaStatus := "healthy"
	keydbStatus := "healthy"

	// 1. PostgreSQL check
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			postgresStatus = "unhealthy"
			status = "degraded"
		} else if err := sqlDB.PingContext(ctx); err != nil {
			postgresStatus = "unhealthy"
			status = "degraded"
		}
	} else {
		postgresStatus = "unhealthy"
		status = "degraded"
	}

	// 2. ScyllaDB check
	if h.scyllaClient != nil && h.scyllaClient.Session != nil {
		err := h.scyllaClient.Session.Query("SELECT now() FROM system.local").WithContext(ctx).Exec()
		if err != nil {
			scyllaStatus = "unhealthy"
			status = "degraded"
		}
	} else {
		scyllaStatus = "unhealthy"
		status = "degraded"
	}

	// 3. KeyDB check
	if h.keydbClient != nil && h.keydbClient.Client != nil {
		err := h.keydbClient.Client.Ping(ctx).Err()
		if err != nil {
			keydbStatus = "unhealthy"
			status = "degraded"
		}
	} else {
		keydbStatus = "unhealthy"
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   status,
		"postgres": postgresStatus,
		"scylla":   scyllaStatus,
		"keydb":    keydbStatus,
	})
}
