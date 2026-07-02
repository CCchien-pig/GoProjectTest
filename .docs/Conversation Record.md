# 對話紀錄更新規則

閱讀Conversation Record.md後把新對話重點整理並更新Conversation Record.md，記得修改對話期間。

# 對話重點記錄

> 對話期間：2026/06/04 — 2026/07/02
> 涵蓋主題：職涯分析 → 專案架構驗證 → Go 學習計畫 → 專題規劃 → 基礎設施建置與資安防護 → 架構師考核標準對齊 → Week 1 & 2 TDD 實作與降級架構落地

---

## 一、專案架構比對（程式碼驗證）

### 3CX（Legacy）

- 單體式架構，Python 2583 行單檔 `app.py`
- .NET Framework MVC/WebAPI、Autofac、多資料庫
- 同步阻塞模式（`.Result`）、硬編碼 secrets
- 部署方式：手動放到 IIS

### USCII（現代架構）

- Go 1.25 微服務，三入口：`cmd/api`、`cmd/scheduler`、`cmd/worker`
- Gin + GORM + PostgreSQL + ScyllaDB + KeyDB
- handler → service → repository 分層
- 白名單欄位防護、PASETO v4 認證、Audit 中間件
- Docker Compose 完整開發環境

### SETOP（大型 Legacy）

- 49 個子目錄、2280 行的 `ViewModelConvertFacade.cs`
- 整合測試是空殼（`Assert.Pass()`）
- Pipeline 用本機 MSBuild 路徑

### PresaleCMSServer（最佳工程實踐）

- Clean Architecture + DDD + CQRS
- 20 個測試專案、SonarCloud CI 門禁
- AWS ECS 部署

---

## 二、外部評論驗證

### 評論者的五個核心論點（全部驗證為「正確」）

1. 「100% 無縫遷移」不成立 — 業務知識遷移率 ~80-90%，但 Go 有 2-4 週學習曲線
2. 9 月接手是最關鍵前提 — 程式碼端有支持證據，但組織決策需自己確認
3. 報告語氣過度樂觀 — 「極佳」「完全」「黃金」是敘事包裝
4. 技能遷移方向可信 — 全文最有價值的部分，有程式碼支撐
5. 取決於能否碰核心開發 — 已在寫 Go concurrency + RecordingTransfer

### 雙方遺漏的事實

- 3CX 的 `ThreeCXCallLogService.cs` 比想像中寫得更好（有 Fail-Safe 降級設計）
- USCII 的 CTI 和 3CX 的 CTI 功能範圍有交集但不完全重疊

---

## 三、架構師的指導方向

### 專題形式

- 原定 Go 考試 → 改為**架構師出專題，一個月完成**
- 題目：「統一設備管理平台 API」（Go + PostgreSQL + ScyllaDB + KeyDB）
- 架構師原話：「理論上大家有 Codex 那就是半天完成、半天驗證，再花三天學習」

### 架構師的職涯建議

1. 常看 developer-roadmap（nilbuild/developer-roadmap）
2. 後端學完 → 懂架構、雲地端
3. 學前端 → 懂得前後端分工的分水嶺
4. 一天一天年年累積，十年之後就會差很多

### 「前後端分水嶺」的含義

- 不是要你去寫前端，而是要你知道**什麼邏輯該放後端、什麼該放前端**
- 例：權限檢查必須在後端做（前端可被繞過）、分頁必須在後端做（大量資料前端會爆）、UI 顯示邏輯放前端做
- 懂得劃這條線的人，才能設計出好的 API

### 「雲地端」的含義

- 你寫的程式碼要能在地端跑（Docker Compose），也要能不改一行 code 就搬到雲端
- 透過 `config.go` + 環境變數 + DI 解耦連線邏輯，就能做到這件事

---

## 四、學習策略

### AI 使用原則
>
> 卡超過 30 分鐘的語法問題，可以問 AI。
> 架構決策和業務邏輯，必須自己想。

