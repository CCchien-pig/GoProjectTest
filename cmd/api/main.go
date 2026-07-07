package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"GoProject/udm/internal/cache"
	"GoProject/udm/internal/config"
	"GoProject/udm/internal/handler"
	"GoProject/udm/internal/keydb"
	"GoProject/udm/internal/model"
	"GoProject/udm/internal/repository"
	"GoProject/udm/internal/routes"
	"GoProject/udm/internal/scylla"
	"GoProject/udm/internal/service"
	"GoProject/udm/pkg/logger"
)

func main() {
	// 初始化結構化日誌
	logger.InitLogger()

	slog.Info("Starting UDM Platform API Server...")

	// 1. 載入環境變數設定
	cfg := config.Load()

	// 2. 初始化 PostgreSQL (GORM)
	db, err := initPostgres(cfg)
	var userRepo repository.UserRepository
	var deviceRepo repository.DeviceRepository
	var alertRuleRepo repository.AlertRuleRepository

	if err != nil {
		// PostgreSQL 為核心資料庫，若失敗則立即中止服務
		slog.Error("CRITICAL: PostgreSQL connection failed", "error", err)
		os.Exit(1)
	} else {
		slog.Info("PostgreSQL connected successfully")
		// 執行 Migration 資料表
		if err := db.AutoMigrate(&model.Role{}, &model.Permission{}, &model.User{}, &model.Device{}, &model.AlertRule{}); err != nil {
			slog.Error("failed to auto migrate tables", "error", err)
			os.Exit(1)
		}
		slog.Info("PostgreSQL auto migration completed")

		userRepo = repository.NewUserRepository(db)
		deviceRepo = repository.NewDeviceRepository(db)
		alertRuleRepo = repository.NewAlertRuleRepository(db)
	}

	// 3. 初始化 ScyllaDB
	var scyllaClient *scylla.Client
	var telemetryRepo scylla.TelemetryRepository
	var alertEventRepo scylla.AlertEventRepository

	hosts := strings.Split(cfg.ScyllaHosts, ",")
	scyllaClient, err = scylla.NewClient(hosts, cfg.ScyllaKeyspace)
	if err != nil {
		slog.Warn("ScyllaDB connection failed. Ingestions will run in degraded mode.", "error", err)
	} else {
		slog.Info("ScyllaDB connected and keyspace initialized successfully")
		telemetryRepo = scylla.NewTelemetryRepository(scyllaClient)
		alertEventRepo = scylla.NewAlertEventRepository(scyllaClient)
	}

	// 4. 初始化 KeyDB
	var keydbClient *keydb.Client
	keydbClient, err = keydb.NewClient(cfg.KeyDBAddr, cfg.KeyDBPassword, cfg.KeyDBClusterMode, cfg.KeyDBUseTLS, cfg.KeyDBCACertPath, cfg.KeyDBInsecure)
	if err != nil {
		slog.Warn("KeyDB connection failed. Caching will run in degraded mode.", "error", err)
	} else {
		slog.Info("KeyDB connected successfully")
	}

	// 5. 組裝 Repositories, Services, and Handlers
	var cacheService cache.Service
	if keydbClient != nil {
		cacheService = cache.NewService(keydbClient.Client)
	}

	userService := service.NewUserService(userRepo)
	deviceService := service.NewDeviceService(deviceRepo, userRepo, telemetryRepo, cacheService)
	alertRuleService := service.NewAlertRuleService(alertRuleRepo, deviceRepo)
	telemetryService := service.NewTelemetryService(telemetryRepo, alertEventRepo, deviceRepo, alertRuleRepo, cacheService)
	statusService := service.NewStatusService(cacheService, telemetryRepo)
	dashboardService := service.NewDashboardService(cacheService, deviceRepo)

	userHandler := handler.NewUserHandler(userService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	alertRuleHandler := handler.NewAlertRuleHandler(alertRuleService)
	telemetryHandler := handler.NewTelemetryHandler(telemetryService)
	statusHandler := handler.NewStatusHandler(statusService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	cacheHandler := handler.NewCacheHandler(cacheService)
	healthHandler := handler.NewHealthHandler(db, scyllaClient, keydbClient)

	// 6. 註冊 Gin 路由
	router := routes.Setup(&routes.Dependencies{
		UserHandler:      userHandler,
		DeviceHandler:    deviceHandler,
		TelemetryHandler: telemetryHandler,
		AlertRuleHandler: alertRuleHandler,
		StatusHandler:    statusHandler,
		DashboardHandler: dashboardHandler,
		CacheHandler:     cacheHandler,
		HealthHandler:    healthHandler,
	})

	// 7. 啟動 HTTP 伺服器
	srv := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("HTTP Server is listening", "port", cfg.APIPort)

	// 8. 實作 Graceful Shutdown（優雅關機）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down API Server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	// 依序關閉並釋放資源
	if keydbClient != nil {
		if err := keydbClient.Close(); err != nil {
			slog.Error("error closing KeyDB", "error", err)
		}
		slog.Info("KeyDB connection closed")
	}
	if scyllaClient != nil {
		scyllaClient.Close()
		slog.Info("ScyllaDB connection closed")
	}

	slog.Info("UDM API Server gracefully stopped")
}

func initPostgres(cfg *config.Config) (*gorm.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, errors.New("empty database connection string")
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.DBMinConns)
	sqlDB.SetMaxOpenConns(cfg.DBMaxConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
