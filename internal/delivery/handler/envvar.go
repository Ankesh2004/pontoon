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

type EnvVarHandler struct {
	envVarUC *usecase.EnvVarUseCase
}

func NewEnvVarHandler(envVarUC *usecase.EnvVarUseCase) *EnvVarHandler {
	return &EnvVarHandler{
		envVarUC: envVarUC,
	}
}

type createEnvVarRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (h *EnvVarHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req createEnvVarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	envVar, err := h.envVarUC.CreateEnvVar(r.Context(), userID, projectID, req.Key, req.Value)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			http.Error(w, "environment variable already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(envVar)
}

func (h *EnvVarHandler) List(w http.ResponseWriter, r *http.Request) {
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

	envVars, err := h.envVarUC.ListEnvVars(r.Context(), userID, projectID)
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

	if envVars == nil {
		envVars = []*domain.EnvVar{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envVars)
}

func (h *EnvVarHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	envVarID := chi.URLParam(r, "envVarId")
	if envVarID == "" {
		http.Error(w, "env var id required", http.StatusBadRequest)
		return
	}

	err := h.envVarUC.DeleteEnvVar(r.Context(), userID, projectID, envVarID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "environment variable not found", http.StatusNotFound)
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
