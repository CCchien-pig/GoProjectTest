package scylla

// internal/scylla/client.go — ScyllaDB 連線
// 參考 USCII internal/scylla/client.go
// TODO: Day 6 實作
// func NewClient(hosts []string, keyspace string) (*gocql.Session, error) {}
// func EnsureSchema(session *gocql.Session) error {}
// EnsureSchema 建立:
//   telemetry   — Partition Key: (device_id, date), Clustering Key: (recorded_at DESC, metric_name ASC), TTL 90天
//   alert_events — Partition Key: (device_id, month), Clustering Key: (triggered_at DESC, rule_id ASC), TTL 365天