- **自己動手**：Partition Key 設計、Cache 策略選擇、Graceful Shutdown 順序、Service 層 if-else 邏輯、測試案例設計
- **AI 輔助**：CQL 語法、go-redis API 用法、Docker Compose YAML、golangci-lint 設定、Makefile 語法

### 測試策略

- **Repository 層**：不 Mock，用 Docker 起真實 DB 做整合測試（驗證 SQL 對不對）
- **Service 層**：用 Mock 替換 Repository，做純粹的單元測試（驗證業務邏輯對不對）
- **關鍵**：沒有 Interface + DI，就無法在測試中把真實 Repo 抽換成 Mock

### 職涯路線建議

- **前 2-3 年**：深耕後端 + 架構（Go 微服務、分散式系統、多 DB、K8s）
- **同時**：用 AI 當工具處理前端需求（不需要深學，但要懂 HTTP 和狀態管理基本觀念）
- **Full Stack 是結果，不是路線**：先把後端打到夠深，前端自然會在需要的時候補上來

---

## 五、專題與 USCII 的對照表

| 專題需求 | USCII 最像的模組 | 參考檔案 |
| :--- | :--- | :--- |
| PostgreSQL 設備主檔 CRUD | 門市/人員主檔 CRUD | `generic_crud.go`、`business_repo.go` |
| PostgreSQL 使用者帳號 | 使用者 + RBAC | `user_handler.go`、`auth_handler.go`、`user_repo.go` |
| ScyllaDB 高頻遙測數據 | 郵件 metadata 儲存 | `scylla/client.go`、`scylla/email_repo.go` |
| KeyDB 快取 + 即時狀態 | Token 快取 + Stream 佇列 | `keydb/dial.go`、`keydb/producer.go`、`keydb/consumer.go` |
| Config 集中管理 | 環境變數管理 | `internal/config/config.go` |
| 統一 API 回傳格式 | Response 包 | `pkg/response/` |
| Trace ID Middleware | 請求追蹤 | `internal/middleware/trace.go` |

### ScyllaDB 的 Partition Key 對照

| USCII（郵件） | 專題（遙測） |
| :--- | :--- |
| Partition Key: `user_id` | Partition Key: `(device_id, date)` |
| Clustering Key: `received_at DESC` | Clustering Key: `(recorded_at DESC, metric_name ASC)` |

---

## 六、Go 三個月學習計畫（原始版，已因專題調整）

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

- 已完成：Learn Go with Tests — concurrency、select（含 `ConfigurableRacer` + timeout 測試）
- 進行中：Learn Go with Tests — context
- 同步完成：RecordingTransfer Go 生產工具（Transfer + Cleanup 兩個 Windows Service）

---

## 七、一個月專題計畫摘要

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

## 八、關鍵結論

1. **從 OP 調到 CC 的損失是真實的但不致命**，PresaleCMSServer 的觀摩機會確實可惜，但 USCII + 專題的成長路徑不比留在 OP 差
2. **USCII 規模小不是扣分項**，面試時「能完整講清楚一個系統怎麼運作」比「待過的系統有多大」重要
3. **架構師出專題 + review 是最佳學習方式**，比考試、比線上課程都有價值
4. **這份專題的難度橫跨 Junior 到 Senior**，但有 USCII 參考 + AI 工具 + 一個月時間，可以完成
5. **扎實完成這份專題學到的，比接需求單寫一年 CRUD 多得多**，因為你在學「為什麼」而不只是「怎麼做」

---

## 九、專案實作歷程 (2026/06/26 起)

### Day 11~14 (6/26~6/29) — 專案起步與基礎設施建置

因原計畫 Day 1~10 尚未動工，進行趕工並確立本地開發架構與資安規範。

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

- 完成 Go Module 初始化 (`github.com/your-name/udm`)。
- 建立符合 USCII 規範的分層目錄結構 (`cmd/api`, `internal/handler`, `internal/service`, `internal/repository` 等)。
- 建立各層級的 Placeholder 原始碼檔案，準備進入後續業務邏輯開發。

---

## 十一、架構師考核標準對齊 (2026/07/01)

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

## 十二、Week 1 & 2 實作里程碑與 TDD 驗證 (2026/07/02)

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
