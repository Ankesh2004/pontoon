package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hibiken/asynq"
	goredis "github.com/redis/go-redis/v9"
	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
	"github.com/Ankesh2004/pontoon/internal/tasks"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

type DeployProcessor struct {
	dockerClient   *docker.Client
	deploymentRepo domain.DeploymentRepository
	projectRepo    domain.ProjectRepository
	envVarRepo     domain.EnvVarRepository
	capacityUC     *usecase.CapacityUseCase
	maxMemoryMB    int
	redisClient    *goredis.Client
}

func NewDeployProcessor(
	dockerClient *docker.Client,
	deploymentRepo domain.DeploymentRepository,
	projectRepo domain.ProjectRepository,
	envVarRepo domain.EnvVarRepository,
	capacityUC *usecase.CapacityUseCase,
	maxMemoryMB int,
	redisClient *goredis.Client,
) *DeployProcessor {
	return &DeployProcessor{
		dockerClient:   dockerClient,
		deploymentRepo: deploymentRepo,
		projectRepo:    projectRepo,
		envVarRepo:     envVarRepo,
		capacityUC:     capacityUC,
		maxMemoryMB:    maxMemoryMB,
		redisClient:    redisClient,
	}
}

func (p *DeployProcessor) ProcessDeployTask(ctx context.Context, t *asynq.Task) error {
	payload, err := tasks.UnmarshalDeployPayload(t.Payload())
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	slog.Info("processing deployment", "deployment_id", payload.DeploymentID)

	// Check capacity before starting
	if err := p.capacityUC.CheckCapacity(ctx, p.maxMemoryMB); err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("capacity check failed: %w", err), "")
	}

	// Update status to cloning
	if err := p.deploymentRepo.UpdateStatus(payload.DeploymentID, domain.DeploymentStatusCloning, ""); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	docker.PublishLog(ctx, p.redisClient, payload.DeploymentID, "==> Cloning repository...")
	docker.PublishLog(ctx, p.redisClient, payload.DeploymentID, "    Repo: "+payload.RepoURL)
	docker.PublishLog(ctx, p.redisClient, payload.DeploymentID, "    Branch: "+payload.Branch)

	// Create temp directory for clone
	tmpDir, err := os.MkdirTemp("", "pontoon-build-*")
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("failed to create temp dir: %w", err), "")
	}
	defer os.RemoveAll(tmpDir)

	// Clone repository
	repoDir := filepath.Join(tmpDir, "repo")
	if err := p.cloneRepo(ctx, payload.RepoURL, payload.Branch, repoDir); err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("failed to clone repo: %w", err), "")
	}
	docker.PublishLog(ctx, p.redisClient, payload.DeploymentID, "==> Clone complete.")

	// Resolve actual commit SHA from the cloned repo
	revCmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	revCmd.Dir = repoDir
	if out, err := revCmd.Output(); err == nil {
		payload.CommitSHA = strings.TrimSpace(string(out))
	}

	// Update status to building
	if err := p.deploymentRepo.UpdateStatus(payload.DeploymentID, domain.DeploymentStatusBuilding, ""); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	docker.PublishLog(ctx, p.redisClient, payload.DeploymentID, "==> Building Docker image...")

	// Get environment variables
	envVars, err := p.envVarRepo.GetByProjectID(payload.ProjectID)
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("failed to get env vars: %w", err), "")
	}

	// Merge env vars
	envMap := make(map[string]string)
	for _, env := range envVars {
		envMap[env.Key] = env.Value
	}

	// Build Docker image
	imageTag := fmt.Sprintf("pontoon/%s:%s", payload.ProjectID, payload.DeploymentID)
	buildOutput, err := p.dockerClient.BuildImage(ctx, docker.BuildConfig{
		WorkDir:      repoDir,
		ImageTag:     imageTag,
		DeploymentID: payload.DeploymentID,
		RedisClient:  p.redisClient,
		EnvVars:      envMap,
	})
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, imageTag, "", fmt.Errorf("failed to build image: %w", err), buildOutput)
	}

	slog.Info("image built", "image", imageTag, "output_length", len(buildOutput))

	// Truncate build logs to 100KB / 500 lines
	truncatedLogs := p.truncateLogs(buildOutput)

	// Update status to running
	if err := p.deploymentRepo.UpdateStatus(payload.DeploymentID, domain.DeploymentStatusRunning, truncatedLogs); err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, imageTag, "", fmt.Errorf("failed to update status: %w", err), truncatedLogs)
	}

	// Get project details
	project, err := p.projectRepo.GetByID(payload.ProjectID)
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, imageTag, "", fmt.Errorf("failed to get project: %w", err), truncatedLogs)
	}

	// Run container
	containerName := fmt.Sprintf("pontoon-%s-%s", payload.ProjectID[:8], payload.DeploymentID[:8])
	containerID, err := p.dockerClient.RunContainer(ctx, docker.RunConfig{
		Image:         imageTag,
		ContainerName: containerName,
		EnvVars:       envMap,
		TenantID:      payload.UserID,
		ProjectID:     payload.ProjectID,
		ProjectName:   project.Name,
		Domain:        payload.Domain,
		MemoryLimitMB: p.maxMemoryMB,
		CPULimit:      0.5,
	})
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, imageTag, "", fmt.Errorf("failed to run container: %w", err), truncatedLogs)
	}

	// Safety net: check if this deployment was superseded while we were building.
	// A newer TriggerDeployment call would have marked us as "failed" in the DB.
	currentDep, err := p.deploymentRepo.GetByID(payload.DeploymentID)
	if err == nil && currentDep.Status == domain.DeploymentStatusFailed {
		slog.Info("deployment was superseded during build, cleaning up", "deployment_id", payload.DeploymentID)
		_ = p.dockerClient.StopContainer(ctx, containerID)
		_ = p.dockerClient.RemoveContainer(ctx, containerID)
		_ = p.dockerClient.RemoveImage(ctx, imageTag)
		return nil
	}

	// Stop existing live deployments for zero-downtime transition
	if oldDeployments, err := p.deploymentRepo.GetByProjectID(payload.ProjectID); err == nil {
		for _, oldDep := range oldDeployments {
			if oldDep.ID != payload.DeploymentID && oldDep.Status == domain.DeploymentStatusLive {
				slog.Info("stopping old deployment", "deployment_id", oldDep.ID, "container_id", oldDep.ContainerID)
				_ = p.dockerClient.StopContainer(ctx, oldDep.ContainerID)
				_ = p.dockerClient.RemoveContainer(ctx, oldDep.ContainerID)
				_ = p.deploymentRepo.UpdateStatus(oldDep.ID, domain.DeploymentStatusStopped, "")
			}
		}
	}

	// Update deployment with success
	deployment := &domain.Deployment{
		ID:            payload.DeploymentID,
		Status:        domain.DeploymentStatusLive,
		DockerImage:   imageTag,
		ContainerID:   containerID,
		ContainerName: containerName,
		CommitSHA:     payload.CommitSHA,
		BuildLogs:     truncatedLogs,
	}

	if err := p.deploymentRepo.Update(deployment); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	slog.Info("deployment successful", "deployment_id", payload.DeploymentID, "container_id", containerID)
	return nil
}

