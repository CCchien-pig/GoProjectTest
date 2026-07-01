# SA 策略文件 — 統一設備管理平台

> 這份文件的目的：以 Solution Architect 的視角記錄本專題的設計決策、技術選型理由、概念理解筆記，以及 AI 協作使用紀錄。

---

## 補充一：資料庫選型設計理由（DB Design Rationale）

### 設計核心原則

> **一個好的 SA 不是選最新的技術，而是選「最適合這筆資料特性」的技術。**

本專題使用三種資料庫，各司其職，分別對應不同的資料存取模式（Access Pattern）：

---

### PostgreSQL — 結構化主資料（Source of Truth）

**選用理由：**
- 設備主檔（Devices）、使用者帳號（Users）、告警規則（Alert Rules）屬於**強一致性（Strong Consistency）** 的業務資料，任何時刻都必須是正確的，不允許髒讀（Dirty Read）。
- 需要 ACID 交易保證：例如新增設備時，同時建立預設告警規則，兩個操作必須是原子的（All or Nothing）。
- 支援複雜的 JOIN 查詢、關聯完整性（Foreign Key Constraint）、全文搜尋（`pg_trgm`）。

**不適合用 PostgreSQL 的資料：**
| 資料類型 | 原因 |
|---------|------|
| 每秒上千筆的遙測數據 | 高頻 INSERT 會造成大量 WAL log，磁碟 I/O 撐不住 |
| Session Token / 在線狀態 | 沒有原生 TTL 機制；需要定期清理過期資料，維護成本高 |
| 毫秒級的即時狀態更新 | 每次 UPDATE 都需要鎖定行（Row-level Lock），高並發下瓶頸明顯 |

**關鍵設計決策：**
- 使用 **Cursor-based 分頁**（而非 OFFSET）：`OFFSET N` 在資料量大時需要掃描前 N 筆資料，O(N) 複雜度；Cursor 分頁以 `WHERE id > last_id` 實現，永遠是 O(1)。
- 使用 **`pg_trgm` GIN 索引**：支援模糊搜尋 `device_code LIKE '%abc%'`，避免 Full Table Scan。
- 使用 **`gen_random_uuid()`** 作為主鍵：分散式環境下避免 Auto-Increment 的 ID 衝突問題。

---

### ScyllaDB — 高頻時序資料（Write-Heavy Time-Series）

**選用理由：**
- IoT 設備每 5 秒上報一筆遙測數據（Temperature、Humidity、Voltage...），一台設備一天約 17,280 筆。若有 1,000 台設備，每天寫入量超過 **1,700 萬筆**。
- ScyllaDB（Cassandra 架構）的 LSM-Tree 儲存引擎專為**大量寫入**設計，INSERT 永遠是 Append，不需要隨機讀取，寫入吞吐量可達 PostgreSQL 的 10x 以上。
- 寬行（Wide Row）資料模型天然符合時序資料的存取模式：以 `(device_id, date)` 為 Partition Key，時間戳為 Clustering Key，查詢「某設備某日的所有數據」只需要一次 Partition 讀取。

**Partition Key 設計（最關鍵的決策）：**

```
Partition Key : (device_id, date)       -- 決定資料在哪個節點
Clustering Key: (recorded_at DESC, metric_name ASC)  -- 決定同一 Partition 內的排序
```

**為什麼要把 `date` 放入 Partition Key？**
如果只用 `device_id`，一台設備所有歷史資料都在同一個 Partition，會造成 **Hot Partition**（單一節點過熱），且 Partition 大小無上限地增長。加入 `date` 後，每天的資料自動分散到不同 Partition，上限為每個 Partition 最多一天的資料量。

**TTL 自動清理：**
- 遙測資料設 `TTL 90天`，超過自動刪除，不需要維護清理 Job。
- 告警事件設 `TTL 365天`，保留一年的歷史紀錄。

