# Code Review — Week 3 & 4 Implementation

**審查範圍**：根據 `implementation_plan.md` 四個 Phase 與 `task.md` 所有完成項，對 `git diff` 中 18 個已修改檔 + 12 個新增檔進行交叉比對 Code Review。

**審查基準**：
- 功能完整性（與 Implementation Plan 對照）
- 錯誤處理與安全性
- 效能與並發安全
- 程式碼可維護性與一致性

---

## 變更檔案清單

| 狀態 | 檔案 | 變更摘要 |
|:---|:---|:---|
| Modified | [main.go](file:///c:/Projects/CC/Go/GoProjectTest/cmd/api/main.go) | DI 接線、slog 替換、cache/status/dashboard/health handler 注入 |
| Modified | [device_service.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service.go) | Cache-Aside + singleflight + 列表快取 + Saga 刪除 |
| Modified | [telemetry_service.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/telemetry_service.go) | Write-Through + Online Status + Alert INCR |
| Modified | [device_handler.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/handler/device_handler.go) | 207 Multi-Status 處理 |
| Modified | [health_handler.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/handler/health_handler.go) | 三 DB Ping 健康檢查 |
| Modified | [client.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/keydb/client.go) | TLS 連線 + CA cert 載入 |
| Modified | [routes.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/routes/routes.go) | 新增 status/dashboard/cache/health 路由 |
| Modified | [trace.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/middleware/trace.go) | Context 注入 logger.RequestIDKey |
| Modified | [config.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/config/config.go) | TLS 參數新增 |
| Modified | [logger.go](file:///c:/Projects/CC/Go/GoProjectTest/pkg/logger/logger.go) | ContextHandler + InitLogger |
| **New** | [cache_service.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/cache/cache_service.go) | Cache Service interface + Redis 實作 |
| **New** | [status_service.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/status_service.go) | 設備即時狀態彙整 |
| **New** | [dashboard_service.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/dashboard_service.go) | Dashboard overview 組裝 |
| **New** | [status_handler.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/handler/status_handler.go) | GET /devices/:id/status |
| **New** | [dashboard_handler.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/handler/dashboard_handler.go) | GET /dashboard/overview |
| **New** | [cache_handler.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/handler/cache_handler.go) | POST /cache/invalidate |
| **New** | [device_service_cache_test.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service_cache_test.go) | Cache-Aside / Null / Invalidation 測試 |
| **New** | [stress_test.go](file:///c:/Projects/CC/Go/GoProjectTest/tests/stress/stress_test.go) | 壓力測試腳本 |
| **New** | [.golangci.yml](file:///c:/Projects/CC/Go/GoProjectTest/.golangci.yml) | Linter 設定 |
| Modified | [README.md](file:///c:/Projects/CC/Go/GoProjectTest/README.md) | 架構圖 + API 總覽 + 啟動步驟 |
| **New** | [api_reference.md](file:///c:/Projects/CC/Go/GoProjectTest/.docs/api_reference.md) | 完整 API 文件 |
| **New** | [stress_test_report.md](file:///c:/Projects/CC/Go/GoProjectTest/.docs/stress_test_report.md) | 壓力測試報告模板 |

---

## Implementation Plan 完整性對照

| Plan 項目 | 狀態 | 備註 |
|:---|:---:|:---|
| 1.1 Cache Service 抽象層 | ✅ | `cache_service.go` 完整覆蓋所有規格方法 |
| 1.2 整合到 device_service / telemetry_service | ✅ | Cache-Aside + Write-Through + INCR 均已落地 |
| 1.3 Dashboard API + Pipeline | ⚠️ | 見 Finding #1 |
| 1.4 Cache Stampede 防護 (singleflight) | ✅ | `FindByID` 使用 `singleflight.Group` |
| 1.5 路由與 DI 更新 | ✅ | `routes.go` + `main.go` 已接線 |
| 2.1 Saga Pattern 刪除 | ✅ | PG Delete → KeyDB Invalidate → 207 Multi-Status |
| 2.2 Health Check | ✅ | PG / ScyllaDB / KeyDB 三方 Ping 含 2s timeout |
| 3.1 單元測試補充 | ✅ | Cache hit/miss/null/invalidation 三組測試 |
| 3.2 壓力測試腳本 | ✅ | stress build tag 隔離 |
| 3.3 golangci-lint | ✅ | 0 issues |
| 3.4 結構化日誌 (slog) | ✅ | 全面替換完成 |
| 4.1 README 更新 | ✅ | Mermaid + API 表 + TLS + 測試說明 |
| 4.2 API 文件 | ✅ | 完整 request/response 範例 |

---

## Findings

### 🔴 Critical

#### Finding #1 — Dashboard 未使用 Pipeline，且設備總數 Fallback 效能問題

**檔案**：[dashboard_service.go:63-70](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/dashboard_service.go#L63-L70)

Implementation Plan 1.3 規格明確指出：**「使用 KeyDB Pipeline 一次取回設備總數、在線數、各 severity 告警總數」**。但目前實作方式是逐個讀取，而非使用 Redis Pipeline 批量取得。

此外，當 `DeviceTotalCount` 快取為 0 時的 Fallback 邏輯呼叫了 `s.deviceRepo.List(ctx, "", 0, "", "", "", "")`。`limit=0` 的語意在 Repository 層取決於實作，若未特殊處理會回傳空集合或全表掃描。**全表掃描在設備量大時是嚴重的效能風險**。

```diff
 // 現狀 — 多次獨立讀取
 if s.cache != nil {
     total, err := s.cache.GetDeviceTotalCount(ctx)
     ...
 }
 if overview.DeviceTotal == 0 {
-    devices, _, err := s.deviceRepo.List(ctx, "", 0, "", "", "", "")
+    // 建議：在 DeviceRepository 新增 Count(ctx) (int64, error) 方法
+    count, err := s.deviceRepo.Count(ctx)
     ...
-    overview.DeviceTotal = int64(len(devices))
+    overview.DeviceTotal = count
 }
```

**建議**：
1. 將多個 KeyDB 讀取合併為 `redis.Pipeline()` 或 `redis.TxPipeline()`，一次往返取得所有值
2. 在 `DeviceRepository` 新增 `Count(ctx) (int64, error)` 方法，使用 `SELECT COUNT(*) FROM devices`，替代 `List` 全表掃描

---

#### Finding #2 — `cache_handler.go` 的 `POST /cache/invalidate` 缺乏權限控制

**檔案**：[cache_handler.go:28-48](file:///c:/Projects/CC/Go/GoProjectTest/internal/handler/cache_handler.go#L28-L48)

`POST /cache/invalidate` 端點接受任意 pattern 並直接使用 `SCAN + DEL` 清除快取。**任何可存取 API 的使用者都能清除全部快取**（例如 `pattern: "*"`），這會造成：
- 突發性的快取擊穿（Cache Avalanche）
- 安全風險：惡意清除導致 DB 瞬間高負載

```diff
 func (h *CacheHandler) Invalidate(c *gin.Context) {
+    // TODO: 加入 RBAC 權限驗證 — 僅 admin 角色可執行
+    // role := c.GetString("user_role")
+    // if role != "admin" {
+    //     response.Forbidden(c, "only admin can invalidate cache")
+    //     return
+    // }
+
     var req InvalidateReq
```

**建議**：
1. 目前已有 RBAC 架構，應在此端點加入 admin 角色限制
2. 至少加入 Pattern 白名單或防止 `*` 全清的保護

---

### 🟡 Major

#### Finding #3 — `InvalidateDeviceAll` 吞掉了 `Del` 錯誤

**檔案**：[cache_service.go:279-283](file:///c:/Projects/CC/Go/GoProjectTest/internal/cache/cache_service.go#L279-L283)

```go
if err := c.client.Del(ctx, keys...).Err(); err != nil {
    slog.ErrorContext(ctx, "failed to invalidate device cache keys", ...)
}
// 也要清除列表快取
return c.InvalidateAllLists(ctx)
```

第 279 行的 `Del` 若失敗只有 log 但不回傳錯誤。然而在 Saga Pattern（`device_service.go:287-290`）中，`InvalidateDeviceAll` 的回傳值決定了是否觸發 `ErrCacheCleanupFailed` → HTTP 207。**Del 失敗時靜默吞掉，會讓呼叫端誤認為 Saga Step 2 成功**。

```diff
 if err := c.client.Del(ctx, keys...).Err(); err != nil {
     slog.ErrorContext(ctx, "failed to invalidate device cache keys", ...)
+    return err
 }
```

---

#### Finding #4 — `CountOnline` 使用 `SCAN` 效能問題

**檔案**：[cache_service.go:128-139](file:///c:/Projects/CC/Go/GoProjectTest/internal/cache/cache_service.go#L128-L139)

`CountOnline` 使用 `SCAN device:online:* COUNT 1000` 逐筆掃描所有 key，在設備量大時（例如 10,000+）效能極差且阻塞 Redis 事件迴圈。

當前已經在 `SetOnlineStatus` 中使用 `SADD dashboard:online_set` 維護在線集合，但 `CountOnline` 卻沒使用這個 Set。

```diff
 func (c *redisCache) CountOnline(ctx context.Context) (int64, error) {
-    var count int64
-    iter := c.client.Scan(ctx, 0, keyDeviceOnline+"*", 1000).Iterator()
-    for iter.Next(ctx) {
-        count++
-    }
-    if err := iter.Err(); err != nil {
-        return 0, err
-    }
-    return count, nil
+    return c.client.SCard(ctx, keyDeviceOnlineSet).Result()
 }
```

---

#### Finding #5 — `ListCacheEntry` 結構重複定義

**檔案**：[device_service.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service.go) — `List` 方法中

`ListCacheEntry` struct 在同一個函式中定義了兩次（一次用於反序列化讀取，一次用於序列化寫入），且與快取層耦合在 Service 層。

```diff
+// ListCacheEntry 設備列表快取格式（移至 cache 或 dto package）
+type ListCacheEntry struct {
+    Devices    []*dto.DeviceResp `json:"devices"`
+    NextCursor string            `json:"next_cursor"`
+}

 func (s *deviceService) List(...) {
-    type ListCacheEntry struct { ... }  // 刪除第一處重複定義
-    type ListCacheEntry struct { ... }  // 刪除第二處重複定義
+    // 直接使用已提取的 ListCacheEntry
 }
```

---

#### Finding #6 — Implementation Plan 規定的「背景 Ticker 同步」未實作

**檔案**：[main.go](file:///c:/Projects/CC/Go/GoProjectTest/cmd/api/main.go)

Implementation Plan 1.3 原文：**「設備總數/在線數由背景 goroutine（`time.Ticker`）定時從 PG/KeyDB 同步」**。但 `main.go` 中沒有啟動任何背景 Ticker 去定期刷新 `dashboard:device_total` 或 `dashboard:online_set`。

目前的設計是：首次 Dashboard 請求觸發 Fallback 查詢 → 寫入快取 (TTL 30s)。這是可以接受的「懶加載」模式，但與計畫不符。

**建議**：考量到專案規模，保持目前的懶加載 + TTL 方式即可。建議在 Implementation Plan 或 README 中註明此設計選擇。

---

### 🟢 Minor

#### Finding #7 — `health_handler.go` 直接依賴基礎設施型別，破壞分層

**檔案**：[health_handler.go:15-18](file:///c:/Projects/CC/Go/GoProjectTest/internal/handler/health_handler.go#L15-L18)

`HealthHandler` 直接注入 `*gorm.DB`、`*scylla.Client`、`*keydb.Client`。Handler 層直接依賴 Infrastructure/ORM 型別，破壞了乾淨的分層架構（Handler → Service → Repository）。

**建議（非必要）**：抽出 `HealthService` interface 包裹三方 Ping 邏輯，Handler 只依賴 interface。這可以在後續重構時處理。

---

#### Finding #8 — `StatusResp` 和 `DashboardOverview` 定義在 Service 層

**檔案**：[status_service.go:14-20](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/status_service.go#L14-L20) / [dashboard_service.go:12-17](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/dashboard_service.go#L12-L17)

DTO 風格的 Response struct 定義在 `service` package 而非 `dto` package，與現有的 `dto.DeviceResp`、`dto.UserResp` 慣例不一致。

**建議**：移至 `internal/dto/` 保持一致性。

---

#### Finding #9 — `stress_test.go` 使用 `math/rand` 全域函式（Go 1.20+ 已改行為）

**檔案**：[stress_test.go:181](file:///c:/Projects/CC/Go/GoProjectTest/tests/stress/stress_test.go#L181)

```go
devID := deviceIDs[rand.Intn(len(deviceIDs))]
```

Go 1.20+ 的 `math/rand` 全域函式已自動使用隨機 seed（不再預設 seed=0），所以功能上沒問題。但多個 goroutine 同時呼叫全域 `rand.Intn` 會有 mutex 競爭。

**建議**：在高併發測試情境中，考慮每個 goroutine 使用 `rand.New(rand.NewSource(...))` 避免鎖競爭。

---

#### Finding #10 — `stress_test.go` P99 計算可能越界

**檔案**：[stress_test.go:238-240](file:///c:/Projects/CC/Go/GoProjectTest/tests/stress/stress_test.go#L238-L240)

```go
p50 := latencies[int(float64(totalRequests)*0.5)]
p95 := latencies[int(float64(totalRequests)*0.95)]
p99 := latencies[int(float64(totalRequests)*0.99)]
```

當 `totalRequests` 恰好使 `0.99 * N` 等於 `N` 時（例如 `N=100`, `0.99*100=99` → index 99 是合法的最後一個元素，但 `0.95*100=95` → index 95 也合法），這裡不會 panic。但更安全的做法是 `min(index, len-1)`。

```diff
-p99 := latencies[int(float64(totalRequests)*0.99)]
+p99Index := int(float64(totalRequests) * 0.99)
+if p99Index >= totalRequests {
+    p99Index = totalRequests - 1
+}
+p99 := latencies[p99Index]
```

---

### 💬 Nit

#### Finding #11 — `InsecureSkipVerify` 在開發環境可接受，但應加註明

**檔案**：[client.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/keydb/client.go) TLS config

`InsecureSkipVerify: insecure` 在本地開發自簽憑證環境中合理。建議在設定 `true` 時印出一次 slog.Warn 提醒這不應用於生產環境。

---

#### Finding #12 — Handler 回應格式不一致

部分 Handler（如 `status_handler.go`、`dashboard_handler.go`）直接使用 `c.JSON(200, gin.H{"code": 200, "message": "ok", "data": resp})`，而非呼叫 `response.Success(c, resp)` 工具函式。

建議統一使用 `pkg/response` 包提供的 helper function 保持一致性。

---

#### Finding #13 — `device_service.go` Create 中 `InvalidateAllLists` 錯誤被忽略但未記錄

**檔案**：[device_service.go:112](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service.go#L112)

```go
_ = s.cache.InvalidateAllLists(ctx)
```

這裡的 `_` 吞掉了錯誤。雖然 Cache 操作靜默失敗是既定設計（Implementation Plan 有明確規定），但建議至少加入 `slog.WarnContext` 記錄，與 `InvalidateDeviceAll` 中的 log 行為保持一致。

---

## 總結

| 嚴重度 | 數量 | 建議行動 |
|:---|:---:|:---|
| 🔴 Critical | 2 | #1 Dashboard 效能 + Pipeline、#2 Cache Invalidate 權限控制 |
| 🟡 Major | 4 | #3 Saga 錯誤吞掉、#4 SCAN 效能、#5 重複定義、#6 Ticker 設計差異 |
| 🟢 Minor | 4 | #7 分層、#8 DTO 位置、#9 rand 鎖、#10 越界 |
| 💬 Nit | 3 | #11 TLS 警告、#12 回應格式、#13 錯誤日誌 |

> [!IMPORTANT]
> **建議優先處理 Finding #1、#2、#3、#4**。這四項影響到功能正確性（Saga 錯誤處理）、效能（SCAN 全掃描 + List Fallback 全表查詢）、與安全性（無權限控制的快取清除）。
>
> 其餘 Minor/Nit 等級的項目可以在後續迭代中逐步修正。

確認後請回覆要修正哪些 Findings，我會著手進行程式碼修正。
