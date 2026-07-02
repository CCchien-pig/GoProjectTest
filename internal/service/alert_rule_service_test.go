package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
)

type mockAlertRuleRepository struct {
	rules          map[uuid.UUID]*model.AlertRule
	rulesByDevice  map[uuid.UUID][]*model.AlertRule
	onCreate       func(rule *model.AlertRule) error
	onFindByID     func(id uuid.UUID) (*model.AlertRule, error)
	onUpdate       func(rule *model.AlertRule) error
	onDelete       func(id uuid.UUID) error
	onFindByDevice func(deviceID uuid.UUID) ([]*model.AlertRule, error)
}

func newMockAlertRuleRepository() *mockAlertRuleRepository {
	return &mockAlertRuleRepository{
		rules:         make(map[uuid.UUID]*model.AlertRule),
		rulesByDevice: make(map[uuid.UUID][]*model.AlertRule),
	}
}

func (m *mockAlertRuleRepository) Create(ctx context.Context, rule *model.AlertRule) error {
	if m.onCreate != nil {
		return m.onCreate(rule)
	}
	rule.ID = uuid.New()
	m.rules[rule.ID] = rule
	m.rulesByDevice[rule.DeviceID] = append(m.rulesByDevice[rule.DeviceID], rule)
	return nil
}

func (m *mockAlertRuleRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AlertRule, error) {
	if m.onFindByID != nil {
		return m.onFindByID(id)
	}
	return m.rules[id], nil
}

func (m *mockAlertRuleRepository) FindByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*model.AlertRule, error) {
	if m.onFindByDevice != nil {
		return m.onFindByDevice(deviceID)
	}
	return m.rulesByDevice[deviceID], nil
}

func (m *mockAlertRuleRepository) Update(ctx context.Context, rule *model.AlertRule) error {
	if m.onUpdate != nil {
		return m.onUpdate(rule)
	}
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockAlertRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.onDelete != nil {
		return m.onDelete(id)
	}
	rule := m.rules[id]
	if rule != nil {
		delete(m.rules, id)
		devRules := m.rulesByDevice[rule.DeviceID]
		for i, r := range devRules {
			if r.ID == id {
				m.rulesByDevice[rule.DeviceID] = append(devRules[:i], devRules[i+1:]...)
				break
			}
		}
	}
	return nil
}

func TestAlertRuleService_Create(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	repo := newMockAlertRuleRepository()
	svc := NewAlertRuleService(repo, deviceRepo)

	dev := &model.Device{ID: uuid.New(), DeviceCode: "DEV-100"}
	deviceRepo.devices[dev.ID] = dev

	req := &dto.CreateAlertRuleReq{
		MetricName: "temperature",
		Operator:   "gt",
		Threshold:  50.0,
		Severity:   "critical",
		IsEnabled:  true,
	}

	resp, err := svc.Create(context.Background(), dev.ID, req)
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	if resp.MetricName != "temperature" || resp.Operator != "gt" || resp.Threshold != 50.0 || resp.Severity != "critical" {
		t.Errorf("unexpected rule response: %+v", resp)
	}

	// ?��?設�?
	_, err = svc.Create(context.Background(), uuid.New(), req)
	if err == nil || err != ErrDeviceNotFound {
		t.Errorf("expected ErrDeviceNotFound, got %v", err)
	}

	// ?��??��?�?	req.Operator = "invalid"
	_, err = svc.Create(context.Background(), dev.ID, req)
	if err == nil || err != ErrInvalidOperator {
		t.Errorf("expected ErrInvalidOperator, got %v", err)
	}

	// ?��??��?等�?
	req.Operator = "gt"
	req.Severity = "epic"
	_, err = svc.Create(context.Background(), dev.ID, req)
	if err == nil || err != ErrInvalidSeverity {
		t.Errorf("expected ErrInvalidSeverity, got %v", err)
	}
}
