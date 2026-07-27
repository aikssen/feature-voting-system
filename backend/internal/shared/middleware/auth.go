// Package middleware provides authentication middleware shared by feature slices.
//
// Two layers, per DECISIONS.md (M9):
//   - Authenticator: parses a Bearer token if present. Absent → continue
//     anonymously (public endpoints can still compute has_voted/is_author).
//     Present-but-invalid → 401 (a clear client error, even on public routes).
//   - RequireAuth: rejects requests that reached it without an identity (401).
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"soundflow/internal/shared/apperr"
	"soundflow/internal/shared/httpx"
	"soundflow/internal/shared/logging"
	"soundflow/internal/shared/token"
)

type ctxKey int

const principalKey ctxKey = iota

// Principal is the authenticated identity stored in the request context.
type Principal struct {
	UserID uuid.UUID
	Name   string
}

// Authenticator parses the Authorization header when present.
func Authenticator(tm *token.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				logging.FromContext(r.Context()).DebugContext(r.Context(), "request is anonymous")
				next.ServeHTTP(w, r)
				return
			}
			raw, ok := bearer(header)
			if !ok {
				// The token itself is never logged — only the fact it was unusable.
				logging.FromContext(r.Context()).WarnContext(r.Context(), "malformed authorization header")
				httpx.WriteError(w, r, apperr.Unauthenticated("Malformed Authorization header."))
				return
			}
			claims, err := tm.Parse(raw)
			if err != nil {
				logging.FromContext(r.Context()).WarnContext(r.Context(), "token rejected", "error", err)
				httpx.WriteError(w, r, apperr.Unauthenticated("Invalid or expired token."))
				return
			}

			// Pin user_id onto the request logger so every downstream line — service,
			// repository error, access log — is attributable without re-deriving it.
			log := logging.FromContext(r.Context()).With("user_id", claims.UserID.String())
			ctx := logging.WithLogger(r.Context(), log)
			ctx = context.WithValue(ctx, principalKey, Principal{UserID: claims.UserID, Name: claims.Name})
			log.DebugContext(ctx, "request authenticated")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth blocks anonymous requests with a 401.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := PrincipalFrom(r.Context()); !ok {
			httpx.WriteError(w, r, apperr.Unauthenticated(""))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// PrincipalFrom returns the authenticated principal, if any.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

// CurrentUserID returns a pointer to the authenticated user id, or nil when
// anonymous — the shape repositories expect for optional-identity queries.
func CurrentUserID(ctx context.Context) *uuid.UUID {
	if p, ok := PrincipalFrom(ctx); ok {
		id := p.UserID
		return &id
	}
	return nil
}

func bearer(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	tok := strings.TrimSpace(header[len(prefix):])
	if tok == "" {
		return "", false
	}
	return tok, true
}
