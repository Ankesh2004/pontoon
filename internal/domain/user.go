package domain

import "time"

type User struct {
	ID             string    `json:"id"`
	GitHubID       int       `json:"github_id"`
	GitHubUsername string    `json:"github_username"`
	Email          string    `json:"email"`
	AvatarURL      string    `json:"avatar_url"`
	AccessToken    string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByGitHubID(githubID int) (*User, error)
	Update(user *User) error
	UpsertByGitHubID(user *User) error
}
