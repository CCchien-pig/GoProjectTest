# Code Review Report - UDM Platform (Week 1 & 2)

> Reviewed: 2026-07-02  
> Scope: All files under `internal/`, `cmd/api/main.go`, `pkg/response/`, `go.mod`

---

## 🔴 High Priority — 邏輯錯誤 / 架構問題

### 1. `device_service.go` — `FindByID` 呼叫有 nil panic 風險

**File**: [`device_service.go:104-108`](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service.go#L104-L108)

```go
// 問題：telemetryRepo 可能是 nil（ScyllaDB 離線時），
// 但沒有先做 nil 檢查就直接呼叫 QueryLatest
telemetries, err := s.telemetryRepo.QueryLatest(ctx, id)  // ← 可能 panic
```

**原因**：`NewDeviceService` 接受 `scylla.TelemetryRepository`（interface），當 ScyllaDB 離線時 `main.go` 傳入 `nil`。Go 中直接呼叫 `nil` interface 會 panic，不是回傳 error。

**修正**：

```go
if s.telemetryRepo != nil {
    telemetries, err := s.telemetryRepo.QueryLatest(ctx, id)
    if err == nil && len(telemetries) > 0 {
        resp.LatestTelemetry = telemetries
    }
}
```

---

### 2. `main.go` — PostgreSQL 離線時，服務依然建立（必定 panic）

**File**: [`main.go:77-80`](file:///c:/Projects/CC/Go/GoProjectTest/cmd/api/main.go#L77-L80)

```go
// 問題：若 PostgreSQL 連線失敗，userRepo/deviceRepo/alertRuleRepo 均為 nil
// 但後面的 Service 仍然建立並使用這些 nil repo
userService := service.NewUserService(userRepo)            // userRepo = nil → panic
deviceService := service.NewDeviceService(deviceRepo, ...) // deviceRepo = nil → panic
```

**原因**：Service 層不像 ScyllaDB 有 nil 判斷，所有方法都直接呼叫 `s.repo.xxx()`，nil repo 會立即 panic。與 ScyllaDB 的「降級」模式不一致。

**修正建議**（兩擇一）：

- **A. 強制中斷**：PostgreSQL 為核心資料庫，連線失敗直接 `log.Fatalf`，不允許降級
- **B. 一致降級**：為所有 Repo 加 nil 檢查，回傳 `ErrDatabaseOffline`

---

### 3. `device_service.go` — `Create` 中 `FindByID` 錯誤被靜默忽略

**File**: [`device_service.go:87`](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service.go#L87)

```go
if req.OwnerID != nil {
    device, _ = s.repo.FindByID(ctx, device.ID)  // ← 錯誤被忽略
}
```

若 FindByID 回傳錯誤（DB 暫時不可用），錯誤會被丟棄，`device` 可能為 nil，後面的 `dto.ToDeviceResp(device)` 會回傳空資料，造成 client 看到空的 device response。

**修正**：

```go
if req.OwnerID != nil {
    if reloaded, err := s.repo.FindByID(ctx, device.ID); err == nil && reloaded != nil {
        device = reloaded
    }
}
```

---

### 4. `device_service.go` — `Update` 的 `FindByID` 錯誤同樣被靜默忽略

**File**: [`device_service.go:152`](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service.go#L152)

```go
device, _ = s.repo.FindByID(ctx, device.ID) // ← 修改後 reload 也忽略錯誤
```

同上問題，需加錯誤處理。

---

### 5. `telemetry_repo.go` — `Query` 中對 SELECT 欄位的 Scan 順序與 SELECT 不符

**File**: [`telemetry_repo.go:67-77`](file:///c:/Projects/CC/Go/GoProjectTest/internal/scylla/telemetry_repo.go#L67-L77)

```sql
-- SELECT 順序: device_id, date, recorded_at, metric_name, value, unit, tags
```

```go
// Scan 順序: &devID, &date, &recordedAt, &mName, &value, &unit, &tags
for iter.Scan(&devID, &date, &recordedAt, &mName, &value, &unit, &tags) {
```

對照 SELECT 欄位順序：`device_id, date, recorded_at, metric_name, value, unit, tags` → Scan 順序相符✓

不過 `QueryLatest` 也是相同 SELECT 格式，但欄位和 Scan 需同步確認（目前看起來正確）。這部分沒問題，可跳過。

---

## 🟡 Medium Priority — 設計缺失 / 業務邏輯問題

### 6. `scylla/client.go` — `EnsureSchema` 使用 `fmt.Sprintf` 拼接 keyspace 名稱（SQL Injection 風險）

**File**: [`client.go:56-97`](file:///c:/Projects/CC/Go/GoProjectTest/internal/scylla/client.go#L56-L97)

```go
err := c.Session.Query(fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s ...`, keyspace)).Exec()
```

CQL 不支援 `?` 佔位符用於 keyspace/table 名稱，這是可接受的做法，但需確保 `keyspace` 值只來自受信任的設定（從 `.env.dev` 讀取），不可由使用者輸入。目前從 `config.go` 讀取是安全的，**請確保不要在任何地方允許 keyspace 由外部輸入**。

---

### 7. `alert_rule.go` model — 缺少 `UpdatedAt` 欄位但有 `BeforeUpdate` 需求

**File**: [`alert_rule.go`](file:///c:/Projects/CC/Go/GoProjectTest/internal/model/alert_rule.go)

`AlertRule` 只有 `CreatedAt`，沒有 `UpdatedAt`。但 `AlertRuleService.Update` 會呼叫 `repo.Update(ctx, rule)` 做 GORM `Save`，資料表沒有 `updated_at` 欄位，查詢時無法知道規則被修改的時間，對運維和 debug 有影響。

**修正**：在 `AlertRule` 加入 `UpdatedAt` 欄位，並加入 `BeforeUpdate` hook（同 User 和 Device 的做法）。

---

### 8. `telemetry_service.go` — 告警評估使用 `fmt.Printf` 而非 log

**File**: [`telemetry_service.go:101`](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/telemetry_service.go#L101)

```go
fmt.Printf("failed to insert alert event: %v\n", err)
```

Service 層不應直接用 `fmt.Printf`，應使用標準 `log` 或結構化 logger（後續 Week 3/4 補強時一定會被考到）。

---

### 9. `telemetry_repo.go` — `Query` 中的排序演算法效能低落（O(n²)）

**File**: [`telemetry_repo.go:99-105`](file:///c:/Projects/CC/Go/GoProjectTest/internal/scylla/telemetry_repo.go#L99-L105)

```go
// 目前是手寫的 bubble sort，時間複雜度 O(n²)
for i := 0; i < len(result); i++ {
    for j := i + 1; j < len(result); j++ {
        if result[i].RecordedAt.Before(result[j].RecordedAt) {
            result[i], result[j] = result[j], result[i]
        }
    }
}
```

當遙測資料量大時（100 萬筆跨 N 天）效能嚴重降低。

**修正**：改用標準庫的 `sort.Slice`：

```go
import "sort"

sort.Slice(result, func(i, j int) bool {
    return result[i].RecordedAt.After(result[j].RecordedAt)
})
```

---

### 10. `dto/device.go` — `LatestTelemetry` 使用 `interface{}` 型別過於寬鬆

**File**: [`device.go:46`](file:///c:/Projects/CC/Go/GoProjectTest/internal/dto/device.go#L46)

```go
LatestTelemetry interface{} `json:"latest_telemetry,omitempty"`
```

直接使用 `interface{}` 代表任何類型都能放入，失去型別安全性，IDE 也無法提供補全。應使用具體型別：

**修正**：

```go
import "GoProject/udm/internal/model"

LatestTelemetry []*model.TelemetryData `json:"latest_telemetry,omitempty"`
```

---

### 11. `device_repo.go` — Cursor 解碼失敗時靜默忽略

**File**: [`device_repo.go:90-94`](file:///c:/Projects/CC/Go/GoProjectTest/internal/repository/device_repo.go#L90-L94)

```go
if cursor != "" {
    cursorTime, cursorID, err := decodeCursor(cursor)
    if err == nil {  // ← 解碼失敗時，直接跳過，等同於從第一頁查詢
        query = query.Where(...)
    }
}
```

當 client 傳入損壞的 cursor，系統會靜默地從第一頁返回資料，而不是告知 client 錯誤。更好的做法是在 Service 層或 Handler 層先做格式驗證，讓 client 知道 cursor 無效。

---

### 12. `scylla/client.go` — `NewClient` 傳入 `hosts []string` 但 `config.go` 的 `ScyllaHosts` 是單一字串

**File**: [`client.go:16`](file:///c:/Projects/CC/Go/GoProjectTest/internal/scylla/client.go#L16) & [`main.go:58`](file:///c:/Projects/CC/Go/GoProjectTest/cmd/api/main.go#L58)

```go
// main.go 呼叫時傳入單元素 slice
scyllaClient, err = scylla.NewClient([]string{cfg.ScyllaHosts}, cfg.ScyllaKeyspace)

// config.go 只存一個字串，但未處理多節點 host 情境
ScyllaHosts string // e.g. "localhost:9042" 或 "host1:9042,host2:9042"
```

若生產環境 ScyllaDB 有多個節點（hosts），目前的設計無法支援逗號分隔的多個 host 字串。

**修正**：在 config 中改用 `strings.Split(cfg.ScyllaHosts, ",")` 處理多節點：

```go
// main.go
hosts := strings.Split(cfg.ScyllaHosts, ",")
scyllaClient, err = scylla.NewClient(hosts, cfg.ScyllaKeyspace)
```

---

## 🟢 Low Priority — 小改善建議

### 13. `go.mod` — 所有依賴都標記為 `indirect`

**File**: [`go.mod`](file:///c:/Projects/CC/Go/GoProjectTest/go.mod)

所有套件都標記為 `// indirect`，表示 `go mod tidy` 尚未執行或 module dependencies 沒有在 Go 程式中直接引入。

**建議**：執行 `go mod tidy` 修正依賴分類。

---

### 14. `user_service.go` — `SoftDelete` 的錯誤沒有 wrap

**File**: [`user_service.go:141`](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/user_service.go#L141)

```go
func (s *userService) SoftDelete(ctx context.Context, id uuid.UUID) error {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return err  // ← 沒有 wrap，失去呼叫位置 context
    }
```

**修正**：`return fmt.Errorf("find user for delete: %w", err)`

---

### 15. `alert_rule_service.go` — `Delete` 的錯誤同上未 wrap

**File**: [`alert_rule_service.go:138`](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/alert_rule_service.go#L138)

```go
if err != nil {
    return err  // ← 未 wrap
}
```

---

### 16. `device_service.go` — `Delete` 的 FindByID 錯誤未 wrap

**File**: [`device_service.go:160`](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service.go#L160)

```go
if err != nil {
    return err  // ← 未 wrap
}
```

---

### 17. `TraceID` middleware — 未優先讀取 `X-Request-ID` Header

**File**: [`trace.go`](file:///c:/Projects/CC/Go/GoProjectTest/internal/middleware/trace.go)

目前每次都產生全新 UUID，即使 upstream（如 API Gateway、前端）已傳入 `X-Request-ID`，也會被覆蓋。

**改善**：

```go
func TraceID() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}
```

---

## 📋 問題彙整表

| # | 嚴重度 | 檔案 | 問題描述 |
|---|--------|------|----------|
| 1 | 🔴 高 | `device_service.go:105` | `telemetryRepo` 為 nil 時呼叫 `QueryLatest` 會 panic |
| 2 | 🔴 高 | `main.go:77-80` | PostgreSQL 離線時 nil Repo 傳入 Service 導致 panic |
| 3 | 🔴 高 | `device_service.go:87` | `Create` 後 `FindByID` 錯誤被 `_` 靜默忽略 |
| 4 | 🔴 高 | `device_service.go:152` | `Update` 後 `FindByID` 錯誤被 `_` 靜默忽略 |
| 5 | 🟡 中 | `scylla/client.go` | CQL keyspace 名稱用 `fmt.Sprintf` 拼接（需確認來源可信） |
| 6 | 🟡 中 | `model/alert_rule.go` | 缺少 `UpdatedAt` 欄位，無法追蹤規則修改時間 |
| 7 | 🟡 中 | `telemetry_service.go:101` | Service 層用 `fmt.Printf` 輸出錯誤，應改用 `log` |
| 8 | 🟡 中 | `scylla/telemetry_repo.go:99` | 排序使用 O(n²) bubble sort，應改用 `sort.Slice` |
| 9 | 🟡 中 | `dto/device.go:46` | `LatestTelemetry` 使用 `interface{}`，缺少型別安全 |
| 10 | 🟡 中 | `device_repo.go:90` | Cursor 解碼失敗時靜默忽略，應返回錯誤 |
| 11 | 🟡 中 | `config.go` + `main.go` | ScyllaDB 多節點配置未處理逗號分隔 hosts |
| 12 | 🟢 低 | `go.mod` | 所有依賴標記為 `indirect`，需執行 `go mod tidy` |
| 13 | 🟢 低 | `user_service.go:141` | SoftDelete FindByID 錯誤未 wrap |
| 14 | 🟢 低 | `alert_rule_service.go:138` | Delete FindByID 錯誤未 wrap |
| 15 | 🟢 低 | `device_service.go:160` | Delete FindByID 錯誤未 wrap |
| 16 | 🟢 低 | `middleware/trace.go` | 未優先讀取上游傳入的 `X-Request-ID` header |

---

## ✅ 正確設計的肯定

- **分層架構清晰**：Handler → Service → Repository，每層職責分明。
- **Interface + DI 設計**：Repository 都使用 Interface，便於測試 Mock。
- **Graceful Shutdown**：`main.go` 實作了標準的 Signal 監聽與資源釋放順序。
- **降級設計** (ScyllaDB)：ScyllaDB 離線時 Service 正確回傳 `ErrScyllaOffline`，Handler 對應回 503。
- **Cursor-based 分頁**：設計正確，利用 `(created_at, id)` 雙欄位確保 stable sort 排序。
- **告警觸發邏輯**：在 `BatchInsert` 後評估規則並寫入 alert event 的設計是正確的。
