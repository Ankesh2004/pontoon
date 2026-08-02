package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Ankesh2004/pontoon/internal/domain"
)

type EnvVarRepo struct {
	pool *pgxpool.Pool
}

func NewEnvVarRepo(pool *pgxpool.Pool) *EnvVarRepo {
	return &EnvVarRepo{pool: pool}
}

func (r *EnvVarRepo) Create(envVar *domain.EnvVar) error {
	query := `
		INSERT INTO env_vars (id, project_id, key, value)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`
	return r.pool.QueryRow(context.Background(), query,
		envVar.ID, envVar.ProjectID, envVar.Key, envVar.Value,
	).Scan(&envVar.CreatedAt)
}

func (r *EnvVarRepo) GetByProjectID(projectID string) ([]*domain.EnvVar, error) {
	query := `
		SELECT id, project_id, key, value, created_at
		FROM env_vars WHERE project_id = $1
		ORDER BY key
	`
	rows, err := r.pool.Query(context.Background(), query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []*domain.EnvVar
	for rows.Next() {
		var envVar domain.EnvVar
		if err := rows.Scan(&envVar.ID, &envVar.ProjectID, &envVar.Key, &envVar.Value, &envVar.CreatedAt); err != nil {
			return nil, err
		}
		envVars = append(envVars, &envVar)
	}
	return envVars, nil
}

func (r *EnvVarRepo) GetByID(id string) (*domain.EnvVar, error) {
	query := `
		SELECT id, project_id, key, value, created_at
		FROM env_vars WHERE id = $1
	`
	var envVar domain.EnvVar
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&envVar.ID, &envVar.ProjectID, &envVar.Key, &envVar.Value, &envVar.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &envVar, nil
}

func (r *EnvVarRepo) Update(envVar *domain.EnvVar) error {
	query := `
		UPDATE env_vars
		SET key = $1, value = $2
		WHERE id = $3
	`
	tag, err := r.pool.Exec(context.Background(), query, envVar.Key, envVar.Value, envVar.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *EnvVarRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), "DELETE FROM env_vars WHERE id = $1", id)
	return err
}
