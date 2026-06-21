package feature

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"soundflow/internal/shared/apperr"
)

// Field bounds (DECISIONS.md — validation summary), enforced post-trim.
const (
	titleMin = 2
	titleMax = 100
	descMin  = 2
	descMax  = 200
)

// Service implements feature-request business rules.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// NewServiceWith injects a clock (used by tests).
func NewServiceWith(repo Repository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

// Create validates and persists a new feature request authored by authorID.
func (s *Service) Create(ctx context.Context, authorID uuid.UUID, req CreateRequest) (FeatureView, error) {
	title := strings.TrimSpace(req.Title)
	desc := strings.TrimSpace(req.Description)

	if err := validateCreate(title, desc); err != nil {
		return FeatureView{}, err
	}

	f := &FeatureRequest{
		ID:              uuid.New(),
		Title:           title,
		Description:     desc,
		NormalizedTitle: normalizeTitle(title),
		AuthorID:        authorID,
		CreatedAt:       s.now().UTC(),
	}
	if err := s.repo.Create(ctx, f); err != nil {
		if errors.Is(err, ErrDuplicate) {
			return FeatureView{}, apperr.DuplicateFeature()
		}
		return FeatureView{}, apperr.Internal(err)
	}

	return FeatureView{
		ID:          f.ID,
		Title:       f.Title,
		Description: f.Description,
		CreatedAt:   f.CreatedAt,
		TotalVotes:  0,
		IsAuthor:    true,
		// Author name is unknown here without an extra lookup; the client knows
		// it created this, and the canonical view comes from the next list load.
		Author: AuthorView{ID: authorID},
	}, nil
}

// List returns a page of feature requests with derived rank.
func (s *Service) List(ctx context.Context, currentUserID *uuid.UUID, search, sort string, page, limit int) (PagedFeatures, error) {
	offset := (page - 1) * limit
	items, total, err := s.repo.List(ctx, currentUserID, ListParams{
		Search: search,
		Sort:   sort,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return PagedFeatures{}, apperr.Internal(err)
	}

	// rank is 1-based within the global sorted result (DECISIONS.md O3 —
	// derived, never persisted).
	for i := range items {
		items[i].Rank = offset + i + 1
	}

	totalPages := (total + limit - 1) / limit
	return PagedFeatures{
		Items:      items,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
	}, nil
}

// Get returns a single feature request, or a 404.
func (s *Service) Get(ctx context.Context, currentUserID *uuid.UUID, id uuid.UUID) (FeatureView, error) {
	v, err := s.repo.GetByID(ctx, currentUserID, id)
	if err != nil {
		if errors.Is(err, ErrFeatureNotFD) {
			return FeatureView{}, apperr.NotFound("Feature request not found.")
		}
		return FeatureView{}, apperr.Internal(err)
	}
	v.Rank = 1
	return *v, nil
}

func validateCreate(title, desc string) error {
	var details []apperr.FieldError
	if n := len([]rune(title)); n < titleMin || n > titleMax {
		details = append(details, apperr.FieldError{Field: "title", Issue: "must be 2–100 characters"})
	}
	if n := len([]rune(desc)); n < descMin || n > descMax {
		details = append(details, apperr.FieldError{Field: "description", Issue: "must be 2–200 characters"})
	}
	if len(details) > 0 {
		return apperr.Validation("One or more fields are invalid.", details...)
	}
	return nil
}
