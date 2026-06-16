# 對話重點記錄

> 對話期間：2026/06/04 — 2026/06/16
> 涵蓋主題：職涯分析 → 專案架構驗證 → Go 學習計畫 → 專題規劃

---

## 一、職涯背景與轉調分析

### 現況
- 從 OP 課（SETOP / PresaleCMSServer）調到 CC 課（3CX → USCII）
- 身份：剛轉職的軟體工程師，第一年

### OP 課「損失」的客觀評估
- **真正的損失只有 PresaleCMSServer**：Clean Architecture + DDD + CQRS、20 個測試專案、SonarCloud CI 門禁、AWS ECS 部署
- **SETOP 的損失被高估了**：Pipeline 用本機 MSBuild 路徑、整合測試是 `Assert.Pass()` 空殼、Docker 只有 1 個 service
- OP 課的 7 次 hotfix 限額 + 薄弱的測試安全網，對新人其實不利

### CC 課的價值判定
- **核心前提**：9 月確實能接手 USCII（已確認 ✅）
- USCII 是所有專案中工程品質最高的：Go 微服務、三入口架構、PostgreSQL + ScyllaDB + KeyDB、PASETO、RBAC、完整 Docker Compose、golangci-lint
- Go 後端工程師的市場供給遠小於 C# 工程師

### 結論
> 損失是真實的但不致命。CC 課值不值得待，100% 取決於 USCII 能否落實。（已確認）

---

## 二、專案架構比對（程式碼驗證）

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

## 三、外部評論驗證

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

## 四、架構師的指導方向

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

## 五、學習策略

### AI 使用原則
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

## 六、專題與 USCII 的對照表

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

## 七、Go 三個月學習計畫（原始版，已因專題調整）

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

## 八、一個月專題計畫摘要

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

## 九、關鍵結論

1. **從 OP 調到 CC 的損失是真實的但不致命**，PresaleCMSServer 的觀摩機會確實可惜，但 USCII + 專題的成長路徑不比留在 OP 差
2. **USCII 規模小不是扣分項**，面試時「能完整講清楚一個系統怎麼運作」比「待過的系統有多大」重要
3. **架構師出專題 + review 是最佳學習方式**，比考試、比線上課程都有價值
4. **這份專題的難度橫跨 Junior 到 Senior**，但有 USCII 參考 + AI 工具 + 一個月時間，可以完成
5. **扎實完成這份專題學到的，比接需求單寫一年 CRUD 多得多**，因為你在學「為什麼」而不只是「怎麼做」
