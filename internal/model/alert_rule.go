package model

// internal/model/alert_rule.go — AlertRule GORM Model
// 對應 PostgreSQL alert_rules 表

// TODO: Day 5 實作
// type AlertRule struct {
//     ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
//     DeviceID   uuid.UUID `gorm:"type:uuid;not null;index"`
//     Device     Device    `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE"`
//     MetricName string    `gorm:"not null"`
//     Operator   string    `gorm:"not null"` // gt, lt, gte, lte, eq
//     Threshold  float64   `gorm:"not null"`
//     Severity   string    `gorm:"not null"` // info, warning, critical
//     IsEnabled  bool      `gorm:"not null;default:true"`
//     CreatedAt  time.Time
//     UpdatedAt  time.Time
// }
