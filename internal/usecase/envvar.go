package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Ankesh2004/pontoon/internal/domain"
)

type EnvVarUseCase struct {
	envVarRepo  domain.EnvVarRepository
	projectRepo domain.ProjectRepository
}

func NewEnvVarUseCase(envVarRepo domain.EnvVarRepository, projectRepo domain.ProjectRepository) *EnvVarUseCase {
	return &EnvVarUseCase{
		envVarRepo:  envVarRepo,
		projectRepo: projectRepo,
	}
}

func (uc *EnvVarUseCase) CreateEnvVar(ctx context.Context, userID, projectID, key, value string) (*domain.EnvVar, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden
	}

	envVar := &domain.EnvVar{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		Key:       key,
		Value:     value,
	}

	if err := uc.envVarRepo.Create(envVar); err != nil {
		return nil, fmt.Errorf("failed to create env var: %w", err)
	}

	return envVar, nil
}

func (uc *EnvVarUseCase) ListEnvVars(ctx context.Context, userID, projectID string) ([]*domain.EnvVar, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden
	}

	envVars, err := uc.envVarRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list env vars: %w", err)
	}

	return envVars, nil
}

func (uc *EnvVarUseCase) DeleteEnvVar(ctx context.Context, userID, projectID, envVarID string) error {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if project.UserID != userID {
		return domain.ErrForbidden
	}

	envVar, err := uc.envVarRepo.GetByID(envVarID)
	if err != nil {
		return fmt.Errorf("failed to get env var: %w", err)
	}

	if envVar.ProjectID != projectID {
		return domain.ErrNotFound
	}

	if err := uc.envVarRepo.Delete(envVarID); err != nil {
		return fmt.Errorf("failed to delete env var: %w", err)
	}

	return nil
}
