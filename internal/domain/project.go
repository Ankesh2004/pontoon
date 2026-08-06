package domain

import "time"

type Project struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`
	RepoURL       string    `json:"repo_url"`
	RepoOwner     string    `json:"repo_owner"`
	RepoName      string    `json:"repo_name"`
	Branch        string    `json:"branch"`
	Domain        string    `json:"domain"`
	WebhookSecret string    `json:"webhook_secret,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProjectRepository interface {
	Create(project *Project) error
	GetByID(id string) (*Project, error)
	GetByUserID(userID string) ([]*Project, error)
	GetAll() ([]*Project, error)
	GetByRepo(owner, name string) (*Project, error)
	Update(project *Project) error
	Delete(id string) error
}
