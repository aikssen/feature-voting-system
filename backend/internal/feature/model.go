package feature

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// FeatureRequest is the domain model.
type FeatureRequest struct {
	ID              uuid.UUID
	Title           string
	Description     string
	NormalizedTitle string
	AuthorID        uuid.UUID
	CreatedAt       time.Time
}

// CreateRequest is the POST /features body.
type CreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// AuthorView is the embedded author projection.
type AuthorView struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// FeatureView is the API projection (DECISIONS.md — FeatureView shape).
type FeatureView struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Author        AuthorView `json:"author"`
	CreatedAt     time.Time  `json:"created_at"`
	TotalVotes    int        `json:"total_votes"`
	TrendingScore float64    `json:"trending_score"`
	HasVoted      bool       `json:"has_voted"`
	IsAuthor      bool       `json:"is_author"`
	Rank          int        `json:"rank"`
}

// ListParams are the normalized discovery parameters passed to the repository.
type ListParams struct {
	Search string
	Sort   string
	Limit  int
	Offset int
}

// PagedFeatures is the Page<FeatureView> wrapper (DECISIONS.md — Page<T>).
type PagedFeatures struct {
	Items      []FeatureView `json:"items"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	Total      int           `json:"total"`
	TotalPages int           `json:"total_pages"`
	HasNext    bool          `json:"has_next"`
}

// normalizeTitle collapses case/whitespace for the duplicate-detection guard
// (DECISIONS.md C1): trim, lowercase, collapse internal whitespace.
func normalizeTitle(title string) string {
	return strings.ToLower(strings.Join(strings.Fields(title), " "))
}
