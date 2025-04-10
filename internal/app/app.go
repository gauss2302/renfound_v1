package app

import (
	"context"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"

	"renfound_v1/config"
	"renfound_v1/infrastructure/auth"
	"renfound_v1/infrastructure/persistence/postgres"
	"renfound_v1/internal/delivery/http/router"
	"renfound_v1/internal/usecase/user"
	"renfound_v1/internal/utils/async"
)

// App represents the application
type App struct {
	cfg        *config.AppConfig
	router     *router.Router
	db         *postgres.Database
	workerPool *async.WorkerPool
	logger     *zap.Logger
}

// NewApp creates a new application
func NewApp() (*App, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	logger := cfg.Logger

	// Create database connection
	db, err := postgres.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	// Create worker pool for async operations
	workerPool := async.NewWorkerPool(10, 100, logger)

	// Create repositories
	userRepo := postgres.NewUserRepository(db, logger)

	// Create auth service
	telegramAuth := auth.NewTelegramAuth(cfg)

	// Create services
	userService := user.NewService(cfg, userRepo, telegramAuth, workerPool)

	// Create router
	r := router.NewRouter(cfg, userService, telegramAuth)
	r.SetupRoutes()

	return &App{
		cfg:        cfg,
		router:     r,
		db:         db,
		workerPool: workerPool,
		logger:     logger,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	// Start server in a goroutine
	go func() {
		if err := a.router.Start(); err != nil {
			a.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down application...")

	// Shutdown worker pool
	a.workerPool.Shutdown(false)

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server with context
	if err := a.router.Shutdown(ctx); err != nil {
		a.logger.Fatal("Server forced to shutdown", zap.Error(err))
		return err
	}

	// Close database connection
	a.db.Close()

	a.logger.Info("Application gracefully stopped")
	return nil
}
