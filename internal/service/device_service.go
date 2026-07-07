package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"GoProject/udm/internal/cache"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
	"GoProject/udm/internal/repository"
	"GoProject/udm/internal/scylla"
)

var (
	// ErrDeviceNotFound 找不到設備錯誤
	ErrDeviceNotFound = errors.New("device not found")
	// ErrDeviceCodeDuplicate 設備編號重複錯誤
	ErrDeviceCodeDuplicate = errors.New("device code already exists")
	// ErrCacheCleanupFailed 設備已刪除但快取清除失敗
	ErrCacheCleanupFailed = errors.New("device deleted from database, but cache cleanup failed")
)

// listCacheEntry 設備列表快取格式（Finding #5: 提取至 package level 避免重複定義）
type listCacheEntry struct {
	Devices    []*dto.DeviceResp `json:"devices"`
	NextCursor string            `json:"next_cursor"`
}


// DeviceService 定義設備業務邏輯介面
type DeviceService interface {
	Create(ctx context.Context, req *dto.CreateDeviceReq) (*dto.DeviceResp, error)
	FindByID(ctx context.Context, id uuid.UUID) (*dto.DeviceResp, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDeviceReq) (*dto.DeviceResp, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, cursor string, limit int, deviceType, status, location, search string) ([]*dto.DeviceResp, string, error)
}

type deviceService struct {
	repo          repository.DeviceRepository
	userRepo      repository.UserRepository
	telemetryRepo scylla.TelemetryRepository
	cache         cache.Service
	sf            singleflight.Group
}

// NewDeviceService 建立 DeviceService 實體
func NewDeviceService(
	repo repository.DeviceRepository,
	userRepo repository.UserRepository,
	telemetryRepo scylla.TelemetryRepository,
	cacheService cache.Service,
) DeviceService {
	return &deviceService{
		repo:          repo,
		userRepo:      userRepo,
		telemetryRepo: telemetryRepo,
		cache:         cacheService,
	}
}

func (s *deviceService) Create(ctx context.Context, req *dto.CreateDeviceReq) (*dto.DeviceResp, error) {
	// 檢查 DeviceCode
	existing, err := s.repo.FindByDeviceCode(ctx, req.DeviceCode)
	if err != nil {
		return nil, fmt.Errorf("find by device code: %w", err)
	}
	if existing != nil {
		return nil, ErrDeviceCodeDuplicate
	}

	// 檢查 Users
	var usersToBind []model.User
	if len(req.UserIDs) > 0 {
		users, err := s.userRepo.FindByIDs(ctx, req.UserIDs)
		if err != nil {
			return nil, fmt.Errorf("find users: %w", err)
		}
		if len(users) != len(req.UserIDs) {
			return nil, ErrUserNotFound
		}
		for _, u := range users {
			usersToBind = append(usersToBind, *u)
		}
	}

	status := req.Status
	if status == "" {
		status = "inactive"
	}

	device := &model.Device{
		DeviceCode: req.DeviceCode,
		Name:       req.Name,
		DeviceType: req.DeviceType,
		Location:   req.Location,
		Metadata:   req.Metadata,
		Status:     status,
		Users:      usersToBind,
	}

	if err := s.repo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("create device: %w", err)
	}

	// 若有設定 Users，重新 Preload 取得完整的 Users 資料
	if len(req.UserIDs) > 0 {
		if reloaded, reloadErr := s.repo.FindByID(ctx, device.ID); reloadErr == nil && reloaded != nil {
			device = reloaded
		}
	}

	// Invalidate list cache (Finding #13: 記錄快取清除失敗 log)
	if s.cache != nil {
		if err := s.cache.InvalidateAllLists(ctx); err != nil {
			slog.WarnContext(ctx, "failed to invalidate list cache after create", "error", err)
		}
	}

	return dto.ToDeviceResp(device), nil
}

func (s *deviceService) FindByID(ctx context.Context, id uuid.UUID) (*dto.DeviceResp, error) {
	idStr := id.String()

	// 1. Try to read from Cache-Aside
	if s.cache != nil {
		cachedBytes, err := s.cache.GetDevice(ctx, idStr)
		if err == nil && cachedBytes != nil {
			if string(cachedBytes) == "null" {
				return nil, ErrDeviceNotFound
			}
			var device model.Device
			if jsonErr := json.Unmarshal(cachedBytes, &device); jsonErr == nil {
				resp := dto.ToDeviceResp(&device)
				s.attachLatestTelemetry(ctx, id, resp)
				return resp, nil
			}
		}
	}

	// 2. Cache miss — Retrieve from DB using singleflight to prevent Cache Stampede
	val, err, _ := s.sf.Do(idStr, func() (interface{}, error) {
		device, dbErr := s.repo.FindByID(ctx, id)
		if dbErr != nil {
			return nil, dbErr
		}
		if device == nil {
			// Write "null" to prevent Cache Penetration
			if s.cache != nil {
				_ = s.cache.SetDeviceNull(ctx, idStr)
			}
			return nil, ErrDeviceNotFound
		}

		// Write to Cache
		if s.cache != nil {
			if deviceBytes, marshalErr := json.Marshal(device); marshalErr == nil {
				_ = s.cache.SetDevice(ctx, idStr, deviceBytes)
			}
		}
		return device, nil
	})

	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("find device singleflight: %w", err)
	}

	device, ok := val.(*model.Device)
	if !ok || device == nil {
		return nil, ErrDeviceNotFound
	}

	resp := dto.ToDeviceResp(device)
	s.attachLatestTelemetry(ctx, id, resp)
	return resp, nil
}

