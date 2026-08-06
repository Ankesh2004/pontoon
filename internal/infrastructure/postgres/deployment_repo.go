package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Ankesh2004/pontoon/internal/domain"
)

type DeploymentRepo struct {
	pool *pgxpool.Pool
}

func NewDeploymentRepo(pool *pgxpool.Pool) *DeploymentRepo {
	return &DeploymentRepo{pool: pool}
}

func (r *DeploymentRepo) Create(deployment *domain.Deployment) error {
	query := `
		INSERT INTO deployments (id, project_id, user_id, status, commit_sha, docker_image, container_id, container_name, memory_limit_mb, build_logs, triggered_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`
	return r.pool.QueryRow(context.Background(), query,
		deployment.ID, deployment.ProjectID, deployment.UserID, deployment.Status,
		deployment.CommitSHA, deployment.DockerImage, deployment.ContainerID,
		deployment.ContainerName, deployment.MemoryLimitMB, deployment.BuildLogs, deployment.TriggeredBy,
	).Scan(&deployment.CreatedAt, &deployment.UpdatedAt)
}

func (r *DeploymentRepo) GetByID(id string) (*domain.Deployment, error) {
	query := `
		SELECT id, project_id, user_id, status, commit_sha, docker_image, container_id, container_name, memory_limit_mb, build_logs, triggered_by, created_at, updated_at
		FROM deployments WHERE id = $1
	`
	var deployment domain.Deployment
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&deployment.ID, &deployment.ProjectID, &deployment.UserID, &deployment.Status,
		&deployment.CommitSHA, &deployment.DockerImage, &deployment.ContainerID,
		&deployment.ContainerName, &deployment.MemoryLimitMB, &deployment.BuildLogs,
		&deployment.TriggeredBy, &deployment.CreatedAt, &deployment.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (r *DeploymentRepo) GetByProjectID(projectID string) ([]*domain.Deployment, error) {
	query := `
		SELECT id, project_id, user_id, status, commit_sha, docker_image, container_id, container_name, memory_limit_mb, build_logs, triggered_by, created_at, updated_at
		FROM deployments WHERE project_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(context.Background(), query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*domain.Deployment
	for rows.Next() {
		var deployment domain.Deployment
		if err := rows.Scan(
			&deployment.ID, &deployment.ProjectID, &deployment.UserID, &deployment.Status,
			&deployment.CommitSHA, &deployment.DockerImage, &deployment.ContainerID,
			&deployment.ContainerName, &deployment.MemoryLimitMB, &deployment.BuildLogs,
			&deployment.TriggeredBy, &deployment.CreatedAt, &deployment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deployments = append(deployments, &deployment)
	}
	return deployments, nil
}

func (r *DeploymentRepo) GetActiveByProjectID(projectID string) ([]*domain.Deployment, error) {
	query := `
		SELECT id, project_id, user_id, status, commit_sha, docker_image, container_id, container_name, memory_limit_mb, build_logs, triggered_by, created_at, updated_at
		FROM deployments WHERE project_id = $1 AND status IN ('pending', 'cloning', 'building', 'running')
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(context.Background(), query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*domain.Deployment
	for rows.Next() {
		var deployment domain.Deployment
		if err := rows.Scan(
			&deployment.ID, &deployment.ProjectID, &deployment.UserID, &deployment.Status,
			&deployment.CommitSHA, &deployment.DockerImage, &deployment.ContainerID,
			&deployment.ContainerName, &deployment.MemoryLimitMB, &deployment.BuildLogs,
			&deployment.TriggeredBy, &deployment.CreatedAt, &deployment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deployments = append(deployments, &deployment)
	}
	return deployments, nil
}

func (r *DeploymentRepo) GetStuck(timeoutMinutes int) ([]*domain.Deployment, error) {
	query := `
		SELECT id, project_id, user_id, status, commit_sha, docker_image, container_id, container_name, memory_limit_mb, build_logs, triggered_by, created_at, updated_at
		FROM deployments 
		WHERE status IN ('cloning', 'building', 'running') 
		AND updated_at < NOW() - ($1 || ' minutes')::INTERVAL
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(context.Background(), query, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*domain.Deployment
	for rows.Next() {
		var deployment domain.Deployment
		if err := rows.Scan(
			&deployment.ID, &deployment.ProjectID, &deployment.UserID, &deployment.Status,
			&deployment.CommitSHA, &deployment.DockerImage, &deployment.ContainerID,
			&deployment.ContainerName, &deployment.MemoryLimitMB, &deployment.BuildLogs,
			&deployment.TriggeredBy, &deployment.CreatedAt, &deployment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deployments = append(deployments, &deployment)
	}
	return deployments, nil
}

func (r *DeploymentRepo) GetByUserID(userID string) ([]*domain.Deployment, error) {
	query := `
		SELECT id, project_id, user_id, status, commit_sha, docker_image, container_id, container_name, memory_limit_mb, build_logs, triggered_by, created_at, updated_at
		FROM deployments WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*domain.Deployment
	for rows.Next() {
		var deployment domain.Deployment
		if err := rows.Scan(
			&deployment.ID, &deployment.ProjectID, &deployment.UserID, &deployment.Status,
			&deployment.CommitSHA, &deployment.DockerImage, &deployment.ContainerID,
			&deployment.ContainerName, &deployment.MemoryLimitMB, &deployment.BuildLogs,
			&deployment.TriggeredBy, &deployment.CreatedAt, &deployment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deployments = append(deployments, &deployment)
	}
	return deployments, nil
}

func (r *DeploymentRepo) Update(deployment *domain.Deployment) error {
	query := `
		UPDATE deployments
		SET status = $2, commit_sha = $3, docker_image = $4, container_id = $5, container_name = $6, build_logs = $7, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(context.Background(), query,
		deployment.ID, deployment.Status, deployment.CommitSHA, deployment.DockerImage,
		deployment.ContainerID, deployment.ContainerName, deployment.BuildLogs)
	return err
}

func (r *DeploymentRepo) UpdateStatus(id string, status domain.DeploymentStatus, logs string) error {
	query := `UPDATE deployments SET status = $2, build_logs = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(context.Background(), query, id, status, logs)
	return err
}

func (r *DeploymentRepo) Delete(id string) error {
	query := `DELETE FROM deployments WHERE id = $1`
	_, err := r.pool.Exec(context.Background(), query, id)
	return err
}

