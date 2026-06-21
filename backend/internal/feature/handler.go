package feature

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"soundflow/internal/shared/apperr"
	"soundflow/internal/shared/httpx"
	"soundflow/internal/shared/middleware"
)

// Handler exposes the feature-request endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts /features. List and Get are public; Create requires auth.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/features", func(r chi.Router) {
		r.Get("/", h.list)
		r.Get("/{id}", h.get)
		r.With(middleware.RequireAuth).Post("/", h.create)
	})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.PrincipalFrom(r.Context())

	var req CreateRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	view, err := h.svc.Create(r.Context(), principal.UserID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, view)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := httpx.ParseListQuery(r)
	cur := middleware.CurrentUserID(r.Context())

	page, err := h.svc.List(r.Context(), cur, q.Search, q.Sort, q.Page, q.Limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, page)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, apperr.NotFound("Feature request not found."))
		return
	}
	cur := middleware.CurrentUserID(r.Context())

	view, err := h.svc.Get(r.Context(), cur, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, view)
}
