# Build stage
FROM golang:1.23-alpine AS builder

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
FROM alpine:3.19 AS api
RUN apk add --no-cache ca-certificates git

WORKDIR /app
COPY --from=api-builder /api .

EXPOSE 8080
CMD ["./api"]

# Worker runtime
FROM alpine:3.19 AS worker
RUN apk add --no-cache ca-certificates git docker-cli

WORKDIR /app
COPY --from=worker-builder /worker .

CMD ["./worker"]
