package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
)

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
}

// NewDeviceService 建立 DeviceService 實體
func NewDeviceService(repo repository.DeviceRepository, userRepo repository.UserRepository, telemetryRepo scylla.TelemetryRepository) DeviceService {
	return &deviceService{
		repo:          repo,
		userRepo:      userRepo,
		telemetryRepo: telemetryRepo,
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
		for _, uid := range req.UserIDs {
			u, err := s.userRepo.FindByID(ctx, uid)
			if err != nil {
				return nil, fmt.Errorf("find user %s: %w", uid, err)
			}
			if u == nil {
				return nil, ErrUserNotFound
			}
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
		reloaded, reloadErr := s.repo.FindByID(ctx, device.ID)
		if reloadErr != nil {
			return nil, fmt.Errorf("reload device after create: %w", reloadErr)
		}
		if reloaded != nil {
			device = reloaded
		}
	}

	return dto.ToDeviceResp(device), nil
}

func (s *deviceService) FindByID(ctx context.Context, id uuid.UUID) (*dto.DeviceResp, error) {
	device, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find device: %w", err)
	}
	if device == nil {
		return nil, ErrDeviceNotFound
	}

	resp := dto.ToDeviceResp(device)

	// 從 ScyllaDB 查出該設備最新遙測資料一併回傳
	if s.telemetryRepo != nil {
		telemetries, telErr := s.telemetryRepo.QueryLatest(ctx, id)
		if telErr == nil && len(telemetries) > 0 {
			resp.LatestTelemetry = telemetries
		}
	}

	return resp, nil
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

	if err := s.repo.Update(ctx, device); err != nil {
		return nil, fmt.Errorf("update device: %w", err)
	}

	// 處理關聯的 Users 更新
	if req.UserIDs != nil {
		var usersToBind []model.User
		if len(req.UserIDs) > 0 {
			for _, uid := range req.UserIDs {
				u, err := s.userRepo.FindByID(ctx, uid)
				if err != nil {
					return nil, fmt.Errorf("find user %s: %w", uid, err)
				}
				if u == nil {
					return nil, ErrUserNotFound
				}
				usersToBind = append(usersToBind, *u)
			}
		}
		if err := s.repo.ReplaceUsers(ctx, device, usersToBind); err != nil {
			return nil, fmt.Errorf("replace device users: %w", err)
		}
	}

	// 重新 Preload 取得 Users 關聯資料
	reloaded, reloadErr := s.repo.FindByID(ctx, device.ID)
	if reloadErr != nil {
		return nil, fmt.Errorf("reload device after update: %w", reloadErr)
	}
	if reloaded != nil {
		device = reloaded
	}

	return dto.ToDeviceResp(device), nil
}

func (s *deviceService) Delete(ctx context.Context, id uuid.UUID) error {
	device, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if device == nil {
		return ErrDeviceNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete device: %w", err)
	}
	return nil
}

func (s *deviceService) List(ctx context.Context, cursor string, limit int, deviceType, status, location, search string) ([]*dto.DeviceResp, string, error) {
	if limit <= 0 {
		limit = 10
	}
	devices, nextCursor, err := s.repo.List(ctx, cursor, limit, deviceType, status, location, search)
	if err != nil {
		return nil, "", fmt.Errorf("list devices: %w", err)
	}
	return dto.ToDeviceRespList(devices), nextCursor, nil
}
