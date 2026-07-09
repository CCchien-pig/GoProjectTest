package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
)

type mockCacheService struct {
	mu               sync.Mutex
	store            map[string][]byte
	deviceNulls      map[string]bool
	latestTelStore   map[string][]byte
	listStore        map[string][]byte
	onlineStore      map[string]bool
	alertCounts      map[string]map[string]int64
	globalAlertCounts map[string]int64

	GetDeviceCalls int
	SetDeviceCalls int
	DelDeviceCalls int
}

func newMockCacheService() *mockCacheService {
	return &mockCacheService{
		store:             make(map[string][]byte),
		deviceNulls:       make(map[string]bool),
		latestTelStore:    make(map[string][]byte),
		listStore:         make(map[string][]byte),
		onlineStore:       make(map[string]bool),
		alertCounts:       make(map[string]map[string]int64),
		globalAlertCounts: make(map[string]int64),
	}
}

func (m *mockCacheService) GetDevice(ctx context.Context, id string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetDeviceCalls++
	if m.deviceNulls[id] {
		return []byte("null"), nil
	}
	return m.store[id], nil
}

func (m *mockCacheService) SetDevice(ctx context.Context, id string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SetDeviceCalls++
	m.store[id] = data
	return nil
}

func (m *mockCacheService) SetDeviceNull(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deviceNulls[id] = true
	return nil
}

func (m *mockCacheService) InvalidateDevice(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DelDeviceCalls++
	delete(m.store, id)
	delete(m.deviceNulls, id)
	return nil
}

func (m *mockCacheService) SetOnlineStatus(ctx context.Context, deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onlineStore[deviceID] = true
	return nil
}

func (m *mockCacheService) IsOnline(ctx context.Context, deviceID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.onlineStore[deviceID], nil
}

func (m *mockCacheService) CountOnline(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.onlineStore)), nil
}

func (m *mockCacheService) SetLatestTelemetry(ctx context.Context, deviceID string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latestTelStore[deviceID] = data
	return nil
}

func (m *mockCacheService) GetLatestTelemetry(ctx context.Context, deviceID string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.latestTelStore[deviceID], nil
}

func (m *mockCacheService) IncrAlertCount(ctx context.Context, deviceID string, severity string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.alertCounts[deviceID] == nil {
		m.alertCounts[deviceID] = make(map[string]int64)
	}
	m.alertCounts[deviceID][severity]++
	// 同步遞增全域計數（模擬 production 行為）
	m.globalAlertCounts[severity]++
	return nil
}

func (m *mockCacheService) GetAlertCounts(ctx context.Context, deviceID string) (map[string]int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.alertCounts[deviceID], nil
}

func (m *mockCacheService) GetGlobalAlertCounts(ctx context.Context) (map[string]int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 回傳副本避免 data race
	result := make(map[string]int64)
	for k, v := range m.globalAlertCounts {
		result[k] = v
	}
	return result, nil
}

func (m *mockCacheService) SyncAlertCount(ctx context.Context, deviceID string, severity string, count int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.alertCounts[deviceID] == nil {
		m.alertCounts[deviceID] = make(map[string]int64)
	}
	m.alertCounts[deviceID][severity] = count
	return nil
}

func (m *mockCacheService) GetDeviceList(ctx context.Context, queryHash string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.listStore[queryHash], nil
}

func (m *mockCacheService) SetDeviceList(ctx context.Context, queryHash string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listStore[queryHash] = data
	return nil
}

func (m *mockCacheService) InvalidateAllLists(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listStore = make(map[string][]byte)
	return nil
}

func (m *mockCacheService) GetDashboard(ctx context.Context) ([]byte, error) {
	return nil, nil
}

func (m *mockCacheService) SetDashboard(ctx context.Context, data []byte) error {
	return nil
}

func (m *mockCacheService) SetDeviceTotalCount(ctx context.Context, count int64) error {
	return nil
}

func (m *mockCacheService) GetDeviceTotalCount(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockCacheService) GetDashboardMetricsPipeline(ctx context.Context) (int64, int64, error) {
	return 0, 0, nil
}

func (m *mockCacheService) InvalidateByPattern(ctx context.Context, pattern string) (int64, error) {
	return 0, nil
}

