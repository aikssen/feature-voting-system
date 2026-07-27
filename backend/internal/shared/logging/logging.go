// Package logging centralises structured logging for the API.
//
// Three things live here (DECISIONS.md D-LOG):
//   - Level configuration from LOG_LEVEL, applied to a single JSON slog handler
//     so every line in stdout is machine-parseable.
//   - The correlation id that ties one browser action to all the server-side work
//     it triggers. The frontend sends it as X-Correlation-ID; if it is missing or
//     malformed the middleware mints one and echoes it back on the response.
//   - The request-scoped logger. Handlers and services pull it off the context
//     with FromContext, so every line they emit already carries the correlation
//     id, method and path without threading a logger through every signature.
//
// Never log tokens, password hashes, or raw passwords (CLAUDE.md guardrail).
package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
)

// CorrelationIDHeader is the wire name of the trace id, shared with the frontend.
const CorrelationIDHeader = "X-Correlation-ID"

// CorrelationIDKey is the slog attribute key every request-scoped line carries.
const CorrelationIDKey = "correlation_id"

// maxCorrelationIDLen bounds what we accept from a client so a hostile header
// cannot bloat every log line it touches.
const maxCorrelationIDLen = 128

type ctxKey int

const (
	loggerKey ctxKey = iota
	correlationIDKey
)

// ParseLevel maps a LOG_LEVEL value to a slog level. Unknown values are an error
// so a typo surfaces at boot instead of silently downgrading observability.
func ParseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid LOG_LEVEL %q (want debug, info, warn or error)", raw)
	}
}

// New builds the application logger: JSON to stdout at the configured level.
func New(rawLevel string) (*slog.Logger, error) {
	level, err := ParseLevel(rawLevel)
	if err != nil {
		return nil, err
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})), nil
}

// WithLogger stores a request-scoped logger on the context.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns the request-scoped logger, falling back to the default one
// so callers outside an HTTP request (boot, shutdown, tests) still log.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

// WithCorrelationID stores the trace id so non-logging code (e.g. outbound calls)
// can propagate it.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationIDFrom returns the request's trace id, or "" outside a request.
func CorrelationIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(correlationIDKey).(string)
	return id
}

// NewCorrelationID mints an id for requests that arrive without one.
func NewCorrelationID() string { return uuid.NewString() }

// NormalizeCorrelationID accepts a client-supplied id only when it is short and
// made of safe characters; anything else is replaced with a fresh id. Returning
// the id verbatim into logs is exactly the injection vector this closes.
func NormalizeCorrelationID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > maxCorrelationIDLen {
		return NewCorrelationID()
	}
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == ':':
		default:
			return NewCorrelationID()
		}
	}
	return raw
}
