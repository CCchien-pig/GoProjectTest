package dto

import (
	"time"

	"github.com/google/uuid"
	"GoProject/udm/internal/model"
)

// PermissionResp 權限回傳 DTO
type PermissionResp struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// RoleResp 角色回傳 DTO
type RoleResp struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Permissions []PermissionResp `json:"permissions,omitempty"`
}

// CreateUserReq 建立使用者請求
type CreateUserReq struct {
	Username string    `json:"username" binding:"required,min=3,max=100"`
	Email    string    `json:"email" binding:"required,email"`
	Password string    `json:"password" binding:"required,min=6"`
	RoleID   uuid.UUID `json:"role_id" binding:"required"`
}

// UpdateUserReq 更新使用者請求
type UpdateUserReq struct {
	Username *string    `json:"username" binding:"omitempty,min=3,max=100"`
	Email    *string    `json:"email" binding:"omitempty,email"`
	RoleID   *uuid.UUID `json:"role_id" binding:"omitempty"`
	IsActive *bool      `json:"is_active"`
}

// UserResp 使用者回傳 DTO
type UserResp struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        *RoleResp `json:"role,omitempty"`
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
	
	var roleResp *RoleResp
	if user.Role != nil {
		perms := make([]PermissionResp, len(user.Role.Permissions))
		for i, p := range user.Role.Permissions {
			perms[i] = PermissionResp{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
			}
		}
		roleResp = &RoleResp{
			ID:          user.Role.ID,
			Name:        user.Role.Name,
			Description: user.Role.Description,
			Permissions: perms,
		}
	}

	return &UserResp{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        roleResp,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		DeviceCount: user.DeviceCount,
	}
}

// ToUserRespList 批次轉換
func ToUserRespList(users []*model.User) []*UserResp {
	list := make([]*UserResp, len(users))
	for i, u := range users {
		list[i] = ToUserResp(u)
	}
	return list
}
