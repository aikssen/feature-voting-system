package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors translated to HTTP status by the service layer.
var (
	ErrEmailTaken   = errors.New("email already registered")
	ErrUserNotFound = errors.New("user not found")
)

const uniqueViolation = "23505"

// Repository abstracts user persistence so the service can be unit-tested with
// a mock (DECISIONS.md O1 — abstraction justified by testability).
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns the Postgres-backed user repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) Create(ctx context.Context, u *User) error {
	const q = `
        INSERT INTO users (id, name, email, password_hash, created_at)
        VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, q, u.ID, u.Name, u.Email, u.PasswordHash, u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return ErrEmailTaken
		}
		return err
	}
	return nil
}

func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
        SELECT id, name, email, password_hash, created_at
        FROM users WHERE email = $1`
	return scanUser(r.pool.QueryRow(ctx, q, email))
}

func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	const q = `
        SELECT id, name, email, password_hash, created_at
        FROM users WHERE id = $1`
	return scanUser(r.pool.QueryRow(ctx, q, id))
}

func scanUser(row pgx.Row) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}
