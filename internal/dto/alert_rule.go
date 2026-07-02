package dto

import (
	"time"

	"github.com/google/uuid"
	"GoProject/udm/internal/model"
)

// CreateAlertRuleReq 建�??�警規�?請�?
type CreateAlertRuleReq struct {
	MetricName string  `json:"metric_name" binding:"required,min=1,max=100"`
	Operator   string  `json:"operator" binding:"required,oneof=gt lt gte lte eq"`
	Threshold  float64 `json:"threshold" binding:"required"`
	Severity   string  `json:"severity" binding:"required,oneof=info warning critical"`
	IsEnabled  bool    `json:"is_enabled"`
}

// UpdateAlertRuleReq ?�新?�警規�?請�?
type UpdateAlertRuleReq struct {
	MetricName *string  `json:"metric_name" binding:"omitempty,min=1,max=100"`
	Operator   *string  `json:"operator" binding:"omitempty,oneof=gt lt gte lte eq"`
	Threshold  *float64 `json:"threshold"`
	Severity   *string  `json:"severity" binding:"omitempty,oneof=info warning critical"`
	IsEnabled  *bool    `json:"is_enabled"`
}

// AlertRuleResp ?�警規�??��? DTO
type AlertRuleResp struct {
	ID         uuid.UUID `json:"id"`
	DeviceID   uuid.UUID `json:"device_id"`
	MetricName string    `json:"metric_name"`
	Operator   string    `json:"operator"`
	Threshold  float64   `json:"threshold"`
	Severity   string    `json:"severity"`
	IsEnabled  bool      `json:"is_enabled"`
	CreatedAt  time.Time `json:"created_at"`
}

// ToAlertRuleResp �?model.AlertRule 轉為 dto.AlertRuleResp
func ToAlertRuleResp(rule *model.AlertRule) *AlertRuleResp {
	if rule == nil {
		return nil
	}
	return &AlertRuleResp{
		ID:         rule.ID,
		DeviceID:   rule.DeviceID,
		MetricName: rule.MetricName,
		Operator:   rule.Operator,
		Threshold:  rule.Threshold,
		Severity:   rule.Severity,
		IsEnabled:  rule.IsEnabled,
		CreatedAt:  rule.CreatedAt,
	}
}

// ToAlertRuleRespList ?��?轉�?
func ToAlertRuleRespList(rules []*model.AlertRule) []*AlertRuleResp {
	list := make([]*AlertRuleResp, len(rules))
	for i, r := range rules {
		list[i] = ToAlertRuleResp(r)
	}
	return list
}
