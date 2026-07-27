package infrastructure

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"soundflow/internal/auth"
	"soundflow/internal/feature"
	"soundflow/internal/shared/httpx"
	"soundflow/internal/shared/logging"
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
	Logger       *slog.Logger
}

// NewRouter assembles the HTTP router and the /api/v1 surface.
func NewRouter(d RouterDeps) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.RealIP)
	// CORS runs before the logger so preflights are answered with the right
	// headers, and before Recoverer's output can escape without them.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: d.CORSOrigins,
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		// The browser may only send X-Correlation-ID if it is allow-listed here,
		// and may only read it back off the response via ExposedHeaders.
		AllowedHeaders:   []string{"Authorization", "Content-Type", logging.CorrelationIDHeader},
		ExposedHeaders:   []string{logging.CorrelationIDHeader},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestLogger(d.Logger))
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

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
