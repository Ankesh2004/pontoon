package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Ankesh2004/pontoon/internal/delivery/middleware"
	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

type LogHandler struct {
	deploymentUC *usecase.DeploymentUseCase
	dockerClient *docker.Client
}

func NewLogHandler(
	deploymentUC *usecase.DeploymentUseCase,
	dockerClient *docker.Client,
) *LogHandler {
	return &LogHandler{
		deploymentUC: deploymentUC,
		dockerClient: dockerClient,
	}
}

func (h *LogHandler) GetRuntimeLogs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deploymentID := chi.URLParam(r, "deploymentId")
	if deploymentID == "" {
		http.Error(w, "deployment_id required", http.StatusBadRequest)
		return
	}

	tailStr := r.URL.Query().Get("tail")
	tail := 100
	if tailStr != "" {
		if t, err := strconv.Atoi(tailStr); err == nil && t > 0 {
			tail = t
		}
	}

	deployment, err := h.deploymentUC.GetDeployment(r.Context(), userID, deploymentID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || err.Error() == domain.ErrNotFound.Error() {
			http.Error(w, "deployment not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrForbidden) || err.Error() == domain.ErrForbidden.Error() {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "failed to get deployment", http.StatusInternalServerError)
		return
	}

	if deployment.ContainerID == "" {
		http.Error(w, "no container running for this deployment", http.StatusNotFound)
		return
	}

	logs, err := h.dockerClient.GetContainerLogs(r.Context(), deployment.ContainerID, tail)
	if err != nil {
		http.Error(w, "failed to get container logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(logs))
}
