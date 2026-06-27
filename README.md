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
   │ GCP 白名單直連 │  │             │  │ GCP 白名單直連    │
   └─────────────┘  └─────────────┘  └─────────────────┘
```

## 資料庫配置

| DB | 位置 | 連線方式 | Port |
|----|------|----------|-----------|
| PostgreSQL | GCP e2-micro | 直連 (GCP 白名單) | <GCP_IP>:5432 |
| KeyDB | GCP e2-micro | 直連 (GCP 白名單) | <GCP_IP>:6379 |
| ScyllaDB | 本地 Docker | 直連 | localhost:9042 |

## 快速啟動

### 1. GCP 設定防火牆白名單
進入 GCP Console -> VPC 網路 -> 防火牆，新增規則允許你目前的外部 IP 存取 tcp:5432,6379。

### 2. GCP 機器：啟動 PostgreSQL + KeyDB
```bash
# 上傳並啟動 GCP 端 compose
scp .docker/docker-compose.gcp.yml <user>@<GCP_IP>:~/
ssh <user>@<GCP_IP> "docker compose up -d"
```

### 3. 本地：啟動 ScyllaDB
```bash
make compose-up
```

### 4. 啟動 API Server
修改 `.env.dev` 填入正確的 `<GCP_EXTERNAL_IP>` 與密碼，然後啟動：
```bash
make run
```

## 環境需求

- Go 1.24+
- Docker + Docker Compose

## 環境變數

複製 `.env.dev` 並填入正確的密碼和 GCP 外部 IP：
```env
DATABASE_URL=postgres://udm:pass@<GCP_EXTERNAL_IP>:5432/udm?sslmode=disable
KEYDB_ADDR=<GCP_EXTERNAL_IP>:6379
SCYLLA_HOSTS=localhost:9042
```
