# 🎯 開發實作考題：統一設備管理平台 API（Unified Device Management Platform）

> **考試時間**：4 週（40 個工作天）  
> **難度**：中級  
> **技術棧**：Go + PostgreSQL + ScyllaDB + KeyDB  

---

## 📋 情境說明

公司正在建設一套 **統一設備管理平台**，用於管理跨廠區的 IoT 設備（感測器、控制器、閘道器等）。平台需要：

1. **PostgreSQL** — 儲存設備主檔（master data）與使用者帳號等結構化資料
2. **ScyllaDB** — 儲存設備回報的高頻遙測數據（telemetry），例如溫度、濕度、電壓等時序資料
3. **KeyDB** — 作為快取層與即時狀態存儲（設備在線狀態、近期告警計數、熱門查詢快取）

---

## 🏗️ 架構概覽

```
┌──────────────────────────────────────────────────────────────┐
│                     REST API (Go)                            │
│                                                              │
│  ┌─────────────┐   ┌──────────────┐   ┌─────────────────┐   │
│  │  Device CRUD │   │ Telemetry    │   │  Alert / Status │   │
│  │  User CRUD   │   │ Ingest+Query │   │  Cache Layer    │   │
│  └──────┬───────┘   └──────┬───────┘   └───────┬─────────┘   │
│         │                  │                   │             │
└─────────┼──────────────────┼───────────────────┼─────────────┘
          │                  │                   │
          ▼                  ▼                   ▼
   ┌─────────────┐   ┌─────────────┐   ┌─────────────────┐
   │ PostgreSQL  │   │  ScyllaDB   │   │     KeyDB       │
   └─────────────┘   └─────────────┘   └─────────────────┘
```

---

## 📝 考題要求

### Part 1：PostgreSQL CRUD（30%）

#### 1.1 資料模型設計

設計並建立以下資料表（至少），需包含適當的索引與約束：

```sql
-- 使用者表
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(100) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'viewer',  -- admin / operator / viewer
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 設備表
CREATE TABLE devices (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_code   VARCHAR(50)  UNIQUE NOT NULL,   -- 設備編號，如 SENSOR-TPE-001
    name          VARCHAR(200) NOT NULL,
    device_type   VARCHAR(50)  NOT NULL,           -- sensor / controller / gateway
    location      VARCHAR(200),                    -- 廠區/區域
    metadata      JSONB DEFAULT '{}',              -- 彈性欄位（韌體版本、IP、型號等）
    owner_id      UUID REFERENCES users(id),
    status        VARCHAR(20)  NOT NULL DEFAULT 'inactive', -- active / inactive / maintenance
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 告警規則表
CREATE TABLE alert_rules (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id     UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    metric_name   VARCHAR(100) NOT NULL,           -- e.g. temperature, voltage
    operator      VARCHAR(10)  NOT NULL,           -- gt, lt, gte, lte, eq
    threshold     DOUBLE PRECISION NOT NULL,
    severity      VARCHAR(20) NOT NULL DEFAULT 'warning', -- info / warning / critical
    is_enabled    BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

#### 1.2 API 端點

| Method | Endpoint | 說明 |
|--------|----------|------|
| `POST` | `/api/v1/users` | 建立使用者 |
| `GET` | `/api/v1/users/:id` | 取得使用者（含其擁有的設備數量） |
| `PUT` | `/api/v1/users/:id` | 更新使用者資訊 |
| `DELETE` | `/api/v1/users/:id` | 軟刪除使用者（`is_active = false`） |
| `POST` | `/api/v1/devices` | 建立設備 |
| `GET` | `/api/v1/devices` | 列表查詢（支援分頁、篩選 `device_type`、`status`、`location`） |
| `GET` | `/api/v1/devices/:id` | 取得設備詳細（含最新遙測資料 from ScyllaDB） |
| `PUT` | `/api/v1/devices/:id` | 更新設備資訊 |
| `DELETE` | `/api/v1/devices/:id` | 刪除設備（級聯刪除 alert_rules） |
| `POST` | `/api/v1/devices/:id/alert-rules` | 新增告警規則 |
| `GET` | `/api/v1/devices/:id/alert-rules` | 取得某設備的所有告警規則 |
| `PUT` | `/api/v1/alert-rules/:id` | 更新告警規則 |
| `DELETE` | `/api/v1/alert-rules/:id` | 刪除告警規則 |

#### 1.3 進階要求

- [ ] 分頁使用 **Cursor-based pagination**（基於 `created_at + id`），不使用 OFFSET
- [ ] `GET /devices` 支援 `?search=xxx` 全文模糊搜尋 device_code 和 name（使用 `pg_trgm`）
- [ ] 所有寫入操作需記錄 `updated_at` 自動更新（trigger 或 application-level）

---

### Part 2：ScyllaDB 時序數據 CRUD（30%）

#### 2.1 資料模型設計

在 ScyllaDB 中設計遙測數據表（需考慮 partition 大小與查詢效能）：

```cql
CREATE KEYSPACE IF NOT EXISTS device_platform
WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': 3};

