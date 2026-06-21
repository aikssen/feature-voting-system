// Command api is the SoundFlow Feature Voting backend entrypoint.
//
// Boot sequence: load config → connect (with retry) → migrate → seed →
// wire slices → serve, with graceful shutdown on SIGINT/SIGTERM.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"soundflow/internal/auth"
	"soundflow/internal/feature"
	"soundflow/internal/infrastructure"
	"soundflow/internal/shared/token"
	"soundflow/internal/vote"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := infrastructure.LoadConfig()
	if err != nil {
		return err
	}

	// Root context cancelled on shutdown signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructure.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := infrastructure.Migrate(ctx, pool); err != nil {
		return err
	}
	if err := infrastructure.Seed(ctx, pool); err != nil {
		return err
	}

	// Dependency injection: build repositories → services → handlers.
	tokenManager := token.NewManager(cfg.JWTSecret, cfg.JWTTTLHours)

	userRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(userRepo, tokenManager)

	featureRepo := feature.NewRepository(pool)
	featureSvc := feature.NewService(featureRepo)

	voteRepo := vote.NewRepository(pool)
	voteSvc := vote.NewService(voteRepo, featureRepo) // featureRepo satisfies AuthorReader

	router := infrastructure.NewRouter(infrastructure.RouterDeps{
		TokenManager: tokenManager,
		Auth:         auth.NewHandler(authSvc),
		Feature:      feature.NewHandler(featureSvc),
		Vote:         vote.NewHandler(voteSvc),
		CORSOrigins:  cfg.CORSAllowedOrigins,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("api listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
