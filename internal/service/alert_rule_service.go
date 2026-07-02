package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-name/udm/internal/dto"
	"github.com/your-name/udm/internal/model"
	"github.com/your-name/udm/internal/repository"
)

var (
	// ErrAlertRuleNotFound 代表找不到告警規則
	ErrAlertRuleNotFound = errors.New("alert rule not found")
	// ErrInvalidOperator 代表不合法的運算子
	ErrInvalidOperator = errors.New("invalid operator, must be gt, lt, gte, lte, eq")
	// ErrInvalidSeverity 代表不合法的嚴重等級
	ErrInvalidSeverity = errors.New("invalid severity, must be info, warning, critical")
)

// AlertRuleService 定義告警規則業務邏輯介面
type AlertRuleService interface {
	Create(ctx context.Context, deviceID uuid.UUID, req *dto.CreateAlertRuleReq) (*dto.AlertRuleResp, error)
	FindByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*dto.AlertRuleResp, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateAlertRuleReq) (*dto.AlertRuleResp, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type alertRuleService struct {
	repo       repository.AlertRuleRepository
	deviceRepo repository.DeviceRepository
}

// NewAlertRuleService 建立 AlertRuleService 實作
func NewAlertRuleService(repo repository.AlertRuleRepository, deviceRepo repository.DeviceRepository) AlertRuleService {
	return &alertRuleService{
		repo:       repo,
		deviceRepo: deviceRepo,
	}
}

func (s *alertRuleService) Create(ctx context.Context, deviceID uuid.UUID, req *dto.CreateAlertRuleReq) (*dto.AlertRuleResp, error) {
	// 檢查設備是否存在
	dev, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("find device: %w", err)
	}
	if dev == nil {
		return nil, ErrDeviceNotFound
	}

	// 驗證運算子
	if !isValidOperator(req.Operator) {
		return nil, ErrInvalidOperator
	}

	// 驗證嚴重等級
	if !isValidSeverity(req.Severity) {
		return nil, ErrInvalidSeverity
	}

	rule := &model.AlertRule{
		DeviceID:   deviceID,
		MetricName: req.MetricName,
		Operator:   req.Operator,
		Threshold:  req.Threshold,
		Severity:   req.Severity,
		IsEnabled:  req.IsEnabled,
	}

	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("create alert rule: %w", err)
	}

	return dto.ToAlertRuleResp(rule), nil
}

func (s *alertRuleService) FindByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*dto.AlertRuleResp, error) {
	// 檢查設備是否存在
	dev, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("find device: %w", err)
	}
	if dev == nil {
		return nil, ErrDeviceNotFound
	}

	rules, err := s.repo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("find alert rules: %w", err)
	}
	return dto.ToAlertRuleRespList(rules), nil
}

func (s *alertRuleService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateAlertRuleReq) (*dto.AlertRuleResp, error) {
	rule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find alert rule: %w", err)
	}
	if rule == nil {
		return nil, ErrAlertRuleNotFound
	}

	if req.MetricName != nil {
		rule.MetricName = *req.MetricName
	}
	if req.Operator != nil {
		if !isValidOperator(*req.Operator) {
			return nil, ErrInvalidOperator
		}
		rule.Operator = *req.Operator
	}
	if req.Threshold != nil {
		rule.Threshold = *req.Threshold
	}
	if req.Severity != nil {
		if !isValidSeverity(*req.Severity) {
			return nil, ErrInvalidSeverity
		}
		rule.Severity = *req.Severity
	}
	if req.IsEnabled != nil {
		rule.IsEnabled = *req.IsEnabled
	}

	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("update alert rule: %w", err)
	}

	return dto.ToAlertRuleResp(rule), nil
}

func (s *alertRuleService) Delete(ctx context.Context, id uuid.UUID) error {
	rule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if rule == nil {
		return ErrAlertRuleNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete alert rule: %w", err)
	}
	return nil
}

func isValidOperator(op string) bool {
	switch op {
	case "gt", "lt", "gte", "lte", "eq":
		return true
	}
	return false
}

func isValidSeverity(sev string) bool {
	switch sev {
	case "info", "warning", "critical":
		return true
	}
	return false
}
