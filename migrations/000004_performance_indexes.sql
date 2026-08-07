-- +goose Up
-- +goose NO TRANSACTION

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_projects_repo_owner_name ON projects(repo_owner, repo_name);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_deployments_project_id_created_at ON deployments(project_id, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_deployments_status_updated_at ON deployments(status, updated_at);

-- +goose Down
-- +goose NO TRANSACTION

DROP INDEX CONCURRENTLY IF EXISTS idx_projects_repo_owner_name;
DROP INDEX CONCURRENTLY IF EXISTS idx_deployments_project_id_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_deployments_status_updated_at;
