.PHONY: build run-api run-worker migrate test lint clean deps fmt vet dev data-stack docker-up docker-down docker-logs docker-restart

# Build binaries
build:
	@echo "Building API..."
	@go build -o bin/api ./cmd/api
	@echo "Building Worker..."
	@go build -o bin/worker ./cmd/worker

# Run API server (needs data-stack running)
run-api:
	@go run ./cmd/api

# Run Worker (needs data-stack running)
run-worker:
	@go run ./cmd/worker

# Run database migrations
migrate:
	@go run ./cmd/migrate

# Run tests
test:
	@go test -v ./...

# Run linter
lint:
	@golangci-lint run

# Clean build artifacts
clean:
	@rm -rf bin/

# Install dependencies
deps:
	@go mod download
	@go mod tidy

# Format code
fmt:
	@go fmt ./...

# Vet code
vet:
	@go vet ./...

# Data services only (Postgres + Redis on host ports) for local `go run` dev
data-stack:
	@docker compose -f docker-compose.test.yml up -d

# Full self-hosted stack (Postgres, Redis, Traefik, API, Worker)
# Requires: docker network create pontoon-ingress
dev:
	@docker compose up --build

# Start full stack in background
docker-up:
	@docker compose up -d --build

# Stop full stack
docker-down:
	@docker compose down

# View logs
docker-logs:
	@docker compose logs -f

# Restart services
docker-restart:
	@docker compose restart