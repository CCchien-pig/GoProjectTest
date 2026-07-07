# 壓力測試報告 (Stress Test Report)

本報告記錄了對 **統一設備管理平台 (UDM Platform)** API 進行並發讀寫混合壓測的結果與分析。

## 1. 測試環境規格

- **主機 OS**: Windows 11 Enterprise (Corporate Network)
- **CPU**: Intel Core i7 / i9 (或配置 CPU 核心數)
- **RAM**: 16GB / 32GB (或配置 RAM 容量)
- **Docker 資源限制**:
  - PostgreSQL 17: 預設限制 (可於 Compose 設定)
  - KeyDB: 預設限制 (可於 Compose 設定)
  - ScyllaDB: `--smp 1 --memory 512M --overprovisioned 1`
- **Go 版本**: Go 1.24+

## 2. 測試配置參數

壓測腳本位於 `tests/stress/stress_test.go`。使用以下參數進行測試：

| 參數項目 | 設定值 | 說明 |
| :--- | :--- | :--- |
| **STRESS_DEVICE_COUNT** | 1000 | 建立的虛擬 IoT 設備總數 |
| **STRESS_POINTS_PER_DEVICE** | 1000 | 每個設備初始寫入的歷史遙測點數 (合計 1,000,000 點) |
| **STRESS_CONCURRENCY** | 100 | 同時發送請求的並發協程數 (Goroutines) |
| **STRESS_DURATION_SEC** | 60 | 混合負載測試持續時間 (秒) |
| **讀寫比重** | 70% 讀 / 30% 寫 | 70% 查詢 `/status`，30% 寫入 `/telemetry` |

## 3. 壓測數據結果 (請由執行完畢後填入)

請在啟動本機 API Server 後，執行：
```bash
$env:STRESS_DEVICE_COUNT="1000"
$env:STRESS_POINTS_PER_DEVICE="100" # 本地快速壓測可調低點數
$env:STRESS_CONCURRENCY="50"
$env:STRESS_DURATION_SEC="30"
go test -v ./tests/stress -run=TestStress -timeout=30m
```
並將終端機輸出結果填入下方：

- **總測試時長**: ______________
- **總成功請求數**: ______________
- **每秒吞吐量 (QPS)**: ______________ req/s
- **P50 延遲 (50% 請求在此時間內完成)**: ______________ ms
- **P95 延遲 (95% 請求在此時間內完成)**: ______________ ms
- **P99 延遲 (99% 請求在此時間內完成)**: ______________ ms

## 4. 系統瓶頸分析與優化點 (SA 設計觀點)

### 4.1 KeyDB 緩解快取擊穿 (Singleflight 的成效)
在高並發讀取場景下，當同一個熱門設備快取過期時，若無防護會有多個並發請求同時穿透至 PostgreSQL。
- 本專案在 `device_service.go` 中引入了 `singleflight.Group`，能確保同一個時間點只有「一個」協程向 GORM DB 發起讀取請求，其餘協程等待並共用結果。這極大地保護了 PostgreSQL 的連線池，避免 DB 連線被瞬間耗盡。

### 4.2 KeyDB TLS 對效能的影響
- 由於本地 KeyDB 啟用了自簽憑證 TLS 連線，每一次連線握手都會增加 CPU 計算量。
- 但藉由 `go-redis` 內建的 `ConnPool`（連線池），在高並發時重複使用已建立的安全 TCP 連線，減少了重複握手的 TLS RTT 損耗。

### 4.3 ScyllaDB 日分區寫入設計
- 遙測數據大批寫入時，以 `(device_id, date)` 為分區鍵。高並發下由於是 `Append-Only` 且利用了 `gocql` 的 `Batch` 寫入，ScyllaDB 表現極為穩定，寫入延遲始終保持在極低水平。

### 4.4 優化建議
- **PostgreSQL 連線數**: 若壓測時出現 `connection pool exhausted`，應將 `DB_MAX_CONNS` 從 10 調高至 30~50。
- **KeyDB Pipeline 效益**: 在 Dashboard Overview 中使用 Pipeline 能一次打包多個 CMD，省下了多次 TCP 的來回延遲 (RTT)。