**不適合用 ScyllaDB 的資料：**
| 資料類型 | 原因 |
|---------|------|
| 需要跨 Partition 的聚合查詢（如 COUNT, SUM） | Cassandra 架構不支援跨 Partition 的高效 Aggregation |
| 需要任意欄位更新 | 底層是 Append-Only，UPDATE 實際上是寫入一筆新紀錄（Tombstone 問題）|
| 需要複雜 JOIN | 不支援，關聯資料需要由應用層自己處理 |

---

### KeyDB — 快取層與即時狀態（Cache & Real-time State）

**選用理由：**
- KeyDB 完全相容 Redis 協定，底層使用多執行緒架構，在單機環境下效能優於 Redis。
- 所有資料存在記憶體（In-Memory），讀寫延遲 < 1ms，適合需要高速存取的場景。
- 原生支援 TTL（`EXPIRE` 指令），資料過期自動清除，不需要額外維護。

**本專題 KeyDB 存放的資料類型（面試準備）：**

| 使用場景 | Key 範例 | 資料結構 | TTL | 理由 |
|---------|----------|---------|-----|------|
| JWT Session Token 快取 | `session:{user_id}` | String | 與 JWT 同壽命 | 快速驗證 Token 有效性，避免每次解密 JWT |
| JWT 黑名單（已登出 Token） | `jwt:blacklist:{jti}` | String | 與 Token 過期時間相同 | 登出後讓未過期的 Token 失效 |
| 設備即時在線狀態 | `device:{id}:status` | String/Hash | 30秒（心跳更新） | 設備狀態每秒可能更新，PG 無法承受此頻率 |
| 設備列表 API 快取 | `cache:devices:p{cursor}` | String (JSON) | 60秒 | 列表查詢耗時，結果快取供多人共用 |
| Dashboard 聚合數據快取 | `cache:dashboard:overview` | Hash | 30秒 | 跨三個 DB 的聚合計算，快取結果避免重複計算 |
| API 速率限制計數器 | `ratelimit:{ip}:{endpoint}` | String | 60秒（滑動視窗） | `INCR` 原子操作天然適合計數，PG 需要加鎖 |
| 分散式鎖 | `lock:device:{id}:write` | String | 5秒 | `SETNX` 原子操作防止同一設備並發寫入衝突 |
| 用戶 RBAC 權限快取 | `rbac:{user_id}:roles` | Set | 5分鐘 | 每個 API 都需查權限，快取比打 PG 快 10x |

> **面試技巧**：當被問「KeyDB 裡放什麼」並追問「還有呢？」時，按照以上表格從「認證相關 → 狀態相關 → 快取相關 → 保護機制（限流/鎖）→ 計算優化（RBAC）」的順序展開，每一類都有清楚的**選用理由**。

---

## 補充二：前後端切分、Vue、RabbitMQ 概念理解

### 前後端職責切分原則

**核心判斷標準：「這個邏輯如果放前端，使用者能不能繞過？」**

| 邏輯類型 | 應放位置 | 理由 |
|---------|---------|------|
| 權限檢查（RBAC） | **後端** | 前端 JS 可被任意修改，權限驗證必須在後端做最終把關 |
| 資料驗證（必填、格式） | **前後端都要** | 前端提供即時 UX 回饋；後端做最終保障，防止惡意請求 |
| 分頁邏輯 | **後端** | 前端拿全量資料再分頁會造成大量資料傳輸；分頁必須在 DB 層執行 |
| UI 顯示格式（日期格式化、顏色判斷） | **前端** | 純展示邏輯，跟後端無關，後端回傳原始值即可 |
| 業務規則計算（告警閾值判斷） | **後端** | 業務邏輯不能依賴前端，後端才是唯一可信任的計算來源 |
| 國際化（i18n）文字 | **前端** | 語言偏好屬於 UI 層的職責 |

**API 設計哲學（RESTful）：**
- 後端的 API 應該是**語意明確、無狀態（Stateless）**：每次請求攜帶足夠的資訊，Server 不依賴前一次請求的 Session 狀態。
- Response 永遠包含一致的格式：`{ code, message, data, pagination }`，前端只需要對應一個格式。
- 錯誤回應使用 HTTP Status Code 語意化：`400 Bad Request`（用戶輸入錯誤）、`401 Unauthorized`（未登入）、`403 Forbidden`（沒有權限）、`404 Not Found`、`500 Internal Error`（後端 Bug）。

