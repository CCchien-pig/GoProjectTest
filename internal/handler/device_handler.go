package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-name/udm/internal/dto"
	"github.com/your-name/udm/internal/service"
	"github.com/your-name/udm/pkg/response"
)

// DeviceHandler 處理設備相關 HTTP 請求
type DeviceHandler struct {
	svc service.DeviceService
}

// NewDeviceHandler 建立 DeviceHandler 實作
func NewDeviceHandler(svc service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// Create 建立設備
func (h *DeviceHandler) Create(c *gin.Context) {
	var req dto.CreateDeviceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrDeviceCodeDuplicate) || errors.Is(err, service.ErrUserNotFound) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, resp)
}

// FindByID 取得設備詳細資訊
func (h *DeviceHandler) FindByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid uuid format")
		return
	}

	resp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, resp)
}

// Update 更新設備資訊
func (h *DeviceHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid uuid format")
		return
	}

	var req dto.UpdateDeviceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) || errors.Is(err, service.ErrUserNotFound) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, resp)
}

// Delete 刪除設備
func (h *DeviceHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid uuid format")
		return
	}

	err = h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "device deleted successfully")
}

// List 取得設備列表（支援 Cursor-based 分頁與模糊搜尋）
func (h *DeviceHandler) List(c *gin.Context) {
	cursor := c.Query("cursor")
	limitStr := c.Query("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	deviceType := c.Query("device_type")
	status := c.Query("status")
	location := c.Query("location")
	search := c.Query("search")

	devices, nextCursor, err := h.svc.List(c.Request.Context(), cursor, limit, deviceType, status, location, search)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	hasMore := nextCursor != ""

	response.OKWithPagination(c, devices, &response.Pagination{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}
