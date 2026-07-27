package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/Ankesh2004/pontoon/internal/config"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/postgres"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Running database migrations...")

	if err := postgres.RunMigrations(ctx, cfg.Database.URL, "./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
	os.Exit(0)
}
