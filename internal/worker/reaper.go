package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/Ankesh2004/pontoon/internal/domain"
)

// ProcessReapStuckTask finds active deployments older than 30 minutes and marks them as failed.
func (p *DeployProcessor) ProcessReapStuckTask(ctx context.Context, t *asynq.Task) error {
	slog.Info("starting stuck deployments reaper job")

	// Get deployments stuck for more than 30 minutes
	stuckDeployments, err := p.deploymentRepo.GetStuck(30)
	if err != nil {
		slog.Error("failed to get stuck deployments", "error", err)
		return fmt.Errorf("failed to get stuck deployments: %w", err)
	}

	if len(stuckDeployments) == 0 {
		slog.Info("no stuck deployments found")
		return nil
	}

	for _, dep := range stuckDeployments {
		slog.Info("reaping stuck deployment", "deployment_id", dep.ID, "project_id", dep.ProjectID, "status", dep.Status)
		
		// 1. Mark as failed
		err := p.deploymentRepo.UpdateStatus(dep.ID, domain.DeploymentStatusFailed, "Deployment timed out or worker crashed unexpectedly.")
		if err != nil {
			slog.Error("failed to update stuck deployment status", "deployment_id", dep.ID, "error", err)
			continue
		}

		// 2. Release the project lock if it still exists
		lockKey := fmt.Sprintf("deploy_lock:%s", dep.ProjectID)
		if err := p.redisClient.Del(ctx, lockKey).Err(); err != nil {
			slog.Warn("failed to delete stuck project lock", "project_id", dep.ProjectID, "error", err)
		}
	}

	slog.Info("reaper job completed", "reaped_count", len(stuckDeployments))
	return nil
}
