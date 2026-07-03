package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
)

type mockTelemetryRepository struct {
	points        []dto.TelemetryPoint
	onBatchInsert func(deviceID uuid.UUID, points []dto.TelemetryPoint) error
}

func (m *mockTelemetryRepository) BatchInsert(ctx context.Context, deviceID uuid.UUID, points []dto.TelemetryPoint) error {
	if m.onBatchInsert != nil {
		return m.onBatchInsert(deviceID, points)
	}
	m.points = append(m.points, points...)
	return nil
}

func (m *mockTelemetryRepository) Query(ctx context.Context, deviceID uuid.UUID, start, end time.Time, metricName string) ([]*model.TelemetryData, error) {
	return nil, nil
}

func (m *mockTelemetryRepository) QueryLatest(ctx context.Context, deviceID uuid.UUID) ([]*model.TelemetryData, error) {
	return nil, nil
}

func (m *mockTelemetryRepository) DeleteByRange(ctx context.Context, deviceID uuid.UUID, start, end time.Time) error {
	return nil
}

type mockAlertEventRepository struct {
	events   []*model.AlertEvent
	onInsert func(event *model.AlertEvent) error
}

func (m *mockAlertEventRepository) Insert(ctx context.Context, event *model.AlertEvent) error {
	if m.onInsert != nil {
		return m.onInsert(event)
	}
	m.events = append(m.events, event)
	return nil
}

func (m *mockAlertEventRepository) QueryByDevice(ctx context.Context, deviceID uuid.UUID, month string, severity string) ([]*model.AlertEvent, error) {
	return m.events, nil
}

func (m *mockAlertEventRepository) Acknowledge(ctx context.Context, deviceID uuid.UUID, month string, triggeredAt time.Time, ruleID uuid.UUID) error {
	return nil
}

func TestTelemetryService_BatchInsert_AlertTrigger(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	alertRuleRepo := newMockAlertRuleRepository()
	telemetryRepo := &mockTelemetryRepository{}
	alertRepo := &mockAlertEventRepository{}

	svc := NewTelemetryService(telemetryRepo, alertRepo, deviceRepo, alertRuleRepo)

	devID := uuid.New()
	dev := &model.Device{ID: devID, DeviceCode: "DEV-100"}
	deviceRepo.devices[devID] = dev

	rule := &model.AlertRule{
		ID:         uuid.New(),
		DeviceID:   devID,
		MetricName: "temperature",
		Operator:   "gt",
		Threshold:  40.0,
		Severity:   "critical",
		IsEnabled:  true,
	}
	alertRuleRepo.rules[rule.ID] = rule
	alertRuleRepo.rulesByDevice[devID] = append(alertRuleRepo.rulesByDevice[devID], rule)

	// 1. 測試沒有觸發告警的遙測寫入
	req := &dto.BatchTelemetryReq{
		Points: []dto.TelemetryPoint{
			{
				RecordedAt: time.Now(),
				MetricName: "temperature",
				Value:      35.0,
				Unit:       "C",
			},
		},
	}

	err := svc.BatchInsert(context.Background(), devID, req)
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	if len(alertRepo.events) != 0 {
		t.Errorf("expected 0 alert events, got %d", len(alertRepo.events))
	}

	// 2. 測試?�觸?��?警�??�測寫入 (45.0 > 40.0)
	reqTrigger := &dto.BatchTelemetryReq{
		Points: []dto.TelemetryPoint{
			{
				RecordedAt: time.Now(),
				MetricName: "temperature",
				Value:      45.0,
				Unit:       "C",
			},
		},
	}

	err = svc.BatchInsert(context.Background(), devID, reqTrigger)
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	if len(alertRepo.events) != 1 {
		t.Fatalf("expected 1 alert event to be triggered, got %d", len(alertRepo.events))
	}

	event := alertRepo.events[0]
	if event.MetricName != "temperature" || event.MetricValue != 45.0 || event.Threshold != 40.0 || event.Severity != "critical" {
		t.Errorf("unexpected alert event data: %+v", event)
	}
}
