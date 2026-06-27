package keydb

// internal/keydb/client.go — KeyDB 連線
// 參考 USCII internal/keydb/dial.go
//
// 連線方式：SSH Tunnel（本地 6380 → GCP:6379）
// .env.dev:  KEYDB_ADDR=localhost:6380
//
// TODO: Day 11 實作
// func NewClient(addr, password string) (*redis.Client, error) {}
