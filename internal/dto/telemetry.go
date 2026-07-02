package dto

import (
	"time"

	"github.com/google/uuid"
)

// TelemetryPoint 單一遙測資料點
type TelemetryPoint struct {
	RecordedAt time.Time         `json:"recorded_at" binding:"required"`
	MetricName string            `json:"metric_name" binding:"required,min=1"`
	Value      float64           `json:"value" binding:"required"`
	Unit       string            `json:"unit" binding:"required"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// BatchTelemetryReq 批次上傳遙測資料請求
type BatchTelemetryReq struct {
	Points []TelemetryPoint `json:"points" binding:"required,dive,required"`
}

// TelemetryQueryResp 遙測資料查詢回應 DTO
type TelemetryQueryResp struct {
	DeviceID  uuid.UUID        `json:"device_id"`
	IsDeleted bool             `json:"is_deleted"`
	Points    []TelemetryPoint `json:"points"`
}
