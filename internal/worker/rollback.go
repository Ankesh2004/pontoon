package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
	"github.com/Ankesh2004/pontoon/internal/tasks"
)

func (p *DeployProcessor) ProcessRollbackTask(ctx context.Context, t *asynq.Task) error {
	payload, err := tasks.UnmarshalRollbackPayload(t.Payload())
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	slog.Info("processing rollback task", "deployment_id", payload.DeploymentID, "project_id", payload.ProjectID, "target_image", payload.TargetImage)

	if err := p.deploymentRepo.UpdateStatus(payload.DeploymentID, domain.DeploymentStatusRunning, ""); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// 1. Check Capacity
	if err := p.capacityUC.CheckCapacity(ctx, p.maxMemoryMB); err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", err, "")
	}

	// 2. Acquire Project Lock to prevent concurrent deployments
	lockKey := fmt.Sprintf("deploy_lock:%s", payload.ProjectID)
	acquired, err := p.redisClient.SetNX(ctx, lockKey, payload.DeploymentID, 5*time.Minute).Result()
	if err != nil || !acquired {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("failed to acquire deployment lock: another deployment is in progress"), "")
	}
	defer p.redisClient.Del(ctx, lockKey)

	// Fetch environment variables
	envVars, err := p.envVarRepo.GetByProjectID(payload.ProjectID)
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("failed to get env vars: %w", err), "")
	}
	envMap := make(map[string]string)
	for _, ev := range envVars {
		envMap[ev.Key] = ev.Value
	}

	project, err := p.projectRepo.GetByID(payload.ProjectID)
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("failed to get project: %w", err), "")
	}

	containerName := fmt.Sprintf("pontoon-%s", payload.DeploymentID)

	// 3. Run Container using the existing image
	slog.Info("starting rollback container", "image", payload.TargetImage, "container_name", containerName)
	containerID, err := p.dockerClient.RunContainer(ctx, docker.RunConfig{
		Image:         payload.TargetImage,
		ContainerName: containerName,
		EnvVars:       envMap,
		TenantID:      payload.UserID,
		ProjectID:     payload.ProjectID,
		ProjectName:   project.Name,
		DeploymentID:  payload.DeploymentID,
		Domain:        payload.Domain,
		MemoryLimitMB: p.maxMemoryMB,
		CPULimit:      1.0,
	})
	if err != nil {
		return p.failDeployment(ctx, payload.DeploymentID, "", "", fmt.Errorf("failed to run container: %w", err), "")
	}

	// 4. LIVENESS CHECK: Wait 3 seconds and ensure it hasn't crashed.
	time.Sleep(StartupLivenessDelay)
	inspect, err := p.dockerClient.ContainerInspect(ctx, containerID)
	if err == nil && !inspect.State.Running {
		crashLogs, _ := p.dockerClient.GetContainerLogs(ctx, containerID, 100)
		finalLogs := "--- RUNTIME CRASH (Startup Failed) ---\n" + crashLogs
		return p.failDeployment(ctx, payload.DeploymentID, "", containerID, fmt.Errorf("container crashed immediately after startup"), finalLogs)
	}

	// Safety net: check if this deployment was superseded while we were waiting
	currentDep, err := p.deploymentRepo.GetByID(payload.DeploymentID)
	if err == nil && currentDep.Status == domain.DeploymentStatusFailed {
		slog.Info("rollback was superseded, cleaning up", "deployment_id", payload.DeploymentID)
		_ = p.dockerClient.StopContainer(ctx, containerID)
		_ = p.dockerClient.RemoveContainer(ctx, containerID)
		return nil
	}

	// 5. Stop existing live deployments for zero-downtime transition
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

	// 6. Update deployment with success
	deployment := &domain.Deployment{
		ID:            payload.DeploymentID,
		Status:        domain.DeploymentStatusLive,
		DockerImage:   payload.TargetImage,
		ContainerID:   containerID,
		ContainerName: containerName,
		CommitSHA:     currentDep.CommitSHA,
		BuildLogs:     "Rollback successful (bypassed build phase)",
	}

	if err := p.deploymentRepo.Update(deployment); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	slog.Info("rollback successful", "deployment_id", payload.DeploymentID, "container_id", containerID)
	return nil
}
