package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"GoProject/udm/internal/service"
	"GoProject/udm/pkg/response"
)

// StatusHandler 設備即時狀態 Handler
type StatusHandler struct {
	svc service.StatusService
}

// NewStatusHandler 建立 StatusHandler
func NewStatusHandler(svc service.StatusService) *StatusHandler {
	return &StatusHandler{svc: svc}
}

// GetStatus 取得設備即時狀態（在線/離線 + 最新遙測 + 告警計數）
// GET /api/v1/devices/:id/status
func (h *StatusHandler) GetStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid device ID format")
		return
	}

	resp, err := h.svc.GetDeviceStatus(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Finding #12: 使用 response.OK 保持回應格式一致性
	response.OK(c, resp)
}
