package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
	"github.com/Ankesh2004/pontoon/internal/tasks"
)

type DeploymentUseCase struct {
	deploymentRepo domain.DeploymentRepository
	projectRepo    domain.ProjectRepository
	asynqClient    *asynq.Client
	dockerClient   *docker.Client
}

func NewDeploymentUseCase(
	deploymentRepo domain.DeploymentRepository,
	projectRepo domain.ProjectRepository,
	asynqClient *asynq.Client,
	dockerClient *docker.Client,
) *DeploymentUseCase {
	return &DeploymentUseCase{
		deploymentRepo: deploymentRepo,
		projectRepo:    projectRepo,
		asynqClient:    asynqClient,
		dockerClient:   dockerClient,
	}
}

type TriggerDeploymentInput struct {
	UserID    string
	ProjectID string
	CommitSHA string
}

func (uc *DeploymentUseCase) TriggerDeployment(ctx context.Context, input TriggerDeploymentInput) (*domain.Deployment, error) {
	project, err := uc.projectRepo.GetByID(input.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project.UserID != input.UserID {
		return nil, domain.ErrForbidden
	}

	// Cancel any in-progress deployments for this project so we don't end
	// up with two containers fighting over the same Traefik Host() rule.
	if activeDeployments, err := uc.deploymentRepo.GetActiveByProjectID(input.ProjectID); err == nil {
		for _, activeDep := range activeDeployments {
			slog.Info("superseding active deployment", "old_deployment_id", activeDep.ID, "project_id", input.ProjectID)
			_ = uc.deploymentRepo.UpdateStatus(activeDep.ID, domain.DeploymentStatusFailed, "Cancelled: superseded by a newer deployment")
		}
	}

	deployment := &domain.Deployment{
		ID:         uuid.New().String(),
		ProjectID:  input.ProjectID,
		UserID:     input.UserID,
		Status:     domain.DeploymentStatusPending,
		CommitSHA:  input.CommitSHA,
	}

	if err := uc.deploymentRepo.Create(deployment); err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	payload := &tasks.DeployPayload{
		DeploymentID: deployment.ID,
		ProjectID:    input.ProjectID,
		UserID:       input.UserID,
		RepoURL:      project.RepoURL,
		Branch:       project.Branch,
		CommitSHA:    input.CommitSHA,
		Domain:       project.Domain,
	}

	payloadBytes, err := payload.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(tasks.TypeDeploy, payloadBytes)
	if _, err := uc.asynqClient.Enqueue(task); err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return deployment, nil
}

func (uc *DeploymentUseCase) GetDeployment(ctx context.Context, userID, deploymentID string) (*domain.Deployment, error) {
	deployment, err := uc.deploymentRepo.GetByID(deploymentID)
	if err != nil {
		return nil, err
	}

	if deployment.UserID != userID {
		return nil, domain.ErrForbidden
	}

	return deployment, nil
}

func (uc *DeploymentUseCase) ListDeployments(ctx context.Context, userID, projectID string) ([]*domain.Deployment, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden
	}

	return uc.deploymentRepo.GetByProjectID(projectID)
}

func (uc *DeploymentUseCase) StopDeployment(ctx context.Context, userID, deploymentID string) error {
	deployment, err := uc.deploymentRepo.GetByID(deploymentID)
	if err != nil {
		return err
	}

	if deployment.UserID != userID {
		return domain.ErrForbidden
	}

	if deployment.ContainerID != "" {
		if err := uc.dockerClient.StopContainer(ctx, deployment.ContainerID); err != nil {
			return fmt.Errorf("failed to stop container: %w", err)
		}
	}

	deployment.Status = domain.DeploymentStatusStopped
	if err := uc.deploymentRepo.Update(deployment); err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}

func (uc *DeploymentUseCase) DeleteDeployment(ctx context.Context, userID, deploymentID string) error {
	deployment, err := uc.deploymentRepo.GetByID(deploymentID)
	if err != nil {
		return err
	}

	if deployment.UserID != userID {
		return domain.ErrForbidden
	}

	if deployment.ContainerID != "" {
		_ = uc.dockerClient.StopContainer(ctx, deployment.ContainerID)
		_ = uc.dockerClient.RemoveContainer(ctx, deployment.ContainerID)
	}

	if deployment.DockerImage != "" {
		// Best effort image removal
		go func() {
			_ = uc.dockerClient.RemoveImage(context.Background(), deployment.DockerImage)
		}()
	}

	if err := uc.deploymentRepo.Delete(deploymentID); err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	return nil
}
