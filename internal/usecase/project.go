package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
)

type ProjectUseCase struct {
	projectRepo    domain.ProjectRepository
	deploymentRepo domain.DeploymentRepository
	dockerClient   *docker.Client
	defaultDomain  string
}

func NewProjectUseCase(
	projectRepo domain.ProjectRepository,
	deploymentRepo domain.DeploymentRepository,
	dockerClient *docker.Client,
	defaultDomain string,
) *ProjectUseCase {
	return &ProjectUseCase{
		projectRepo:    projectRepo,
		deploymentRepo: deploymentRepo,
		dockerClient:   dockerClient,
		defaultDomain:  defaultDomain,
	}
}

type CreateProjectInput struct {
	UserID    string
	Name      string
	RepoURL   string
	Branch    string
}

func (uc *ProjectUseCase) CreateProject(ctx context.Context, input CreateProjectInput) (*domain.Project, error) {
	owner, name, err := parseRepoURL(input.RepoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repo URL: %w", err)
	}

	webhookSecret, err := generateWebhookSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate webhook secret: %w", err)
	}

	project := &domain.Project{
		ID:            uuid.New().String(),
		UserID:        input.UserID,
		Name:          input.Name,
		RepoURL:       input.RepoURL,
		RepoOwner:     owner,
		RepoName:      name,
		Branch:        input.Branch,
		Domain:        fmt.Sprintf("%s.%s", input.Name, uc.defaultDomain),
		WebhookSecret: webhookSecret,
	}

	if err := uc.projectRepo.Create(project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

func generateWebhookSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (uc *ProjectUseCase) GetProject(ctx context.Context, userID, projectID string) (*domain.Project, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, err
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden
	}

	return project, nil
}

func (uc *ProjectUseCase) ListProjects(ctx context.Context, userID string) ([]*domain.Project, error) {
	return uc.projectRepo.GetByUserID(userID)
}

type UpdateProjectInput struct {
	Name   string
	Branch string
}

func (uc *ProjectUseCase) UpdateProject(ctx context.Context, userID, projectID string, input UpdateProjectInput) (*domain.Project, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, err
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden
	}

	if input.Name != "" {
		project.Name = input.Name
		project.Domain = fmt.Sprintf("%s.%s", input.Name, uc.defaultDomain)
	}
	if input.Branch != "" {
		project.Branch = input.Branch
	}

	if err := uc.projectRepo.Update(project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

func (uc *ProjectUseCase) DeleteProject(ctx context.Context, userID, projectID string) error {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return err
	}

	if project.UserID != userID {
		return domain.ErrForbidden
	}

	deployments, err := uc.deploymentRepo.GetByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get deployments: %w", err)
	}

	for _, deployment := range deployments {
		if deployment.ContainerID != "" && deployment.Status == domain.DeploymentStatusLive {
			slog.Info("stopping container for project deletion", "deployment_id", deployment.ID, "container_id", deployment.ContainerID)
			
			if err := uc.dockerClient.StopContainer(ctx, deployment.ContainerID); err != nil {
				slog.Warn("failed to stop container", "error", err, "container_id", deployment.ContainerID)
			}
			
			if err := uc.dockerClient.RemoveContainer(ctx, deployment.ContainerID); err != nil {
				slog.Warn("failed to remove container", "error", err, "container_id", deployment.ContainerID)
			}
		}
	}

	return uc.projectRepo.Delete(projectID)
}

func parseRepoURL(repoURL string) (owner, name string, err error) {
	// Handle GitHub URLs: https://github.com/owner/repo.git or https://github.com/owner/repo
	// Also handle SSH: git@github.com:owner/repo.git
	
	// Simple parsing for now - can be enhanced later
	var path string
	
	if len(repoURL) > 19 && repoURL[:19] == "https://github.com/" {
		path = repoURL[19:]
	} else if len(repoURL) > 15 && repoURL[:15] == "http://github.com/" {
		path = repoURL[15:]
	} else if len(repoURL) > 19 && repoURL[:19] == "git@github.com:" {
		path = repoURL[19:]
	} else {
		return "", "", fmt.Errorf("unsupported repo URL format: %s", repoURL)
	}
	
	// Remove .git suffix if present
	if len(path) > 4 && path[len(path)-4:] == ".git" {
		path = path[:len(path)-4]
	}
	
	// Split owner/repo
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			owner = path[:i]
			name = path[i+1:]
			break
		}
	}
	
	if owner == "" || name == "" {
		return "", "", fmt.Errorf("invalid repo URL format: %s", repoURL)
	}
	
	return owner, name, nil
}
