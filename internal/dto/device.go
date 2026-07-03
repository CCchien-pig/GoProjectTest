package dto

import (
	"time"

	"github.com/google/uuid"
	"GoProject/udm/internal/model"
)

// CreateDeviceReq 建立設備請求
type CreateDeviceReq struct {
	DeviceCode string                 `json:"device_code" binding:"required,min=3,max=50"`
	Name       string                 `json:"name" binding:"required,min=2,max=200"`
	DeviceType string                 `json:"device_type" binding:"required,oneof=sensor controller gateway"`
	Location   string                 `json:"location" binding:"omitempty,max=200"`
	Metadata   map[string]interface{} `json:"metadata" binding:"omitempty"`
	UserIDs    []uuid.UUID            `json:"user_ids" binding:"omitempty"`
	Status     string                 `json:"status" binding:"omitempty,oneof=active inactive maintenance"`
}

// UpdateDeviceReq 更新設備請求
type UpdateDeviceReq struct {
	Name       *string                `json:"name" binding:"omitempty,min=2,max=200"`
	DeviceType *string                `json:"device_type" binding:"omitempty,oneof=sensor controller gateway"`
	Location   *string                `json:"location" binding:"omitempty,max=200"`
	Metadata   map[string]interface{} `json:"metadata" binding:"omitempty"`
	UserIDs    []uuid.UUID            `json:"user_ids" binding:"omitempty"`
	Status     *string                `json:"status" binding:"omitempty,oneof=active inactive maintenance"`
}

// DeviceResp 設備回傳 DTO
type DeviceResp struct {
	ID         uuid.UUID              `json:"id"`
	DeviceCode string                 `json:"device_code"`
	Name       string                 `json:"name"`
	DeviceType string                 `json:"device_type"`
	Location   string                 `json:"location"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Users      []*UserResp            `json:"users,omitempty"`

	// ScyllaDB 遙測欄位（Week 2 新增）
	LatestTelemetry []*model.TelemetryData `json:"latest_telemetry,omitempty"`
}

// ToDeviceResp 將 model.Device 轉為 dto.DeviceResp
func ToDeviceResp(device *model.Device) *DeviceResp {
	if device == nil {
		return nil
	}
	
	var userResps []*UserResp
	if len(device.Users) > 0 {
		userResps = make([]*UserResp, len(device.Users))
		for i, u := range device.Users {
			// Create a copy to avoid loop variable issues
			userCopy := u
			userResps[i] = ToUserResp(&userCopy)
		}
	}

	return &DeviceResp{
		ID:         device.ID,
		DeviceCode: device.DeviceCode,
		Name:       device.Name,
		DeviceType: device.DeviceType,
		Location:   device.Location,
		Metadata:   device.Metadata,
		Status:     device.Status,
		CreatedAt:  device.CreatedAt,
		UpdatedAt:  device.UpdatedAt,
		Users:      userResps,
	}
}

// ToDeviceRespList 批次轉換
func ToDeviceRespList(devices []*model.Device) []*DeviceResp {
	list := make([]*DeviceResp, len(devices))
	for i, d := range devices {
		list[i] = ToDeviceResp(d)
	}
	return list
}
