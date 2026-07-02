package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
)

type mockAlertRuleService struct {
	onCreate       func(deviceID uuid.UUID, req *dto.CreateAlertRuleReq) (*dto.AlertRuleResp, error)
	onFindByDevice func(deviceID uuid.UUID) ([]*dto.AlertRuleResp, error)
	onUpdate       func(id uuid.UUID, req *dto.UpdateAlertRuleReq) (*dto.AlertRuleResp, error)
	onDelete       func(id uuid.UUID) error
}

func (m *mockAlertRuleService) Create(ctx context.Context, deviceID uuid.UUID, req *dto.CreateAlertRuleReq) (*dto.AlertRuleResp, error) {
	if m.onCreate != nil {
		return m.onCreate(deviceID, req)
	}
	return &dto.AlertRuleResp{
		ID:         uuid.New(),
		DeviceID:   deviceID,
		MetricName: req.MetricName,
		Operator:   req.Operator,
		Threshold:  req.Threshold,
		Severity:   req.Severity,
		IsEnabled:  req.IsEnabled,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *mockAlertRuleService) FindByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*dto.AlertRuleResp, error) {
	if m.onFindByDevice != nil {
		return m.onFindByDevice(deviceID)
	}
	return []*dto.AlertRuleResp{
		{ID: uuid.New(), DeviceID: deviceID, MetricName: "temperature", Operator: "gt", Threshold: 50.0},
	}, nil
}

func (m *mockAlertRuleService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateAlertRuleReq) (*dto.AlertRuleResp, error) {
	if m.onUpdate != nil {
		return m.onUpdate(id, req)
	}
	return &dto.AlertRuleResp{
		ID:        id,
		IsEnabled: true,
	}, nil
}

func (m *mockAlertRuleService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.onDelete != nil {
		return m.onDelete(id)
	}
	return nil
}

func TestAlertRuleHandler_Create(t *testing.T) {
	svc := &mockAlertRuleService{}
	h := NewAlertRuleHandler(svc)

	r := gin.New()
	r.POST("/devices/:id/alert-rules", h.Create)

	deviceID := uuid.New()
	reqPayload := dto.CreateAlertRuleReq{
		MetricName: "voltage",
		Operator:   "lt",
		Threshold:  12.0,
		Severity:   "warning",
		IsEnabled:  true,
	}

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/devices/"+deviceID.String()+"/alert-rules", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}
}

func TestAlertRuleHandler_FindByDeviceID(t *testing.T) {
	svc := &mockAlertRuleService{}
	h := NewAlertRuleHandler(svc)

	r := gin.New()
	r.GET("/devices/:id/alert-rules", h.FindByDeviceID)

	deviceID := uuid.New()
	req, _ := http.NewRequest("GET", "/devices/"+deviceID.String()+"/alert-rules", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}
}
