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
)

type DeployProcessor struct {
	dockerClient   *docker.Client
	deploymentRepo domain.DeploymentRepository
	projectRepo    domain.ProjectRepository
	envVarRepo     domain.EnvVarRepository
}

func NewDeployProcessor(
	dockerClient *docker.Client,
	deploymentRepo domain.DeploymentRepository,
	projectRepo domain.ProjectRepository,
	envVarRepo domain.EnvVarRepository,
) *DeployProcessor {
	return &DeployProcessor{
		dockerClient:   dockerClient,
		deploymentRepo: deploymentRepo,
		projectRepo:    projectRepo,
		envVarRepo:     envVarRepo,
	}
}

func (p *DeployProcessor) ProcessDeployTask(ctx context.Context, t *asynq.Task) error {
	payload, err := UnmarshalDeployPayload(t.Payload())
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	slog.Info("processing deployment", "deployment_id", payload.DeploymentID)

	// Update status to cloning
	if err := p.deploymentRepo.UpdateStatus(payload.DeploymentID, domain.DeploymentStatusCloning, ""); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Create temp directory for clone
	tmpDir, err := os.MkdirTemp("", "pontoon-build-*")
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, fmt.Errorf("failed to create temp dir: %w", err))
	}
	defer os.RemoveAll(tmpDir)

	// Clone repository
	repoDir := filepath.Join(tmpDir, "repo")
	if err := p.cloneRepo(ctx, payload.RepoURL, payload.Branch, repoDir); err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, fmt.Errorf("failed to clone repo: %w", err))
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
		return p.failDeployment(ctx, payload.DeploymentID, fmt.Errorf("failed to build image: %w", err))
	}

	slog.Info("image built", "image", imageTag, "output_length", len(buildOutput))

	// Update status to running
	if err := p.deploymentRepo.UpdateStatus(payload.DeploymentID, domain.DeploymentStatusRunning, ""); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Get project details
	project, err := p.projectRepo.GetByID(payload.ProjectID)
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, fmt.Errorf("failed to get project: %w", err))
	}

	// Get environment variables
	envVars, err := p.envVarRepo.GetByProjectID(payload.ProjectID)
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, fmt.Errorf("failed to get env vars: %w", err))
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
		MemoryLimitMB: 128,
		CPULimit:      0.5,
	})
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, fmt.Errorf("failed to run container: %w", err))
	}

	// Update deployment with success
	deployment := &domain.Deployment{
		ID:            payload.DeploymentID,
		Status:        domain.DeploymentStatusLive,
		DockerImage:   imageTag,
		ContainerID:   containerID,
		ContainerName: containerName,
		CommitSHA:     payload.CommitSHA,
		BuildLogs:     buildOutput,
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

func (p *DeployProcessor) failDeployment(ctx context.Context, deploymentID string, err error) error {
	slog.Error("deployment failed", "deployment_id", deploymentID, "error", err)
	
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

func (p *DeployProcessor) truncateLogs(logs string, maxBytes int) string {
	if len(logs) <= maxBytes {
		return logs
	}
	
	// Keep last portion of logs
	truncated := logs[len(logs)-maxBytes:]
	
	// Find first newline to avoid cutting mid-line
	if idx := strings.Index(truncated, "\n"); idx != -1 {
		truncated = truncated[idx+1:]
	}
	
	return "[LOGS TRUNCATED - showing last " + fmt.Sprintf("%d", maxBytes) + " bytes]\n" + truncated
}
