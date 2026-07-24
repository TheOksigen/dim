package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"exam-results-platform/api/internal/cache"
	"exam-results-platform/api/internal/config"
	"exam-results-platform/api/internal/httpapi"
	"exam-results-platform/api/internal/results"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("server configuration is invalid")
		os.Exit(1)
	}

	startupContext, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()
	pool, err := newPool(startupContext, cfg)
	if err != nil {
		logger.Error("database is unavailable during startup")
		os.Exit(1)
	}
	defer pool.Close()

	redisCache, err := cache.New(startupContext, cfg.RedisURL)
	if err != nil {
		logger.Warn("redis cache is unavailable; database lookups remain enabled")
		redisCache = nil
	}
	if redisCache != nil {
		defer func() { _ = redisCache.Close() }()
	}

	app := httpapi.New(cfg, results.NewRepository(pool), redisCache, logger)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- app.Listen(cfg.HTTPAddress)
	}()

	logger.Info("exam results API started", "environment", cfg.AppEnv)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-signals:
		logger.Info("shutdown signal received")
	case err := <-serverErrors:
		if err != nil {
			logger.Error("http server stopped unexpectedly")
		}
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := app.ShutdownWithContext(shutdownContext); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("graceful shutdown did not complete")
	}
	logger.Info("exam results API stopped")
}

func newPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = cfg.MaxDBConns
	poolConfig.MinConns = cfg.MinDBConns
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnLifetimeJitter = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute
	poolConfig.PingTimeout = cfg.DBTimeout
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["default_transaction_read_only"] = "on"

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
