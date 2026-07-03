# 實作計畫：修復 Code Review (W1_2 & RBAC) 提出之所有問題

本計畫旨在根據 `code_review_W1_2.md` 與 `RBACcode_review0703v1.md` 兩份 Code Review 文件，全面修復系統架構與業務邏輯上的缺失。

## User Review Required

> [!CAUTION]
> 針對 `users` 表格的 `role_id` 加上 `NOT NULL` 約束後，未來新增使用者**必須**指定合法的角色 ID。這代表系統不可存在「無角色」的使用者。若這與您的業務邏輯相符，請批准此計畫。

> [!WARNING]
> 為了讓資料庫的變更（如 `NOT NULL`、新增 `updated_at` 欄位以及 Seed data）生效，修復完成後需要再次執行 `make compose-down-v` 與 `make compose-up` 來重建資料庫。這將會清空目前本機的假資料。

## 提出的變更計畫

### 1. 資料庫變更 (Migrations & Seed)
- **[MODIFY]** `migrations/000001_create_users.up.sql`
  - 為 `users.role_id` 加上 `NOT NULL`。
  - 為 `roles` 與 `permissions` 補上 `updated_at` 欄位。
- **[MODIFY]** `migrations/000003_create_alert_rules.up.sql`
  - 為 `alert_rules` 補上 `updated_at` 欄位。
- **[NEW]** `migrations/000005_seed_rbac.up.sql`
  - 建立初始的 RBAC 種子資料（寫入預設的角色與權限）。

### 2. 資料模型 (Models & DTOs)
- **[MODIFY]** `internal/model/rbac.go`
  - 刪除多餘的 `RolePermission` struct。
- **[MODIFY]** `internal/model/alert_rule.go`
  - 新增 `UpdatedAt time.Time` 欄位。
  - 實作 GORM 的 `BeforeUpdate` Hook 來自動更新時間。
- **[MODIFY]** `internal/dto/device.go`
  - 將 `LatestTelemetry` 的型別由 `interface{}` 改為 `[]*model.TelemetryData`，增強型別安全。

### 3. 儲存庫層 (Repositories)
- **[MODIFY]** `internal/repository/user_repo.go`
  - 實作 `FindByIDs(ctx, ids []uuid.UUID)` 方法，支援透過 `IN` 語法批次查詢使用者，解決 N+1 問題。
- **[MODIFY]** `internal/repository/device_repo.go`
  - 移除多餘的 `Preload("Users")`。
  - 在 `List` 的 Cursor 解析失敗時，回傳 Error 而非靜默忽略。
  - 實作 `UpdateWithUsers` 方法，將設備更新與使用者關聯更新包裝在同一個 GORM `Transaction` 中。
- **[MODIFY]** `internal/scylla/telemetry_repo.go`
  - 將 `Query` 內效能低落的 Bubble Sort 改為標準函式庫的 `sort.Slice`。

### 4. 業務邏輯與介面層 (Services & Middleware)
- **[MODIFY]** `internal/service/device_service.go`
  - 修正 `FindByID` 在處理 `telemetryRepo` nil 時的 panic 風險。
  - 修正 `Create` 與 `Update` 中 `s.repo.FindByID` 錯誤被 `_` 靜默忽略的問題。
  - 呼叫 `userRepo.FindByIDs` 取代 for 迴圈查詢，解決 N+1 問題。
  - 改用新的 `UpdateWithUsers` 確保 `Update` 具備事務安全性。
  - 補充註解說明 `UserIDs` 的 PATCH 語意（nil/空陣列/值）。
  - 修復 `Delete` 中 `FindByID` 錯誤未被 wrap 的問題。
- **[MODIFY]** `internal/service/user_service.go` & `internal/service/alert_rule_service.go`
  - 修復 `SoftDelete` / `Delete` 錯誤回傳沒有使用 `fmt.Errorf` 進行 wrap 的問題。
- **[MODIFY]** `internal/service/telemetry_service.go`
  - 將 `fmt.Printf` 替換為 `log.Printf`。
- **[MODIFY]** `internal/middleware/trace.go`
  - 優先讀取 Client 傳入的 `X-Request-ID` Header，若無才產生新的 UUID。
- **[MODIFY]** `cmd/api/main.go` & `internal/config/config.go`
  - 在建立 Scylla Client 時使用 `strings.Split(cfg.ScyllaHosts, ",")` 來支援多節點。

### 5. 依賴管理
- 執行 `go mod tidy` 整理 `go.mod` 內被錯誤標記為 indirect 的套件。

## 驗證計畫
### Automated Tests
- 執行 `go test ./...` 確保單元測試全數通過（尤其是剛剛修正的方法）。
- 執行 `golangci-lint run` 確保無其他語法與風格錯誤。

### Manual Verification
- 停止並重建 DB Volumes (`make compose-down-v` & `make compose-up`)。
- 在終端機重啟 `make run`。
- 確認 Seed 資料已正確載入 DB。
