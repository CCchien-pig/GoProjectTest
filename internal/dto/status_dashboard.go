package dto

import (
	"github.com/google/uuid"
	"GoProject/udm/internal/model"
)

// StatusResp 設備即時狀態回傳 DTO
// Finding #8: 移至 dto package 與 DeviceResp 等保持一致的分層慣例
type StatusResp struct {
	DeviceID    uuid.UUID              `json:"device_id"`
	IsOnline    bool                   `json:"is_online"`
	Latest      []*model.TelemetryData `json:"latest_telemetry,omitempty"`
	AlertCounts map[string]int64       `json:"alert_counts,omitempty"`
}

// DashboardOverview 儀表板摘要 DTO
// Finding #8: 移至 dto package 統一管理所有 response 結構
type DashboardOverview struct {
	DeviceTotal  int64            `json:"device_total"`
	DeviceOnline int64            `json:"device_online"`
	AlertCounts  map[string]int64 `json:"alert_counts"` // info, warning, critical
}
