package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type LogPublisher struct {
	client *redis.Client
}

type LogMessage struct {
	DeploymentID string `json:"deployment_id"`
	Line         string `json:"line"`
	Timestamp    int64  `json:"timestamp"`
}

func NewLogPublisher(redisURL string) (*LogPublisher, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &LogPublisher{client: client}, nil
}

func (p *LogPublisher) PublishLog(ctx context.Context, deploymentID, line string) error {
	msg := LogMessage{
		DeploymentID: deploymentID,
		Line:         line,
		Timestamp:    0,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal log message: %w", err)
	}

	channel := fmt.Sprintf("deployment:%s:logs", deploymentID)
	return p.client.Publish(ctx, channel, data).Err()
}

func (p *LogPublisher) Close() error {
	return p.client.Close()
}
