package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
)

type mockDeviceRepository struct {
	devices       map[uuid.UUID]*model.Device
	devicesByCode map[string]*model.Device
	onCreate      func(device *model.Device) error
	onFindByID    func(id uuid.UUID) (*model.Device, error)
	onUpdate      func(device *model.Device) error
	onDelete      func(id uuid.UUID) error
	onList        func(cursor string, limit int, dt, st, loc, q string) ([]*model.Device, string, error)
}

func newMockDeviceRepository() *mockDeviceRepository {
	return &mockDeviceRepository{
		devices:       make(map[uuid.UUID]*model.Device),
		devicesByCode: make(map[string]*model.Device),
	}
}

func (m *mockDeviceRepository) Create(ctx context.Context, device *model.Device) error {
	if m.onCreate != nil {
		return m.onCreate(device)
	}
	device.ID = uuid.New()
	m.devices[device.ID] = device
	m.devicesByCode[device.DeviceCode] = device
	return nil
}

func (m *mockDeviceRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Device, error) {
	if m.onFindByID != nil {
		return m.onFindByID(id)
	}
	return m.devices[id], nil
}

func (m *mockDeviceRepository) FindByDeviceCode(ctx context.Context, code string) (*model.Device, error) {
	return m.devicesByCode[code], nil
}

func (m *mockDeviceRepository) Update(ctx context.Context, device *model.Device) error {
	if m.onUpdate != nil {
		return m.onUpdate(device)
	}
	m.devices[device.ID] = device
	m.devicesByCode[device.DeviceCode] = device
	return nil
}

func (m *mockDeviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.onDelete != nil {
		return m.onDelete(id)
	}
	delete(m.devices, id)
	return nil
}

func (m *mockDeviceRepository) UpdateWithUsers(ctx context.Context, device *model.Device, users []model.User) error {
	m.devices[device.ID] = device
	m.devicesByCode[device.DeviceCode] = device
	return nil
}

func (m *mockDeviceRepository) List(ctx context.Context, cursor string, limit int, dt, st, loc, q string) ([]*model.Device, string, error) {
	if m.onList != nil {
		return m.onList(cursor, limit, dt, st, loc, q)
	}
	var list []*model.Device
	for _, d := range m.devices {
		list = append(list, d)
	}
	return list, "", nil
}

func (m *mockDeviceRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.devices)), nil
}

func TestDeviceService_Create(t *testing.T) {
	userRepo := newMockUserRepository()
	repo := newMockDeviceRepository()
	svc := NewDeviceService(repo, userRepo, &mockTelemetryRepository{}, nil)

	u := &model.User{ID: uuid.New(), Username: "owner"}
	userRepo.users[u.ID] = u

	req := &dto.CreateDeviceReq{
		DeviceCode: "DEV-001",
		Name:       "Test Device",
		DeviceType: "sensor",
		Location:   "TPE",
		UserIDs:    []uuid.UUID{u.ID},
		Status:     "active",
	}

	resp, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to create device: %v", err)
	}

	if resp.DeviceCode != "DEV-001" || resp.Name != "Test Device" || resp.Status != "active" {
		t.Errorf("unexpected device response: %+v", resp)
	}

	// ??編?
	_, err = svc.Create(context.Background(), req)
	if err == nil || err != ErrDeviceCodeDuplicate {
		t.Errorf("expected ErrDeviceCodeDuplicate, got %v", err)
	}

	// ?? Owner
	fakeID := uuid.New()
	req.DeviceCode = "DEV-002"
	req.UserIDs = []uuid.UUID{fakeID}
	_, err = svc.Create(context.Background(), req)
	if err == nil || err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
