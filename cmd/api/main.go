package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ankesh2004/pontoon/internal/config"
	"github.com/Ankesh2004/pontoon/internal/delivery"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/github"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/postgres"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/redis"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	slog.Info("starting pontoon API server")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	pool, err := postgres.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		return err
	}
	defer pool.Close()

	slog.Info("database connection established")

	userRepo := postgres.NewUserRepo(pool)
	projectRepo := postgres.NewProjectRepo(pool)
	deploymentRepo := postgres.NewDeploymentRepo(pool)

	oauthService := github.NewOAuthService(github.OAuthConfig{
		ClientID:     cfg.GitHub.ClientID,
		ClientSecret: cfg.GitHub.ClientSecret,
		RedirectURL:  fmt.Sprintf("http://localhost%s/auth/callback", cfg.API.Addr),
	})

	asynqClient, err := redis.NewAsynqClient(cfg.Redis.URL)
	if err != nil {
		return err
	}
	defer asynqClient.Close()

	authUC := usecase.NewAuthUseCase(oauthService, userRepo, cfg.JWT.Secret)
	projectUC := usecase.NewProjectUseCase(projectRepo, cfg.Domain.Default)
	deploymentUC := usecase.NewDeploymentUseCase(deploymentRepo, projectRepo, asynqClient)

	router := delivery.NewRouter(authUC, projectUC, deploymentUC)

	server := &http.Server{
		Addr:         cfg.API.Addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("API server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		slog.Info("shutting down API server")
	case sig := <-sigCh:
		slog.Info("received signal", "signal", sig)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		return err
	}

	slog.Info("API server stopped")
	return nil
}
