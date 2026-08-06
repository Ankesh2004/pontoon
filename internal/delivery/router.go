package delivery

import (
	"crypto/sha256"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"

	"github.com/Ankesh2004/pontoon/internal/config"
	"github.com/Ankesh2004/pontoon/internal/delivery/handler"
	mw "github.com/Ankesh2004/pontoon/internal/delivery/middleware"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

const (
	RateLimitGlobal   = 100
	RateLimitAuth     = 10
	RateLimitDeploy   = 5
	RateLimitWebhooks = 30
)

type Router struct {
	*chi.Mux
}

func NewRouter(
	cfg *config.Config,
	authUC *usecase.AuthUseCase,
	userUC *usecase.UserUseCase,
	projectUC *usecase.ProjectUseCase,
	deploymentUC *usecase.DeploymentUseCase,
	envVarUC *usecase.EnvVarUseCase,
	webhookUC *usecase.WebhookUseCase,
	redisClient *goredis.Client,
	dockerClient *docker.Client,
	dbPool *pgxpool.Pool, // Added DB pool for AI handler
) *Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mw.CORS(cfg.CORS.AllowedOrigins))

	authHandler := handler.NewAuthHandler(authUC, userUC, redisClient, cfg.FrontendURL)
	projectHandler := handler.NewProjectHandler(projectUC)
	deploymentHandler := handler.NewDeploymentHandler(deploymentUC)
	envVarHandler := handler.NewEnvVarHandler(envVarUC)
	webhookHandler := handler.NewWebhookHandler(webhookUC)
	wsHandler := handler.NewWebSocketHandler(deploymentUC, redisClient)
	logHandler := handler.NewLogHandler(deploymentUC, dockerClient)
	aiHandler := handler.NewAIHandler(dbPool) // Initialize AI Handler

	// Extract hostnames for CSRF trusted origins
	var trustedOrigins []string
	for _, origin := range cfg.CORS.AllowedOrigins {
		if u, err := url.Parse(origin); err == nil && u.Host != "" {
			trustedOrigins = append(trustedOrigins, u.Host)
		} else {
			trustedOrigins = append(trustedOrigins, origin)
		}
	}

	// Generate 32-byte key from JWT secret for CSRF
	csrfKey := sha256.Sum256([]byte(cfg.JWT.Secret))
	csrfMiddleware := csrf.Protect(
		csrfKey[:],
		csrf.Secure(false), // Disable Secure for localhost testing (set to true in production with HTTPS)
		csrf.Path("/"),
		csrf.TrustedOrigins(trustedOrigins),
	)

	r.Route("/auth", func(r chi.Router) {
		r.Use(httprate.LimitByIP(RateLimitAuth, 1*time.Minute))
		r.Get("/github", authHandler.Login)
		r.Get("/callback", authHandler.Callback)
		r.Get("/logout", authHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(authUC))
			r.Use(csrfMiddleware)
			r.Get("/me", authHandler.Me)
			r.Get("/csrf", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-CSRF-Token", csrf.Token(r))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"token":"` + csrf.Token(r) + `"}`))
			})
			r.Post("/ws-ticket", authHandler.WSTicket)
		})
	})

	r.With(httprate.LimitByIP(RateLimitWebhooks, 1*time.Minute)).Post("/webhooks/github", webhookHandler.Handle)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(httprate.LimitByIP(RateLimitGlobal, 1*time.Minute))
		r.Use(mw.Auth(authUC))
		r.Use(csrfMiddleware)

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

			r.With(httprate.LimitByIP(RateLimitDeploy, 1*time.Minute)).Post("/{id}/deploy", deploymentHandler.Trigger)
			r.Get("/{id}/deployments", deploymentHandler.List)
			r.Post("/{id}/deployments/{deploymentId}/rollback", deploymentHandler.Rollback)

			r.Post("/{id}/env", envVarHandler.Create)
			r.Get("/{id}/env", envVarHandler.List)
			r.Put("/{id}/env/{envVarId}", envVarHandler.Update)
			r.Delete("/{id}/env/{envVarId}", envVarHandler.Delete)
		})

		r.Route("/deployments", func(r chi.Router) {
			r.Get("/{deploymentId}", deploymentHandler.Get)
			r.Post("/{deploymentId}/stop", deploymentHandler.Stop)
			r.Delete("/{deploymentId}", deploymentHandler.Delete)
			r.Get("/{deploymentId}/logs", logHandler.GetRuntimeLogs)
			r.Post("/{deploymentId}/recover", aiHandler.RecoverBuild) // Trigger AI Recovery
		})

		r.Route("/pipelines", func(r chi.Router) {
			r.Get("/{pipelineId}", aiHandler.GetPipeline)
			r.Post("/{pipelineId}/approve", aiHandler.ApprovePipeline)
			r.Post("/{pipelineId}/reject", aiHandler.RejectPipeline)
		})
	})

	r.Get("/api/v1/deployments/{deploymentId}/ws", wsHandler.StreamLogs)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return &Router{r}
}
