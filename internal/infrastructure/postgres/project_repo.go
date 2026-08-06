package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Ankesh2004/pontoon/internal/domain"
)

type ProjectRepo struct {
	pool *pgxpool.Pool
}

func NewProjectRepo(pool *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{pool: pool}
}

func (r *ProjectRepo) Create(project *domain.Project) error {
	query := `
		INSERT INTO projects (id, user_id, name, repo_url, repo_owner, repo_name, branch, domain, webhook_secret)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`
	return r.pool.QueryRow(context.Background(), query,
		project.ID, project.UserID, project.Name, project.RepoURL,
		project.RepoOwner, project.RepoName, project.Branch, project.Domain, project.WebhookSecret,
	).Scan(&project.CreatedAt, &project.UpdatedAt)
}

func (r *ProjectRepo) GetByID(id string) (*domain.Project, error) {
	query := `
		SELECT id, user_id, name, repo_url, repo_owner, repo_name, branch, domain, webhook_secret, created_at, updated_at
		FROM projects WHERE id = $1
	`
	var project domain.Project
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&project.ID, &project.UserID, &project.Name, &project.RepoURL,
		&project.RepoOwner, &project.RepoName, &project.Branch, &project.Domain,
		&project.WebhookSecret, &project.CreatedAt, &project.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepo) GetByUserID(userID string) ([]*domain.Project, error) {
	query := `
		SELECT id, user_id, name, repo_url, repo_owner, repo_name, branch, domain, webhook_secret, created_at, updated_at
		FROM projects WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var project domain.Project
		if err := rows.Scan(
			&project.ID, &project.UserID, &project.Name, &project.RepoURL,
			&project.RepoOwner, &project.RepoName, &project.Branch, &project.Domain,
			&project.WebhookSecret, &project.CreatedAt, &project.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, &project)
	}
	return projects, nil
}

func (r *ProjectRepo) GetAll() ([]*domain.Project, error) {
	query := `
		SELECT id, user_id, name, repo_url, repo_owner, repo_name, branch, domain, webhook_secret, created_at, updated_at
		FROM projects
	`
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var project domain.Project
		if err := rows.Scan(
			&project.ID, &project.UserID, &project.Name, &project.RepoURL,
			&project.RepoOwner, &project.RepoName, &project.Branch, &project.Domain,
			&project.WebhookSecret, &project.CreatedAt, &project.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, &project)
	}
	return projects, nil
}

func (r *ProjectRepo) GetByRepo(owner, name string) (*domain.Project, error) {
	query := `
		SELECT id, user_id, name, repo_url, repo_owner, repo_name, branch, domain, webhook_secret, created_at, updated_at
		FROM projects WHERE repo_owner = $1 AND repo_name = $2
	`
	var project domain.Project
	err := r.pool.QueryRow(context.Background(), query, owner, name).Scan(
		&project.ID, &project.UserID, &project.Name, &project.RepoURL,
		&project.RepoOwner, &project.RepoName, &project.Branch, &project.Domain,
		&project.WebhookSecret, &project.CreatedAt, &project.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepo) Update(project *domain.Project) error {
	query := `
		UPDATE projects
		SET name = $2, branch = $3, domain = $4, webhook_secret = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(context.Background(), query,
		project.ID, project.Name, project.Branch, project.Domain, project.WebhookSecret)
	return err
}

func (r *ProjectRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), "DELETE FROM projects WHERE id = $1", id)
	return err
}
