package scylla

// internal/scylla/telemetry_repo.go
// TODO: Day 7~8 實作
// BatchInsert(deviceID, []TelemetryPoint)       — CQL Batch INSERT，上限 100 筆
// Query(deviceID, start, end, metricName)       — 跨日分區查詢
// QueryLatest(deviceID)                         — 最新各 metric 數據
// DeleteByRange(deviceID, start, end)
