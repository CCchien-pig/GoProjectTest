package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
)

type mockTelemetryService struct {
	onBatchInsert           func(deviceID uuid.UUID, req *dto.BatchTelemetryReq) error
	onQuery                 func(deviceID uuid.UUID, start, end time.Time, metricName string) (*dto.TelemetryQueryResp, error)
	onQueryLatest           func(deviceID uuid.UUID) ([]*model.TelemetryData, error)
	onDeleteByRange         func(deviceID uuid.UUID, start, end time.Time) error
	onQueryAlertEvents      func(deviceID uuid.UUID, month, severity string) ([]*model.AlertEvent, error)
	onAcknowledgeAlertEvent func(deviceID uuid.UUID, month string, triggeredAt time.Time, ruleID uuid.UUID) error
}

func (m *mockTelemetryService) BatchInsert(ctx context.Context, deviceID uuid.UUID, req *dto.BatchTelemetryReq) error {
	if m.onBatchInsert != nil {
		return m.onBatchInsert(deviceID, req)
	}
	return nil
}

func (m *mockTelemetryService) Query(ctx context.Context, deviceID uuid.UUID, start, end time.Time, metricName string) (*dto.TelemetryQueryResp, error) {
	if m.onQuery != nil {
		return m.onQuery(deviceID, start, end, metricName)
	}
	return &dto.TelemetryQueryResp{DeviceID: deviceID}, nil
}

func (m *mockTelemetryService) QueryLatest(ctx context.Context, deviceID uuid.UUID) ([]*model.TelemetryData, error) {
	if m.onQueryLatest != nil {
		return m.onQueryLatest(deviceID)
	}
	return nil, nil
}

func (m *mockTelemetryService) DeleteByRange(ctx context.Context, deviceID uuid.UUID, start, end time.Time) error {
	if m.onDeleteByRange != nil {
		return m.onDeleteByRange(deviceID, start, end)
	}
	return nil
}

func (m *mockTelemetryService) QueryAlertEvents(ctx context.Context, deviceID uuid.UUID, month string, severity string) ([]*model.AlertEvent, error) {
	if m.onQueryAlertEvents != nil {
		return m.onQueryAlertEvents(deviceID, month, severity)
	}
	return nil, nil
}

func (m *mockTelemetryService) AcknowledgeAlertEvent(ctx context.Context, deviceID uuid.UUID, month string, triggeredAt time.Time, ruleID uuid.UUID) error {
	if m.onAcknowledgeAlertEvent != nil {
		return m.onAcknowledgeAlertEvent(deviceID, month, triggeredAt, ruleID)
	}
	return nil
}

func TestTelemetryHandler_BatchIngest(t *testing.T) {
	svc := &mockTelemetryService{}
	h := NewTelemetryHandler(svc)

	r := gin.New()
	r.POST("/devices/:id/telemetry", h.BatchIngest)

	devID := uuid.New()
	reqPayload := dto.BatchTelemetryReq{
		Points: []dto.TelemetryPoint{
			{RecordedAt: time.Now(), MetricName: "cpu", Value: 23.5, Unit: "%"},
		},
	}

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/devices/"+devID.String()+"/telemetry", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}
}

func TestTelemetryHandler_Query(t *testing.T) {
	svc := &mockTelemetryService{}
	h := NewTelemetryHandler(svc)

	r := gin.New()
	r.GET("/devices/:id/telemetry", h.Query)

	devID := uuid.New()
	start := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	end := time.Now().Format(time.RFC3339)

	q := url.Values{}
	q.Add("start", start)
	q.Add("end", end)

	req, _ := http.NewRequest("GET", "/devices/"+devID.String()+"/telemetry?"+q.Encode(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}
}
