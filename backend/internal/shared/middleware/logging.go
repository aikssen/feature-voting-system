package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"

	"soundflow/internal/shared/logging"
)

// RequestLogger is the observability entry point for every request.
//
// It establishes the correlation id (from the client's X-Correlation-ID header
// when usable, otherwise freshly minted), echoes it back so the browser can tie
// its own log lines to the server's, publishes a request-scoped logger on the
// context, and emits one access-log line per request.
//
// This replaces chi's stock Logger, which writes unstructured text and would
// interleave unparseable lines into the JSON stream.
func RequestLogger(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			correlationID := logging.NormalizeCorrelationID(r.Header.Get(logging.CorrelationIDHeader))
			w.Header().Set(logging.CorrelationIDHeader, correlationID)

			reqLogger := base.With(
				logging.CorrelationIDKey, correlationID,
				"method", r.Method,
				"path", r.URL.Path,
			)

			ctx := logging.WithCorrelationID(r.Context(), correlationID)
			ctx = logging.WithLogger(ctx, reqLogger)
			r = r.WithContext(ctx)

			reqLogger.DebugContext(ctx, "request started", "query", r.URL.RawQuery, "remote_addr", r.RemoteAddr)

			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			defer func() {
				status := ww.Status()
				if status == 0 { // handler wrote nothing; net/http will send a 200
					status = http.StatusOK
				}
				reqLogger.Log(ctx, levelForStatus(status), "request completed",
					"status", status,
					"bytes", ww.BytesWritten(),
					"duration_ms", time.Since(start).Milliseconds(),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// levelForStatus keeps the access log readable at LOG_LEVEL=warn: only failures
// survive, and only 5xx counts as an error the operator must act on.
func levelForStatus(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
