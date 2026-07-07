# API 參考文件 (API Reference)

本文件提供了 **統一設備管理平台 (UDM Platform)** API 的詳細規格、請求與回應範例。

---

## 1. 系統健康檢查

### 1.1 `GET /health`
取得伺服器以及 PostgreSQL、ScyllaDB、KeyDB 連線狀態。

- **回應範例 (200 OK - 全部正常)**:
  ```json
  {
    "status": "healthy",
    "postgres": "healthy",
    "scylla": "healthy",
    "keydb": "healthy"
  }
  ```
- **回應範例 (200 OK - 快取降級模式)**:
  ```json
  {
    "status": "degraded",
    "postgres": "healthy",
    "scylla": "healthy",
    "keydb": "unhealthy"
  }
  ```

---

## 2. 使用者管理 (Users)

### 2.1 `POST /api/v1/users`
建立新使用者。

- **請求 Body**:
  ```json
  {
    "username": "operator01",
    "email": "op01@company.com",
    "password": "securepassword",
    "role": "operator"
  }
  ```
- **回應範例 (200 OK)**:
  ```json
  {
    "code": 200,
    "message": "ok",
    "data": {
      "id": "e9dbf074-b52b-47e1-bfa0-79841f3d82d4",
      "username": "operator01",
      "email": "op01@company.com",
      "role": "operator",
      "is_active": true,
      "created_at": "2026-07-02T18:00:00Z"
    }
  }
  ```

### 2.2 `GET /api/v1/users/:id`
取得單一使用者詳細資料 (含擁有的設備數量)。

- **回應範例 (200 OK)**:
  ```json
  {
    "code": 200,
    "message": "ok",
    "data": {
      "id": "e9dbf074-b52b-47e1-bfa0-79841f3d82d4",
      "username": "operator01",
      "email": "op01@company.com",
      "role": "operator",
      "is_active": true,
      "device_count": 5,
      "created_at": "2026-07-02T18:00:00Z"
    }
  }
  ```

---

## 3. 設備管理 (Devices)

### 3.1 `POST /api/v1/devices`
建立設備。

- **請求 Body**:
  ```json
  {
    "device_code": "SENSOR-TPE-001",
    "name": "溫度感測器 #1",
    "device_type": "sensor",
    "location": "Taipei-A區",
    "metadata": {
      "firmware": "v1.2.0",
      "ip": "192.168.1.50"
    },
    "user_ids": ["e9dbf074-b52b-47e1-bfa0-79841f3d82d4"]
  }
  ```
- **回應範例 (200 OK)**:
  ```json
  {
    "code": 200,
    "message": "ok",
    "data": {
      "id": "d0e1c2b3-a4f5-5678-abcd-1234567890ef",
      "device_code": "SENSOR-TPE-001",
      "name": "溫度感測器 #1",
      "device_type": "sensor",
      "location": "Taipei-A區",
      "status": "inactive",
      "metadata": {
        "firmware": "v1.2.0",
        "ip": "192.168.1.50"
      },
      "users": [
        {
          "id": "e9dbf074-b52b-47e1-bfa0-79841f3d82d4",
          "username": "operator01",
          "email": "op01@company.com",
          "role": "operator"
        }
      ]
    }
  }
  ```

### 3.2 `GET /api/v1/devices`
分頁查詢設備清單，支援多種篩選與全文檢索 (`?search=SENSOR-TPE`)。

- **查詢參數**:
  - `cursor`: 前一頁回傳的 next_cursor
  - `limit`: 每頁數量 (預設 10)
  - `device_type`: sensor / controller / gateway
  - `status`: active / inactive / maintenance
  - `location`: 廠區名稱
  - `search`: 模糊搜尋 device_code 與 name
- **回應範例 (200 OK)**:
  ```json
  {
    "code": 200,
    "message": "ok",
    "data": [
      {
        "id": "d0e1c2b3-a4f5-5678-abcd-1234567890ef",
        "device_code": "SENSOR-TPE-001",
        "name": "溫度感測器 #1",
        "device_type": "sensor",
        "location": "Taipei-A區",
        "status": "active"
      }
    ],
    "pagination": {
      "next_cursor": "eyJjcmVhdGVkX2F0IjoiMjAyNi0wNy0wMlQxODowMDowMFoiLCJpZCI6ImQwZTFjMmIzLWE0ZjUtNTY3OC1hYmNkLTEyMzQ1Njc4OTBlZiJ9",
      "has_more": true,
      "limit": 10
    }
  }
  ```

### 3.3 `DELETE /api/v1/devices/:id`
刪除設備。此操作執行 **Saga 跨 DB 一致性事務**：

- **正常回應 (200 OK)**:
  ```json
  {
    "code": 200,
    "message": "device deleted successfully"
  }
  ```
- **快取清理失敗的局部成功回應 (207 Multi-Status)**:
  ```json
  {
    "code": 207,
    "message": "partial success: device deleted from PostgreSQL, but KeyDB cache cleanup failed",
    "data": "cache_cleanup_failed"
  }
  ```

---

## 4. 即時狀態與儀表板 (KeyDB)

### 4.1 `GET /api/v1/devices/:id/status`
取得設備的即時狀態（在線/離線 + 最新一筆遙測 + 近期告警累計次數）。

- **回應範例 (200 OK)**:
  ```json
  {
    "code": 200,
    "message": "ok",
    "data": {
      "device_id": "d0e1c2b3-a4f5-5678-abcd-1234567890ef",
      "is_online": true,
      "latest_telemetry": [
        {
          "metric_name": "temperature",
          "value": 26.5,
          "unit": "C",
          "recorded_at": "2026-07-02T18:05:00Z"
        }
      ],
      "alert_counts": {
        "info": 0,
        "warning": 2,
        "critical": 1
      }
    }
  }
  ```

### 4.2 `GET /api/v1/dashboard/overview`
儀表板總覽，一次性從 KeyDB Pipeline 讀取所有聚合統計值，不讀取主資料庫。

- **回應範例 (200 OK)**:
  ```json
  {
    "code": 200,
    "message": "ok",
    "data": {
      "device_total": 1250,
      "device_online": 980,
      "alert_counts": {
        "info": 12,
        "warning": 45,
        "critical": 8
      }
    }
  }
  ```

---

## 5. 時序遙測數據 (ScyllaDB)

### 5.1 `POST /api/v1/devices/:id/telemetry`
批量上傳遙測數據（限制單次上限 100 筆）。會自動判定是否觸發告警規則。

- **請求 Body**:
  ```json
  {
    "points": [
      {
        "recorded_at": "2026-07-02T18:10:00Z",
        "metric_name": "temperature",
        "value": 55.4,
        "unit": "C",
        "tags": {
          "sensor_type": "pt100"
        }
      }
    ]
  }
  ```
- **回應範例 (200 OK)**:
  ```json
  {
    "code": 200,
    "message": "ok"
  }
  ```