func (m *mockCacheService) InvalidateDeviceAll(ctx context.Context, deviceID string) error {
	_ = m.InvalidateDevice(ctx, deviceID)
	_ = m.InvalidateAllLists(ctx)
	return nil
}

func (m *mockCacheService) Ping(ctx context.Context) error {
	return nil
}

func TestDeviceService_CacheAside(t *testing.T) {
	repo := newMockDeviceRepository()
	userRepo := newMockUserRepository()
	telRepo := &mockTelemetryRepository{}
	cacheSvc := newMockCacheService()
	svc := NewDeviceService(repo, userRepo, telRepo, cacheSvc)

	devID := uuid.New()
	dev := &model.Device{
		ID:         devID,
		DeviceCode: "DEV-CACHE-1",
		Name:       "Cache Test",
		DeviceType: "sensor",
		Status:     "active",
	}
	repo.devices[devID] = dev

	// 1. First query: Cache Miss
	resp, err := svc.FindByID(context.Background(), devID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if resp.Name != "Cache Test" {
		t.Errorf("Expected Name Cache Test, got %s", resp.Name)
	}
	if cacheSvc.GetDeviceCalls != 1 {
		t.Errorf("Expected 1 GetDevice call, got %d", cacheSvc.GetDeviceCalls)
	}
	if cacheSvc.SetDeviceCalls != 1 {
		t.Errorf("Expected 1 SetDevice call, got %d", cacheSvc.SetDeviceCalls)
	}

	// 2. Second query: Cache Hit
	delete(repo.devices, devID) // Delete from mock DB to ensure it comes from cache
	resp, err = svc.FindByID(context.Background(), devID)
	if err != nil {
		t.Fatalf("FindByID from cache failed: %v", err)
	}
	if resp.Name != "Cache Test" {
		t.Errorf("Expected Name Cache Test, got %s", resp.Name)
	}
	if cacheSvc.GetDeviceCalls != 2 {
		t.Errorf("Expected 2 GetDevice calls, got %d", cacheSvc.GetDeviceCalls)
	}
	if cacheSvc.SetDeviceCalls != 1 {
		t.Errorf("Expected no additional SetDevice calls, got %d", cacheSvc.SetDeviceCalls)
	}
}

func TestDeviceService_CachePenetrationNull(t *testing.T) {
	repo := newMockDeviceRepository()
	userRepo := newMockUserRepository()
	telRepo := &mockTelemetryRepository{}
	cacheSvc := newMockCacheService()
	svc := NewDeviceService(repo, userRepo, telRepo, cacheSvc)

	fakeID := uuid.New()

	// Query non-existing device
	_, err := svc.FindByID(context.Background(), fakeID)
	if !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("Expected ErrDeviceNotFound, got %v", err)
	}

	// Verify "null" was cached to prevent penetration
	if !cacheSvc.deviceNulls[fakeID.String()] {
		t.Errorf("Expected null to be cached for device %s", fakeID)
	}

	// Query again: Should hit "null" cache immediately
	_, err = svc.FindByID(context.Background(), fakeID)
	if !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("Expected ErrDeviceNotFound on second query, got %v", err)
	}
}

func TestDeviceService_Invalidation(t *testing.T) {
	repo := newMockDeviceRepository()
	userRepo := newMockUserRepository()
	telRepo := &mockTelemetryRepository{}
	cacheSvc := newMockCacheService()
	svc := NewDeviceService(repo, userRepo, telRepo, cacheSvc)

	devID := uuid.New()
	dev := &model.Device{
		ID:         devID,
		DeviceCode: "DEV-123",
		Name:       "Old Name",
		DeviceType: "sensor",
		Status:     "active",
	}
	repo.devices[devID] = dev

	// Populate cache
	_, _ = svc.FindByID(context.Background(), devID)

	// Update device
	newName := "New Name"
	req := &dto.UpdateDeviceReq{
		Name: &newName,
	}
	_, err := svc.Update(context.Background(), devID, req)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify cache was invalidated (deleted)
	if _, exists := cacheSvc.store[devID.String()]; exists {
		t.Error("Expected cache key to be invalidated after update")
	}
}