func (p *DeployProcessor) cloneRepo(ctx context.Context, repoURL, branch, destDir string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", branch, repoURL, destDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}
	return nil
}

func (p *DeployProcessor) failDeployment(ctx context.Context, deploymentID, imageTag, containerID string, err error, buildLogs string) error {
	slog.Error("deployment failed", "deployment_id", deploymentID, "error", err)

	// Cleanup container if it was created
	if containerID != "" {
		slog.Info("cleaning up container", "container_id", containerID)
		if stopErr := p.dockerClient.StopContainer(ctx, containerID); stopErr != nil {
			slog.Warn("failed to stop container during cleanup", "error", stopErr)
		}
		if removeErr := p.dockerClient.RemoveContainer(ctx, containerID); removeErr != nil {
			slog.Warn("failed to remove container during cleanup", "error", removeErr)
		}
	}

	// Cleanup image if it was built
	if imageTag != "" {
		slog.Info("cleaning up image", "image", imageTag)
		if removeErr := p.dockerClient.RemoveImage(ctx, imageTag); removeErr != nil {
			slog.Warn("failed to remove image during cleanup", "error", removeErr)
		}
	}

	finalLogs := err.Error()
	if buildLogs != "" {
		finalLogs = err.Error() + "\n\nBuild Logs:\n" + buildLogs
	}
	finalLogs = p.truncateLogs(finalLogs)

	if updateErr := p.deploymentRepo.UpdateStatus(deploymentID, domain.DeploymentStatusFailed, finalLogs); updateErr != nil {
		slog.Error("failed to update deployment status", "error", updateErr)
	}

	return err
}

const (
	maxLogBytes = 100 * 1024 // 100KB
	maxLogLines = 500
)

func (p *DeployProcessor) truncateLogs(logs string) string {
	truncated := logs

	// Enforce 500-line limit first
	lines := strings.Split(truncated, "\n")
	if len(lines) > maxLogLines {
		lines = lines[len(lines)-maxLogLines:]
		truncated = strings.Join(lines, "\n")
	}

	// Then enforce 100KB byte limit
	if len(truncated) > maxLogBytes {
		truncated = truncated[len(truncated)-maxLogBytes:]
		// Avoid cutting mid-line
		if idx := strings.Index(truncated, "\n"); idx != -1 {
			truncated = truncated[idx+1:]
		}
	}

	if truncated != logs {
		return fmt.Sprintf("[LOGS TRUNCATED - max %d bytes / %d lines]\n", maxLogBytes, maxLogLines) + truncated
	}
	return truncated
}
