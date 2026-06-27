package model

// internal/model/device.go — Device GORM Model
// 對應 PostgreSQL devices 表

// TODO: Day 4 實作
// type Device struct {
//     ID         uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
//     DeviceCode string          `gorm:"not null;uniqueIndex"`
//     Name       string          `gorm:"not null"`
//     DeviceType string          `gorm:"not null;index"`
//     Location   string          `gorm:"index"`
//     Metadata   datatypes.JSON  `gorm:"type:jsonb"`
//     OwnerID    uuid.UUID       `gorm:"type:uuid;not null;index"`
//     Owner      User            `gorm:"foreignKey:OwnerID"`
//     Status     string          `gorm:"not null;default:'active';index"`
//     CreatedAt  time.Time
//     UpdatedAt  time.Time
// }
