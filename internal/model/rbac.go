package model

import (
	"time"

	"github.com/google/uuid"
)

// Role 代表角色表
type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string       `gorm:"type:varchar(50);unique;not null" json:"name"`
	Description string       `gorm:"type:varchar(200)" json:"description"`
	CreatedAt   time.Time    `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`

	// 關聯
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// Permission 代表權限表
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(50);unique;not null" json:"name"`
	Description string    `gorm:"type:varchar(200)" json:"description"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
}

// RolePermission 代表中間表 (如果有需要獨立設定 GORM 也可以，但通常 many2many 就夠了)
// 但若需要操作獨立表，可以定義 struct
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

// TableName 指定資料表名稱
func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
