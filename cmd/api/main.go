package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/your-name/udm/internal/config"
	"github.com/your-name/udm/internal/handler"
	"github.com/your-name/udm/internal/keydb"
	"github.com/your-name/udm/internal/model"
	"github.com/your-name/udm/internal/repository"
	"github.com/your-name/udm/internal/routes"
	"github.com/your-name/udm/internal/scylla"
	"github.com/your-name/udm/internal/service"
)

func main() {
	log.Println("Starting UDM Platform API Server...")

	// 1. 載入環境變數設定
	cfg := config.Load()

	// 2. 初始化 PostgreSQL (GORM)
	db, err := initPostgres(cfg)
	var userRepo repository.UserRepository
	var deviceRepo repository.DeviceRepository
	var alertRuleRepo repository.AlertRuleRepository

	if err != nil {
		// PostgreSQL 為核心資料庫，連線失敗則立即中止啟動
		// （與 ScyllaDB/KeyDB 的降級處理不同，因為主檔 CRUD 全部依賴 PostgreSQL）
		log.Fatalf("CRITICAL: PostgreSQL connection failed: %v\n", err)
	} else {
		log.Println("PostgreSQL connected successfully")
		// 自動 Migration 資料表
		if err := db.AutoMigrate(&model.User{}, &model.Device{}, &model.AlertRule{}); err != nil {
			log.Fatalf("failed to auto migrate tables: %v", err)
		}
		log.Println("PostgreSQL auto migration completed")

		userRepo = repository.NewUserRepository(db)
		deviceRepo = repository.NewDeviceRepository(db)
		alertRuleRepo = repository.NewAlertRuleRepository(db)
	}

	// 3. 初始化 ScyllaDB
	var scyllaClient *scylla.Client
	var telemetryRepo scylla.TelemetryRepository
	var alertEventRepo scylla.AlertEventRepository

	scyllaClient, err = scylla.NewClient([]string{cfg.ScyllaHosts}, cfg.ScyllaKeyspace)
	if err != nil {
		log.Printf("ScyllaDB connection failed: %v. Ingestions will run in degraded mode.\n", err)
	} else {
		log.Println("ScyllaDB connected and keyspace initialized successfully")
		telemetryRepo = scylla.NewTelemetryRepository(scyllaClient)
		alertEventRepo = scylla.NewAlertEventRepository(scyllaClient)
	}

	// 4. 初始化 KeyDB
	var keydbClient *keydb.Client
	keydbClient, err = keydb.NewClient(cfg.KeyDBAddr, cfg.KeyDBPassword, cfg.KeyDBClusterMode)
	if err != nil {
		log.Printf("KeyDB connection failed: %v. Caching will run in degraded mode.\n", err)
	} else {
		log.Println("KeyDB connected successfully")
	}

	// 5. 組裝 Repositories, Services, and Handlers
	userService := service.NewUserService(userRepo)
	deviceService := service.NewDeviceService(deviceRepo, userRepo, telemetryRepo)
	alertRuleService := service.NewAlertRuleService(alertRuleRepo, deviceRepo)
	telemetryService := service.NewTelemetryService(telemetryRepo, alertEventRepo, deviceRepo, alertRuleRepo)

	userHandler := handler.NewUserHandler(userService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	alertRuleHandler := handler.NewAlertRuleHandler(alertRuleService)
	telemetryHandler := handler.NewTelemetryHandler(telemetryService)

	// 6. 註冊 Gin 路由
	router := routes.Setup(&routes.Dependencies{
		UserHandler:      userHandler,
		DeviceHandler:    deviceHandler,
		TelemetryHandler: telemetryHandler,
		AlertRuleHandler: alertRuleHandler,
	})

	// 7. 啟動 HTTP 伺服器
	srv := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %v", err)
		}
	}()
	log.Printf("HTTP Server is listening on port %s", cfg.APIPort)

	// 8. 實作 Graceful Shutdown (優雅關閉)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API Server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// 依序關閉連線釋放資源
	if keydbClient != nil {
		if err := keydbClient.Close(); err != nil {
			log.Printf("error closing KeyDB: %v", err)
		}
		log.Println("KeyDB connection closed")
	}
	if scyllaClient != nil {
		scyllaClient.Close()
		log.Println("ScyllaDB connection closed")
	}

	log.Println("UDM API Server gracefully stopped")
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
