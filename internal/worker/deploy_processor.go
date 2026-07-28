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
}

func NewDeployProcessor(
	dockerClient *docker.Client,
	deploymentRepo domain.DeploymentRepository,
	projectRepo domain.ProjectRepository,
	envVarRepo domain.EnvVarRepository,
	capacityUC *usecase.CapacityUseCase,
	maxMemoryMB int,
) *DeployProcessor {
	return &DeployProcessor{
		dockerClient:   dockerClient,
		deploymentRepo: deploymentRepo,
		projectRepo:    projectRepo,
		envVarRepo:     envVarRepo,
		capacityUC:     capacityUC,
		maxMemoryMB:    maxMemoryMB,
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
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("capacity check failed: %w", err))
	}

	// Update status to cloning
	if err := p.deploymentRepo.UpdateStatus(payload.DeploymentID, domain.DeploymentStatusCloning, ""); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Create temp directory for clone
	tmpDir, err := os.MkdirTemp("", "pontoon-build-*")
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("failed to create temp dir: %w", err))
	}
	defer os.RemoveAll(tmpDir)

	// Clone repository
	repoDir := filepath.Join(tmpDir, "repo")
	if err := p.cloneRepo(ctx, payload.RepoURL, payload.Branch, repoDir); err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("failed to clone repo: %w", err))
	}

	// Update status to building
	if err := p.deploymentRepo.UpdateStatus(payload.DeploymentID, domain.DeploymentStatusBuilding, ""); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Build Docker image
	imageTag := fmt.Sprintf("pontoon/%s:%s", payload.ProjectID, payload.DeploymentID)
	buildOutput, err := p.dockerClient.BuildImage(ctx, docker.BuildConfig{
		WorkDir:  repoDir,
		ImageTag: imageTag,
	})
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, imageTag, "", fmt.Errorf("failed to build image: %w", err))
	}

	slog.Info("image built", "image", imageTag, "output_length", len(buildOutput))

	// Truncate build logs to 100KB / 500 lines
	truncatedLogs := p.truncateLogs(buildOutput)

	// Update status to running
	if err := p.deploymentRepo.UpdateStatus(payload.DeploymentID, domain.DeploymentStatusRunning, truncatedLogs); err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, imageTag, "", fmt.Errorf("failed to update status: %w", err))
	}

	// Get project details
	project, err := p.projectRepo.GetByID(payload.ProjectID)
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, imageTag, "", fmt.Errorf("failed to get project: %w", err))
	}

	// Get environment variables
	envVars, err := p.envVarRepo.GetByProjectID(payload.ProjectID)
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, imageTag, "", fmt.Errorf("failed to get env vars: %w", err))
	}

	// Merge env vars
	envMap := make(map[string]string)
	for _, env := range envVars {
		envMap[env.Key] = env.Value
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
		return p.failDeployment(ctx, payload.DeploymentID, imageTag, "", fmt.Errorf("failed to run container: %w", err))
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

func (p *DeployProcessor) failDeployment(ctx context.Context, deploymentID, imageTag, containerID string, err error) error {
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

	// Truncate error message to fit in database
	errMsg := err.Error()
	if len(errMsg) > 1000 {
		errMsg = errMsg[:1000] + "... (truncated)"
	}

	if updateErr := p.deploymentRepo.UpdateStatus(deploymentID, domain.DeploymentStatusFailed, errMsg); updateErr != nil {
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
