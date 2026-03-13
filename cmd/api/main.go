package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nagicantsleep/k-map/internal/api"
	"github.com/nagicantsleep/k-map/internal/auth"
	"github.com/nagicantsleep/k-map/internal/config"
	"github.com/nagicantsleep/k-map/internal/geocode"
	"github.com/nagicantsleep/k-map/internal/proximity"
	"github.com/nagicantsleep/k-map/internal/storage"
	"github.com/nagicantsleep/k-map/internal/telemetry"
	"github.com/nagicantsleep/k-map/migrations"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)

		return 1
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := storage.NewPostgresPool(cfg.Postgres.DSN)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)

		return 1
	}
	defer db.Close()

	if err := storage.RunMigrations(db, migrations.FS); err != nil {
		logger.Error("failed to run migrations", "error", err)

		return 1
	}

	readinessChecker, err := api.NewReadinessChecker(cfg)
	if err != nil {
		logger.Error("failed to build readiness checker", "error", err)

		return 1
	}

	reg := prometheus.NewRegistry()
	metrics := telemetry.NewMetrics(reg)

	nominatimClient := geocode.NewNominatimClient(cfg.Nominatim.BaseURL, cfg.HTTP.WriteTimeout)

	cache := storage.NewCache(cfg.Redis.Address, cfg.Redis.CacheTTL)
	cachedGeocoder := geocode.NewCachedGeocoder(nominatimClient, cache, logger)

	proximitySvc := proximity.NewService(cachedGeocoder)

	authRepo := auth.NewRepository(db)
	rateLimiter := auth.NewRateLimiter(cfg.Redis.Address, cfg.RateLimit.RequestsPerMinute)
	usageRecorder := auth.NewUsageRecorder(db)

	handler := api.NewHandler(api.HandlerOptions{
		Logger:              logger,
		ReadinessChecker:    readinessChecker,
		Geocoder:            cachedGeocoder,
		Proximity:           proximitySvc,
		Metrics:             metrics,
		MetricsRegistry:     reg,
		AuthMiddleware:      auth.AuthMiddleware(authRepo),
		RateLimitMiddleware: auth.RateLimitMiddleware(rateLimiter),
		UsageMiddleware:     auth.UsageMiddleware(usageRecorder, logger),
	})
	server := api.NewServer(cfg.HTTP, handler)

	logger.Info("starting api server", "http_addr", cfg.HTTP.Address)

	if err := api.Run(ctx, server, cfg.HTTP.ShutdownTimeout); err != nil {
		logger.Error("api server stopped with error", "error", err)

		return 1
	}

	logger.Info("api server stopped cleanly")

	return 0
}
