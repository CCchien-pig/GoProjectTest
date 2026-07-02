package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-name/udm/internal/handler"
	"github.com/your-name/udm/internal/middleware"
)

// Dependencies 封裝所有 API 路由所需的 Handler 實作
type Dependencies struct {
	UserHandler      *handler.UserHandler
	DeviceHandler    *handler.DeviceHandler
	TelemetryHandler *handler.TelemetryHandler
	AlertRuleHandler *handler.AlertRuleHandler
}

// Setup 初始化 Gin 引擎路由與中間件
func Setup(deps *Dependencies) *gin.Engine {
	r := gin.New()

	// 載入基本中間件與自訂的 TraceID 中間件
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.TraceID())

	// 簡易健康檢查 (後續降級處理會進行擴充)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	v1 := r.Group("/api/v1")
	{
		// 1. Users CRUD 端點
		users := v1.Group("/users")
		{
			users.POST("", deps.UserHandler.Create)
			users.GET("/:id", deps.UserHandler.FindByID)
			users.PUT("/:id", deps.UserHandler.Update)
			users.DELETE("/:id", deps.UserHandler.SoftDelete)
		}

		// 2. Devices & 子資源端點
		devices := v1.Group("/devices")
		{
			devices.POST("", deps.DeviceHandler.Create)
			devices.GET("", deps.DeviceHandler.List)
			devices.GET("/:id", deps.DeviceHandler.FindByID)
			devices.PUT("/:id", deps.DeviceHandler.Update)
			devices.DELETE("/:id", deps.DeviceHandler.Delete)

			// Alert Rules 子資源
			devices.POST("/:id/alert-rules", deps.AlertRuleHandler.Create)
			devices.GET("/:id/alert-rules", deps.AlertRuleHandler.FindByDeviceID)

			// Telemetry & Alert Events 子資源 (ScyllaDB 時序數據)
			devices.POST("/:id/telemetry", deps.TelemetryHandler.BatchIngest)
			devices.GET("/:id/telemetry", deps.TelemetryHandler.Query)
			devices.GET("/:id/telemetry/latest", deps.TelemetryHandler.QueryLatest)
			devices.DELETE("/:id/telemetry", deps.TelemetryHandler.DeleteByRange)
			devices.GET("/:id/alert-events", deps.TelemetryHandler.QueryAlertEvents)
		}

		// 3. Alert Rules 獨立更新/刪除端點
		alertRules := v1.Group("/alert-rules")
		{
			alertRules.PUT("/:id", deps.AlertRuleHandler.Update)
			alertRules.DELETE("/:id", deps.AlertRuleHandler.Delete)
		}

		// 4. Alert Event ACK 專屬端點 (配合多集群主鍵)
		v1.PUT("/alert-events/:device_id/:month/:triggered_at/:rule_id/ack", deps.TelemetryHandler.AcknowledgeAlertEvent)
	}

	return r
}
