package handler

import (
	"github.com/gin-gonic/gin"
	"GoProject/udm/internal/service"
	"GoProject/udm/pkg/response"
)

// DashboardHandler 儀表板 Handler
type DashboardHandler struct {
	svc service.DashboardService
}

// NewDashboardHandler 建立 DashboardHandler
func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// GetOverview 取得儀表板統計摘要
// GET /api/v1/dashboard/overview
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	resp, err := h.svc.GetOverview(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Finding #12: 使用 response.OK 保持回應格式一致性
	response.OK(c, resp)
}
