package model

// internal/model/user.go — User GORM Model
// 對應 PostgreSQL users 表

// TODO: Day 3 實作
// type User struct {
//     ID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
//     Username     string     `gorm:"not null;uniqueIndex"`
//     Email        string     `gorm:"not null;uniqueIndex"`
//     PasswordHash string     `gorm:"not null"`
//     Role         string     `gorm:"not null;default:'user'"`
//     IsActive     bool       `gorm:"not null;default:true"`
//     CreatedAt    time.Time
//     UpdatedAt    time.Time
// }
