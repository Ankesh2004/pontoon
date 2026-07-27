package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Ankesh2004/pontoon/internal/domain"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, github_id, github_username, email, avatar_url, access_token)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(context.Background(), query,
		user.ID, user.GitHubID, user.GitHubUsername, user.Email, user.AvatarURL, user.AccessToken)
	return err
}

func (r *UserRepo) GetByID(id string) (*domain.User, error) {
	query := `
		SELECT id, github_id, github_username, email, avatar_url, access_token, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user domain.User
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&user.ID, &user.GitHubID, &user.GitHubUsername, &user.Email,
		&user.AvatarURL, &user.AccessToken, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByGitHubID(githubID int) (*domain.User, error) {
	query := `
		SELECT id, github_id, github_username, email, avatar_url, access_token, created_at, updated_at
		FROM users WHERE github_id = $1
	`
	var user domain.User
	err := r.pool.QueryRow(context.Background(), query, githubID).Scan(
		&user.ID, &user.GitHubID, &user.GitHubUsername, &user.Email,
		&user.AvatarURL, &user.AccessToken, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET github_username = $2, email = $3, avatar_url = $4, access_token = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(context.Background(), query,
		user.ID, user.GitHubUsername, user.Email, user.AvatarURL, user.AccessToken)
	return err
}

func (r *UserRepo) UpsertByGitHubID(user *domain.User) error {
	query := `
		INSERT INTO users (id, github_id, github_username, email, avatar_url, access_token)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (github_id) DO UPDATE
		SET github_username = EXCLUDED.github_username,
		    email = EXCLUDED.email,
		    avatar_url = EXCLUDED.avatar_url,
		    access_token = EXCLUDED.access_token,
		    updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(context.Background(), query,
		user.ID, user.GitHubID, user.GitHubUsername, user.Email, user.AvatarURL, user.AccessToken,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}
