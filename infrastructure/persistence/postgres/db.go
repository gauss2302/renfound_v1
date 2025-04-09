package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"renfound_v1/config"
	"time"
)

type Database struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewDatabase(cfg *config.AppConfig) (*Database, error) {
	logger := cfg.Logger.With(zap.String("component", "database"))

	poolConfig, err := pgxpool.ParseConfig(cfg.Config.DB.URL)
	if err != nil {
		logger.Error("Failed to parse db Url", zap.Error(err))
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Pool Configuration
	poolConfig.MaxConns = 25
	poolConfig.MaxConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("Failed to create database connection pool", zap.Error(err))
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected successfully to db")

	return &Database{
		Pool:   pool,
		logger: logger,
	}, nil
}

func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.logger.Info("Db connection pool closed")
	}
}
