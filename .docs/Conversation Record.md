# 對話紀錄更新規則

閱讀Conversation Record.md後把新對話重點整理並更新Conversation Record.md，記得修改對話期間。

# 對話重點記錄

> 對話期間：2026/06/04 — 2026/07/02
> 涵蓋主題：專案架構驗證 → Go 學習計畫 → 專題規劃 → 基礎設施建置與資安防護 → 架構師考核標準對齊 → Week 1 & 2 TDD 實作與降級架構落地

---

## 一、Go 三個月學習計畫（原始版，已因專題調整）

| 月 | 主題 |
| :--- | :--- |
| 第 1 個月 | Go 語法、測試、並行、Gin、GORM、分層架構 |
| 第 2 個月 | Docker、Compose、Queue/Worker/Scheduler、日誌、Graceful Shutdown |
| 第 3 個月 | 認證/RBAC、資料匯入、外部服務整合、整合驗收 |

### 調整後的策略

- 原計畫從「學習路線」降級為「知識清單」
- 專題變成主線，遇到不懂的概念回頭翻對應週的學習重點
- 專題結束後，用清單對照哪些沒學到，再補

### 目前進度（截至 6/16）

- 已完成：Learn Go with Tests — concurrency、select（含 `ConfigurableRacer` + timeout 測試及context
- 同步完成：RecordingTransfer Go 生產工具（Transfer + Cleanup 兩個 Windows Service）

---

## 二、一個月專題計畫摘要