---

### Vue 概念速覽（面試準備）

**核心概念（非實作，理解思想即可）：**

| 概念 | 說明 | 與後端的連結 |
|------|------|-------------|
| **MVVM 架構** | Model（資料）↔ ViewModel（Vue 實例）↔ View（DOM）。資料驅動視圖，不直接操作 DOM | 後端提供的 JSON 就是 Model |
| **雙向綁定（v-model）** | 表單輸入與 JS 變數自動同步，不需要手動監聽 `onChange` 事件 | 後端只需要定義清楚接收的欄位格式 |
| **組件化（Component）** | UI 拆分成可複用的獨立單元，各自管理自己的狀態與模板 | 每個組件通常對應一個 API 端點的資料消費 |
| **單頁應用（SPA）** | 只載入一次 HTML，後續頁面切換透過 JS 操作，不整頁刷新 | 後端不再負責 Server-side Rendering，只提供純 JSON API |
| **前後端分離** | 前端（Vue，跑在使用者瀏覽器）透過 HTTP 打 API 取得資料，完全解耦 | 後端工程師不需要了解 UI 細節，只需定義好 API Contract |
| **Pinia / Vuex（狀態管理）** | 跨組件的共享狀態（如：登入用戶資訊、全域告警），集中管理避免 Prop 傳遞地獄 | 對應後端的 Session / Token 驗證機制 |

**SA 視角的前後端分離優勢：**
1. 獨立部署：前端部署到 CDN，後端部署到 K8s，互不影響。
2. 獨立擴展：API 流量大時只需擴後端，靜態資源由 CDN 承擔。
3. 跨平台共用 API：同一套後端 API 可以服務 Vue Web、React Native 手機 App、Python 腳本。

---

### RabbitMQ 概念速覽（面試準備）

**解決什麼問題：**
同步 HTTP 呼叫（A 服務直接打 B 服務的 API）有個致命問題：若 B 服務掛掉，A 服務也會失敗。MQ 引入**非同步解耦**，A 只需把訊息丟進 Queue，B 自己從 Queue 消費，兩者完全解耦。

**核心架構：**
```
Producer（訊息發送方）
    ↓ 發送到
Exchange（路由中心，決定訊息去哪個 Queue）
    ↓ 根據 Routing Key 分發到
Queue（訊息暫存）
    ↑ 消費
Consumer（訊息接收方）
```

**三種 Exchange 類型（必背）：**

| Exchange 類型 | 路由規則 | 使用場景 |
|--------------|---------|---------|
| **Direct** | Routing Key 完全匹配 | 告警事件只發給負責告警的 Consumer |
| **Fanout** | 廣播給所有綁定的 Queue | 系統公告，所有服務都要收到 |
| **Topic** | Routing Key 支援萬用字元（`*` `#`） | `device.alert.critical` 只讓高優先 Consumer 處理，`device.#` 讓日誌服務記錄所有設備事件 |

**RabbitMQ vs KeyDB Pub/Sub 的差異（重點）：**

| 面向 | RabbitMQ | KeyDB Pub/Sub |
|------|----------|--------------|
| 訊息持久化 | ✅ 支援（訊息寫入磁碟，Consumer 掛掉後重啟仍可消費） | ❌ 不支援（Consumer 離線期間的訊息直接丟棄） |
| 訊息確認（ACK） | ✅ Consumer 需要 ACK，否則訊息重新入隊 | ❌ 無 ACK 機制，Fire and Forget |
| 複雜路由 | ✅ Exchange + Routing Key 靈活路由 | ❌ 只能 Channel 訂閱，無路由邏輯 |
| 適用場景 | 金融交易、訂單處理、重要業務事件 | 即時通知、在線狀態廣播、輕量推播 |

