package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashabaranov/go-openai"
)

// Orchestrator manages the execution of the AI multi-agent DAG pipeline.
type Orchestrator struct {
	groqClient *openai.Client
	agents     []Agent
	db         *pgxpool.Pool
}

// NewOrchestrator initializes the AI Orchestrator with Groq client, DB pool, and defined agents.
func NewOrchestrator(apiKey string, db *pgxpool.Pool) *Orchestrator {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"
	client := openai.NewClientWithConfig(config)

	// Define the DAG pipeline (order matters here for sequential execution)
	agents := []Agent{
		NewInspectorAgent(client),
		NewDiagnosticianAgent(client),
		NewArchitectAgent(client),
		NewSecOpsAgent(client),
	}

	return &Orchestrator{
		groqClient: client,
		agents:     agents,
		db:         db,
	}
}

// RunPipeline executes the DAG pipeline, updating the state in the database after every stage.
func (o *Orchestrator) RunPipeline(ctx context.Context, pCtx *PipelineContext) error {
	// 1. Initial State Persistence
	err := o.persistContext(ctx, pCtx, StatusRunning)
	if err != nil {
		return fmt.Errorf("failed to initialize pipeline state: %w", err)
	}

	for _, agent := range o.agents {
		// 2. Execute agent with 3x retry and exponential backoff
		err := o.executeWithRetry(ctx, func() error {
			return agent.Execute(ctx, pCtx)
		}, 3)

		if err != nil {
			o.persistContext(ctx, pCtx, StatusFailed)
			return fmt.Errorf("pipeline aborted at %s: %w", agent.Name(), err)
		}

		// 3. Confidence Check: If confidence drops below 80, fail/pause
		if pCtx.ConfidenceScore != nil && *pCtx.ConfidenceScore < 80 {
			// Alternatively, could set to waiting_approval with a warning, 
			// but for fail-safe, we fail it if it's too uncertain.
			o.persistContext(ctx, pCtx, StatusFailed)
			return fmt.Errorf("pipeline aborted at %s due to low confidence: %d%%", agent.Name(), *pCtx.ConfidenceScore)
		}

		// 4. Save state after each agent succeeds (Resilience)
		err = o.persistContext(ctx, pCtx, StatusRunning)
		if err != nil {
			return fmt.Errorf("failed to persist state after %s: %w", agent.Name(), err)
		}
	}

	// 5. Pipeline finishes, move to "waiting_approval" UI state
	err = o.persistContext(ctx, pCtx, StatusWaitingApproval)
	if err != nil {
		return fmt.Errorf("failed to transition pipeline to waiting_approval: %w", err)
	}

	return nil
}

// executeWithRetry runs a function with exponential backoff on failure
func (o *Orchestrator) executeWithRetry(ctx context.Context, operation func() error, maxRetries int) error {
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		// Check if context is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if attempt < maxRetries {
			sleepDur := time.Duration(1<<attempt) * time.Second // 2s, 4s, 8s...
			time.Sleep(sleepDur)
		}
	}
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

// persistContext saves the current pipeline context as JSONB to PostgreSQL
func (o *Orchestrator) persistContext(ctx context.Context, pCtx *PipelineContext, status PipelineStatus) error {
	ctxJSON, err := json.Marshal(pCtx)
	if err != nil {
		return fmt.Errorf("failed to marshal pipeline context: %w", err)
	}

	query := `
		INSERT INTO ai_pipelines (id, project_id, deployment_id, status, context)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE 
		SET status = EXCLUDED.status, context = EXCLUDED.context;
	`
	_, err = o.db.Exec(ctx, query, pCtx.PipelineID, pCtx.ProjectID, pCtx.DeploymentID, status, ctxJSON)
	return err
}
