# 統一設備管理平台 — 一個月實作計畫（逐天版）

> **期間**：2026/06/15 (一) — 2026/07/14 (一)
> **參考來源**：[考題全文](file:///c:/Projects/CC/GoProjectTest/.docs/Unified%20Device%20Management%20Platform..md)
> **核心參考專案**：USCII (`c:/Projects/CC/USCII/golang/`)

---

## 全局注意事項（每一天都適用）

- **不要用全域變數**：所有 DB client 透過 DI 注入給 Service / Repository
- **統一回傳格式**：所有 API 使用 `pkg/response/` 的 `OK()` / `BadRequest()` / `NotFound()`
- **錯誤包裝**：所有 error 用 `fmt.Errorf("context: %w", err)` 向上傳遞
- **設定集中化**：連線字串、port、密碼全部從 `.env` 讀取，不寫死

---

## Week 1：基礎架構 + PostgreSQL CRUD (佔分 30%)

### Day 1 — 6/15 (一)：專案骨架與基礎設施

**目標**：`docker compose up` 能把三個 DB 全部跑起來

- [ ] 建立專案目錄結構：
  ```
  cmd/api/main.go
  internal/config/
  internal/handler/
  internal/service/
  internal/repository/
  internal/middleware/
  internal/scylla/
  internal/keydb/
  internal/model/
  pkg/response/
  migrations/
  .docker/
  ```
- [ ] 寫 `internal/config/config.go`（參考 USCII config.go）
  - `Config struct`：PostgreSQL DSN、ScyllaDB Hosts/Keyspace、KeyDB Addr、API Port
  - `Load()` 函式，用 `godotenv` 讀 `.env`
- [ ] 寫 `.env.dev`（開發用環境變數）
- [ ] 寫 `pkg/response/response.go`：統一 JSON 回傳格式
  ```json
  {"code": 200, "message": "ok", "data": {...}, "pagination": {...}}
  ```
- [ ] 寫 `internal/middleware/trace.go`（Request ID，參考 USCII trace.go）

### Day 2 — 6/16 (二)：Docker Compose + PostgreSQL Schema

**目標**：三個 DB 容器 + PostgreSQL Schema 建好

- [ ] 寫 `.docker/docker-compose.dev.yml`（參考 USCII docker-compose.dev.yml）
  - `postgres:17-alpine` + healthcheck
  - `scylladb/scylla:latest` + `--smp 1 --memory 512M` + healthcheck
  - `scylladb-init`（建立 keyspace）
  - `eqalpha/keydb:latest` + healthcheck
- [ ] 寫 PostgreSQL Migration 檔案（`migrations/`）：
  - `001_create_users.up.sql` — users 表（UUID PK、username、email、password_hash、role、is_active）
  - `002_create_devices.up.sql` — devices 表（UUID PK、device_code UNIQUE、name、device_type、location、metadata JSONB、owner_id FK、status）
  - `003_create_alert_rules.up.sql` — alert_rules 表（UUID PK、device_id FK ON DELETE CASCADE、metric_name、operator、threshold、severity、is_enabled）
  - `004_add_indexes.up.sql` — `pg_trgm` GIN 索引（`device_code`、`name`）、device_type / status / location 索引
- [ ] 驗證：`docker compose up` → 連進 psql 確認三張表都存在

### Day 3 — 6/17 (三)：Users CRUD + GORM 基礎

**目標**：User 的完整 CRUD API 可用 Postman 測試

- [ ] 寫 `internal/model/user.go`（GORM model，對應 users 表）
- [ ] 寫 `internal/repository/user_repo.go`（interface + 實作）
  - `Create()`、`FindByID()`、`Update()`、`SoftDelete()`（將 is_active 設為 false）
  - `FindByID` 需回傳該使用者擁有的設備數量（`SELECT count(*) FROM devices WHERE owner_id = ?`）
- [ ] 寫 `internal/service/user_service.go`
  - 密碼用 `bcrypt` 雜湊後存入
- [ ] 寫 `internal/handler/user_handler.go`
  - `POST /api/v1/users`
  - `GET /api/v1/users/:id`（含設備數量）
  - `PUT /api/v1/users/:id`
  - `DELETE /api/v1/users/:id`（軟刪除）
- [ ] 在 `cmd/api/main.go` 組裝 DI 鏈：`config → gorm.DB → UserRepo → UserService → UserHandler → gin.Router`

### Day 4 — 6/18 (四)：Devices CRUD + 分頁 + 搜尋

**目標**：Device 完整 CRUD，含 Cursor-based 分頁和 pg_trgm 模糊搜尋

- [ ] 寫 `internal/model/device.go`
- [ ] 寫 `internal/repository/device_repo.go`
  - `Create()`、`FindByID()`、`Update()`、`Delete()`（真刪除，CASCADE 會自動刪 alert_rules）
  - **`List()` — 實作 Cursor-based Pagination**：
    - 用 `WHERE (created_at, id) < (?, ?)` 取代 OFFSET
    - 支援 `device_type`、`status`、`location` 篩選
    - 支援 `?search=xxx` — 用 `pg_trgm` 的 `ILIKE '%xxx%'` 搜尋 device_code 和 name
    - 回傳 `next_cursor` 供前端下一頁使用
- [ ] 寫 `internal/service/device_service.go`
- [ ] 寫 `internal/handler/device_handler.go`
  - `POST /api/v1/devices`
  - `GET /api/v1/devices`（分頁 + 篩選 + 搜尋）
  - `GET /api/v1/devices/:id`（Week 2 會擴充加入最新遙測）
  - `PUT /api/v1/devices/:id`
  - `DELETE /api/v1/devices/:id`

### Day 5 — 6/19 (五)：Alert Rules CRUD + updated_at 自動更新

**目標**：告警規則 CRUD 完成，PostgreSQL 部分全部收工

- [ ] 寫 `internal/model/alert_rule.go`
- [ ] 寫 `internal/repository/alert_rule_repo.go`
  - `Create()`、`FindByDeviceID()`、`Update()`、`Delete()`
- [ ] 寫 `internal/service/alert_rule_service.go`
  - 驗證 `operator` 只能是 `gt, lt, gte, lte, eq`
  - 驗證 `severity` 只能是 `info, warning, critical`
- [ ] 寫 `internal/handler/alert_rule_handler.go`
  - `POST /api/v1/devices/:id/alert-rules`
  - `GET /api/v1/devices/:id/alert-rules`
  - `PUT /api/v1/alert-rules/:id`
  - `DELETE /api/v1/alert-rules/:id`
- [ ] 實作 `updated_at` 自動更新：
  - 方案 A：PostgreSQL Trigger
  - 方案 B：GORM 的 `BeforeUpdate` hook（更簡單，推薦）
- [ ] **里程碑驗收**：用 Postman 完整測試 users、devices、alert_rules 三組 CRUD

---

## Week 2：ScyllaDB 時序數據 (佔分 30%)

### Day 6 — 6/21 (日)：ScyllaDB 連線 + Schema

**目標**：ScyllaDB 連線成功，兩張表建好

- [ ] 寫 `internal/scylla/client.go`（參考 USCII scylla/client.go）
  - `NewClient()` + `Close()` + `EnsureSchema()`
- [ ] `EnsureSchema()` 建立兩張表：
  - `telemetry` — Partition Key: `(device_id, date)`，Clustering Key: `(recorded_at DESC, metric_name ASC)`，TTL 90 天
  - `alert_events` — Partition Key: `(device_id, month)`，Clustering Key: `(triggered_at DESC, rule_id ASC)`，TTL 365 天

> **重要**：Partition Key 是 `(device_id, date)` 而不是單純的 `device_id`！這是為了避免單一設備的資料量過大導致 hot partition。每天一個分區，查詢跨天時需要拆分多個分區查詢再合併。

- [ ] 在 `main.go` 加入 ScyllaDB client 的初始化和 DI 注入

### Day 7 — 6/22 (一)：遙測數據寫入 API（批次 + 告警觸發）

**目標**：`POST /telemetry` 能批次寫入，並自動觸發告警

- [ ] 寫 `internal/scylla/telemetry_repo.go`
  - `BatchInsert(deviceID, []TelemetryPoint)` — 使用 CQL Batch INSERT，上限 100 筆
  - 使用 Prepared Statement
- [ ] 寫 `internal/service/telemetry_service.go`
  - 接收批次遙測資料
  - **核心邏輯：寫入遙測後，自動比對 `alert_rules`**
    1. 從 PostgreSQL 查出該 device 的所有 enabled alert_rules
    2. 逐筆遙測數據比對 `metric_name` + `operator` + `threshold`
    3. 若觸發 → 寫入 ScyllaDB `alert_events` 表
- [ ] 寫 `internal/handler/telemetry_handler.go`
  - `POST /api/v1/devices/:id/telemetry` — 批次寫入

### Day 8 — 6/23 (二)：遙測數據查詢（跨日分區 + 刪除標記過濾）

**目標**：查詢 API 能正確處理跨天的時間範圍，並過濾已刪除設備的資料

- [ ] 在 `telemetry_repo.go` 加入：
  - `Query(deviceID, start, end, metricName)` — **跨日分區查詢**：
    1. 根據 `start` 和 `end` 計算涉及哪些 `date` 分區
    2. 對每個分區發出獨立查詢
    3. 合併結果，按 `recorded_at DESC` 排序
  - `QueryLatest(deviceID)` — 取最新一筆各 metric 的數據（只查今天和昨天兩個分區）
  - `DeleteByRange(deviceID, start, end)` — 範圍刪除
- [ ] **設備刪除標記過濾 (`is_deleted` flag)**：
  - 查詢遙測時，先從 PG 確認設備是否存在
  - 若設備已被刪除，回傳歷史資料但附帶 `is_deleted: true` 標記，告知前端這屬於歷史稽核資料
- [ ] 在 handler 加入：
  - `GET /api/v1/devices/:id/telemetry`（必須帶 `start` / `end` 參數）
  - `GET /api/v1/devices/:id/telemetry/latest`
  - `DELETE /api/v1/devices/:id/telemetry`

### Day 9 — 6/24 (三)：告警事件 CRUD + 擴充設備詳情

**目標**：告警事件完整 CRUD，設備詳情含最新遙測

- [ ] 在 `internal/scylla/alert_event_repo.go` 加入：
  - `Insert(alertEvent)` — 寫入告警事件
  - `QueryByDevice(deviceID, month, severity)` — 查詢告警事件，支援 severity 篩選
  - `Acknowledge(deviceID, month, triggeredAt, ruleID)` — 確認告警
- [ ] 在 handler 加入：
  - `POST /api/v1/devices/:id/alert-events`（也可由遙測寫入自動觸發）
  - `GET /api/v1/devices/:id/alert-events`
  - `PUT /api/v1/alert-events/:device_id/:month/:triggered_at/:rule_id/ack`
- [ ] **擴充 `GET /api/v1/devices/:id`**：回傳設備詳情時，從 ScyllaDB 查最新遙測一併回傳
- [ ] **里程碑驗收**：用 Postman 測試完整的遙測寫入 → 告警自動觸發 → 查詢告警事件流程

### Day 10 — 6/25 (四)：Buffer / 補進度

- [ ] 回顧 Week 1-2 所有 API，補齊遺漏的 edge case
- [ ] 確保所有 error 回傳都有正確的 HTTP status code 和統一格式
- [ ] 如果提前完成，開始預習 KeyDB 的 go-redis/v9 API

---

## Week 3：KeyDB 快取與即時狀態 (佔分 25%)

### Day 11 — 6/26 (五)：KeyDB 連線 (TLS) + Cache-Aside + 在線狀態

**目標**：設備詳情快取 + 在線狀態判斷可運作，並完成 TLS 連線設定

- [ ] 寫 `internal/keydb/client.go`（參考 USCII keydb/dial.go）
  - `NewClient()` 建立連線
- [ ] **KeyDB 啟用 TLS 連線**：
  - Docker Compose 中 KeyDB 啟用 TLS（掛載自簽憑證）
  - Go 端 `redis.Options` 設定 `TLSConfig`
  - 驗證：不帶 TLS 連線會被拒絕
- [ ] 實作 **Cache-Aside**（`device:{device_id}`，TTL 5 min）：
  - 修改 `GET /api/v1/devices/:id`：先查 KeyDB → miss 時查 PG → 寫回 KeyDB
  - 修改 `PUT /api/v1/devices/:id`：更新 PG 後 invalidate 對應 cache key
  - Cache 穿透防護：PG 找不到時，存空值 + TTL 30 秒
- [ ] 實作 **在線狀態**（`device:online:{device_id}`，TTL 3 min）：
  - 修改 `POST /telemetry`：每次遙測寫入時 `SET + EXPIRE`
  - `GET /api/v1/devices/:id/status` — 讀取在線狀態

### Day 12 — 6/27 (六)：Write-Through + 告警計數 (含定時校正) + 列表快取

**目標**：三種額外快取策略實作完成，包含背景定時校正機制

- [ ] 實作 **Write-Through**（`telemetry:latest:{device_id}`，TTL 30 sec）：
  - 修改 `POST /telemetry`：寫入 ScyllaDB 後，同步更新 KeyDB 最新遙測快取
  - `GET /devices/:id/telemetry/latest` 先查 KeyDB，miss 才查 ScyllaDB
- [ ] 實作 **告警計數與定時校正**（`alert:count:{device_id}:{severity}`，TTL 10 min）：
  - 告警事件寫入時，對對應 key 執行 `INCR`
  - **定時校正機制**：用簡單的 `time.Ticker` + goroutine 實作背景任務，定時從 ScyllaDB `alert_events` 重新 COUNT 該設備的告警數，並寫回 KeyDB，防止 `INCR` 偏差
  - `GET /devices/:id/status` 回傳時包含各 severity 的告警計數
- [ ] 實作 **設備列表快取**（`devices:list:{hash_of_query_params}`，TTL 2 min）：
  - `GET /devices` 查詢結果寫入 KeyDB
  - 任何設備的 CREATE / UPDATE / DELETE 操作後，刪除所有 `devices:list:*` key

### Day 13 — 6/28 (日)：Dashboard API + Pipeline

**目標**：Dashboard 一次從 KeyDB 取回所有摘要

- [ ] 實作 `GET /api/v1/dashboard/overview`：
  - 使用 **KeyDB Pipeline** 一次取回：
    - 設備總數（可從 PG 定時同步到 KeyDB）
    - 在線設備數（掃描 `device:online:*` 的 key 數量，或維護一個計數器）
    - 各 severity 告警總數
  - 全部從 KeyDB 讀取，不打 PG / ScyllaDB

### Day 14 — 6/29 (一)：Cache Stampede 防護 + Invalidate API

**目標**：管理端點實作與防止快取擊穿

- [ ] 實作 `POST /api/v1/cache/invalidate`：
  - 接受 key pattern，清除匹配的 cache（管理用途）
- [ ] 實作 **Cache Stampede 防護**：
  - 使用 `golang.org/x/sync/singleflight` 包
  - 當多個 request 同時 cache miss 時，只有一個 goroutine 去查 DB，其他等結果

### Day 15 — 6/30 (二)： buffer 可回顧或趕進度

---

## Week 4：跨 DB 一致性、測試、品質與文件 (佔分 15%)

### Day 16 — 7/1 (三)：跨 DB 刪除一致性 (Saga Pattern)

**目標**：設備刪除時三個 DB 協調一致，並實作軟刪除標記

- [ ] 實作**跨 DB 刪除一致性**（Saga Pattern）：
  - 設備刪除時依序執行：
    1. PostgreSQL：刪除設備 + alert_rules（CASCADE）
    2. KeyDB：清除 `device:{id}`、`telemetry:latest:{id}`、`device:online:{id}`、`alert:count:{id}:*`、invalidate list cache
    3. ScyllaDB：**不刪除**遙測歷史資料（保留供稽核），透過 PG 查詢時附加的 `is_deleted` flag 來處理歷史資料標記
  - 若任一步驟失敗：記錄 log + 回傳部分成功的狀態碼

### Day 17 — 7/2 (四)：降級處理 + Health Check

**目標**：某 DB 掛掉不全崩，實作健康檢查

- [ ] 實作**降級處理**：
  - KeyDB 掛掉 → bypass cache，直接查 PG / ScyllaDB（catch connection error → fallback）
  - ScyllaDB 掛掉 → 遙測寫入返回 `503 Service Unavailable`，設備 CRUD 仍正常
  - PostgreSQL 掛掉 → 所有寫入返回 `503`，讀取嘗試從 KeyDB 快取提供
- [ ] 實作 `GET /health`：
  ```json
  {
    "status": "degraded",
    "postgres": "healthy",
    "scylladb": "unhealthy",
    "keydb": "healthy"
  }
  ```

### Day 18 — 7/3 (五)：Buffer / 補進度

- [ ] 回顧 Week 3-4 所有快取與一致性邏輯，確保 TTL、invalidation、降級都正確
- [ ] 用 Postman 完整測試所有 API endpoint
- [ ] **里程碑驗收**：
  - 用 Postman 驗證五種 KeyDB 快取策略（Cache-Aside、Write-Through、列表快取、在線狀態、告警計數）各自的 hit / miss / invalidation 行為
  - 呼叫 `GET /health`，確認三個 DB 的狀態都正確回報
  - 手動把某個 DB 容器停掉（`docker stop`），確認 API 不全崩、降級邏輯生效後再把容器啟回來
  - 所有功能開發完成，進入收尾測試階段

### Day 19 — 7/4 (六)：單元測試 — Service 層

**目標**：核心業務邏輯有 Mock 測試覆蓋

- [ ] 定義 Repository Interface（如果之前沒做）
- [ ] 為以下核心邏輯寫單元測試（使用 Mock Repository）：
  - **告警規則比對邏輯**：不同 operator (gt/lt/gte/lte/eq) × threshold 的排列組合
  - **快取策略邏輯**：cache hit / miss / invalidation 的行為驗證
  - **Cursor-based 分頁邏輯**：cursor 解析、邊界條件

### Day 20 — 7/5 (日)：單元測試 — 補充 + 整合測試

- [ ] 補充 Service 層測試（降級邏輯、跨 DB 刪除的補償機制）
- [ ] 寫整合測試框架：
  - 使用 Docker Compose 啟動三個 DB
  - 跑完整的 CRUD flow：建立使用者 → 建立設備 → 設定告警規則 → 寫入遙測 → 觸發告警 → 查詢告警

### Day 21 — 7/6 (一)：壓力測試腳本 + 報告

**目標**：產出可量化的效能報告

- [ ] 寫壓力測試腳本（Go test / k6 / 或自寫腳本）：
  1. 批量建立 1,000 個設備
  2. 對每個設備寫入 1,000 筆遙測數據
  3. 混合讀寫負載（70% 讀 / 30% 寫）
- [ ] 執行壓力測試，記錄結果：
  - QPS（每秒請求數）
  - P50 / P95 / P99 延遲
  - 測試環境規格（CPU、RAM、Docker 配置）
- [ ] 撰寫壓力測試報告（含瓶頸分析）

### Day 22 - 7/7 (二)： golangci-lint + Graceful Shutdown + 結構化日誌

**目標**：程式碼品質達到生產標準

- [ ] 安裝並設定 `golangci-lint`，消滅所有 warning
- [ ] 在 `main.go` 實作 **Graceful Shutdown**：
  - 捕捉 `SIGTERM` / `SIGINT`
  - 依序關閉：HTTP Server → KeyDB → ScyllaDB → PostgreSQL
  - 每個關閉步驟印出 log
- [ ] 將所有 `fmt.Println` / `log.Println` 替換為**結構化日誌**：
  - 使用 `log/slog`（Go 標準庫）或 `zerolog`
  - 所有 log 都帶 request_id（從 middleware 注入的 trace ID）
- [ ] 寫 `Makefile`：
  ```makefile
  build:    go build -o bin/api ./cmd/api/
  test:     go test ./... -v -race -cover
  lint:     golangci-lint run
  run:      go run ./cmd/api/
  compose:  docker compose --env-file .env.dev -f .docker/docker-compose.dev.yml up -d
  ```

### Day 23~25 — 7/8~7/10 (三)~(五)：README + API 文件 + 簡報準備

- [ ] 撰寫 `README.md`：
  - 架構圖（可用 Mermaid）
  - 環境需求
  - 啟動步驟：`docker compose up` → `go run ./cmd/api/`
  - API 端點總覽
  - 測試執行方式
- [ ] **建立 API 文件**：放棄使用 swaggo 自動產生，改用 **Postman Collection**，將開發過程中寫好的 Postman requests 整理分類後匯出成 JSON 提供
- [ ] 確認交付物清單：

| 交付物 | 狀態 |
| :--- | :--- |
| 完整 Go 原始碼（可 `go build` 成功） | [ ] |
| `README.md`（架構說明 + 啟動步驟） | [ ] |
| `docker-compose.yml` | [ ] |
| SQL Migration 檔案（`migrations/`） | [ ] |
| CQL Schema（`EnsureSchema` 或獨立 `.cql` 檔） | [ ] |
| API 文件（Postman Collection） | [ ] |
| 壓力測試報告 | [ ] |
| `Makefile` | [ ] |

- [ ] 準備簡報講稿


### Day 26~28 — 7/11~7/13 (六)~(一)：全流程彩排 + 收尾

**目標**：模擬「架構師第一次拿到這個專案」的完整操作體驗，提前抓出任何整合問題

#### 全流程彩排步驟（依序執行，不跳過）

- [ ] **Step 1 — 全新環境啟動**：
  - 先把所有容器和 volume 清除：`docker compose down -v`
  - 重新執行：`docker compose up -d`
  - 確認三個 DB 的 healthcheck 全部變成 healthy（不能有任何 container 是 restarting）
- [ ] **Step 2 — Migration 驗證**：
  - 確認 PostgreSQL Migration 有自動執行（或手動跑 `make migrate`）
  - 連進 psql 確認 `users`、`devices`、`alert_rules` 三張表都存在且 schema 正確
  - 確認 ScyllaDB keyspace 和兩張 CQL 表（`telemetry`、`alert_events`）都建立成功
- [ ] **Step 3 — 編譯 + 啟動 API**：
  - 先執行 `go build ./cmd/api/` 確認編譯無誤（這是 Step 2 的 psql 之後才做的事，psql 不會幫你抓 Go 編譯錯誤）
  - 執行 `make run`（或 `go run ./cmd/api/`）
  - 確認 `main.go` 的啟動 log 依序印出（這是 GORM/gocql/go-redis 的連線，不是 psql）：
    - `PostgreSQL connected`（GORM `db.Ping()` 成功）
    - `ScyllaDB connected, EnsureSchema done`
    - `KeyDB connected`
    - `HTTP Server listening on :8080`
  - 若任何一行沒出現 → 表示環境變數或連線設定有問題，在這裡修，不要帶著問題進 Step 4
- [ ] **Step 4 — 呼叫 `GET /health`**：
  - 三個 DB 狀態全部顯示 `healthy`
- [ ] **Step 5 — 核心業務流程 Postman 完整跑一遍**：
  1. 建立使用者（`POST /users`）
  2. 建立設備（`POST /devices`，帶 owner_id）
  3. 設定告警規則（`POST /devices/:id/alert-rules`，設 temperature > 50 = warning）
  4. 寫入遙測批次（`POST /devices/:id/telemetry`，包含 temperature=55）
  5. 確認告警自動觸發（`GET /devices/:id/alert-events`，應看到一筆 warning）
  6. 查詢 KeyDB 在線狀態（`GET /devices/:id/status`）
  7. 查詢 Dashboard（`GET /dashboard/overview`）
  8. 查詢設備清單（`GET /devices`，驗證 Cursor-based 分頁）
  9. 搜尋設備（`GET /devices?search=SENSOR`，驗證 pg_trgm）
  10. 刪除設備（`DELETE /devices/:id`），再確認 KeyDB 相關 key 已清除
- [ ] **Step 6 — 降級驗證**：
  - 停掉 KeyDB：`docker stop <keydb_container>`
  - 確認 `GET /devices/:id` 仍能回傳資料（bypass cache 直查 PG）
  - 確認 `GET /health` 回報 KeyDB `unhealthy`
  - 把 KeyDB 重啟：`docker start <keydb_container>`，確認恢復正常
- [ ] **Step 7 — Graceful Shutdown 驗證**：
  - 對 API process 發送 `Ctrl+C`
  - 確認 log 依序印出：HTTP Server 關閉 → KeyDB 關閉 → ScyllaDB 關閉 → PostgreSQL 關閉
  - 確認程式正常退出（exit code 0）
- [ ] **Step 8 — 最終除錯**：修復彩排過程中發現的任何問題

### Day 29 — 7/14 (二)：專題報告

- [ ] 7/14 報告專題

---

## 各週驗收里程碑

| 週 | 驗收標準 |
| :--- | :--- |
| **Week 1 結束** | Postman 能完整測試 users / devices / alert_rules 三組 CRUD，Cursor-based 分頁和 pg_trgm 搜尋正常運作 |
| **Week 2 結束** | 遙測批次寫入 → 自動觸發告警 → 跨日查詢遙測 → 查詢告警事件，完整鏈路可跑通 |
| **Week 3 結束** | 五種 KeyDB 快取策略全部可驗證、Dashboard API 用 Pipeline 一次取值、Health Check 正確反映各 DB 狀態、某 DB 掛掉時 API 不全崩 |
| **Week 4 結束** | 測試覆蓋核心邏輯、壓力測試報告產出、Lint 零 warning、README 完整、可一鍵啟動整個系統 |

---

## USCII 參考檔案速查表

| 你要做的功能 | 參考 USCII 的檔案 |
| :--- | :--- |
| **main.go DI 組裝順序** | `cmd/api/main.go`（config → DB clients → Repos → Services → Handlers → Router → Server） |
| **PostgreSQL 連線（GORM）** | `cmd/api/main.go`（`gorm.Open` + `db.Ping` 的初始化寫法） |
| Config 集中管理 | `internal/config/config.go` |
| 通用 CRUD 路由 | `internal/handler/generic_crud.go` |
| GORM Repository | `internal/repository/business_repo.go` |
| ScyllaDB 連線 | `internal/scylla/client.go` |
| ScyllaDB CQL 操作 | `internal/scylla/email_repo.go` |
| KeyDB 連線 | `internal/keydb/dial.go` |
| KeyDB Stream/Queue | `internal/keydb/producer.go` |
| 統一回傳格式 | `pkg/response/` 目錄 |
| Trace ID Middleware | `internal/middleware/trace.go` |
| Docker Compose | `.docker/docker-compose.dev.yml` |
