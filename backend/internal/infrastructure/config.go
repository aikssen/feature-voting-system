package infrastructure

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds runtime configuration loaded from the environment.
// See DECISIONS.md (D-ENV) for the authoritative variable list.
type Config struct {
	DatabaseURL        string
	JWTSecret          string
	JWTTTLHours        int
	Port               string
	LogLevel           string
	CORSAllowedOrigins []string
}

// LoadConfig reads configuration from the environment and validates the
// mandatory fields. The app must refuse to boot without a JWT secret.
//
// BACKEND_PORT is the port this process binds to. Under Docker Compose the
// container always binds 3000 (Compose overrides the value from .env) and the
// .env knob only picks the published host port; natively it is the real port.
func LoadConfig() (Config, error) {
	cfg := Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTTTLHours:        envInt("JWT_TTL_HOURS", 24),
		Port:               envString("BACKEND_PORT", "3000"),
		LogLevel:           envString("LOG_LEVEL", "info"),
		CORSAllowedOrigins: envList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
	}

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

func envString(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key))); err == nil && v > 0 {
		return v
	}
	return fallback
}

func envList(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
