package feature

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"soundflow/internal/shared/httpx"
)

var (
	ErrDuplicate    = errors.New("duplicate feature request")
	ErrFeatureNotFD = errors.New("feature request not found")
)

const uniqueViolation = "23505"

// trendingExpr is the frozen v1 trending score (DECISIONS.md D-TREND):
//
//	votes / (age_hours + 2) ^ 1.5
const trendingExpr = `COUNT(v.id)::float8 / power(EXTRACT(EPOCH FROM (now() - f.created_at)) / 3600.0 + 2, 1.5)`

// selectColumns is shared by List and GetByID so both projections stay identical.
// $1 is the optional current user id (nullable) used for has_voted / is_author.
const selectColumns = `
        f.id, f.title, f.description, f.created_at,
        u.id   AS author_id,
        u.name AS author_name,
        COUNT(v.id)                                      AS total_votes,
        ` + trendingExpr + `                             AS trending_score,
        COUNT(*) FILTER (WHERE v.user_id = $1::uuid) > 0 AS has_voted,
        COALESCE(f.author_id = $1::uuid, false)          AS is_author`

// Repository abstracts feature persistence. The vote slice depends only on the
// AuthorReader subset to stay decoupled.
type Repository interface {
	Create(ctx context.Context, f *FeatureRequest) error
	List(ctx context.Context, currentUserID *uuid.UUID, p ListParams) ([]FeatureView, int, error)
	GetByID(ctx context.Context, currentUserID *uuid.UUID, id uuid.UUID) (*FeatureView, error)
	GetAuthorID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) Create(ctx context.Context, f *FeatureRequest) error {
	const q = `
        INSERT INTO feature_requests (id, title, description, normalized_title, author_id, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, q, f.ID, f.Title, f.Description, f.NormalizedTitle, f.AuthorID, f.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

func (r *postgresRepository) List(ctx context.Context, currentUserID *uuid.UUID, p ListParams) ([]FeatureView, int, error) {
	cur := nullableUUID(currentUserID)

	// Total count for pagination metadata (same search predicate, no joins).
	var total int
	const countQ = `
        SELECT COUNT(*) FROM feature_requests f
        WHERE ($1 = '' OR f.title ILIKE '%' || $1 || '%' OR f.description ILIKE '%' || $1 || '%')`
	if err := r.pool.QueryRow(ctx, countQ, p.Search).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []FeatureView{}, 0, nil
	}

	// $1 current user, $2 search, $3 limit, $4 offset.
	q := fmt.Sprintf(`
        SELECT %s
        FROM feature_requests f
        JOIN users u ON u.id = f.author_id
        LEFT JOIN votes v ON v.feature_request_id = f.id
        WHERE ($2 = '' OR f.title ILIKE '%%' || $2 || '%%' OR f.description ILIKE '%%' || $2 || '%%')
        GROUP BY f.id, u.id
        ORDER BY %s
        LIMIT $3 OFFSET $4`, selectColumns, orderBy(p.Sort))

	rows, err := r.pool.Query(ctx, q, cur, p.Search, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	views := make([]FeatureView, 0, p.Limit)
	for rows.Next() {
		v, err := scanView(rows)
		if err != nil {
			return nil, 0, err
		}
		views = append(views, v)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return views, total, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, currentUserID *uuid.UUID, id uuid.UUID) (*FeatureView, error) {
	cur := nullableUUID(currentUserID)
	// $1 current user, $2 feature id.
	q := fmt.Sprintf(`
        SELECT %s
        FROM feature_requests f
        JOIN users u ON u.id = f.author_id
        LEFT JOIN votes v ON v.feature_request_id = f.id
        WHERE f.id = $2::uuid
        GROUP BY f.id, u.id`, selectColumns)

	v, err := scanView(r.pool.QueryRow(ctx, q, cur, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFeatureNotFD
		}
		return nil, err
	}
	return &v, nil
}

func (r *postgresRepository) GetAuthorID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	const q = `SELECT author_id FROM feature_requests WHERE id = $1`
	var authorID uuid.UUID
	if err := r.pool.QueryRow(ctx, q, id).Scan(&authorID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrFeatureNotFD
		}
		return uuid.Nil, err
	}
	return authorID, nil
}

func scanView(row pgx.Row) (FeatureView, error) {
	var v FeatureView
	err := row.Scan(
		&v.ID, &v.Title, &v.Description, &v.CreatedAt,
		&v.Author.ID, &v.Author.Name,
		&v.TotalVotes, &v.TrendingScore, &v.HasVoted, &v.IsAuthor,
	)
	return v, err
}

// orderBy maps the validated sort token to a SQL clause. The input is already
// constrained to a known set by httpx.ParseListQuery, so there is no injection
// surface; the switch is a second guard.
func orderBy(sort string) string {
	switch sort {
	case httpx.SortNewest:
		return "f.created_at DESC"
	case httpx.SortMostVoted:
		return "total_votes DESC, f.created_at DESC"
	default: // trending
		return "trending_score DESC, f.created_at DESC"
	}
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return id.String()
}