func (s *deviceService) attachLatestTelemetry(ctx context.Context, id uuid.UUID, resp *dto.DeviceResp) {
	idStr := id.String()
	// 先從 KeyDB 查詢最新遙測快取 (Write-Through 模式)
	var gotFromCache bool
	if s.cache != nil {
		cachedTel, err := s.cache.GetLatestTelemetry(ctx, idStr)
		if err == nil && cachedTel != nil {
			var telemetries []*model.TelemetryData
			if jsonErr := json.Unmarshal(cachedTel, &telemetries); jsonErr == nil {
				resp.LatestTelemetry = telemetries
				gotFromCache = true
			}
		}
	}

	// Cache miss or error — 從 ScyllaDB 查出該設備最新遙測資料一併回傳
	if !gotFromCache && s.telemetryRepo != nil {
		telemetries, telErr := s.telemetryRepo.QueryLatest(ctx, id)
		if telErr == nil && len(telemetries) > 0 {
			resp.LatestTelemetry = telemetries
			// 同步寫回 KeyDB 快取
			if s.cache != nil {
				if telBytes, marshalErr := json.Marshal(telemetries); marshalErr == nil {
					_ = s.cache.SetLatestTelemetry(ctx, idStr, telBytes)
				}
			}
		}
	}
}

func (s *deviceService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDeviceReq) (*dto.DeviceResp, error) {
	device, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find device: %w", err)
	}
	if device == nil {
		return nil, ErrDeviceNotFound
	}

	if req.Name != nil {
		device.Name = *req.Name
	}
	if req.DeviceType != nil {
		device.DeviceType = *req.DeviceType
	}
	if req.Location != nil {
		device.Location = *req.Location
	}
	if req.Metadata != nil {
		device.Metadata = req.Metadata
	}
	if req.Status != nil {
		device.Status = *req.Status
	}

	// 處理關聯的 Users 更新
	if req.UserIDs != nil {
		var usersToBind []model.User
		if len(req.UserIDs) > 0 {
			users, err := s.userRepo.FindByIDs(ctx, req.UserIDs)
			if err != nil {
				return nil, fmt.Errorf("find users: %w", err)
			}
			if len(users) != len(req.UserIDs) {
				return nil, ErrUserNotFound
			}
			for _, u := range users {
				usersToBind = append(usersToBind, *u)
			}
		}
		
		if err := s.repo.UpdateWithUsers(ctx, device, usersToBind); err != nil {
			return nil, fmt.Errorf("update device with users: %w", err)
		}
	} else {
		if err := s.repo.Update(ctx, device); err != nil {
			return nil, fmt.Errorf("update device: %w", err)
		}
	}

	// 重新 Preload 取得 Users 關聯資料
	if reloaded, reloadErr := s.repo.FindByID(ctx, device.ID); reloadErr == nil && reloaded != nil {
		device = reloaded
	}

	// Invalidate caches (Finding #13: 記錄快取清除失敗 log)
	if s.cache != nil {
		if err := s.cache.InvalidateDevice(ctx, id.String()); err != nil {
			slog.WarnContext(ctx, "failed to invalidate device cache after update", "device_id", id, "error", err)
		}
		if err := s.cache.InvalidateAllLists(ctx); err != nil {
			slog.WarnContext(ctx, "failed to invalidate list cache after update", "error", err)
		}
	}

	return dto.ToDeviceResp(device), nil
}

func (s *deviceService) Delete(ctx context.Context, id uuid.UUID) error {
	device, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find device for delete: %w", err)
	}
	if device == nil {
		return ErrDeviceNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete device: %w", err)
	}

	// Invalidate caches (Saga Step 2)
	if s.cache != nil {
		if err := s.cache.InvalidateDeviceAll(ctx, id.String()); err != nil {
			slog.ErrorContext(ctx, "Saga Warning: device deleted from PostgreSQL but KeyDB cache invalidation failed", "device_id", id, "error", err)
			return ErrCacheCleanupFailed
		}
	}

	return nil
}

func (s *deviceService) List(ctx context.Context, cursor string, limit int, deviceType, status, location, search string) ([]*dto.DeviceResp, string, error) {
	if limit <= 0 {
		limit = 10
	}

	// 1. Generate query hash
	queryHash := cache.HashQuery(cursor, fmt.Sprintf("%d", limit), deviceType, status, location, search)

	// 2. Try to read from list cache (Finding #5: 使用 package-level listCacheEntry)
	if s.cache != nil {
		cachedBytes, err := s.cache.GetDeviceList(ctx, queryHash)
		if err == nil && cachedBytes != nil {
			var entry listCacheEntry
			if jsonErr := json.Unmarshal(cachedBytes, &entry); jsonErr == nil {
				return entry.Devices, entry.NextCursor, nil
			}
		}
	}

	// 3. Cache miss — Retrieve from DB
	devices, nextCursor, err := s.repo.List(ctx, cursor, limit, deviceType, status, location, search)
	if err != nil {
		return nil, "", fmt.Errorf("list devices: %w", err)
	}

	respList := dto.ToDeviceRespList(devices)

	// 4. Write back to list cache
	if s.cache != nil {
		entry := listCacheEntry{
			Devices:    respList,
			NextCursor: nextCursor,
		}
		if entryBytes, marshalErr := json.Marshal(entry); marshalErr == nil {
			_ = s.cache.SetDeviceList(ctx, queryHash, entryBytes)
		}
	}

	return respList, nextCursor, nil
}
