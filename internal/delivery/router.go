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

func NewRouter(authUC *usecase.AuthUseCase) *Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	authHandler := handler.NewAuthHandler(authUC)

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
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return &Router{r}
}
