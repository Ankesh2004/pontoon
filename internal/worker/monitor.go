package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/hibiken/asynq"

	"github.com/Ankesh2004/pontoon/internal/domain"
)

func (p *DeployProcessor) ProcessMonitorHealthTask(ctx context.Context, t *asynq.Task) error {
	slog.Info("starting health monitor task")

	// Find all dead/exited containers managed by pontoon
	f := filters.NewArgs()
	f.Add("label", "pontoon.managed=true")
	f.Add("status", "exited")
	f.Add("status", "dead")

	containers, err := p.dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, ctr := range containers {
		deploymentID, ok := ctr.Labels["pontoon.deployment-id"]
		if !ok || deploymentID == "" {
			continue // Legacy container without label
		}

		// Check the deployment in the database
		deployment, err := p.deploymentRepo.GetByID(deploymentID)
		if err != nil {
			slog.Warn("health monitor: failed to get deployment", "deployment_id", deploymentID, "error", err)
			continue
		}

		// If the database thinks it's live, but it's actually exited, it crashed!
		if deployment.Status == domain.DeploymentStatusLive {
			slog.Warn("health monitor: detected crashed deployment", "deployment_id", deployment.ID, "container_id", ctr.ID)

			// Try to grab the tail of the logs
			var runtimeError string
			logs, logErr := p.dockerClient.GetContainerLogs(ctx, ctr.ID, 100)
			if logErr == nil && len(logs) > 0 {
				runtimeError = string(logs)
			}

			// Update the deployment status
			deployment.Status = domain.DeploymentStatusFailed
			
			// We append the runtime error to the build logs so the user can see it in the UI
			if runtimeError != "" {
				deployment.BuildLogs = deployment.BuildLogs + "\n\n--- RUNTIME CRASH LOGS ---\n" + runtimeError
			} else {
				deployment.BuildLogs = deployment.BuildLogs + "\n\n--- RUNTIME CRASH ---\nContainer exited unexpectedly with no logs."
			}

			if err := p.deploymentRepo.Update(deployment); err != nil {
				slog.Error("health monitor: failed to update crashed deployment status", "deployment_id", deployment.ID, "error", err)
			} else {
				slog.Info("health monitor: marked deployment as failed", "deployment_id", deployment.ID)
			}
		}
	}

	return nil
}
