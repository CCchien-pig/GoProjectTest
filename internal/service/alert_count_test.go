package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
)

// ─── 單一設備告警查詢測試 (StatusService) ───────────────────────────────

// TestStatusService_AlertCounts_FromCache 驗證 GetDeviceStatus 能正確從
// KeyDB 讀取並回傳指定設備的各等級告警計數（不查 DB）
func TestStatusService_AlertCounts_FromCache(t *testing.T) {
	cacheSvc := newMockCacheService()
	svc := NewStatusService(cacheSvc, nil)

	deviceID := uuid.New()
	idStr := deviceID.String()

	// 模擬已有 3 筆 critical、1 筆 warning 的告警計數
	cacheSvc.alertCounts[idStr] = map[string]int64{
		"critical": 3,
		"warning":  1,
		"info":     0,
	}

	resp, err := svc.GetDeviceStatus(context.Background(), deviceID)
	if err != nil {
		t.Fatalf("GetDeviceStatus failed: %v", err)
	}

	if resp.AlertCounts["critical"] != 3 {
		t.Errorf("expected critical=3, got %d", resp.AlertCounts["critical"])
	}
	if resp.AlertCounts["warning"] != 1 {
		t.Errorf("expected warning=1, got %d", resp.AlertCounts["warning"])
	}
}

// TestStatusService_AlertCounts_IncrByTelemetry 驗證遙測寫入觸發告警後，
// 設備級別的告警計數會在 KeyDB 中正確遞增
func TestStatusService_AlertCounts_IncrByTelemetry(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	alertRuleRepo := newMockAlertRuleRepository()
	telRepo := &mockTelemetryRepository{}
	alertEventRepo := &mockAlertEventRepository{}
	cacheSvc := newMockCacheService()

	telSvc := NewTelemetryService(telRepo, alertEventRepo, deviceRepo, alertRuleRepo, cacheSvc)
	statusSvc := NewStatusService(cacheSvc, nil)

	devID := uuid.New()
	deviceRepo.devices[devID] = &model.Device{ID: devID, DeviceCode: "DEV-ALERT-01"}

	rule := &model.AlertRule{
		ID:         uuid.New(),
		DeviceID:   devID,
		MetricName: "voltage",
		Operator:   "lt", // 低於閾值觸發
		Threshold:  220.0,
		Severity:   "warning",
		IsEnabled:  true,
	}
	alertRuleRepo.rules[rule.ID] = rule
	alertRuleRepo.rulesByDevice[devID] = []*model.AlertRule{rule}

	// 上報一筆正常數據（230V > 220V，不觸發告警）
	_ = telSvc.BatchInsert(context.Background(), devID, &dto.BatchTelemetryReq{
		Points: []dto.TelemetryPoint{{RecordedAt: time.Now(), MetricName: "voltage", Value: 230.0}},
	})

	resp, _ := statusSvc.GetDeviceStatus(context.Background(), devID)
	if resp.AlertCounts["warning"] != 0 {
		t.Errorf("expected warning=0 before trigger, got %d", resp.AlertCounts["warning"])
	}

	// 上報一筆異常數據（210V < 220V，觸發 warning）
	_ = telSvc.BatchInsert(context.Background(), devID, &dto.BatchTelemetryReq{
		Points: []dto.TelemetryPoint{{RecordedAt: time.Now(), MetricName: "voltage", Value: 210.0}},
	})

	resp, _ = statusSvc.GetDeviceStatus(context.Background(), devID)
	if resp.AlertCounts["warning"] != 1 {
		t.Errorf("expected warning=1 after trigger, got %d", resp.AlertCounts["warning"])
	}
}

// ─── Dashboard 全域告警查詢測試 (DashboardService) ──────────────────────

