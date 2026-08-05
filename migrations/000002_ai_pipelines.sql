-- +goose Up
CREATE TABLE IF NOT EXISTS ai_pipelines (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    deployment_id UUID,
    status VARCHAR(50) DEFAULT 'running', -- running, waiting_approval, succeeded, failed, rejected
    context JSONB, -- The aggregated data passing through the pipeline (including ConfidenceScore)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS ai_pipelines;
