package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"golang.org/x/oauth2"

	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/tasks"
	gh "github.com/Ankesh2004/pontoon/internal/infrastructure/github"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
)

type ProjectUseCase struct {
	projectRepo    domain.ProjectRepository
	deploymentRepo domain.DeploymentRepository
	envVarRepo     domain.EnvVarRepository
	userRepo       domain.UserRepository
	asynqClient    *asynq.Client
	dockerClient   *docker.Client
	defaultDomain  string
	apiURL         string
}

func NewProjectUseCase(
	projectRepo domain.ProjectRepository,
	deploymentRepo domain.DeploymentRepository,
	envVarRepo domain.EnvVarRepository,
	userRepo domain.UserRepository,
	asynqClient *asynq.Client,
	dockerClient *docker.Client,
	defaultDomain string,
	apiURL string,
) *ProjectUseCase {
	return &ProjectUseCase{
		projectRepo:    projectRepo,
		deploymentRepo: deploymentRepo,
		envVarRepo:     envVarRepo,
		userRepo:       userRepo,
		asynqClient:    asynqClient,
		dockerClient:   dockerClient,
		defaultDomain:  defaultDomain,
		apiURL:         apiURL,
	}
}

type CreateProjectInput struct {
	UserID    string
	Name      string
	RepoURL   string
	Branch    string
	Port      string
}

func (uc *ProjectUseCase) CreateProject(ctx context.Context, input CreateProjectInput) (*domain.Project, error) {
	user, err := uc.userRepo.GetByID(input.UserID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

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
		Domain:        fmt.Sprintf("%s-%s.%s", input.Name, strings.ToLower(user.GitHubUsername), uc.defaultDomain),
		WebhookSecret: webhookSecret,
	}

	if err := uc.projectRepo.Create(project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Create default PORT env var
	portVal := input.Port
	if portVal == "" {
		portVal = "8080"
	}
	
	_ = uc.envVarRepo.Create(&domain.EnvVar{
		ID:        uuid.New().String(),
		ProjectID: project.ID,
		Key:       "PORT",
		Value:     portVal,
	})

	// Try to auto-register webhook
	go func(u *domain.User) {
		if u.AccessToken == "" {
			slog.Warn("skipping webhook auto-registration: could not get user access token", "project_id", project.ID)
			return
		}

		// Create github client with user's token
		token := &oauth2.Token{AccessToken: u.AccessToken}
		ghClient := gh.NewClient(context.Background(), token)

		webhookURL := fmt.Sprintf("%s/webhooks/github?project_id=%s", uc.apiURL, project.ID)
		
		err = ghClient.CreateWebhook(context.Background(), owner, name, webhookURL, webhookSecret)
		if err != nil {
			slog.Warn("could not auto-register webhook", "project_id", project.ID, "error", err)
			return
		}

		slog.Info("successfully auto-registered webhook", "project_id", project.ID)
	}(user)

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

	user, err := uc.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if input.Name != "" {
		project.Name = input.Name
		project.Domain = fmt.Sprintf("%s-%s.%s", input.Name, strings.ToLower(user.GitHubUsername), uc.defaultDomain)
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
		if deployment.ContainerID != "" {
			slog.Info("stopping container for project deletion", "deployment_id", deployment.ID, "container_id", deployment.ContainerID)
			
			if err := uc.dockerClient.StopContainer(ctx, deployment.ContainerID); err != nil {
				slog.Warn("failed to stop container", "error", err, "container_id", deployment.ContainerID)
			}
			
			if err := uc.dockerClient.RemoveContainer(ctx, deployment.ContainerID); err != nil {
				slog.Warn("failed to remove container", "error", err, "container_id", deployment.ContainerID)
			}
		}

		if deployment.DockerImage != "" {
			// Best effort image cleanup
			go func(img string) {
				_ = uc.dockerClient.RemoveImage(context.Background(), img)
			}(deployment.DockerImage)
		}
	}

	// Enqueue webhook deletion task
	payload := &tasks.DeleteWebhookPayload{
		UserID:     project.UserID,
		RepoOwner:  project.RepoOwner,
		RepoName:   project.RepoName,
		WebhookURL: fmt.Sprintf("%s/webhooks/github?project_id=%s", uc.apiURL, project.ID),
	}
	
	if payloadBytes, err := payload.Marshal(); err == nil {
		task := asynq.NewTask(tasks.TypeDeleteWebhook, payloadBytes)
		if _, err := uc.asynqClient.Enqueue(task); err != nil {
			slog.Warn("failed to enqueue webhook deletion task", "project_id", project.ID, "error", err)
		}
	} else {
		slog.Warn("failed to marshal webhook deletion payload", "error", err)
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