-- 遙測數據表（按設備 + 天分區）
CREATE TABLE device_platform.telemetry (
    device_id    UUID,
    date         DATE,           -- partition by day，避免 partition 過大
    recorded_at  TIMESTAMP,
    metric_name  TEXT,           -- temperature, humidity, voltage, etc.
    value        DOUBLE,
    unit         TEXT,           -- °C, %, V, etc.
    tags         MAP<TEXT, TEXT>, -- 額外標籤
    PRIMARY KEY ((device_id, date), recorded_at, metric_name)
) WITH CLUSTERING ORDER BY (recorded_at DESC, metric_name ASC)
  AND default_time_to_live = 7776000;  -- 90 天自動過期

-- 告警事件表
CREATE TABLE device_platform.alert_events (
    device_id    UUID,
    month        TEXT,           -- '2026-06' 按月分區
    triggered_at TIMESTAMP,
    rule_id      UUID,
    metric_name  TEXT,
    metric_value DOUBLE,
    threshold    DOUBLE,
    severity     TEXT,
    acknowledged BOOLEAN,
    PRIMARY KEY ((device_id, month), triggered_at, rule_id)
) WITH CLUSTERING ORDER BY (triggered_at DESC, rule_id ASC);
```

#### 2.2 API 端點

| Method | Endpoint | 說明 |
|--------|----------|------|
| `POST` | `/api/v1/devices/:id/telemetry` | 批次寫入遙測數據（支援一次寫入多筆） |
| `GET` | `/api/v1/devices/:id/telemetry` | 查詢遙測資料（必須帶 `start` / `end` 時間範圍） |
| `GET` | `/api/v1/devices/:id/telemetry/latest` | 取得設備最新一筆各 metric 的數據 |
| `DELETE` | `/api/v1/devices/:id/telemetry` | 刪除指定時間範圍內的遙測資料 |
| `POST` | `/api/v1/devices/:id/alert-events` | 寫入告警事件（可由遙測寫入時自動觸發） |
| `GET` | `/api/v1/devices/:id/alert-events` | 查詢告警事件（支援 `severity` 篩選） |
| `PUT` | `/api/v1/alert-events/:device_id/:month/:triggered_at/:rule_id/ack` | 確認（acknowledge）告警 |

#### 2.3 進階要求

- [ ] 遙測寫入 API 需**批次處理**（Batch INSERT），一次最多接受 100 筆
- [ ] 查詢 API 需支援**跨日分區**查詢（自動拆分多個 partition 查詢並合併結果）
- [ ] 遙測寫入時，自動比對 PostgreSQL 中的 `alert_rules`，若觸發則寫入 `alert_events`
- [ ] 實作 **TTL 管理**：遙測數據 90 天、告警事件 365 天

---

### Part 3：KeyDB 快取與即時狀態（25%）

#### 3.1 快取策略實作

| 快取項目 | Key 格式 | TTL | 策略 |
|----------|----------|-----|------|
| 設備詳情快取 | `device:{device_id}` | 5 min | **Cache-Aside**：讀取時先查 KeyDB，miss 則查 PG 後寫入 |
| 設備列表快取 | `devices:list:{hash_of_query_params}` | 2 min | 寫入/更新/刪除設備時 **invalidate** 所有 list cache |
| 設備最新遙測 | `telemetry:latest:{device_id}` | 30 sec | 每次遙測寫入時 **Write-Through** 同步更新 |
| 設備在線狀態 | `device:online:{device_id}` | 3 min | 遙測寫入時 `SET + EXPIRE`，過期即為離線 |
| 告警計數 | `alert:count:{device_id}:{severity}` | 10 min | 告警事件寫入時 `INCR`，定時從 ScyllaDB 校正 |

#### 3.2 API 端點

| Method | Endpoint | 說明 |
|--------|----------|------|
| `GET` | `/api/v1/devices/:id/status` | 取得設備即時狀態（在線/離線 + 最新遙測 + 告警計數） |
| `GET` | `/api/v1/dashboard/overview` | 儀表板摘要（設備總數、在線數、各嚴重等級告警數）— **全部從 KeyDB 讀取** |
| `POST` | `/api/v1/cache/invalidate` | 手動清除指定 key pattern 的快取（管理用途） |

#### 3.3 進階要求

- [ ] 使用 **TLS 連線** 連接 KeyDB（對應 on-prem 的 TLS cluster 設定）
- [ ] 實作 **Cache Stampede 防護**：使用 singleflight 或分散式鎖避免大量 cache miss 同時回源
- [ ] `GET /dashboard/overview` 需使用 **Pipeline** 一次取回多個 key，減少 round-trip
- [ ] 設備刪除時，需同步清理 KeyDB 中所有相關 key（`device:*`, `telemetry:latest:*`, `alert:count:*`）

---

### Part 4：整合與品質要求（15%）

#### 4.1 跨資料庫一致性

- [ ] 設備刪除時，需在一個邏輯事務中完成：
  1. PostgreSQL：刪除設備及關聯的 alert_rules
  2. KeyDB：清除所有相關快取 key
  3. ScyllaDB：**不刪除** 遙測歷史資料（保留供稽核），但標記設備已刪除
- [ ] 實作 **Saga Pattern** 或 **補償機制**：若任一步驟失敗，確保最終一致性

#### 4.2 錯誤處理與韌性

- [ ] 任一資料庫不可用時，API 不得整體崩潰：
  - KeyDB 掛掉 → 快取功能降級（bypass cache），直接查 PG/ScyllaDB
  - ScyllaDB 掛掉 → 遙測寫入返回 503，設備 CRUD 仍可運作
  - PostgreSQL 掛掉 → 所有寫入返回 503，讀取從 KeyDB 快取嘗試提供
- [ ] 實作 **Health Check** 端點 `GET /health`，回報三個資料庫的連線狀態

#### 4.3 測試要求

- [ ] **單元測試**：核心業務邏輯（告警規則比對、快取策略、分頁邏輯）
- [ ] **整合測試**：使用 Docker Compose 或 Testcontainers 啟動三個資料庫，測試完整 CRUD flow
- [ ] **壓力測試腳本**：
  - 批量建立 1,000 個設備
  - 對每個設備寫入 1,000 筆遙測數據
  - 混合讀寫負載測試（70% 讀 / 30% 寫）
  - 報告：QPS、P50/P95/P99 延遲

#### 4.4 程式碼品質

- [ ] 專案結構遵循 Go 標準佈局（`cmd/`, `internal/`, `pkg/`）
- [ ] 使用 **介面抽象** 資料庫存取層（Repository pattern）
- [ ] 統一 JSON 回應格式（含 `code`, `message`, `data`, `pagination`）
- [ ] 使用 **結構化日誌**（如 `slog` 或 `zerolog`）
- [ ] 提供 `Makefile` 或 `Taskfile` 整合常用指令（`build`, `test`, `lint`, `run`）

---

## 📦 交付物

| 項目 | 說明 |
|------|------|
| 原始碼 | 完整 Go 專案，可 `go build` 成功 |
| `README.md` | 含架構說明、環境需求、啟動步驟、API 文件連結 |
| `docker-compose.yml` | 本地開發用（含 PG + ScyllaDB + KeyDB） |
| SQL Migration | PostgreSQL schema migration 檔案（可用 `golang-migrate` 或 `goose`） |
| CQL Schema | ScyllaDB keyspace + table 建立腳本 |
| API 文件 | Swagger/OpenAPI 或 Postman Collection |
| 壓力測試報告 | 含測試環境規格、測試結果、瓶頸分析 |

---

## 📊 評分標準

| 分類 | 配分 | 評分重點 |
|------|------|----------|
| **PostgreSQL CRUD** | 30% | Schema 設計合理性、分頁實作、pg_trgm 搜尋、交易處理 |
| **ScyllaDB CRUD** | 30% | Partition 設計、跨日查詢、批次寫入、TTL 管理、告警自動觸發 |
| **KeyDB 快取** | 25% | 快取策略正確性、TLS 連線、Cache stampede 防護、Pipeline 使用 |
| **整合與品質** | 15% | 跨 DB 一致性、降級處理、測試覆蓋率、程式碼結構 |

---

## ⏰ 時程建議

| 天數 | 建議進度 |
|------|----------|
| Day 1-2 | 環境建置（Docker Compose）、專案骨架、DB schema 建立 |
| Day 3-4 | PostgreSQL CRUD 完成（Users + Devices + Alert Rules） |
| Day 5-6 | ScyllaDB 遙測寫入/查詢、告警自動觸發 |
| Day 7-8 | KeyDB 快取策略實作、Dashboard API |
| Day 9 | 跨 DB 一致性、降級處理、錯誤處理 |
| Day 10 | 測試、壓力測試、文件整理、最終交付 |

---

> [!TIP]
> 本考題的核心考察點不只是 CRUD 本身，而是考生如何在 **三種不同特性的資料庫** 之間做出合理的架構選擇：  
> - **PostgreSQL** = 強一致、關聯式、事務型  
> - **ScyllaDB** = 高吞吐、分區鍵設計、最終一致  
> - **KeyDB** = 低延遲、揮發性、快取策略  
>