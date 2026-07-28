package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	API      APIConfig
	Database DatabaseConfig
	Redis    RedisConfig
	GitHub   GitHubConfig
	JWT      JWTConfig
	Domain   DomainConfig
	Worker   WorkerConfig
	Webhook  WebhookConfig
	CORS     CORSConfig
}

type CORSConfig struct {
	AllowedOrigins []string
}

type APIConfig struct {
	Addr string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
}

type JWTConfig struct {
	Secret string
}

type DomainConfig struct {
	Default string
}

type WorkerConfig struct {
	MaxContainerMemoryMB  int
	TotalContainerMemoryMB int
}

type WebhookConfig struct {
	Secret string
}

func Load() (*Config, error) {
	cfg := &Config{
		API: APIConfig{
			Addr: getEnv("API_ADDR", ":8080"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", ""),
		},
		GitHub: GitHubConfig{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
		Domain: DomainConfig{
			Default: getEnv("DEFAULT_DOMAIN", "pontoon.example.com"),
		},
		Worker: WorkerConfig{
			MaxContainerMemoryMB:  getEnvInt("MAX_CONTAINER_MEMORY_MB", 256),
			TotalContainerMemoryMB: getEnvInt("TOTAL_CONTAINER_MEMORY_MB", 2048),
		},
		Webhook: WebhookConfig{
			Secret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	if c.GitHub.ClientID == "" {
		return fmt.Errorf("GITHUB_CLIENT_ID is required")
	}
	if c.GitHub.ClientSecret == "" {
		return fmt.Errorf("GITHUB_CLIENT_SECRET is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvList(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	parts := strings.Split(valueStr, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
