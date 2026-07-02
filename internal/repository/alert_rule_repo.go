package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/your-name/udm/internal/model"
	"gorm.io/gorm"
)

// AlertRuleRepository 定義對 alert_rules 資料表的資料存取介面
type AlertRuleRepository interface {
	Create(ctx context.Context, rule *model.AlertRule) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.AlertRule, error)
	FindByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*model.AlertRule, error)
	Update(ctx context.Context, rule *model.AlertRule) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type gormAlertRuleRepository struct {
	db *gorm.DB
}

// NewAlertRuleRepository 建立 GORM 的 AlertRuleRepository 實作
func NewAlertRuleRepository(db *gorm.DB) AlertRuleRepository {
	return &gormAlertRuleRepository{db: db}
}

func (r *gormAlertRuleRepository) Create(ctx context.Context, rule *model.AlertRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *gormAlertRuleRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AlertRule, error) {
	var rule model.AlertRule
	if err := r.db.WithContext(ctx).First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rule, nil
}

func (r *gormAlertRuleRepository) FindByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*model.AlertRule, error) {
	var rules []*model.AlertRule
	if err := r.db.WithContext(ctx).Where("device_id = ?", deviceID).Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *gormAlertRuleRepository) Update(ctx context.Context, rule *model.AlertRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *gormAlertRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.AlertRule{}, "id = ?", id).Error
}
