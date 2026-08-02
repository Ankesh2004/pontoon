# Build stage — needs to match the go version in go.mod
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build API binary
FROM builder AS api-builder
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Build Worker binary
FROM builder AS worker-builder
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# API runtime
FROM alpine:3.21 AS api
RUN apk add --no-cache ca-certificates git

WORKDIR /app
COPY --from=api-builder /api .
# the app runs migrations on startup, so it needs the sql files
COPY --from=api-builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./api"]

# Worker runtime
FROM alpine:3.21 AS worker
RUN apk add --no-cache ca-certificates git docker-cli docker-cli-buildx curl bash tar
RUN curl -sSL https://nixpacks.com/install.sh | bash

WORKDIR /app
COPY --from=worker-builder /worker .
# worker also runs migrations on startup
COPY --from=worker-builder /app/migrations ./migrations

CMD ["./worker"]
