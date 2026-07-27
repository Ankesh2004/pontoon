package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	"github.com/Ankesh2004/pontoon/internal/delivery/handler"
	mw "github.com/Ankesh2004/pontoon/internal/delivery/middleware"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

type Router struct {
	*chi.Mux
}

func NewRouter(
	authUC *usecase.AuthUseCase,
	projectUC *usecase.ProjectUseCase,
	deploymentUC *usecase.DeploymentUseCase,
	envVarUC *usecase.EnvVarUseCase,
	webhookUC *usecase.WebhookUseCase,
	redisClient *redis.Client,
	dockerClient *docker.Client,
) *Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	authHandler := handler.NewAuthHandler(authUC)
	projectHandler := handler.NewProjectHandler(projectUC)
	deploymentHandler := handler.NewDeploymentHandler(deploymentUC)
	envVarHandler := handler.NewEnvVarHandler(envVarUC)
	webhookHandler := handler.NewWebhookHandler(webhookUC)
	wsHandler := handler.NewWebSocketHandler(deploymentUC, redisClient)
	logHandler := handler.NewLogHandler(deploymentUC, dockerClient)

	r.Route("/auth", func(r chi.Router) {
		r.Get("/github", authHandler.Login)
		r.Get("/callback", authHandler.Callback)
	})

	r.Post("/webhooks/github", webhookHandler.Handle)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.Auth(authUC))

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		r.Route("/projects", func(r chi.Router) {
			r.Post("/", projectHandler.Create)
			r.Get("/", projectHandler.List)
			r.Get("/{id}", projectHandler.Get)
			r.Put("/{id}", projectHandler.Update)
			r.Delete("/{id}", projectHandler.Delete)

			r.Post("/{id}/deploy", deploymentHandler.Trigger)
			r.Get("/{id}/deployments", deploymentHandler.List)

			r.Post("/{id}/env", envVarHandler.Create)
			r.Get("/{id}/env", envVarHandler.List)
			r.Delete("/{id}/env/{envVarId}", envVarHandler.Delete)
		})

		r.Route("/deployments", func(r chi.Router) {
			r.Get("/{deploymentId}", deploymentHandler.Get)
			r.Get("/{deploymentId}/logs", logHandler.GetRuntimeLogs)
		})

		r.Get("/ws/logs", wsHandler.StreamLogs)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return &Router{r}
}
