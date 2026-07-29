package usecase

import (
	"context"
	"fmt"

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
		if err := uc.dockerClient.RemoveContainer(ctx, deployment.ContainerID); err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}

	if err := uc.deploymentRepo.UpdateStatus(deploymentID, domain.DeploymentStatusStopped, deployment.BuildLogs); err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}
