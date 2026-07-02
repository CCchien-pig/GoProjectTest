package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-name/udm/internal/model"
)

// CreateDeviceReq 建立設備請求
type CreateDeviceReq struct {
	DeviceCode string                 `json:"device_code" binding:"required,min=3,max=50"`
	Name       string                 `json:"name" binding:"required,min=2,max=200"`
	DeviceType string                 `json:"device_type" binding:"required,oneof=sensor controller gateway"`
	Location   string                 `json:"location" binding:"omitempty,max=200"`
	Metadata   map[string]interface{} `json:"metadata" binding:"omitempty"`
	OwnerID    *uuid.UUID             `json:"owner_id" binding:"omitempty"`
	Status     string                 `json:"status" binding:"omitempty,oneof=active inactive maintenance"`
}

// UpdateDeviceReq 更新設備請求
type UpdateDeviceReq struct {
	Name       *string                 `json:"name" binding:"omitempty,min=2,max=200"`
	DeviceType *string                 `json:"device_type" binding:"omitempty,oneof=sensor controller gateway"`
	Location   *string                 `json:"location" binding:"omitempty,max=200"`
	Metadata   map[string]interface{}  `json:"metadata" binding:"omitempty"`
	OwnerID    *uuid.UUID              `json:"owner_id" binding:"omitempty"`
	Status     *string                 `json:"status" binding:"omitempty,oneof=active inactive maintenance"`
}

// DeviceResp 設備回應 DTO
type DeviceResp struct {
	ID         uuid.UUID              `json:"id"`
	DeviceCode string                 `json:"device_code"`
	Name       string                 `json:"name"`
	DeviceType string                 `json:"device_type"`
	Location   string                 `json:"location"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	OwnerID    *uuid.UUID             `json:"owner_id,omitempty"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Owner      *UserResp              `json:"owner,omitempty"`

	// ScyllaDB 遙測欄位 (Week 2 擴充用)
	LatestTelemetry interface{} `json:"latest_telemetry,omitempty"`
}

// ToDeviceResp 將 model.Device 轉為 dto.DeviceResp
func ToDeviceResp(device *model.Device) *DeviceResp {
	if device == nil {
		return nil
	}
	var ownerResp *UserResp
	if device.Owner != nil {
		ownerResp = ToUserResp(device.Owner)
	}

	return &DeviceResp{
		ID:         device.ID,
		DeviceCode: device.DeviceCode,
		Name:       device.Name,
		DeviceType: device.DeviceType,
		Location:   device.Location,
		Metadata:   device.Metadata,
		OwnerID:    device.OwnerID,
		Status:     device.Status,
		CreatedAt:  device.CreatedAt,
		UpdatedAt:  device.UpdatedAt,
		Owner:      ownerResp,
	}
}

// ToDeviceRespList 批量轉換
func ToDeviceRespList(devices []*model.Device) []*DeviceResp {
	list := make([]*DeviceResp, len(devices))
	for i, d := range devices {
		list[i] = ToDeviceResp(d)
	}
	return list
}
