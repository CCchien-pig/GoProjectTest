package dto

import (
	"time"

	"github.com/google/uuid"
	"GoProject/udm/internal/model"
)

// CreateUserReq 建立使用者請求
type CreateUserReq struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
}

// UpdateUserReq 更新使用者請求
type UpdateUserReq struct {
	Username *string `json:"username" binding:"omitempty,min=3,max=100"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Role     *string `json:"role" binding:"omitempty,oneof=admin operator viewer"`
	IsActive *bool   `json:"is_active"`
}

// UserResp 使用者回傳 DTO
type UserResp struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeviceCount int64     `json:"device_count,omitempty"`
}

// ToUserResp 將 model.User 轉為 dto.UserResp
func ToUserResp(user *model.User) *UserResp {
	if user == nil {
		return nil
	}
	return &UserResp{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		DeviceCount: user.DeviceCount,
	}
}
