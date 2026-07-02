package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-name/udm/internal/dto"
	"github.com/your-name/udm/internal/service"
	"github.com/your-name/udm/pkg/response"
)

// AlertRuleHandler 處理告警規則相關 HTTP 請求
type AlertRuleHandler struct {
	svc service.AlertRuleService
}

// NewAlertRuleHandler 建立 AlertRuleHandler 實作
func NewAlertRuleHandler(svc service.AlertRuleService) *AlertRuleHandler {
	return &AlertRuleHandler{svc: svc}
}

// Create 建立告警規則
func (h *AlertRuleHandler) Create(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(c, "invalid device id format")
		return
	}

	var req dto.CreateAlertRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), deviceID, &req)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) || errors.Is(err, service.ErrInvalidOperator) || errors.Is(err, service.ErrInvalidSeverity) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, resp)
}

// FindByDeviceID 取得某設備的所有告警規則
func (h *AlertRuleHandler) FindByDeviceID(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(c, "invalid device id format")
		return
	}

	resp, err := h.svc.FindByDeviceID(c.Request.Context(), deviceID)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, resp)
}

// Update 更新告警規則
func (h *AlertRuleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid uuid format")
		return
	}

	var req dto.UpdateAlertRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrAlertRuleNotFound) || errors.Is(err, service.ErrInvalidOperator) || errors.Is(err, service.ErrInvalidSeverity) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, resp)
}

// Delete 刪除告警規則
func (h *AlertRuleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid uuid format")
		return
	}

	err = h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrAlertRuleNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "alert rule deleted successfully")
}
