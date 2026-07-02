package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 對應 users 資料表的 GORM 模型
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"type:varchar(100);unique;not null" json:"username"`
	Email        string    `gorm:"type:varchar(255);unique;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null;column:password_hash" json:"-"`
	Role         string    `gorm:"type:varchar(20);not null;default:'viewer'" json:"role"` // admin / operator / viewer
	IsActive     bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;not null;default:now()" json:"updated_at"`

	// 額外欄位：只用於回傳，不對應 DB
	DeviceCount int64 `gorm:"-" json:"device_count,omitempty"`
}

// TableName 指定資料表名稱
func (User) TableName() string {
	return "users"
}

// BeforeUpdate GORM 更新 hook
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return nil
}
