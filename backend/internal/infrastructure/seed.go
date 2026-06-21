package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"soundflow/internal/auth"
)

// seedPassword is the shared password for all seeded demo accounts. It satisfies
// the password policy (4–12 chars, ≥1 special).
const seedPassword = "Pa$s"

type seedUser struct {
	name  string
	email string
}

type seedFeature struct {
	title       string
	description string
	authorEmail string
	ageHours    int
}

// Seed populates demo data, but only when the database is empty (idempotent,
// per DECISIONS.md). Safe to call on every boot.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM feature_requests`).Scan(&count); err != nil {
		return fmt.Errorf("check seed state: %w", err)
	}
	if count > 0 {
		return nil // already seeded
	}

	hash, err := auth.BcryptHasher{}.Hash(seedPassword)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}

	users := []seedUser{
		{"Ever", "ever@example.com"},
		{"Mia", "mia@example.com"},
		{"Leo", "leo@example.com"},
		{"Nina", "nina@example.com"},
		{"Sam", "sam@example.com"},
	}
	features := []seedFeature{
		{"Offline Downloads", "Download playlists for flights and travel.", "ever@example.com", 48},
		{"Spatial Audio", "Immersive 3D sound for headphones.", "mia@example.com", 5},
		{"Collaborative Playlists", "Build playlists together in real time.", "leo@example.com", 120},
		{"Sleep Timer", "Auto-stop playback after a set time.", "ever@example.com", 12},
		{"Lyrics Sync", "Real-time synced lyrics while listening.", "mia@example.com", 72},
		{"Crossfade", "Smooth transitions between tracks.", "leo@example.com", 1},
	}
	// (featureTitle, voterEmail) — authors never vote on their own request.
	votes := [][2]string{
		{"Collaborative Playlists", "ever@example.com"},
		{"Collaborative Playlists", "mia@example.com"},
		{"Collaborative Playlists", "nina@example.com"},
		{"Collaborative Playlists", "sam@example.com"},
		{"Offline Downloads", "mia@example.com"},
		{"Offline Downloads", "leo@example.com"},
		{"Offline Downloads", "nina@example.com"},
		{"Offline Downloads", "sam@example.com"},
		{"Spatial Audio", "ever@example.com"},
		{"Spatial Audio", "leo@example.com"},
		{"Spatial Audio", "nina@example.com"},
		{"Lyrics Sync", "ever@example.com"},
		{"Lyrics Sync", "sam@example.com"},
		{"Sleep Timer", "leo@example.com"},
		{"Crossfade", "ever@example.com"},
	}

	now := time.Now().UTC()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback(ctx)

	userIDs := make(map[string]uuid.UUID, len(users))
	for _, u := range users {
		id := uuid.New()
		userIDs[u.email] = id
		if _, err := tx.Exec(ctx,
			`INSERT INTO users (id, name, email, password_hash, created_at) VALUES ($1,$2,$3,$4,$5)`,
			id, u.name, u.email, hash, now,
		); err != nil {
			return fmt.Errorf("seed user %s: %w", u.email, err)
		}
	}

	featureIDs := make(map[string]uuid.UUID, len(features))
	for _, f := range features {
		id := uuid.New()
		featureIDs[f.title] = id
		createdAt := now.Add(-time.Duration(f.ageHours) * time.Hour)
		if _, err := tx.Exec(ctx,
			`INSERT INTO feature_requests (id, title, description, normalized_title, author_id, created_at)
             VALUES ($1,$2,$3,$4,$5,$6)`,
			id, f.title, f.description, normalizeTitle(f.title), userIDs[f.authorEmail], createdAt,
		); err != nil {
			return fmt.Errorf("seed feature %q: %w", f.title, err)
		}
	}

	for _, v := range votes {
		if _, err := tx.Exec(ctx,
			`INSERT INTO votes (id, feature_request_id, user_id, created_at) VALUES ($1,$2,$3,$4)`,
			uuid.New(), featureIDs[v[0]], userIDs[v[1]], now,
		); err != nil {
			return fmt.Errorf("seed vote %q/%s: %w", v[0], v[1], err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed: %w", err)
	}
	slog.Info("seeded demo data", "users", len(users), "features", len(features), "votes", len(votes))
	return nil
}

// normalizeTitle mirrors feature.normalizeTitle for the seed's duplicate guard.
func normalizeTitle(title string) string {
	return strings.ToLower(strings.Join(strings.Fields(title), " "))
}
