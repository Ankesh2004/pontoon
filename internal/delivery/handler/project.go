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

type ProjectHandler struct {
	projectUC *usecase.ProjectUseCase
}

func NewProjectHandler(projectUC *usecase.ProjectUseCase) *ProjectHandler {
	return &ProjectHandler{
		projectUC: projectUC,
	}
}

type createProjectRequest struct {
	Name    string `json:"name"`
	RepoURL string `json:"repo_url"`
	Branch  string `json:"branch"`
	Port    string `json:"port"`
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.RepoURL == "" {
		http.Error(w, "name and repo_url are required", http.StatusBadRequest)
		return
	}

	if req.Branch == "" {
		req.Branch = "main"
	}

	project, err := h.projectUC.CreateProject(r.Context(), usecase.CreateProjectInput{
		UserID:  userID,
		Name:    req.Name,
		RepoURL: req.RepoURL,
		Branch:  req.Branch,
		Port:    req.Port,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	projects, err := h.projectUC.ListProjects(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if projects == nil {
		projects = []*domain.Project{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	project, err := h.projectUC.GetProject(r.Context(), userID, projectID)
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
	json.NewEncoder(w).Encode(project)
}

type updateProjectRequest struct {
	Name   string `json:"name"`
	Branch string `json:"branch"`
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	project, err := h.projectUC.UpdateProject(r.Context(), userID, projectID, usecase.UpdateProjectInput{
		Name:   req.Name,
		Branch: req.Branch,
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
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	err := h.projectUC.DeleteProject(r.Context(), userID, projectID)
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

	w.WriteHeader(http.StatusNoContent)
}
