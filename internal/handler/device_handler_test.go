package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/service"
)

type mockDeviceService struct {
	onFindByID func(id uuid.UUID) (*dto.DeviceResp, error)
	onUpdate   func(id uuid.UUID, req *dto.UpdateDeviceReq) (*dto.DeviceResp, error)
	onDelete   func(id uuid.UUID) error
	onList     func(cursor string, limit int, dt, st, loc, q string) ([]*dto.DeviceResp, string, error)
}

func (m *mockDeviceService) Create(ctx context.Context, req *dto.CreateDeviceReq) (*dto.DeviceResp, error) {
	return &dto.DeviceResp{
		ID:         uuid.New(),
		DeviceCode: req.DeviceCode,
		Name:       req.Name,
		DeviceType: req.DeviceType,
		Location:   req.Location,
		Metadata:   req.Metadata,

		Status:     req.Status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

func (m *mockDeviceService) FindByID(ctx context.Context, id uuid.UUID) (*dto.DeviceResp, error) {
	if m.onFindByID != nil {
		return m.onFindByID(id)
	}
	return &dto.DeviceResp{
		ID:         id,
		DeviceCode: "DEV-001",
		Name:       "Mock Device",
		DeviceType: "sensor",
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

func (m *mockDeviceService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDeviceReq) (*dto.DeviceResp, error) {
	if m.onUpdate != nil {
		return m.onUpdate(id, req)
	}
	name := "Updated Mock"
	if req.Name != nil {
		name = *req.Name
	}
	return &dto.DeviceResp{
		ID:         id,
		DeviceCode: "DEV-001",
		Name:       name,
		DeviceType: "sensor",
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

func (m *mockDeviceService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.onDelete != nil {
		return m.onDelete(id)
	}
	return nil
}

func (m *mockDeviceService) List(ctx context.Context, cursor string, limit int, dt, st, loc, q string) ([]*dto.DeviceResp, string, error) {
	if m.onList != nil {
		return m.onList(cursor, limit, dt, st, loc, q)
	}
	return []*dto.DeviceResp{
		{ID: uuid.New(), DeviceCode: "DEV-001", Name: "Mock 1"},
		{ID: uuid.New(), DeviceCode: "DEV-002", Name: "Mock 2"},
	}, "", nil
}

func TestDeviceHandler_Create(t *testing.T) {
	svc := &mockDeviceService{}
	h := NewDeviceHandler(svc)

	r := gin.New()
	r.POST("/devices", h.Create)

	reqPayload := dto.CreateDeviceReq{
		DeviceCode: "DEV-100",
		Name:       "Ingested Sensor",
		DeviceType: "sensor",
	}

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}
}

func TestDeviceHandler_FindByID(t *testing.T) {
	uid := uuid.New()
	svc := &mockDeviceService{
		onFindByID: func(id uuid.UUID) (*dto.DeviceResp, error) {
			if id != uid {
				return nil, service.ErrDeviceNotFound
			}
			return &dto.DeviceResp{ID: id, DeviceCode: "FOUND-01"}, nil
		},
	}
	h := NewDeviceHandler(svc)

	r := gin.New()
	r.GET("/devices/:id", h.FindByID)

	req, _ := http.NewRequest("GET", "/devices/"+uid.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	req2, _ := http.NewRequest("GET", "/devices/"+uuid.New().String(), nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Errorf("expected status NotFound, got %d", w2.Code)
	}
}
