package main

// cmd/api/main.go — API Server 進入點
//
// 啟動順序（DI 鏈）：
//   config → gorm.DB → ScyllaClient → KeyDB → Repos → Services → Handlers → Router → Server
//
// 注意：PostgreSQL, ScyllaDB, KeyDB 皆在本地 Docker 運行
//   請先執行：make compose-up

func main() {
	// TODO: Day 3 實作
	// 1. config.Load()
	// 2. initDatabase()     — GORM + PostgreSQL (本地 Docker:5432)
	// 3. initScylla()       — ScyllaDB (本地 Docker:9042)
	// 4. initKeyDB()        — KeyDB (本地 Docker:6379)
	// 5. runMigrations()
	// 6. 組裝 Repos / Services / Handlers
	// 7. routes.Setup()
	// 8. http.Server + Graceful Shutdown
}
