package infrastructure

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// schemaSQL is the idempotent schema. Every statement uses IF NOT EXISTS so the
// migration can run on every boot (DECISIONS.md — data model deltas).
const schemaSQL = `
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id            uuid PRIMARY KEY,
    name          text  NOT NULL,
    email         citext NOT NULL,
    password_hash text  NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users (email);

CREATE TABLE IF NOT EXISTS feature_requests (
    id               uuid PRIMARY KEY,
    title            text NOT NULL,
    description      text NOT NULL,
    normalized_title text NOT NULL,
    author_id        uuid NOT NULL REFERENCES users (id),
    created_at       timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_feature_requests_normalized_title ON feature_requests (normalized_title);
CREATE INDEX IF NOT EXISTS ix_feature_requests_created_at ON feature_requests (created_at);

CREATE TABLE IF NOT EXISTS votes (
    id                 uuid PRIMARY KEY,
    feature_request_id uuid NOT NULL REFERENCES feature_requests (id),
    user_id            uuid NOT NULL REFERENCES users (id),
    created_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ux_votes_user_feature UNIQUE (user_id, feature_request_id)
);
CREATE INDEX IF NOT EXISTS ix_votes_feature_request_id ON votes (feature_request_id);
CREATE INDEX IF NOT EXISTS ix_votes_user_id ON votes (user_id);
`

// Migrate applies the schema. Safe to call on every startup.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
