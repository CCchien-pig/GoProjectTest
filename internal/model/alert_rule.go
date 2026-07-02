package model

import (
	"time"

	"github.com/google/uuid"
)

// AlertRule �?�� alert_rules 資�?表�? GORM 模�?
type AlertRule struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DeviceID   uuid.UUID `gorm:"type:uuid;column:device_id;not null" json:"device_id"`
	MetricName string    `gorm:"type:varchar(100);column:metric_name;not null" json:"metric_name"` // e.g. temperature, voltage
	Operator   string    `gorm:"type:varchar(10);not null" json:"operator"`                        // gt, lt, gte, lte, eq
	Threshold  float64   `gorm:"type:double precision;not null" json:"threshold"`
	Severity   string    `gorm:"type:varchar(20);not null;default:'warning'" json:"severity"` // info / warning / critical
	IsEnabled  bool      `gorm:"type:boolean;column:is_enabled;not null;default:true" json:"is_enabled"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`

	// 關聯
	Device *Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
}

// TableName 指定資料表名稱
func (AlertRule) TableName() string {
	return "alert_rules"
}
