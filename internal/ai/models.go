package ai

import (
	"context"

	"github.com/google/uuid"
)

// PipelineStatus represents the current state of a pipeline
type PipelineStatus string

const (
	StatusRunning         PipelineStatus = "running"
	StatusWaitingApproval PipelineStatus = "waiting_approval"
	StatusSucceeded       PipelineStatus = "succeeded"
	StatusFailed          PipelineStatus = "failed"
	StatusRejected        PipelineStatus = "rejected"
)

// PipelineContext represents the shared state passed between agents in the DAG
type PipelineContext struct {
	PipelineID      uuid.UUID `json:"pipeline_id"`
	ProjectID       uuid.UUID `json:"project_id"`
	DeploymentID    uuid.UUID `json:"deployment_id"`
	RawLogs         string    `json:"raw_logs,omitempty"`
	ParsedError     *string   `json:"parsed_error,omitempty"`
	RootCause       *string   `json:"root_cause,omitempty"`
	ProposedPatch   *string   `json:"proposed_patch,omitempty"`
	SecurityPassed  bool      `json:"security_passed"`
	ConfidenceScore *int      `json:"confidence_score,omitempty"` // Aggregated confidence score
}

// Agent represents a node in the DAG pipeline
type Agent interface {
	// Execute takes the current pipeline context, performs its specific LLM task,
	// updates the context in place, and returns an error if it fails.
	Execute(ctx context.Context, pCtx *PipelineContext) error
	
	// Name returns the identifier of the agent for logging and tracing
	Name() string
}
