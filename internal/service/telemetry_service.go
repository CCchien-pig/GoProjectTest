package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-name/udm/internal/dto"
	"github.com/your-name/udm/internal/model"
	"github.com/your-name/udm/internal/repository"
	"github.com/your-name/udm/internal/scylla"
)

var (
	// ErrScyllaOffline 代表 ScyllaDB 斷線/離線
	ErrScyllaOffline = errors.New("scylla time-series database is offline")
)

// TelemetryService 定義遙測與告警事件業務邏輯介面
type TelemetryService interface {
	BatchInsert(ctx context.Context, deviceID uuid.UUID, req *dto.BatchTelemetryReq) error
	Query(ctx context.Context, deviceID uuid.UUID, start, end time.Time, metricName string) (*dto.TelemetryQueryResp, error)
	QueryLatest(ctx context.Context, deviceID uuid.UUID) ([]*model.TelemetryData, error)
	DeleteByRange(ctx context.Context, deviceID uuid.UUID, start, end time.Time) error
	QueryAlertEvents(ctx context.Context, deviceID uuid.UUID, month string, severity string) ([]*model.AlertEvent, error)
	AcknowledgeAlertEvent(ctx context.Context, deviceID uuid.UUID, month string, triggeredAt time.Time, ruleID uuid.UUID) error
}

type telemetryService struct {
	telemetryRepo scylla.TelemetryRepository
	alertRepo     scylla.AlertEventRepository
	deviceRepo    repository.DeviceRepository
	alertRuleRepo repository.AlertRuleRepository
}

// NewTelemetryService 建立 TelemetryService 實作
func NewTelemetryService(
	telemetryRepo scylla.TelemetryRepository,
	alertRepo scylla.AlertEventRepository,
	deviceRepo repository.DeviceRepository,
	alertRuleRepo repository.AlertRuleRepository,
) TelemetryService {
	return &telemetryService{
		telemetryRepo: telemetryRepo,
		alertRepo:     alertRepo,
		deviceRepo:    deviceRepo,
		alertRuleRepo: alertRuleRepo,
	}
}

func (s *telemetryService) BatchInsert(ctx context.Context, deviceID uuid.UUID, req *dto.BatchTelemetryReq) error {
	// 1. 檢查設備是否存在
	dev, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("find device: %w", err)
	}
	if dev == nil {
		return ErrDeviceNotFound
	}

	// 2. 寫入 ScyllaDB 遙測時序表 (若 ScyllaDB 斷線，回報錯誤降級)
	if s.telemetryRepo == nil {
		return ErrScyllaOffline
	}
	if err := s.telemetryRepo.BatchInsert(ctx, deviceID, req.Points); err != nil {
		return fmt.Errorf("batch insert telemetry: %w", err)
	}

	// 3. 告警評估邏輯：查出該設備的所有告警規則
	rules, err := s.alertRuleRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("find alert rules: %w", err)
	}

	for _, rule := range rules {
		if !rule.IsEnabled {
			continue
		}

		for _, p := range req.Points {
			if p.MetricName == rule.MetricName {
				// 比對閾值
				if checkThreshold(p.Value, rule.Threshold, rule.Operator) {
					// 觸發告警，寫入 ScyllaDB
					event := &model.AlertEvent{
						DeviceID:     deviceID,
						Month:        p.RecordedAt.Format("2006-01"),
						TriggeredAt:  p.RecordedAt,
						RuleID:       rule.ID,
						MetricName:   p.MetricName,
						MetricValue:  p.Value,
						Threshold:    rule.Threshold,
						Severity:     rule.Severity,
						Acknowledged: false,
					}
					if s.alertRepo != nil {
						if err := s.alertRepo.Insert(ctx, event); err != nil {
							// 記錄 log，不應因為告警寫入失敗而導致遙測 API 崩潰
							fmt.Printf("failed to insert alert event: %v\n", err)
						}
					}
				}
			}
		}
	}

	return nil
}

func (s *telemetryService) Query(ctx context.Context, deviceID uuid.UUID, start, end time.Time, metricName string) (*dto.TelemetryQueryResp, error) {
	// 確認設備是否存在（若不存在代表已被刪除，但我們依然回傳歷史資料並加上 is_deleted 標記）
	dev, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("find device: %w", err)
	}

	isDeleted := (dev == nil)

	if s.telemetryRepo == nil {
		return nil, ErrScyllaOffline
	}
	data, err := s.telemetryRepo.Query(ctx, deviceID, start, end, metricName)
	if err != nil {
		return nil, fmt.Errorf("query telemetry: %w", err)
	}

	points := make([]dto.TelemetryPoint, len(data))
	for i, d := range data {
		points[i] = dto.TelemetryPoint{
			RecordedAt: d.RecordedAt,
			MetricName: d.MetricName,
			Value:      d.Value,
			Unit:       d.Unit,
			Tags:       d.Tags,
		}
	}

	return &dto.TelemetryQueryResp{
		DeviceID:  deviceID,
		IsDeleted: isDeleted,
		Points:    points,
	}, nil
}

func (s *telemetryService) QueryLatest(ctx context.Context, deviceID uuid.UUID) ([]*model.TelemetryData, error) {
	dev, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("find device: %w", err)
	}
	if dev == nil {
		return nil, ErrDeviceNotFound
	}

	if s.telemetryRepo == nil {
		return nil, ErrScyllaOffline
	}
	return s.telemetryRepo.QueryLatest(ctx, deviceID)
}

func (s *telemetryService) DeleteByRange(ctx context.Context, deviceID uuid.UUID, start, end time.Time) error {
	dev, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("find device: %w", err)
	}
	if dev == nil {
		return ErrDeviceNotFound
	}

	if s.telemetryRepo == nil {
		return ErrScyllaOffline
	}
	return s.telemetryRepo.DeleteByRange(ctx, deviceID, start, end)
}

func (s *telemetryService) QueryAlertEvents(ctx context.Context, deviceID uuid.UUID, month string, severity string) ([]*model.AlertEvent, error) {
	dev, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("find device: %w", err)
	}
	if dev == nil {
		return nil, ErrDeviceNotFound
	}

	if month == "" {
		month = time.Now().Format("2006-01")
	}

	if s.alertRepo == nil {
		return nil, ErrScyllaOffline
	}
	return s.alertRepo.QueryByDevice(ctx, deviceID, month, severity)
}

func (s *telemetryService) AcknowledgeAlertEvent(ctx context.Context, deviceID uuid.UUID, month string, triggeredAt time.Time, ruleID uuid.UUID) error {
	if s.alertRepo == nil {
		return ErrScyllaOffline
	}
	return s.alertRepo.Acknowledge(ctx, deviceID, month, triggeredAt, ruleID)
}

func checkThreshold(val, threshold float64, op string) bool {
	switch op {
	case "gt":
		return val > threshold
	case "lt":
		return val < threshold
	case "gte":
		return val >= threshold
	case "lte":
		return val <= threshold
	case "eq":
		return val == threshold
	}
	return false
}