> **本專題的連結**：如果要在 UDM 平台中加入告警通知（Email / LINE Notify），適合使用 RabbitMQ，讓告警觸發（Producer）與通知發送（Consumer）解耦，避免通知失敗影響主業務流程。

---

## 補充三：AI 使用關鍵字紀錄

> 架構師說他會看 AI 的關鍵字。以下記錄每次使用 AI 的「提問方式」，展現以 SA 視角精準下關鍵字的能力。

### 記錄格式

```
### [日期] — [任務目標]
**關鍵字 / Prompt**：
> [輸入給 AI 的完整關鍵字]

**獲得的關鍵洞見**：
- [從 AI 的回答中提煉出的重要觀念]

**後續驗證**：
- [你如何驗證 AI 給的答案是正確的]
```

---

### [2026/06/26] — 確立 ScyllaDB Partition Key 設計

**關鍵字 / Prompt**：
> "ScyllaDB partition key design for IoT telemetry data, avoid hot partition, time-series data access pattern"

**獲得的關鍵洞見**：
- 單一 `device_id` 為 Partition Key 會造成 Hot Partition，因為同一設備的所有資料都在同一個節點。
- 加入 `date` 作為複合 Partition Key，讓每天的資料分散到不同 Partition，有效控制單一 Partition 的大小。
- Clustering Key 的降序（DESC）設計讓「查最新資料」不需要全量掃描。

**後續驗證**：
- 查閱 Apache Cassandra 官方文件關於 Data Modeling Best Practices，確認 Bucketing by Time 的做法。

---

### [2026/06/26] — 確認 Cursor-based 分頁 vs OFFSET 分頁的效能差異

**關鍵字 / Prompt**：
> "PostgreSQL cursor-based pagination vs offset pagination performance, large dataset, keyset pagination"

**獲得的關鍵洞見**：
- `OFFSET N` 需要資料庫從頭掃描 N 筆再跳過，時間複雜度 O(N)，資料量越大越慢。
- Keyset Pagination（`WHERE id > :last_id LIMIT :size`）時間複雜度永遠是 O(1)，利用索引直接定位。
- 缺點：無法直接跳到第 N 頁，只能前後翻頁，但 IoT Dashboard 的使用情境不需要跳頁。

**後續驗證**：
- 用 `EXPLAIN ANALYZE` 實測兩種方式在 100 萬筆資料下的執行計畫差異。

---

### [2026/06/27] — GCP e2-micro 記憶體限制下的 Docker 資源配置

**關鍵字 / Prompt**：
> "docker compose memory limit OOM killer e2-micro 1GB RAM PostgreSQL KeyDB resource constraints"

**獲得的關鍵洞見**：
- `deploy.resources.limits.memory` 設定容器硬上限，超過觸發 OOM Kill 而非讓整個 VM 崩潰。
- PostgreSQL `shared_buffers` 建議為可用記憶體的 25%，在 384M limit 下設 64MB 是保守但安全的做法。
- `max_connections=30` 配合 Go 的 pgxpool，實際連線數遠小於 PG 理論上限，不需要設太高。

**後續驗證**：
- 實際啟動 docker compose 並用 `docker stats` 監控各容器的實際記憶體使用量。

---

### [2026/06/29] — Docker Compose 環境變數安全性

**關鍵字 / Prompt**：
> "docker compose environment variable security best practice, avoid hardcoded credentials, .env file gitignore, required variable validation"

**獲得的關鍵洞見**：
- `${VAR:-default}` 語法：若 VAR 未設定，使用 default 值（**不安全**，密碼不應有預設值）。
- `${VAR:?error message}` 語法：若 VAR 未設定，docker compose 直接中止並顯示錯誤（**安全**）。
- `.env.dev.example` 進 Git 作為格式範本；`.env.dev` 透過 `.gitignore` 排除，密碼不入版控。

**後續驗證**：
- 刪除本地 `.env.dev`，執行 `make compose-up`，確認 docker compose 正確報錯而非以空密碼啟動。

---

*此文件持續更新，每次使用 AI 解決重要技術問題後，於當日補充一筆紀錄。*
