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

	"github.com/joho/godotenv"

	"github.com/Ankesh2004/pontoon/internal/config"
	"github.com/Ankesh2004/pontoon/internal/delivery"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/github"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/postgres"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/redis"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using environment variables")
	}

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

	// Run migrations on startup
	slog.Info("running database migrations")
	if err := postgres.RunMigrations(ctx, cfg.Database.URL, "./migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	slog.Info("migrations completed successfully")

	pool, err := postgres.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		return err
	}
	defer pool.Close()

	slog.Info("database connection established")

	userRepo := postgres.NewUserRepo(pool)
	projectRepo := postgres.NewProjectRepo(pool)
	deploymentRepo := postgres.NewDeploymentRepo(pool)
	envVarRepo := postgres.NewEnvVarRepo(pool)

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

	redisClient, err := redis.NewRedisClient(cfg.Redis.URL)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	dockerClient, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	authUC := usecase.NewAuthUseCase(oauthService, userRepo, cfg.JWT.Secret)
	userUC := usecase.NewUserUseCase(userRepo)
	projectUC := usecase.NewProjectUseCase(projectRepo, deploymentRepo, envVarRepo, dockerClient, cfg.Domain.Default)
	deploymentUC := usecase.NewDeploymentUseCase(deploymentRepo, projectRepo, asynqClient, dockerClient)
	envVarUC := usecase.NewEnvVarUseCase(envVarRepo, projectRepo)
	webhookUC := usecase.NewWebhookUseCase(projectRepo, deploymentRepo, envVarRepo, asynqClient)

	router := delivery.NewRouter(cfg, authUC, userUC, projectUC, deploymentUC, envVarUC, webhookUC, redisClient, dockerClient)

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
