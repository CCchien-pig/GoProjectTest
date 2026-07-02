package handler

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/service"
	"GoProject/udm/pkg/response"
)

// TelemetryHandler 處理遙測與告警事件相關 HTTP 請求
type TelemetryHandler struct {
	svc service.TelemetryService
}

// NewTelemetryHandler 建立 TelemetryHandler 實體
func NewTelemetryHandler(svc service.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{svc: svc}
}

// BatchIngest 批次上傳遙測數據
func (h *TelemetryHandler) BatchIngest(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(c, "invalid device id format")
		return
	}

	var req dto.BatchTelemetryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.svc.BatchInsert(c.Request.Context(), deviceID, &req)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.BadRequest(c, err.Error())
			return
		}
		handleError(c, err)
		return
	}

	response.OK(c, "telemetry ingested successfully")
}

// Query 查詢遙測時序資料（需帶時間範圍）
func (h *TelemetryHandler) Query(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(c, "invalid device id format")
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		response.BadRequest(c, "start and end query parameters are required")
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		response.BadRequest(c, "invalid start time format, must be RFC3339")
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		response.BadRequest(c, "invalid end time format, must be RFC3339")
		return
	}

	metricName := c.Query("metric_name")

	resp, err := h.svc.Query(c.Request.Context(), deviceID, start, end, metricName)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, resp)
}

// QueryLatest 查詢各指標最新一筆遙測資料
func (h *TelemetryHandler) QueryLatest(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(c, "invalid device id format")
		return
	}

	resp, err := h.svc.QueryLatest(c.Request.Context(), deviceID)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		handleError(c, err)
		return
	}

	response.OK(c, resp)
}

// DeleteByRange 刪除指定時間範圍的遙測資料
func (h *TelemetryHandler) DeleteByRange(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(c, "invalid device id format")
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		response.BadRequest(c, "start and end query parameters are required")
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		response.BadRequest(c, "invalid start time format, must be RFC3339")
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		response.BadRequest(c, "invalid end time format, must be RFC3339")
		return
	}

	err = h.svc.DeleteByRange(c.Request.Context(), deviceID, start, end)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		handleError(c, err)
		return
	}

	response.OK(c, "telemetry deleted successfully")
}

// QueryAlertEvents 查詢告警事件歷史資料
func (h *TelemetryHandler) QueryAlertEvents(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(c, "invalid device id format")
		return
	}

	month := c.Query("month")
	severity := c.Query("severity")

	resp, err := h.svc.QueryAlertEvents(c.Request.Context(), deviceID, month, severity)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		handleError(c, err)
		return
	}

	response.OK(c, resp)
}

// AcknowledgeAlertEvent 確認/消除告警事件
func (h *TelemetryHandler) AcknowledgeAlertEvent(c *gin.Context) {
	deviceIDStr := c.Param("device_id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(c, "invalid device id format")
		return
	}

	month := c.Param("month")
	triggeredAtStr := c.Param("triggered_at")
	ruleIDStr := c.Param("rule_id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		response.BadRequest(c, "invalid rule id format")
		return
	}

	triggeredAt, err := time.Parse(time.RFC3339, triggeredAtStr)
	if err != nil {
		response.BadRequest(c, "invalid triggered_at format, must be RFC3339")
		return
	}

	err = h.svc.AcknowledgeAlertEvent(c.Request.Context(), deviceID, month, triggeredAt, ruleID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, "alert event acknowledged successfully")
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrScyllaOffline) {
		response.ServiceUnavailable(c, err.Error())
		return
	}
	response.InternalError(c, err.Error())
}
