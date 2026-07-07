package routes

import (
	"github.com/gin-gonic/gin"
	"GoProject/udm/internal/handler"
	"GoProject/udm/internal/middleware"
)

// Dependencies 封裝所有 API 路由所需的 Handler 實體
type Dependencies struct {
	UserHandler      *handler.UserHandler
	DeviceHandler    *handler.DeviceHandler
	TelemetryHandler *handler.TelemetryHandler
	AlertRuleHandler *handler.AlertRuleHandler
	StatusHandler    *handler.StatusHandler
	DashboardHandler *handler.DashboardHandler
	CacheHandler     *handler.CacheHandler
	HealthHandler    *handler.HealthHandler
}

// Setup 設定 Gin 引擎路由與中間件
func Setup(deps *Dependencies) *gin.Engine {
	r := gin.New()

	// 載入基本中間件：日誌、Panic 恢復、TraceID 中間件
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.TraceID())

	// 健康檢查（包含各資料庫的健康狀態）
	r.GET("/health", deps.HealthHandler.Health)

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

			// Telemetry & Alert Events 子資源（ScyllaDB 相關端點）
			devices.POST("/:id/telemetry", deps.TelemetryHandler.BatchIngest)
			devices.GET("/:id/telemetry", deps.TelemetryHandler.Query)
			devices.GET("/:id/telemetry/latest", deps.TelemetryHandler.QueryLatest)
			devices.DELETE("/:id/telemetry", deps.TelemetryHandler.DeleteByRange)
			devices.GET("/:id/alert-events", deps.TelemetryHandler.QueryAlertEvents)

			// 即時狀態子資源
			devices.GET("/:id/status", deps.StatusHandler.GetStatus)
		}

		// 3. Alert Rules 獨立更新/刪除端點
		alertRules := v1.Group("/alert-rules")
		{
			alertRules.PUT("/:id", deps.AlertRuleHandler.Update)
			alertRules.DELETE("/:id", deps.AlertRuleHandler.Delete)
		}

		// 4. Alert Event ACK 專屬端點（含多個路徑參數）
		v1.PUT("/alert-events/:device_id/:month/:triggered_at/:rule_id/ack", deps.TelemetryHandler.AcknowledgeAlertEvent)

		// 5. 快取與儀表板端點
		v1.GET("/dashboard/overview", deps.DashboardHandler.GetOverview)
		v1.POST("/cache/invalidate", deps.CacheHandler.Invalidate)
	}

	return r
}
