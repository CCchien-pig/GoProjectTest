package main

// cmd/api/main.go — API Server 進入點
//
// 啟動順序（DI 鏈）：
//   config → gorm.DB → ScyllaClient → KeyDB → Repos → Services → Handlers → Router → Server
//
// 注意：PostgreSQL 和 KeyDB 透過 SSH Tunnel 連線
//   本地 localhost:5433 → GCP:5432 (PostgreSQL)
//   本地 localhost:6380 → GCP:6379 (KeyDB)
//   啟動前請先執行：bash .docker/ssh/start_tunnels.sh start

func main() {
	// TODO: Day 3 實作
	// 1. config.Load()
	// 2. initDatabase()     — GORM + PostgreSQL (via SSH Tunnel → GCP:5432)
	// 3. initScylla()       — ScyllaDB (本地 Docker)
	// 4. initKeyDB()        — KeyDB (via SSH Tunnel → GCP:6379)
	// 5. runMigrations()
	// 6. 組裝 Repos / Services / Handlers
	// 7. routes.Setup()
	// 8. http.Server + Graceful Shutdown
}
