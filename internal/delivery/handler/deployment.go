package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Ankesh2004/pontoon/internal/delivery/middleware"
	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

type DeploymentHandler struct {
	deploymentUC *usecase.DeploymentUseCase
}

func NewDeploymentHandler(deploymentUC *usecase.DeploymentUseCase) *DeploymentHandler {
	return &DeploymentHandler{
		deploymentUC: deploymentUC,
	}
}

type triggerDeploymentRequest struct {
	CommitSHA string `json:"commit_sha"`
}

func (h *DeploymentHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		http.Error(w, "project id required", http.StatusBadRequest)
		return
	}

	var req triggerDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.CommitSHA = "HEAD"
	}
	if req.CommitSHA == "" {
		req.CommitSHA = "HEAD"
	}

	deployment, err := h.deploymentUC.TriggerDeployment(r.Context(), usecase.TriggerDeploymentInput{
		UserID:    userID,
		ProjectID: projectID,
		CommitSHA: req.CommitSHA,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(deployment)
}

func (h *DeploymentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		http.Error(w, "project id required", http.StatusBadRequest)
		return
	}

	deployments, err := h.deploymentUC.ListDeployments(r.Context(), userID, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if deployments == nil {
		deployments = []*domain.Deployment{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

func (h *DeploymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deploymentID := chi.URLParam(r, "deploymentId")
	if deploymentID == "" {
		http.Error(w, "deployment id required", http.StatusBadRequest)
		return
	}

	deployment, err := h.deploymentUC.GetDeployment(r.Context(), userID, deploymentID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "deployment not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}

func (h *DeploymentHandler) Stop(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deploymentID := chi.URLParam(r, "deploymentId")
	if deploymentID == "" {
		http.Error(w, "deployment id required", http.StatusBadRequest)
		return
	}

	if err := h.deploymentUC.StopDeployment(r.Context(), userID, deploymentID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "deployment not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"stopped"}`))
}

func (h *DeploymentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deploymentID := chi.URLParam(r, "deploymentId")
	if deploymentID == "" {
		http.Error(w, "deployment id required", http.StatusBadRequest)
		return
	}

	if err := h.deploymentUC.DeleteDeployment(r.Context(), userID, deploymentID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "deployment not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
