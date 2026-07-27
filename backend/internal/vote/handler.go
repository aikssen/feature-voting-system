package vote

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"soundflow/internal/shared/apperr"
	"soundflow/internal/shared/httpx"
	"soundflow/internal/shared/middleware"
)

// Handler exposes the voting endpoint.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the vote action under /features/{id}/vote (auth required).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.With(middleware.RequireAuth).Post("/features/{id}/vote", h.vote)
}

func (h *Handler) vote(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.PrincipalFrom(r.Context())

	featureID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, r, apperr.NotFound("Feature request not found."))
		return
	}

	res, err := h.svc.Vote(r.Context(), featureID, principal.UserID)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, res)
}
