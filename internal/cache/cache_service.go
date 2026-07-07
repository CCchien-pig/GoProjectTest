package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// TTL 常數定義
const (
	DeviceTTL          = 5 * time.Minute
	DeviceNullTTL      = 30 * time.Second // Cache 穿透防護：空值短 TTL
	OnlineStatusTTL    = 3 * time.Minute
	LatestTelemetryTTL = 30 * time.Second
	AlertCountTTL      = 10 * time.Minute
	DeviceListTTL      = 2 * time.Minute
	DashboardTTL       = 30 * time.Second
)

// Key prefix 常數
const (
	keyDevice           = "device:"
	keyDeviceOnline     = "device:online:"
	keyTelemetryLatest  = "telemetry:latest:"
	keyAlertCount       = "alert:count:"
	keyDeviceList       = "devices:list:"
	keyDashboard        = "dashboard:overview"
	keyDeviceTotalCount = "dashboard:device_total"
	keyDeviceOnlineSet  = "dashboard:online_set"
)

// Service 定義快取操作介面
type Service interface {
	// Device Cache-Aside
	GetDevice(ctx context.Context, id string) ([]byte, error)
	SetDevice(ctx context.Context, id string, data []byte) error
	SetDeviceNull(ctx context.Context, id string) error
	InvalidateDevice(ctx context.Context, id string) error

	// 在線狀態
	SetOnlineStatus(ctx context.Context, deviceID string) error
	IsOnline(ctx context.Context, deviceID string) (bool, error)
	CountOnline(ctx context.Context) (int64, error)

	// Write-Through 最新遙測
	SetLatestTelemetry(ctx context.Context, deviceID string, data []byte) error
	GetLatestTelemetry(ctx context.Context, deviceID string) ([]byte, error)

	// 告警計數
	IncrAlertCount(ctx context.Context, deviceID string, severity string) error
	GetAlertCounts(ctx context.Context, deviceID string) (map[string]int64, error)
	SyncAlertCount(ctx context.Context, deviceID string, severity string, count int64) error

	// 設備列表快取
	GetDeviceList(ctx context.Context, queryHash string) ([]byte, error)
	SetDeviceList(ctx context.Context, queryHash string, data []byte) error
	InvalidateAllLists(ctx context.Context) error

	// Dashboard
	GetDashboard(ctx context.Context) ([]byte, error)
	SetDashboard(ctx context.Context, data []byte) error
	SetDeviceTotalCount(ctx context.Context, count int64) error
	GetDeviceTotalCount(ctx context.Context) (int64, error)
	GetDashboardMetricsPipeline(ctx context.Context) (int64, int64, error)

	// 管理
	InvalidateByPattern(ctx context.Context, pattern string) (int64, error)
	InvalidateDeviceAll(ctx context.Context, deviceID string) error

	// Ping (Health Check)
	Ping(ctx context.Context) error
}

type redisCache struct {
	client redis.UniversalClient
}

// NewService 建立 Cache Service 實體
func NewService(client redis.UniversalClient) Service {
	return &redisCache{client: client}
}

// ── Device Cache-Aside ──────────────────────────────────────

func (c *redisCache) GetDevice(ctx context.Context, id string) ([]byte, error) {
	val, err := c.client.Get(ctx, keyDevice+id).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *redisCache) SetDevice(ctx context.Context, id string, data []byte) error {
	return c.client.Set(ctx, keyDevice+id, data, DeviceTTL).Err()
}

func (c *redisCache) SetDeviceNull(ctx context.Context, id string) error {
	return c.client.Set(ctx, keyDevice+id, []byte("null"), DeviceNullTTL).Err()
}

func (c *redisCache) InvalidateDevice(ctx context.Context, id string) error {
	return c.client.Del(ctx, keyDevice+id).Err()
}

// ── 在線狀態 ────────────────────────────────────────────────

