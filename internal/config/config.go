package config

// internal/config/config.go — 環境變數集中管理
// 參考 USCII internal/config/config.go

import (
	"os"
	"strconv"
)

// Config 包含所有服務的設定
type Config struct {
	AppEnv  string
	APIPort string

	// PostgreSQL 本地 Docker
	DatabaseURL string
	DBMaxConns  int
	DBMinConns  int

	// ScyllaDB 本地 Docker
	ScyllaHosts    string
	ScyllaKeyspace string

	// KeyDB 本地 Docker
	KeyDBAddr        string
	KeyDBPassword    string
	KeyDBClusterMode bool
	KeyDBUseTLS      bool
	KeyDBCACertPath  string
	KeyDBInsecure    bool
}

// Load 從 .env.dev 載入設定
func Load() *Config {
	// 嘗試載入 .env.dev，在環境變數已設定或找不到檔案時不崩潰
	loadEnvFile(".env.dev")

	dbMaxConns, _ := strconv.Atoi(getEnv("DB_MAX_CONNS", "10"))
	dbMinConns, _ := strconv.Atoi(getEnv("DB_MIN_CONNS", "2"))
	keyDBClusterMode, _ := strconv.ParseBool(getEnv("KEYDB_CLUSTER_MODE", "false"))
	keyDBUseTLS, _ := strconv.ParseBool(getEnv("KEYDB_USE_TLS", "false"))
	keyDBInsecure, _ := strconv.ParseBool(getEnv("KEYDB_INSECURE_SKIP_VERIFY", "false"))

	return &Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		APIPort:          getEnv("API_PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		DBMaxConns:       dbMaxConns,
		DBMinConns:       dbMinConns,
		ScyllaHosts:      getEnv("SCYLLA_HOSTS", "localhost:9042"),
		ScyllaKeyspace:   getEnv("SCYLLA_KEYSPACE", "udm"),
		KeyDBAddr:        getEnv("KEYDB_ADDR", "localhost:6379"),
		KeyDBPassword:    getEnv("KEYDB_PASSWORD", ""),
		KeyDBClusterMode: keyDBClusterMode,
		KeyDBUseTLS:      keyDBUseTLS,
		KeyDBCACertPath:  getEnv("KEYDB_CA_CERT_PATH", ""),
		KeyDBInsecure:    keyDBInsecure,
	}
}

// loadEnvFile 嘗試載入 .env 檔案，不存在時忽略（不 panic）
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	// 簡易解析 KEY=VALUE 格式
	import_godotenv_style(filename)
}

func import_godotenv_style(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	lines := splitLines(string(data))
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		for i, c := range line {
			if c == '=' {
				key := line[:i]
				val := line[i+1:]
				if _, exists := os.LookupEnv(key); !exists {
					_ = os.Setenv(key, val)
				}
				break
			}
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
