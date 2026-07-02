package config

// internal/config/config.go — 環境變數集中管理
// 參考 USCII internal/config/config.go

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

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
func Load() *Config {
	// 嘗試載入 .env.dev，在環境變數已設定或找不到檔案時不崩潰
	_ = godotenv.Load(".env.dev")

	dbMaxConns, _ := strconv.Atoi(getEnv("DB_MAX_CONNS", "10"))
	dbMinConns, _ := strconv.Atoi(getEnv("DB_MIN_CONNS", "2"))
	keyDBClusterMode, _ := strconv.ParseBool(getEnv("KEYDB_CLUSTER_MODE", "false"))

	return &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		APIPort:            getEnv("API_PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		DBMaxConns:         dbMaxConns,
		DBMinConns:         dbMinConns,
		ScyllaHosts:        getEnv("SCYLLA_HOSTS", "localhost:9042"),
		ScyllaKeyspace:     getEnv("SCYLLA_KEYSPACE", "udm"),
		KeyDBAddr:          getEnv("KEYDB_ADDR", "localhost:6379"),
		KeyDBPassword:      getEnv("KEYDB_PASSWORD", ""),
		KeyDBClusterMode:   keyDBClusterMode,
		MigrationSourceURL: getEnv("MIGRATION_SOURCE_URL", "file://migrations"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
