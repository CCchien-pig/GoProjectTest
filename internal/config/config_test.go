package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Setup temporary environment variables
	_ = os.Setenv("APP_ENV", "test")
	_ = os.Setenv("API_PORT", "9090")
	_ = os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	_ = os.Setenv("DB_MAX_CONNS", "20")
	_ = os.Setenv("DB_MIN_CONNS", "5")
	_ = os.Setenv("SCYLLA_HOSTS", "localhost:9042")
	_ = os.Setenv("SCYLLA_KEYSPACE", "test_keyspace")
	_ = os.Setenv("KEYDB_ADDR", "localhost:6379")
	_ = os.Setenv("KEYDB_PASSWORD", "secret")
	_ = os.Setenv("KEYDB_CLUSTER_MODE", "true")

	defer func() {
		_ = os.Unsetenv("APP_ENV")
		_ = os.Unsetenv("API_PORT")
		_ = os.Unsetenv("DATABASE_URL")
		_ = os.Unsetenv("DB_MAX_CONNS")
		_ = os.Unsetenv("DB_MIN_CONNS")
		_ = os.Unsetenv("SCYLLA_HOSTS")
		_ = os.Unsetenv("SCYLLA_KEYSPACE")
		_ = os.Unsetenv("KEYDB_ADDR")
		_ = os.Unsetenv("KEYDB_PASSWORD")
		_ = os.Unsetenv("KEYDB_CLUSTER_MODE")
	}()

	cfg := Load()

	if cfg.AppEnv != "test" {
		t.Errorf("expected AppEnv to be test, got %s", cfg.AppEnv)
	}
	if cfg.APIPort != "9090" {
		t.Errorf("expected APIPort to be 9090, got %s", cfg.APIPort)
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("expected DatabaseURL matches, got %s", cfg.DatabaseURL)
	}
	if cfg.DBMaxConns != 20 {
		t.Errorf("expected DBMaxConns 20, got %d", cfg.DBMaxConns)
	}
	if cfg.DBMinConns != 5 {
		t.Errorf("expected DBMinConns 5, got %d", cfg.DBMinConns)
	}
	if cfg.ScyllaHosts != "localhost:9042" {
		t.Errorf("expected ScyllaHosts matches, got %s", cfg.ScyllaHosts)
	}
	if cfg.ScyllaKeyspace != "test_keyspace" {
		t.Errorf("expected ScyllaKeyspace matches, got %s", cfg.ScyllaKeyspace)
	}
	if cfg.KeyDBAddr != "localhost:6379" {
		t.Errorf("expected KeyDBAddr matches, got %s", cfg.KeyDBAddr)
	}
	if cfg.KeyDBPassword != "secret" {
		t.Errorf("expected KeyDBPassword matches, got %s", cfg.KeyDBPassword)
	}
	if !cfg.KeyDBClusterMode {
		t.Errorf("expected KeyDBClusterMode to be true, got false")
	}

}
