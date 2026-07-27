.PHONY: build run-api run-worker migrate test lint clean dev dev-local docker-build docker-up docker-down

# Build binaries
build:
	@echo "Building API..."
	@go build -o bin/api ./cmd/api
	@echo "Building Worker..."
	@go build -o bin/worker ./cmd/worker

# Run API server
run-api:
	@go run ./cmd/api

# Run Worker
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

# Development with Docker Compose (local)
dev-local:
	@docker compose -f docker-compose.local.yml up --build

# Development with Docker Compose (production-like)
dev:
	@docker compose up --build

# Build Docker images
docker-build:
	@docker compose build

# Start Docker Compose
docker-up:
	@docker compose up -d

# Stop Docker Compose
docker-down:
	@docker compose down

# View logs
docker-logs:
	@docker compose logs -f

# Restart services
docker-restart:
	@docker compose restart
