package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Ankesh2004/pontoon/internal/config"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/postgres"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/redis"
	"github.com/Ankesh2004/pontoon/internal/usecase"
	"github.com/Ankesh2004/pontoon/internal/worker"
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
	slog.Info("starting pontoon worker")

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

	dockerClient, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	slog.Info("docker connection established")

	if err := dockerClient.EnsureIngressNetwork(ctx); err != nil {
		return err
	}
	slog.Info("ingress network ensured")

	deploymentRepo := postgres.NewDeploymentRepo(pool)
	projectRepo := postgres.NewProjectRepo(pool)
	envVarRepo := postgres.NewEnvVarRepo(pool, cfg.EncryptionKey)

	redisClient, err := redis.NewRedisClient(cfg.Redis.URL)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	slog.Info("redis connection established")

	capacityUC := usecase.NewCapacityUseCase(
		dockerClient,
		cfg.Worker.MaxContainerMemoryMB,
		cfg.Worker.TotalContainerMemoryMB,
	)

	deployProcessor := worker.NewDeployProcessor(
		dockerClient,
		deploymentRepo,
		projectRepo,
		envVarRepo,
		capacityUC,
		cfg.Worker.MaxContainerMemoryMB,
		redisClient,
	)

	processor, err := worker.NewProcessor(cfg.Redis.URL, 2, deployProcessor)
	if err != nil {
		return err
	}

	go func() {
		if err := processor.Start(); err != nil {
			slog.Error("processor error", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("worker ready")

	select {
	case <-ctx.Done():
		slog.Info("shutting down worker")
	case sig := <-sigCh:
		slog.Info("received signal", "signal", sig)
	}

	processor.Stop()
	slog.Info("worker stopped")
	return nil
}
