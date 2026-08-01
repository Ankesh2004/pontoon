package docker

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	goredis "github.com/redis/go-redis/v9"
)

type BuildConfig struct {
	RepoURL      string
	Branch       string
	ImageTag     string
	WorkDir      string
	DeploymentID string
	RedisClient  *goredis.Client
}

func (c *Client) BuildImage(ctx context.Context, cfg BuildConfig) (string, error) {
	buildContext, err := createBuildContext(cfg.WorkDir)
	if err != nil {
		return "", fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContext.Close()

	buildOptions := types.ImageBuildOptions{
		Tags:       []string{cfg.ImageTag},
		Dockerfile: "Dockerfile",
		Remove:     true,
		NoCache:    false,
	}

	resp, err := c.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	var logs bytes.Buffer
	scanner := bufio.NewScanner(resp.Body)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		var buildOutput struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		
		var logText string
		if err := json.Unmarshal([]byte(line), &buildOutput); err == nil {
			if buildOutput.Stream != "" {
				logText = buildOutput.Stream
			} else if buildOutput.Error != "" {
				logText = "ERROR: " + buildOutput.Error + "\n"
			}
		}
		
		if logText == "" {
			logText = line + "\n"
		}

		logs.WriteString(logText)
		
		// Publish to Redis in real-time if deployment ID and Redis client are provided
		if cfg.DeploymentID != "" && cfg.RedisClient != nil {
			streamLine := strings.TrimSuffix(logText, "\n")
			if streamLine != "" {
				PublishLog(ctx, cfg.RedisClient, cfg.DeploymentID, streamLine)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read build response: %w", err)
	}

	return logs.String(), nil
}

func PublishLog(ctx context.Context, redisClient *goredis.Client, deploymentID, line string) {
	channel := fmt.Sprintf("deployment:%s:logs", deploymentID)
	histKey  := fmt.Sprintf("deployment:%s:log_history", deploymentID)
	message := map[string]interface{}{
		"deployment_id": deploymentID,
		"line":          line,
		"timestamp":     time.Now().Unix(),
	}

	if data, err := json.Marshal(message); err == nil {
		// fire-and-forget pub/sub for live clients
		redisClient.Publish(ctx, channel, data)
		// also persist to a list so late joiners can replay
		redisClient.RPush(ctx, histKey, data)
		redisClient.Expire(ctx, histKey, time.Hour)
	}
}

func createBuildContext(contextDir string) (io.ReadCloser, error) {
	if _, err := os.Stat(contextDir); err != nil {
		return nil, fmt.Errorf("context directory does not exist: %w", err)
	}

	return archive.Tar(contextDir, archive.Uncompressed)
}

func (c *Client) RemoveImage(ctx context.Context, imageID string) error {
	_, err := c.ImageRemove(ctx, imageID, types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	return err
}

func createTarFromDir(srcDir string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})

	return buf, err
}
