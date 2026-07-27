package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ankesh2004/pontoon/internal/config"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/postgres"
	"github.com/Ankesh2004/pontoon/internal/worker"
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
	slog.Info("starting pontoon worker")

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
	envVarRepo := postgres.NewEnvVarRepo(pool)

	deployProcessor := worker.NewDeployProcessor(
		dockerClient,
		deploymentRepo,
		projectRepo,
		envVarRepo,
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
