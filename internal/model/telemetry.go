package model

import (
	"time"

	"github.com/google/uuid"
)

// TelemetryData 代表 ScyllaDB 中 telemetry 表的時序資料結構
type TelemetryData struct {
	DeviceID   uuid.UUID         `json:"device_id"`
	Date       string            `json:"date"`
	RecordedAt time.Time         `json:"recorded_at"`
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Unit       string            `json:"unit"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// AlertEvent 代表 ScyllaDB 中 alert_events 表的資料結構
type AlertEvent struct {
	DeviceID    uuid.UUID `json:"device_id"`
	Month       string    `json:"month"` // e.g. "2026-07"
	TriggeredAt time.Time `json:"triggered_at"`
	RuleID      uuid.UUID `json:"rule_id"`
	MetricName  string    `json:"metric_name"`
	MetricValue float64   `json:"metric_value"`
	Threshold   float64   `json:"threshold"`
	Severity    string    `json:"severity"`
	Acknowledged bool      `json:"acknowledged"`
}
