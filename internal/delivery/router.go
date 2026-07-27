package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/Ankesh2004/pontoon/internal/delivery/handler"
	mw "github.com/Ankesh2004/pontoon/internal/delivery/middleware"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

type Router struct {
	*chi.Mux
}

func NewRouter(authUC *usecase.AuthUseCase, projectUC *usecase.ProjectUseCase) *Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	authHandler := handler.NewAuthHandler(authUC)
	projectHandler := handler.NewProjectHandler(projectUC)

	r.Route("/auth", func(r chi.Router) {
		r.Get("/github", authHandler.Login)
		r.Get("/callback", authHandler.Callback)
	})

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
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return &Router{r}
}
