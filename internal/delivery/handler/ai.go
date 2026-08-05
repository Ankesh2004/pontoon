package handler

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ankesh2004/pontoon/internal/ai"
)

type AIHandler struct {
	orchestrator *ai.Orchestrator
	db           *pgxpool.Pool
}

func NewAIHandler(db *pgxpool.Pool) *AIHandler {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		// Log warning or handle gracefully, for now we let it be empty (it will fail on use)
	}
	orch := ai.NewOrchestrator(apiKey, db)
	return &AIHandler{
		orchestrator: orch,
		db:           db,
	}
}

// RecoverBuild triggers a background AI recovery pipeline
func (h *AIHandler) RecoverBuild(w http.ResponseWriter, r *http.Request) {
	deploymentIDStr := chi.URLParam(r, "deploymentId")
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "invalid deployment ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ProjectID uuid.UUID `json:"project_id"`
		RawLogs   string    `json:"raw_logs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	pipelineID := uuid.New()
	pCtx := &ai.PipelineContext{
		PipelineID:   pipelineID,
		ProjectID:    req.ProjectID,
		DeploymentID: deploymentID,
		RawLogs:      req.RawLogs,
	}

	// Trigger pipeline in background
	go func() {
		_ = h.orchestrator.RunPipeline(r.Context(), pCtx)
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pipeline_id": pipelineID,
		"status":      ai.StatusRunning,
	})
}

// GetPipeline returns the current status and context of a pipeline
func (h *AIHandler) GetPipeline(w http.ResponseWriter, r *http.Request) {
	pipelineIDStr := chi.URLParam(r, "pipelineId")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		http.Error(w, "invalid pipeline ID", http.StatusBadRequest)
		return
	}

	query := `SELECT status, context FROM ai_pipelines WHERE id = $1`
	var status string
	var ctxBytes []byte
	err = h.db.QueryRow(r.Context(), query, pipelineID).Scan(&status, &ctxBytes)
	if err != nil {
		http.Error(w, "pipeline not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(string(ctxBytes))) // context is already JSONB, we could enrich it with status
}

// ApprovePipeline accepts the proposed patch and applies it
func (h *AIHandler) ApprovePipeline(w http.ResponseWriter, r *http.Request) {
	pipelineIDStr := chi.URLParam(r, "pipelineId")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		http.Error(w, "invalid pipeline ID", http.StatusBadRequest)
		return
	}

	// In a real implementation, this would trigger a GitHub commit with the ProposedPatch
	// For now, we just mark it as succeeded
	query := `UPDATE ai_pipelines SET status = $1 WHERE id = $2`
	_, err = h.db.Exec(r.Context(), query, ai.StatusSucceeded, pipelineID)
	if err != nil {
		http.Error(w, "failed to update pipeline", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "succeeded"})
}

// RejectPipeline marks the pipeline as rejected
func (h *AIHandler) RejectPipeline(w http.ResponseWriter, r *http.Request) {
	pipelineIDStr := chi.URLParam(r, "pipelineId")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		http.Error(w, "invalid pipeline ID", http.StatusBadRequest)
		return
	}

	query := `UPDATE ai_pipelines SET status = $1 WHERE id = $2`
	_, err = h.db.Exec(r.Context(), query, ai.StatusRejected, pipelineID)
	if err != nil {
		http.Error(w, "failed to update pipeline", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "rejected"})
}
