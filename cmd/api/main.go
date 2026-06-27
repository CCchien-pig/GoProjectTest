package main

// cmd/api/main.go — API Server 進入點
//
// 啟動順序（DI 鏈）：
//   config → gorm.DB → ScyllaClient → KeyDB → Repos → Services → Handlers → Router → Server
//
// 注意：PostgreSQL 和 KeyDB 直連 GCP 外部 IP
//   務必確認 GCP 防火牆已將本地 IP 加入白名單，開放 5432 與 6379 port

func main() {
	// TODO: Day 3 實作
	// 1. config.Load()
	// 2. initDatabase()     — GORM + PostgreSQL (直連 GCP:5432)
	// 3. initScylla()       — ScyllaDB (本地 Docker)
	// 4. initKeyDB()        — KeyDB (直連 GCP:6379)
	// 5. runMigrations()
	// 6. 組裝 Repos / Services / Handlers
	// 7. routes.Setup()
	// 8. http.Server + Graceful Shutdown
}
