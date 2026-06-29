package config

// internal/config/config.go — 環境變數集中管理
// 參考 USCII internal/config/config.go

// Config 包含所有服務的設定
type Config struct {
	AppEnv  string
	APIPort string

	// PostgreSQL — 本地 Docker
	DatabaseURL string
	DBMaxConns  int
	DBMinConns  int

	// ScyllaDB — 本地 Docker
	ScyllaHosts    string
	ScyllaKeyspace string

	// KeyDB — 本地 Docker
	KeyDBAddr        string
	KeyDBPassword    string
	KeyDBClusterMode bool

	// Migration 來源（golang-migrate）
	MigrationSourceURL string
}

// Load 從 .env.dev 載入設定
// TODO: Day 1 實作
func Load() *Config {
	panic("TODO: implement Load()")
}
