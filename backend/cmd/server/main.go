package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"leadtrack/backend/internal/config"
	httphandler "leadtrack/backend/internal/http"
	"leadtrack/backend/internal/repository"
	"leadtrack/backend/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	loc, err := time.LoadLocation(cfg.AppTimezone)
	if err != nil {
		slog.Error("failed to load timezone", "timezone", cfg.AppTimezone, "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	slog.Info("running database migrations")
	if err := repository.RunMigrations(ctx, pool); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations completed successfully")

	repo := repository.New(pool)
	svc := service.New(repo)
	handler := httphandler.New(svc, loc)
	router := httphandler.NewRouter(handler, cfg.CorsOrigin)

	slog.Info("starting server", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
