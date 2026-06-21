package vote

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrAlreadyVoted is returned when the (user, feature) unique constraint trips
// (DECISIONS.md R1 — the DB constraint is the source of truth).
var ErrAlreadyVoted = errors.New("already voted")

const uniqueViolation = "23505"

// Repository abstracts vote persistence.
type Repository interface {
	Create(ctx context.Context, v *Vote) error
	CountByFeature(ctx context.Context, featureID uuid.UUID) (int, error)
}

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) Create(ctx context.Context, v *Vote) error {
	const q = `
        INSERT INTO votes (id, feature_request_id, user_id, created_at)
        VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, q, v.ID, v.FeatureRequestID, v.UserID, v.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return ErrAlreadyVoted
		}
		return err
	}
	return nil
}

func (r *postgresRepository) CountByFeature(ctx context.Context, featureID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM votes WHERE feature_request_id = $1`
	var count int
	if err := r.pool.QueryRow(ctx, q, featureID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
