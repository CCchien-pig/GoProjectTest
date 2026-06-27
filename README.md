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
   │ GCP e2-micro│  │  本地 Docker │  │  GCP e2-micro   │
   │ SSH Tunnel  │  │             │  │  SSH Tunnel     │
   └─────────────┘  └─────────────┘  └─────────────────┘
```

## 資料庫配置

| DB | 位置 | 連線方式 | 本地 Port |
|----|------|----------|-----------|
| PostgreSQL | GCP e2-micro | SSH Tunnel | localhost:5433 |
| KeyDB | GCP e2-micro | SSH Tunnel | localhost:6380 |
| ScyllaDB | 本地 Docker | 直連 | localhost:9042 |

## 快速啟動

### 1. GCP 機器：啟動 PostgreSQL + KeyDB
```bash
# 上傳並啟動 GCP 端 compose
scp .docker/docker-compose.gcp.yml <user>@<GCP_IP>:~/
ssh <user>@<GCP_IP> "docker compose up -d"
```

### 2. 本地：建立 SSH Tunnel
```bash
# 一鍵啟動兩條 tunnel
bash .docker/ssh/start_tunnels.sh start
# 確認狀態
bash .docker/ssh/start_tunnels.sh status
```

### 3. 本地：啟動 ScyllaDB
```bash
make compose-up
```

### 4. 啟動 API Server
```bash
make run
```

## 環境需求

- Go 1.24+
- Docker + Docker Compose
- SSH Key（連 GCP 用）

## 環境變數

複製 `.env.dev` 並填入正確的密碼和 GCP IP 後的 SSH Tunnel Port 設定：
```env
DATABASE_URL=postgres://udm:pass@localhost:5433/udm?sslmode=disable
KEYDB_ADDR=localhost:6380
SCYLLA_HOSTS=localhost:9042
```