> 完整逐天版本請見 [MonthSchedule.md](file:///c:/Projects/CC/GoProjectTest/.docs/MonthSchedule.md)

| 週 | 重點 | 驗收標準 |
| :--- | :--- | :--- |
| Week 1 | 專案骨架 + PostgreSQL 三張表 CRUD + Cursor 分頁 + pg_trgm 搜尋 | Postman 完整測試三組 CRUD |
| Week 2 | ScyllaDB 兩張表 + 批次寫入 + 跨日分區 + 遙測→告警自動觸發 | 完整鏈路可跑通 |
| Week 3 | KeyDB 五種快取策略 + Dashboard Pipeline + Stampede + 降級 + Health | 所有功能開發完成 |
| Week 4 | 單元測試 + 整合測試 + 壓力測試報告 + Lint + 文件 + 簡報 | 可一鍵啟動整個系統 |

### 三個容易遺漏的基礎設施

1. `internal/config/config.go` — 集中化設定管理（Day 1）
2. `pkg/response/response.go` — 統一 API 回傳格式（Day 1）
3. `internal/middleware/trace.go` — Request ID 追蹤（Day 1）

---

---

## 三、專案實作歷程 (2026/06/26 起)

### Day 11~14 (6/26~6/29) — 專案起步與基礎設施建置

**1. 架構決策 (改採 All-in-One 本地 Docker)**

- 因應公司網域防火牆限制，放棄 GCP 部署方案。
- **PostgreSQL, KeyDB, ScyllaDB**：全部整合至 `.docker/docker-compose.dev.yml` 在本地 Docker 運行。
- **優勢**：100% 離線開發、無網路相依性，且透過 Git 同步，可確保公司與家裡開發環境完全一致。

**2. 開發環境與指令簡化**

- 刪除 `docker-compose.gcp.yml` 與 SSH Tunnel 腳本。
- `Makefile` 提供單一啟動指令 `make compose-up` 即可一次帶起所有依賴庫。

**3. 資安防護與環境變數管理 (6/29 修正)**

- **移除硬編碼密碼**：禁止在 `docker-compose.dev.yml` 中使用 fallback 預設密碼，改用 `:?` 語法 (`${POSTGRES_PASSWORD:?錯誤訊息}`) 強制檢查，避免以空密碼或不安全的預設密碼啟動服務。
- **.env 管理規範**：建立 `.env.dev.example` 作為 Git 版控的公開範本，包含密碼的 `.env.dev` 則透過 `.gitignore` (`!.env.dev.example`) 嚴格排除。

**4. 專案骨架初始化**

- 完成 Go Module 初始化 (`GoProject/udm`)。
- 建立符合 USCII 規範的分層目錄結構 (`cmd/api`, `internal/handler`, `internal/service`, `internal/repository` 等)。
- 建立各層級的 Placeholder 原始碼檔案，準備進入後續業務邏輯開發。

---

## 四、架構師考核標準對齊 (2026/07/01)

### 考核標準轉向：從「實作成果」到「SA 觀念與過程」

收到架構師最新指示，考核重點不僅在程式碼寫不寫得出來，更在於**為什麼這樣設計**以及**如何善用工具**：

1. **SA 思維的關鍵字能力**：能下對關鍵字讓 AI 處理 80% 的繁瑣工作。
2. **設計觀念的辯證**：必須清楚三個 DB 的特性與適用場景（能列舉多種 KeyDB 使用案例並回答「還有呢？」）。
3. **架構視野**：了解 API 規劃哲學、前後端分離界線，以及 Vue 和 RabbitMQ 的核心概念。

### 應對策略：建立 SA_Strategy.md

針對上述要求，於 `.docs/SA_Strategy.md` 建立專屬文件作為應考的「大腦」，內容涵蓋：

- **DB 選型設計理由**：詳列 PostgreSQL, ScyllaDB (含 Partition Key 設計防禦), KeyDB (列出 8 種應用場景與 TTL) 的適用性。
- **前後端切分與其他概念**：以 SA 視角釐清邏輯切分標準（權限/分頁必放後端），並速覽 Vue 與 RabbitMQ 核心概念。
- **AI 使用關鍵字紀錄**：建立 Prompt 紀錄表，留下精確的技術英文 Prompt、獲得的洞見與後續驗證方式，作為「善用 AI 工具」的直接證據。

---

## 五、Week 1 & 2 實作里程碑與 TDD 驗證 (2026/07/02)

完成了 Week 1 與 Week 2 計畫的所有內容，全程遵循「先寫測試、再寫代碼、測試通過、完善優化」的輕量級 TDD 開發流程，代碼全部成功編譯且單元與整合測試 100% 通過。

### 1. PostgreSQL 設備主檔 CRUD 與進階搜尋 (Week 1)

- **基礎設施**：載入 `.env.dev` 動態設定、 Trace ID 中介層與 RESTful 統一 JSON 回傳格式。
- **Cursor 游標分頁**：設備查詢 API 實作以 `(created_at, id)` 雙欄位游標分頁以保證大數據下的 O(1) 效能。
- **模糊搜尋與自動更新**：透過 PostgreSQL GIN 索引與 pg_trgm 運算子實作高效模糊搜尋，並透過 GORM `BeforeUpdate` hook 自動更新修改時間。

### 2. ScyllaDB 遙測數據與告警觸發 (Week 2)

- **時序資料庫 Schema**：建立 `telemetry` (TTL 90 天) 與 `alert_events` (TTL 365 天) 資料表，並封裝 ScyllaDB Client。
- **告警評估與觸發**：批次遙測上傳時，自動拉取 GORM 設備的告警規則比對閥值，若觸發則自動寫入 ScyllaDB `alert_events` 表。
- **日分區跨天查詢**：考量 Scan Penalty，在 Repository 實作中自動拆分查詢日期，個別發起 partition-key 查詢並在記憶體中進行高性能排序。
- **降級與高可用 (Degraded Mode)**：在 `main.go` 啟動階段對資料庫進行連線偵測與容錯處理。當 ScyllaDB 斷線時，設備與使用者 CRUD 依然完全可用，遙測 API 則能優雅回傳 HTTP `503 Service Unavailable` 說明系統降級。
- **整合布線與測試**：實作 `routes.go` 與 `cmd/api/main.go` 完成所有模組的依賴注入與 Graceful Shutdown，整合測試與編譯成功。

### 3. 高優先級 Bug 修復與本地環境驗證 (2026/07/02)

- **修正 Nil Panic**：修復 `device_service.go` 中因 ScyllaDB 降級導致的 `telemetryRepo` 為 `nil` 的 panic 問題。
- **修正 PostgreSQL 容錯邏輯**：PostgreSQL 為核心資料庫不可降級，將離線警告改為 `log.Fatalf` 立即中止，避免引發後續層級的 Panic。
- **ScyllaDB 架構升級**：為支援 ScyllaDB 6.0 的 Tablet Replication 新特性，將 `docker-compose.dev.yml` 及 `client.go` 的 Keyspace 策略由 `SimpleStrategy` 全面升級為符合生產環境標準的 `NetworkTopologyStrategy`。
- **環境自動化驗證**：引導並使用 `winget` 安裝 Windows `make` 工具，成功背景啟動 `docker compose` 與 `go run ./cmd/api/`，驗證三個資料庫 (PostgreSQL, ScyllaDB, KeyDB) 全數完美連線。

---

## 六、Week 3 & 4 實作里程碑與 Code Review 修正 (2026/07/08)

完成了 Week 3 與 Week 4 計畫的所有內容，主要聚焦於 KeyDB 進階快取機制、分散式事務 (Saga Pattern)、RBAC 權限控管與系統壓力測試，並通過了嚴格的 Code Review。

### 1. KeyDB 進階快取與高併發優化 (Week 3)

- **多維度快取策略**：實作了 Device Cache-Aside (5m)、Telemetry Write-Through (30s) 以及 Alert Counts 獨立快取 (10m)。
- **Dashboard 雙層快取架構**：放棄原定的 Ticker 背景輪詢，全面改用「懶加載 (Lazy-Loading) + TTL」策略，大幅降低系統 Idle Overhead。
- **O(1) 在線狀態統計**：將原本 O(N) 的 `SCAN` 改為透過 KeyDB Set 維護心跳，並用 `SCard` 以 O(1) 複雜度取得在線設備數。
- **Redis Pipeline 實作**：在 `dashboard_service.go` 中實作 Pipeline，一次網路往返 (Roundtrip) 即可取得多個指標數據。

### 2. Saga Pattern 與 RBAC 權限控管 (Week 4)

- **Saga 分散式事務**：在設備刪除流程中實作 Saga Pattern。先刪除 PostgreSQL 資料，若後續 KeyDB 快取清除失敗，則回傳 HTTP 207 Multi-Status (ErrCacheCleanupFailed) 警示。
- **RBAC 管理員驗證**：為 `/cache/invalidate` 端點加上 `role == "admin"` 的權限守衛，防止未授權使用者引發 Cache Avalanche。

### 3. 壓力測試與 Code Review 修正

- **壓力測試防護**：修正了 `stress_test.go` 中全域 `rand` 的 Mutex 鎖競爭問題 (改用 thread-local rand)，並加入 P95/P99 百分位數的安全邊界檢查。
- **Code Review 13 項修正**：完成了包含效能優化、錯誤處理 (Saga 錯誤吞併修正)、DTO 分層 (DashboardOverview) 以及 API 回應格式統一 (`response.OK`) 等 13 項嚴格的 Code Review 指標。
