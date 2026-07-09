package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"GoProject/udm/internal/cache"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/repository"
)

// DashboardService 定義儀表板業務邏輯介面
type DashboardService interface {
	GetOverview(ctx context.Context) (*dto.DashboardOverview, error)
}

type dashboardService struct {
	cache      cache.Service
	deviceRepo repository.DeviceRepository
}

// NewDashboardService 建立 DashboardService 實體
func NewDashboardService(cacheService cache.Service, deviceRepo repository.DeviceRepository) DashboardService {
	return &dashboardService{
		cache:      cacheService,
		deviceRepo: deviceRepo,
	}
}

func (s *dashboardService) GetOverview(ctx context.Context) (*dto.DashboardOverview, error) {
	// 嘗試從 KeyDB 快取取得整包 dashboard 結果
	if s.cache != nil {
		data, err := s.cache.GetDashboard(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get dashboard from cache", "error", err)
		} else if data != nil {
			var overview dto.DashboardOverview
			if jsonErr := json.Unmarshal(data, &overview); jsonErr == nil {
				return &overview, nil
			}
		}
	}

	// Cache miss — 組裝 DashboardOverview
	overview := &dto.DashboardOverview{
		AlertCounts: make(map[string]int64),
	}

	// 全域告警計數：從 KeyDB 讀取（獨立事件驅動，不查 DB）
	// 每次遙測觸發告警時，IncrAlertCount 已同步維護 alert:count:global:{severity}
	if s.cache != nil {
		if counts, err := s.cache.GetGlobalAlertCounts(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to get global alert counts from cache", "error", err)
		} else {
			overview.AlertCounts = counts
		}
	}

	// 使用 Pipeline 一次從 KeyDB 取得在線數 + 設備總數 (Finding #1 完整實作)
	if s.cache != nil {
		onlineCount, total, err := s.cache.GetDashboardMetricsPipeline(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get dashboard metrics via pipeline", "error", err)
		} else {
			overview.DeviceOnline = onlineCount
			if total > 0 {
				overview.DeviceTotal = total
			}
		}
	}

	// 設備總數 Fallback：若快取中無資料，改用 DB COUNT(*) 查詢 (Finding #1 — 替換 List 全表掃描)
	if overview.DeviceTotal == 0 && s.deviceRepo != nil {
		count, err := s.deviceRepo.Count(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to count devices from DB", "error", err)
		} else {
			overview.DeviceTotal = count
			// 同步寫回 KeyDB 快取
			if s.cache != nil {
				if err := s.cache.SetDeviceTotalCount(ctx, count); err != nil {
					slog.ErrorContext(ctx, "failed to cache device total count", "error", err)
				}
			}
		}
	}

	// 寫回快取
	if s.cache != nil {
		data, _ := json.Marshal(overview)
		if err := s.cache.SetDashboard(ctx, data); err != nil {
			slog.ErrorContext(ctx, "failed to set dashboard cache", "error", err)
		}
	}

	return overview, nil
}
