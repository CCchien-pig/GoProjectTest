package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Setup temporary environment variables
	os.Setenv("APP_ENV", "test")
	os.Setenv("API_PORT", "9090")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("DB_MAX_CONNS", "20")
	os.Setenv("DB_MIN_CONNS", "5")
	os.Setenv("SCYLLA_HOSTS", "localhost:9042")
	os.Setenv("SCYLLA_KEYSPACE", "test_keyspace")
	os.Setenv("KEYDB_ADDR", "localhost:6379")
	os.Setenv("KEYDB_PASSWORD", "secret")
	os.Setenv("KEYDB_CLUSTER_MODE", "true")
	os.Setenv("MIGRATION_SOURCE_URL", "file://migrations")

	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("API_PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("DB_MAX_CONNS")
		os.Unsetenv("DB_MIN_CONNS")
		os.Unsetenv("SCYLLA_HOSTS")
		os.Unsetenv("SCYLLA_KEYSPACE")
		os.Unsetenv("KEYDB_ADDR")
		os.Unsetenv("KEYDB_PASSWORD")
		os.Unsetenv("KEYDB_CLUSTER_MODE")
		os.Unsetenv("MIGRATION_SOURCE_URL")
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
	if cfg.MigrationSourceURL != "file://migrations" {
		t.Errorf("expected MigrationSourceURL matches, got %s", cfg.MigrationSourceURL)
	}
}
