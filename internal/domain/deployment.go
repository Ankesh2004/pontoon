package domain

import "time"

type DeploymentStatus string

const (
	DeploymentStatusPending  DeploymentStatus = "pending"
	DeploymentStatusCloning  DeploymentStatus = "cloning"
	DeploymentStatusBuilding DeploymentStatus = "building"
	DeploymentStatusRunning  DeploymentStatus = "running"
	DeploymentStatusLive     DeploymentStatus = "live"
	DeploymentStatusStopped  DeploymentStatus = "stopped"
	DeploymentStatusFailed   DeploymentStatus = "failed"
)

type Deployment struct {
	ID             string           `json:"id"`
	ProjectID      string           `json:"project_id"`
	UserID         string           `json:"user_id"`
	Status         DeploymentStatus `json:"status"`
	CommitSHA      string           `json:"commit_sha"`
	DockerImage    string           `json:"docker_image"`
	ContainerID    string           `json:"container_id"`
	ContainerName  string           `json:"container_name"`
	MemoryLimitMB  int              `json:"memory_limit_mb"`
	BuildLogs      string           `json:"build_logs"`
	TriggeredBy    string           `json:"triggered_by"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type DeploymentRepository interface {
	Create(deployment *Deployment) error
	GetByID(id string) (*Deployment, error)
	GetByProjectID(projectID string) ([]*Deployment, error)
	GetActiveByProjectID(projectID string) ([]*Deployment, error)
	GetStuck(timeoutMinutes int) ([]*Deployment, error)
	GetByUserID(userID string) ([]*Deployment, error)
	Update(deployment *Deployment) error
	UpdateStatus(id string, status DeploymentStatus, logs string) error
	Delete(id string) error
}

type EnvVar struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type EnvVarRepository interface {
	Create(envVar *EnvVar) error
	GetByProjectID(projectID string) ([]*EnvVar, error)
	GetByID(id string) (*EnvVar, error)
	Update(envVar *EnvVar) error
	Delete(id string) error
}
