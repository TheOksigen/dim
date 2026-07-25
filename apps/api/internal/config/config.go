package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv           string
	HTTPAddress      string
	DatabaseURL      string
	RedisURL         string
	MaxDBConns       int32
	MinDBConns       int32
	DBTimeout        time.Duration
	RequestTimeout   time.Duration
	MaxBodyBytes     int
	RateLimitMax     int
	RateLimitWindow  time.Duration
	CacheTTL         time.Duration
	NegativeCacheTTL time.Duration
}

func Load() (Config, error) {
	// .env is optional in containers, where variables are injected by the runtime.
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:           valueOr("APP_ENV", "production"),
		HTTPAddress:      valueOr("HTTP_ADDR", ":8080"),
		DatabaseURL:      strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisURL:         strings.TrimSpace(os.Getenv("REDIS_URL")),
		MaxDBConns:       32,
		MinDBConns:       4,
		DBTimeout:        time.Second,
		RequestTimeout:   3 * time.Second,
		MaxBodyBytes:     1024,
		RateLimitMax:     60,
		RateLimitWindow:  time.Minute,
		CacheTTL:         15 * time.Minute,
		NegativeCacheTTL: 30 * time.Second,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	var err error
	if cfg.MaxDBConns, err = int32FromEnv("API_MAX_DB_CONNS", cfg.MaxDBConns); err != nil {
		return Config{}, err
	}
	if cfg.MinDBConns, err = int32FromEnv("API_MIN_DB_CONNS", cfg.MinDBConns); err != nil {
		return Config{}, err
	}
	if cfg.MinDBConns < 0 || cfg.MaxDBConns < 1 || cfg.MinDBConns > cfg.MaxDBConns {
		return Config{}, fmt.Errorf("database pool bounds are invalid")
	}
	if cfg.DBTimeout, err = durationFromEnv("API_DB_TIMEOUT", cfg.DBTimeout); err != nil {
		return Config{}, err
	}
	if cfg.RequestTimeout, err = durationFromEnv("API_REQUEST_TIMEOUT", cfg.RequestTimeout); err != nil {
		return Config{}, err
	}
	if cfg.RateLimitWindow, err = durationFromEnv("API_RATE_LIMIT_WINDOW", cfg.RateLimitWindow); err != nil {
		return Config{}, err
	}
	if cfg.CacheTTL, err = durationFromEnv("API_CACHE_TTL", cfg.CacheTTL); err != nil {
		return Config{}, err
	}
	if cfg.NegativeCacheTTL, err = durationFromEnv("API_NEGATIVE_CACHE_TTL", cfg.NegativeCacheTTL); err != nil {
		return Config{}, err
	}
	if cfg.MaxBodyBytes, err = intFromEnv("API_MAX_BODY_BYTES", cfg.MaxBodyBytes); err != nil {
		return Config{}, err
	}
	if cfg.RateLimitMax, err = intFromEnv("API_RATE_LIMIT_MAX", cfg.RateLimitMax); err != nil {
		return Config{}, err
	}

	if cfg.DBTimeout <= 0 || cfg.RequestTimeout <= 0 || cfg.MaxBodyBytes < 1 || cfg.RateLimitMax < 1 || cfg.RateLimitWindow <= 0 || cfg.CacheTTL <= 0 || cfg.NegativeCacheTTL <= 0 {
		return Config{}, fmt.Errorf("timeout, cache, rate-limit, and body values must be positive")
	}

	return cfg, nil
}

func valueOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration", key)
	}
	return parsed, nil
}

func intFromEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return parsed, nil
}

func int32FromEnv(key string, fallback int32) (int32, error) {
	value, err := intFromEnv(key, int(fallback))
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}
