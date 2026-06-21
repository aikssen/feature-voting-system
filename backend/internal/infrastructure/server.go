package infrastructure

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"soundflow/internal/auth"
	"soundflow/internal/feature"
	"soundflow/internal/shared/httpx"
	"soundflow/internal/shared/middleware"
	"soundflow/internal/shared/token"
	"soundflow/internal/vote"
)

// RouterDeps are the handlers + config the router needs.
type RouterDeps struct {
	TokenManager *token.Manager
	Auth         *auth.Handler
	Feature      *feature.Handler
	Vote         *vote.Handler
	CORSOrigins  []string
}

// NewRouter assembles the HTTP router and the /api/v1 surface.
func NewRouter(d RouterDeps) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   d.CORSOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", health)

		// Public auth endpoints (no token parsing).
		d.Auth.RegisterRoutes(r)

		// Everything else parses an optional bearer token; specific routes
		// additionally require it via middleware.RequireAuth.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticator(d.TokenManager))
			d.Feature.RegisterRoutes(r)
			d.Vote.RegisterRoutes(r)
		})
	})

	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
