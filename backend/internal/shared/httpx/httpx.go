// Package httpx contains HTTP plumbing shared across feature slices:
// JSON decoding/encoding, the error-envelope writer, and list-query parsing.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"soundflow/internal/shared/apperr"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// envelope is the frozen error wire format.
type envelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details []apperr.FieldError `json:"details,omitempty"`
}

// DecodeJSON strictly decodes the request body into dst, rejecting unknown
// fields and oversized payloads with a 400.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if r.Body == nil {
		return apperr.BadRequest("Request body is required.")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return apperr.BadRequest("Request body is required.")
		}
		return apperr.BadRequest("Request body is not valid JSON.")
	}
	if dec.More() {
		return apperr.BadRequest("Request body must contain a single JSON object.")
	}
	return nil
}

// WriteJSON serializes v as JSON with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// WriteError maps any error to the frozen envelope. Unrecognized errors become
// a 500 with their cause logged, never leaked.
func WriteError(w http.ResponseWriter, err error) {
	ae, ok := apperr.As(err)
	if !ok {
		ae = apperr.Internal(err)
	}
	if ae.Status >= 500 {
		slog.Error("request failed", "code", ae.Code, "error", ae.Error())
	}
	WriteJSON(w, ae.Status, envelope{Error: errorBody{
		Code:    ae.Code,
		Message: ae.Message,
		Details: ae.Details,
	}})
}

// ListQuery holds the normalized discovery parameters.
type ListQuery struct {
	Search string
	Sort   string
	Page   int
	Limit  int
}

// Sort + pagination bounds, per DECISIONS.md (M7).
const (
	SortNewest    = "newest"
	SortMostVoted = "most_voted"
	SortTrending  = "trending"

	defaultLimit = 20
	maxLimit     = 50
)

// ParseListQuery reads search/sort/page/limit leniently: unknown sort falls
// back to trending, page floors at 1, limit clamps to 1..50.
func ParseListQuery(r *http.Request) ListQuery {
	q := r.URL.Query()

	sort := strings.ToLower(strings.TrimSpace(q.Get("sort")))
	switch sort {
	case SortNewest, SortMostVoted, SortTrending:
	default:
		sort = SortTrending
	}

	page := 1
	if v, err := strconv.Atoi(q.Get("page")); err == nil && v > 1 {
		page = v
	}

	limit := defaultLimit
	if v, err := strconv.Atoi(q.Get("limit")); err == nil {
		switch {
		case v < 1:
			limit = 1
		case v > maxLimit:
			limit = maxLimit
		default:
			limit = v
		}
	}

	return ListQuery{
		Search: strings.TrimSpace(q.Get("search")),
		Sort:   sort,
		Page:   page,
		Limit:  limit,
	}
}
