package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"exam-results-platform/api/internal/cache"
	"exam-results-platform/api/internal/config"
	"exam-results-platform/api/internal/results"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

const (
	resultCachePrefix = "exam-result:v1:"
	missCachePrefix   = "exam-result-miss:v1:"
)

type App struct {
	config     config.Config
	repository *results.Repository
	cache      *cache.Redis
	logger     *slog.Logger
}

type lookupRequest struct {
	FINCode string `json:"finCode"`
}

type lookupResponse struct {
	Result results.Result `json:"result"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func New(cfg config.Config, repository *results.Repository, redisCache *cache.Redis, logger *slog.Logger) *fiber.App {
	handler := &App{config: cfg, repository: repository, cache: redisCache, logger: logger}
	app := fiber.New(fiber.Config{
		AppName:      "exam-results-api",
		BodyLimit:    cfg.MaxBodyBytes,
		ReadTimeout:  cfg.RequestTimeout,
		WriteTimeout: cfg.RequestTimeout,
		IdleTimeout:  15 * time.Second,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return handler.writeError(c, fiber.StatusInternalServerError, "internal_error", "Xidmətə qoşulmaq mümkün olmadı. Yenidən cəhd edin.")
		},
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: commaSeparated(cfg.CORSOrigins),
		AllowMethods: []string{fiber.MethodPost, fiber.MethodGet, fiber.MethodOptions},
		AllowHeaders: []string{"Content-Type", "X-Request-Id"},
		MaxAge:       600,
	}))
	app.Use(func(c fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, private")
		c.Set("Pragma", "no-cache")
		c.Set("X-Content-Type-Options", "nosniff")
		return c.Next()
	})

	app.Get("/healthz", handler.liveness)
	app.Get("/readyz", handler.readiness)

	lookupRateLimit := limiter.New(limiter.Config{
		Max:        cfg.RateLimitMax,
		Expiration: cfg.RateLimitWindow,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
	})
	app.Post("/api/v1/results/lookup", lookupRateLimit, handler.lookup)

	return app
}

func commaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func (a *App) liveness(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (a *App) readiness(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), a.config.DBTimeout)
	defer cancel()
	if err := a.repository.Ping(ctx); err != nil {
		return a.writeError(c, fiber.StatusServiceUnavailable, "not_ready", "Xidmət hazırda əlçatan deyil.")
	}
	return c.JSON(fiber.Map{"status": "ready"})
}

func (a *App) lookup(c fiber.Ctx) error {
	startedAt := time.Now()
	request := new(lookupRequest)
	if err := c.Bind().Body(request); err != nil {
		return a.completeWithError(c, startedAt, "invalid_request", "none", fiber.StatusBadRequest, "invalid_request", "Göndərilən məlumat düzgün deyil.")
	}

	fin, err := results.NormalizeFIN(request.FINCode)
	if err != nil {
		return a.completeWithError(c, startedAt, "invalid_fin", "none", fiber.StatusBadRequest, "invalid_fin", "FIN kodu 7 böyük hərf və ya rəqəmdən ibarət olmalıdır.")
	}

	ctx, cancel := context.WithTimeout(c.Context(), a.config.DBTimeout)
	defer cancel()
	cacheKey := resultCachePrefix + fin
	missKey := missCachePrefix + fin

	if payload, found := a.cacheGet(ctx, cacheKey); found {
		a.logLookup("success", "cache", startedAt)
		c.Set("Content-Type", "application/json; charset=utf-8")
		return c.Send(payload)
	}
	if _, found := a.cacheGet(ctx, missKey); found {
		return a.completeWithError(c, startedAt, "not_found", "negative_cache", fiber.StatusNotFound, "not_found", "Nəticə tapılmadı. FIN kodunu yenidən yoxlayın.")
	}

	result, err := a.repository.FindByFIN(ctx, fin)
	if errors.Is(err, results.ErrNotFound) {
		a.cacheSet(ctx, missKey, []byte("1"), a.config.NegativeCacheTTL)
		return a.completeWithError(c, startedAt, "not_found", "database", fiber.StatusNotFound, "not_found", "Nəticə tapılmadı. FIN kodunu yenidən yoxlayın.")
	}
	if err != nil {
		return a.completeWithError(c, startedAt, "unavailable", "database", fiber.StatusServiceUnavailable, "service_unavailable", "Xidmət hazırda əlçatan deyil. Yenidən cəhd edin.")
	}

	payload, err := json.Marshal(lookupResponse{Result: result})
	if err != nil {
		return a.completeWithError(c, startedAt, "internal_error", "serialization", fiber.StatusInternalServerError, "internal_error", "Xidmətə qoşulmaq mümkün olmadı. Yenidən cəhd edin.")
	}
	a.cacheSet(ctx, cacheKey, payload, a.config.CacheTTL)
	a.logLookup("success", "database", startedAt)
	c.Set("Content-Type", "application/json; charset=utf-8")
	return c.Send(payload)
}

func (a *App) cacheGet(ctx context.Context, key string) ([]byte, bool) {
	if a.cache == nil {
		return nil, false
	}
	payload, found, err := a.cache.Get(ctx, key)
	if err != nil || !found {
		return nil, false
	}
	if strings.HasPrefix(key, resultCachePrefix) && !json.Valid(payload) {
		a.cache.Delete(ctx, key)
		return nil, false
	}
	return payload, true
}

func (a *App) cacheSet(ctx context.Context, key string, payload []byte, ttl time.Duration) {
	if a.cache != nil {
		_ = a.cache.Set(ctx, key, payload, ttl)
	}
}

func (a *App) completeWithError(c fiber.Ctx, startedAt time.Time, outcome, source string, status int, code, message string) error {
	a.logLookup(outcome, source, startedAt)
	return a.writeError(c, status, code, message)
}

func (a *App) writeError(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(errorEnvelope{Error: apiError{Code: code, Message: message}})
}

func (a *App) logLookup(outcome, source string, startedAt time.Time) {
	a.logger.Info("lookup completed",
		"outcome", outcome,
		"source", source,
		"duration_ms", time.Since(startedAt).Milliseconds(),
	)
}
