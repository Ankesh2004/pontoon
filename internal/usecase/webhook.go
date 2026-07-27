package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/tasks"
)

type WebhookUseCase struct {
	projectRepo    domain.ProjectRepository
	deploymentRepo domain.DeploymentRepository
	envVarRepo     domain.EnvVarRepository
	asynqClient    *asynq.Client
}

func NewWebhookUseCase(
	projectRepo domain.ProjectRepository,
	deploymentRepo domain.DeploymentRepository,
	envVarRepo domain.EnvVarRepository,
	asynqClient *asynq.Client,
) *WebhookUseCase {
	return &WebhookUseCase{
		projectRepo:    projectRepo,
		deploymentRepo: deploymentRepo,
		envVarRepo:     envVarRepo,
		asynqClient:    asynqClient,
	}
}

func (uc *WebhookUseCase) GetWebhookSecret(projectID string) (string, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	if project.WebhookSecret == "" {
		return "", fmt.Errorf("webhook secret not configured for project")
	}

	return project.WebhookSecret, nil
}

func (uc *WebhookUseCase) TriggerDeploymentFromWebhook(
	ctx context.Context,
	projectID string,
	branch string,
	commitSHA string,
) (*domain.Deployment, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project.Branch != branch {
		return nil, fmt.Errorf("branch mismatch: expected %s, got %s", project.Branch, branch)
	}

	envVars, err := uc.envVarRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get env vars: %w", err)
	}

	envMap := make(map[string]string)
	for _, env := range envVars {
		envMap[env.Key] = env.Value
	}

	deployment := &domain.Deployment{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		UserID:      project.UserID,
		CommitSHA:   commitSHA,
		Status:      domain.DeploymentStatusPending,
		TriggeredBy: "webhook",
	}

	if err := uc.deploymentRepo.Create(deployment); err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	payload := &tasks.DeployPayload{
		DeploymentID: deployment.ID,
		ProjectID:    projectID,
		UserID:       project.UserID,
		RepoURL:      project.RepoURL,
		Branch:       project.Branch,
		CommitSHA:    commitSHA,
		Domain:       project.Domain,
		EnvVars:      envMap,
	}

	payloadBytes, err := payload.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(tasks.TypeDeploy, payloadBytes)
	if _, err := uc.asynqClient.Enqueue(task); err != nil {
		return nil, fmt.Errorf("failed to enqueue deployment task: %w", err)
	}

	return deployment, nil
}
