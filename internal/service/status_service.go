package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"GoProject/udm/internal/cache"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
	"GoProject/udm/internal/scylla"
)

// StatusService 定義設備即時狀態業務邏輯介面
type StatusService interface {
	GetDeviceStatus(ctx context.Context, deviceID uuid.UUID) (*dto.StatusResp, error)
}

type statusService struct {
	cache         cache.Service
	telemetryRepo scylla.TelemetryRepository
}

// NewStatusService 建立 StatusService 實體
func NewStatusService(cacheService cache.Service, telemetryRepo scylla.TelemetryRepository) StatusService {
	return &statusService{
		cache:         cacheService,
		telemetryRepo: telemetryRepo,
	}
}

func (s *statusService) GetDeviceStatus(ctx context.Context, deviceID uuid.UUID) (*dto.StatusResp, error) {
	idStr := deviceID.String()
	resp := &dto.StatusResp{DeviceID: deviceID}

	// 1. 在線狀態
	if s.cache != nil {
		online, err := s.cache.IsOnline(ctx, idStr)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check online status", "error", err)
		} else {
			resp.IsOnline = online
		}
	}

	// 2. 最新遙測（先嘗試 KeyDB，miss 才查 ScyllaDB）
	var gotFromCache bool
	if s.cache != nil {
		data, err := s.cache.GetLatestTelemetry(ctx, idStr)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get latest telemetry from cache", "error", err)
		} else if data != nil {
			var telemetries []*model.TelemetryData
			if jsonErr := json.Unmarshal(data, &telemetries); jsonErr == nil {
				resp.Latest = telemetries
				gotFromCache = true
			}
		}
	}

	if !gotFromCache && s.telemetryRepo != nil {
		telemetries, err := s.telemetryRepo.QueryLatest(ctx, deviceID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to query latest telemetry from ScyllaDB", "error", err)
		} else {
			resp.Latest = telemetries
		}
	}

	// 3. 告警計數
	if s.cache != nil {
		counts, err := s.cache.GetAlertCounts(ctx, idStr)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get alert counts from cache", "error", err)
		} else {
			resp.AlertCounts = counts
		}
	}

	return resp, nil
}
