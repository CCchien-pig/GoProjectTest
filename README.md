# 統一設備管理平台（Unified Device Management Platform）

## 架構說明

```
┌──────────────────────────────────────────────────────────────┐
│                     REST API (Go + Gin)                       │
└─────────────────────────┬────────────────────────────────────┘
          │               │                │
          ▼               ▼                ▼
   ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐
   │ PostgreSQL  │  │  ScyllaDB   │  │     KeyDB       │
   │ 本地 Docker │  │ 本地 Docker │  │  本地 Docker    │
   └─────────────┘  └─────────────┘  └─────────────────┘
```

## 資料庫配置

| DB | 位置 | 連線方式 | Local Port |
|----|------|----------|-----------|
| PostgreSQL | 本地 Docker | 直連 | localhost:5432 |
| KeyDB | 本地 Docker | 直連 | localhost:6379 |
| ScyllaDB | 本地 Docker | 直連 | localhost:9042 |

## 快速啟動

### 0. 設定環境變數
```bash
# 從範本複製環境變數檔
# Windows (PowerShell)
copy .env.dev.example .env.dev
# Linux / macOS
cp .env.dev.example .env.dev
```
然後用文字編輯器打開 `.env.dev`，將 `POSTGRES_PASSWORD` 改為你自己的密碼。

### 1. 啟動所有資料庫 (PostgreSQL + KeyDB + ScyllaDB)
```bash
make compose-up
```

### 2. 啟動 API Server
```bash
make run
```

### 3. 停止與清理
```bash
# 停止容器
make compose-down

# 停止容器並清空資料（慎用）
make compose-down-v
```

## 環境需求

- Go 1.24+
- Docker + Docker Compose

## 環境變數

詳見 [.env.dev.example](file:///c:/Projects/CC/Go/GoProjectTest/.env.dev.example) 了解所有可配置項目。實際的 `.env.dev` 已被 `.gitignore` 排除，不會進入版本控制。
