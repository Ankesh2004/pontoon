package worker

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

// ProcessPruneImagesTask finds deployments beyond the 5th most recent for each project
// and deletes their corresponding Docker images to save disk space.
func (p *DeployProcessor) ProcessPruneImagesTask(ctx context.Context, t *asynq.Task) error {
	slog.Info("starting image pruner job")

	projects, err := p.projectRepo.GetAll()
	if err != nil {
		slog.Error("failed to get all projects for pruning", "error", err)
		return err
	}

	totalPruned := 0

	for _, project := range projects {
		deployments, err := p.deploymentRepo.GetByProjectID(project.ID)
		if err != nil {
			slog.Warn("failed to get deployments for project", "project_id", project.ID, "error", err)
			continue
		}

		// Keep the first 5 (which are ordered by created_at DESC)
		retentionLimit := 5
		if len(deployments) <= retentionLimit {
			continue
		}

		// Prune the rest
		for i := retentionLimit; i < len(deployments); i++ {
			dep := deployments[i]
			
			// Only prune if it has an image and it's not live (just in case)
			if dep.DockerImage != "" && dep.Status != "live" {
				slog.Info("pruning old docker image", "project_id", project.ID, "deployment_id", dep.ID, "image", dep.DockerImage)
				
				err := p.dockerClient.RemoveImage(ctx, dep.DockerImage)
				if err != nil {
					slog.Warn("failed to remove docker image", "image", dep.DockerImage, "error", err)
					// If the image doesn't exist anymore, Docker returns an error, but we should still clear the DB field
				}
				
				// Clear the image field in DB so we don't try to prune it again or rollback to it
				dep.DockerImage = ""
				if err := p.deploymentRepo.Update(dep); err != nil {
					slog.Warn("failed to clear docker_image field in db after pruning", "deployment_id", dep.ID, "error", err)
				} else {
					totalPruned++
				}
			}
		}
	}

	slog.Info("image pruner job completed", "total_pruned", totalPruned)
	return nil
}