// TestDashboardService_AlertCounts_FromGlobalCache 驗證 Dashboard 回傳的告警計數
// 是全域性的（不分設備），且完全從 KeyDB 讀取（不查 DB）
func TestDashboardService_AlertCounts_FromGlobalCache(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	cacheSvc := newMockCacheService()

	// 直接在 Mock 的全域計數器裡預置資料
	cacheSvc.globalAlertCounts["critical"] = 5
	cacheSvc.globalAlertCounts["warning"] = 10
	cacheSvc.globalAlertCounts["info"] = 2

	dashSvc := NewDashboardService(cacheSvc, deviceRepo)
	overview, err := dashSvc.GetOverview(context.Background())
	if err != nil {
		t.Fatalf("GetOverview failed: %v", err)
	}

	if overview.AlertCounts["critical"] != 5 {
		t.Errorf("expected critical=5, got %d", overview.AlertCounts["critical"])
	}
	if overview.AlertCounts["warning"] != 10 {
		t.Errorf("expected warning=10, got %d", overview.AlertCounts["warning"])
	}
	if overview.AlertCounts["info"] != 2 {
		t.Errorf("expected info=2, got %d", overview.AlertCounts["info"])
	}
}

// TestDashboardService_AlertCounts_AggregatesAcrossDevices 驗證多台設備觸發告警後，
// Dashboard 顯示的是所有設備的加總（全域計數器，非單一設備）
func TestDashboardService_AlertCounts_AggregatesAcrossDevices(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	alertRuleRepo := newMockAlertRuleRepository()
	alertEventRepo := &mockAlertEventRepository{}
	telRepo := &mockTelemetryRepository{}
	cacheSvc := newMockCacheService()

	telSvc := NewTelemetryService(telRepo, alertEventRepo, deviceRepo, alertRuleRepo, cacheSvc)
	dashSvc := NewDashboardService(cacheSvc, deviceRepo)

	// 建立兩台設備，各有一條告警規則
	dev1ID := uuid.New()
	dev2ID := uuid.New()
	deviceRepo.devices[dev1ID] = &model.Device{ID: dev1ID, DeviceCode: "DEV-A"}
	deviceRepo.devices[dev2ID] = &model.Device{ID: dev2ID, DeviceCode: "DEV-B"}

	rule1 := &model.AlertRule{ID: uuid.New(), DeviceID: dev1ID, MetricName: "temp", Operator: "gt", Threshold: 50.0, Severity: "critical", IsEnabled: true}
	rule2 := &model.AlertRule{ID: uuid.New(), DeviceID: dev2ID, MetricName: "temp", Operator: "gt", Threshold: 50.0, Severity: "critical", IsEnabled: true}

	alertRuleRepo.rules[rule1.ID] = rule1
	alertRuleRepo.rulesByDevice[dev1ID] = []*model.AlertRule{rule1}
	alertRuleRepo.rules[rule2.ID] = rule2
	alertRuleRepo.rulesByDevice[dev2ID] = []*model.AlertRule{rule2}

	// 設備 A 上報一次超標 → critical +1（設備A=1，全域=1）
	_ = telSvc.BatchInsert(context.Background(), dev1ID, &dto.BatchTelemetryReq{
		Points: []dto.TelemetryPoint{{RecordedAt: time.Now(), MetricName: "temp", Value: 60.0}},
	})
	// 設備 B 上報兩次超標 → critical +2（設備B=2，全域=3）
	_ = telSvc.BatchInsert(context.Background(), dev2ID, &dto.BatchTelemetryReq{
		Points: []dto.TelemetryPoint{{RecordedAt: time.Now(), MetricName: "temp", Value: 61.0}},
	})
	_ = telSvc.BatchInsert(context.Background(), dev2ID, &dto.BatchTelemetryReq{
		Points: []dto.TelemetryPoint{{RecordedAt: time.Now(), MetricName: "temp", Value: 62.0}},
	})

	// Dashboard 取得的應為全域合計：3 次 critical
	overview, err := dashSvc.GetOverview(context.Background())
	if err != nil {
		t.Fatalf("GetOverview failed: %v", err)
	}
	if overview.AlertCounts["critical"] != 3 {
		t.Errorf("expected global critical=3 (1 from DevA + 2 from DevB), got %d", overview.AlertCounts["critical"])
	}

	// 驗證單一設備計數器仍然獨立且正確
	if cacheSvc.alertCounts[dev1ID.String()]["critical"] != 1 {
		t.Errorf("expected devA critical=1, got %d", cacheSvc.alertCounts[dev1ID.String()]["critical"])
	}
	if cacheSvc.alertCounts[dev2ID.String()]["critical"] != 2 {
		t.Errorf("expected devB critical=2, got %d", cacheSvc.alertCounts[dev2ID.String()]["critical"])
	}
}
