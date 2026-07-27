package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/Ankesh2004/pontoon/internal/domain"
)

type DeploymentUseCase struct {
	deploymentRepo domain.DeploymentRepository
	projectRepo    domain.ProjectRepository
	asynqClient    *asynq.Client
}

func NewDeploymentUseCase(
	deploymentRepo domain.DeploymentRepository,
	projectRepo domain.ProjectRepository,
	asynqClient *asynq.Client,
) *DeploymentUseCase {
	return &DeploymentUseCase{
		deploymentRepo: deploymentRepo,
		projectRepo:    projectRepo,
		asynqClient:    asynqClient,
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

	payload := map[string]interface{}{
		"deployment_id": deployment.ID,
		"project_id":    input.ProjectID,
		"user_id":       input.UserID,
		"repo_url":      project.RepoURL,
		"branch":        project.Branch,
		"commit_sha":    input.CommitSHA,
		"domain":        project.Domain,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask("deploy", payloadBytes)
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
