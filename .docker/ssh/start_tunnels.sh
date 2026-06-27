#!/bin/bash
# .docker/ssh/start_tunnels.sh
# SSH Tunnel 一鍵啟動腳本
# 建立本地端口到 GCP e2-micro 上 PostgreSQL 和 KeyDB 的加密隧道
#
# ── 使用前設定 ─────────────────────────────────────────────────
#  1. 將此腳本的 GCP_USER / GCP_IP / SSH_KEY 填入正確值
#  2. 或透過環境變數傳入：
#       GCP_USER=ubuntu GCP_IP=34.x.x.x ./start_tunnels.sh
#
# ── 端口對應 ───────────────────────────────────────────────────
#  本地 5433 → GCP localhost:5432 (PostgreSQL)
#  本地 6380 → GCP localhost:6379 (KeyDB)
#
# ── .env.dev 對應設定 ──────────────────────────────────────────
#  DATABASE_URL=postgres://udm:pass@localhost:5433/udm?sslmode=disable
#  KEYDB_ADDR=localhost:6380
#
# ────────────────────────────────────────────────────────────────

set -e

GCP_USER="${GCP_USER:-ubuntu}"
GCP_IP="${GCP_IP:?請設定 GCP_IP 環境變數}"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/gcp_key}"
LOCAL_PG_PORT="${LOCAL_PG_PORT:-5433}"
LOCAL_KEYDB_PORT="${LOCAL_KEYDB_PORT:-6380}"

PID_FILE="/tmp/udm_ssh_tunnels.pid"

start_tunnels() {
    echo "=== UDM SSH Tunnel 啟動 ==="
    echo "GCP:  ${GCP_USER}@${GCP_IP}"
    echo "SSH Key: ${SSH_KEY}"

    # 檢查是否已有 tunnel 在跑
    if [ -f "$PID_FILE" ]; then
        echo "[警告] 偵測到已有 tunnel PID 檔，嘗試先停止..."
        stop_tunnels
    fi

    # 啟動 PostgreSQL tunnel
    ssh -f -N \
        -L "${LOCAL_PG_PORT}:localhost:5432" \
        -i "${SSH_KEY}" \
        -o StrictHostKeyChecking=no \
        -o ServerAliveInterval=60 \
        -o ServerAliveCountMax=3 \
        -o ExitOnForwardFailure=yes \
        "${GCP_USER}@${GCP_IP}"
    PG_PID=$!
    echo "[OK] PostgreSQL tunnel: localhost:${LOCAL_PG_PORT} -> GCP:5432 (PID: $PG_PID)"

    # 啟動 KeyDB tunnel
    ssh -f -N \
        -L "${LOCAL_KEYDB_PORT}:localhost:6379" \
        -i "${SSH_KEY}" \
        -o StrictHostKeyChecking=no \
        -o ServerAliveInterval=60 \
        -o ServerAliveCountMax=3 \
        -o ExitOnForwardFailure=yes \
        "${GCP_USER}@${GCP_IP}"
    KD_PID=$!
    echo "[OK] KeyDB tunnel: localhost:${LOCAL_KEYDB_PORT} -> GCP:6379 (PID: $KD_PID)"

    # 儲存 PID
    echo "${PG_PID} ${KD_PID}" > "$PID_FILE"
    echo ""
    echo "Tunnels 已啟動，PID 儲存於: ${PID_FILE}"
    echo "使用 './start_tunnels.sh stop' 停止"
}

stop_tunnels() {
    if [ -f "$PID_FILE" ]; then
        read -r PG_PID KD_PID < "$PID_FILE"
        kill "${PG_PID}" 2>/dev/null && echo "[OK] PostgreSQL tunnel 已停止 (PID: $PG_PID)" || true
        kill "${KD_PID}" 2>/dev/null && echo "[OK] KeyDB tunnel 已停止 (PID: $KD_PID)" || true
        rm -f "$PID_FILE"
    else
        echo "找不到 PID 檔，嘗試用 pkill 清除..."
        pkill -f "ssh.*5433.*localhost.*5432" 2>/dev/null || true
        pkill -f "ssh.*6380.*localhost.*6379" 2>/dev/null || true
        echo "清除完成"
    fi
}

status_tunnels() {
    echo "=== SSH Tunnel 狀態 ==="
    if [ -f "$PID_FILE" ]; then
        read -r PG_PID KD_PID < "$PID_FILE"
        if kill -0 "${PG_PID}" 2>/dev/null; then
            echo "[RUNNING] PostgreSQL tunnel (PID: $PG_PID)"
        else
            echo "[STOPPED] PostgreSQL tunnel"
        fi
        if kill -0 "${KD_PID}" 2>/dev/null; then
            echo "[RUNNING] KeyDB tunnel (PID: $KD_PID)"
        else
            echo "[STOPPED] KeyDB tunnel"
        fi
    else
        echo "[STOPPED] 無 tunnel 運行中"
    fi
    echo ""
    echo "連線測試："
    nc -z localhost "${LOCAL_PG_PORT}" 2>/dev/null && echo "  PostgreSQL (localhost:${LOCAL_PG_PORT}): OK" || echo "  PostgreSQL (localhost:${LOCAL_PG_PORT}): FAIL"
    nc -z localhost "${LOCAL_KEYDB_PORT}" 2>/dev/null && echo "  KeyDB (localhost:${LOCAL_KEYDB_PORT}): OK" || echo "  KeyDB (localhost:${LOCAL_KEYDB_PORT}): FAIL"
}

case "${1:-start}" in
    start)  start_tunnels ;;
    stop)   stop_tunnels ;;
    status) status_tunnels ;;
    restart)
        stop_tunnels
        sleep 1
        start_tunnels
        ;;
    *)
        echo "用法: $0 [start|stop|status|restart]"
        exit 1
        ;;
esac