func (c *redisCache) SetOnlineStatus(ctx context.Context, deviceID string) error {
	pipe := c.client.Pipeline()
	pipe.Set(ctx, keyDeviceOnline+deviceID, "1", OnlineStatusTTL)
	pipe.SAdd(ctx, keyDeviceOnlineSet, deviceID)
	pipe.Expire(ctx, keyDeviceOnlineSet, OnlineStatusTTL+time.Minute) // 稍長於在線 TTL
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) IsOnline(ctx context.Context, deviceID string) (bool, error) {
	exists, err := c.client.Exists(ctx, keyDeviceOnline+deviceID).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (c *redisCache) CountOnline(ctx context.Context) (int64, error) {
	// Finding #4: 使用 SCARD 取代 SCAN 逐筆掃描，O(1) 操作
	// keyDeviceOnlineSet 由 SetOnlineStatus 透過 SADD 維護，已與在線 key 保持同步
	return c.client.SCard(ctx, keyDeviceOnlineSet).Result()
}

// ── Write-Through 最新遙測 ──────────────────────────────────

func (c *redisCache) SetLatestTelemetry(ctx context.Context, deviceID string, data []byte) error {
	return c.client.Set(ctx, keyTelemetryLatest+deviceID, data, LatestTelemetryTTL).Err()
}

func (c *redisCache) GetLatestTelemetry(ctx context.Context, deviceID string) ([]byte, error) {
	val, err := c.client.Get(ctx, keyTelemetryLatest+deviceID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

// ── 告警計數 ────────────────────────────────────────────────

func (c *redisCache) IncrAlertCount(ctx context.Context, deviceID string, severity string) error {
	key := fmt.Sprintf("%s%s:%s", keyAlertCount, deviceID, severity)
	pipe := c.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, AlertCountTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) GetAlertCounts(ctx context.Context, deviceID string) (map[string]int64, error) {
	counts := make(map[string]int64)
	severities := []string{"info", "warning", "critical"}

	keys := make([]string, len(severities))
	for i, sev := range severities {
		keys[i] = fmt.Sprintf("%s%s:%s", keyAlertCount, deviceID, sev)
	}

	results, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	for i, sev := range severities {
		if results[i] != nil {
			if v, ok := results[i].(string); ok {
				var n int64
				_, _ = fmt.Sscanf(v, "%d", &n)
				counts[sev] = n
			}
		}
	}
	return counts, nil
}

func (c *redisCache) SyncAlertCount(ctx context.Context, deviceID string, severity string, count int64) error {
	key := fmt.Sprintf("%s%s:%s", keyAlertCount, deviceID, severity)
	return c.client.Set(ctx, key, count, AlertCountTTL).Err()
}

// ── 設備列表快取 ────────────────────────────────────────────

func (c *redisCache) GetDeviceList(ctx context.Context, queryHash string) ([]byte, error) {
	val, err := c.client.Get(ctx, keyDeviceList+queryHash).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *redisCache) SetDeviceList(ctx context.Context, queryHash string, data []byte) error {
	return c.client.Set(ctx, keyDeviceList+queryHash, data, DeviceListTTL).Err()
}

func (c *redisCache) InvalidateAllLists(ctx context.Context) error {
	return c.deleteByPattern(ctx, keyDeviceList+"*")
}

// ── Dashboard ───────────────────────────────────────────────

func (c *redisCache) GetDashboard(ctx context.Context) ([]byte, error) {
	val, err := c.client.Get(ctx, keyDashboard).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *redisCache) SetDashboard(ctx context.Context, data []byte) error {
	return c.client.Set(ctx, keyDashboard, data, DashboardTTL).Err()
}

func (c *redisCache) SetDeviceTotalCount(ctx context.Context, count int64) error {
	return c.client.Set(ctx, keyDeviceTotalCount, count, 5*time.Minute).Err()
}

func (c *redisCache) GetDeviceTotalCount(ctx context.Context) (int64, error) {
	val, err := c.client.Get(ctx, keyDeviceTotalCount).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var n int64
	_, _ = fmt.Sscanf(val, "%d", &n)
	return n, nil
}

// GetDashboardMetricsPipeline 使用 Pipeline 一次取得多個 Dashboard 基礎指標 (Finding #1 完整實作)
func (c *redisCache) GetDashboardMetricsPipeline(ctx context.Context) (int64, int64, error) {
	pipe := c.client.Pipeline()

	// 1. 在線設備數量 (SCARD)
	onlineCmd := pipe.SCard(ctx, keyDeviceOnlineSet)
	// 2. 設備總數快取 (GET)
	totalCmd := pipe.Get(ctx, keyDeviceTotalCount)

	_, err := pipe.Exec(ctx)
	// pipeline 中如果有某個 key 不存在，Exec 也會回傳 error (redis.Nil)。
	// 但我們允許 total 不存在 (fallback)，因此過濾掉 redis.Nil
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}

	onlineCount := onlineCmd.Val()
	
	// totalCount 可能是 nil
	var totalCount int64
	if totalStr := totalCmd.Val(); totalStr != "" {
		_, _ = fmt.Sscanf(totalStr, "%d", &totalCount)
	}

	return onlineCount, totalCount, nil
}

// ── 管理 ────────────────────────────────────────────────────

func (c *redisCache) InvalidateByPattern(ctx context.Context, pattern string) (int64, error) {
	var total int64
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return 0, err
	}
	if len(keys) > 0 {
		deleted, err := c.client.Del(ctx, keys...).Result()
		if err != nil {
			return 0, err
		}
		total = deleted
	}
	return total, nil
}

func (c *redisCache) InvalidateDeviceAll(ctx context.Context, deviceID string) error {
	keys := []string{
		keyDevice + deviceID,
		keyTelemetryLatest + deviceID,
		keyDeviceOnline + deviceID,
	}
	// 告警計數 keys
	for _, sev := range []string{"info", "warning", "critical"} {
		keys = append(keys, fmt.Sprintf("%s%s:%s", keyAlertCount, deviceID, sev))
	}

	// Finding #3: 回傳 Del 錯誤，讓 Saga 呼叫端能正確判斷快取清除是否成功
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		slog.ErrorContext(ctx, "failed to invalidate device cache keys", "device_id", deviceID, "error", err)
		return err
	}
	// 也要清除列表快取
	return c.InvalidateAllLists(ctx)
}

func (c *redisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// ── Helpers ─────────────────────────────────────────────────

func (c *redisCache) deleteByPattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

// HashQuery 將查詢參數雜湊為 cache key
func HashQuery(params ...string) string {
	data, _ := json.Marshal(params)
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:8])
}
