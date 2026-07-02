package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Device 代表 devices 資料表的 GORM 模型
type Device struct {
	ID         uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DeviceCode string                 `gorm:"type:varchar(50);unique;not null;column:device_code" json:"device_code"`
	Name       string                 `gorm:"type:varchar(200);not null" json:"name"`
	DeviceType string                 `gorm:"type:varchar(50);not null;column:device_type" json:"device_type"` // sensor / controller / gateway
	Location   string                 `gorm:"type:varchar(200)" json:"location"`
	Metadata   map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"metadata"` // 彈性欄位
	OwnerID    *uuid.UUID             `gorm:"type:uuid;column:owner_id" json:"owner_id,omitempty"`
	Status     string                 `gorm:"type:varchar(20);not null;default:'inactive'" json:"status"` // active / inactive / maintenance
	CreatedAt  time.Time              `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time              `gorm:"type:timestamptz;not null;default:now()" json:"updated_at"`

	// 關聯
	Owner *User `gorm:"foreignKey:OwnerID;constraint:OnDelete:SET NULL;" json:"owner,omitempty"`
}

// TableName 指定資料表名稱
func (Device) TableName() string {
	return "devices"
}

// BeforeUpdate GORM 更新 hook
func (d *Device) BeforeUpdate(tx *gorm.DB) (err error) {
	d.UpdatedAt = time.Now()
	return nil
}
