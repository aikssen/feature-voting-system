package vote

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"soundflow/internal/feature"
	"soundflow/internal/shared/apperr"
)

// AuthorReader is the narrow slice of the feature repository the vote service
// needs. Declaring it here (consumer-side) keeps the slices decoupled.
type AuthorReader interface {
	GetAuthorID(ctx context.Context, featureID uuid.UUID) (uuid.UUID, error)
}

// Service enforces the voting business rules (DECISIONS.md D-VOTE).
type Service struct {
	repo     Repository
	features AuthorReader
	now      func() time.Time
}

func NewService(repo Repository, features AuthorReader) *Service {
	return &Service{repo: repo, features: features, now: time.Now}
}

// NewServiceWith injects a clock (used by tests).
func NewServiceWith(repo Repository, features AuthorReader, now func() time.Time) *Service {
	return &Service{repo: repo, features: features, now: now}
}

// Vote records a vote by userID on featureID, enforcing: feature must exist
// (404), no self-voting (403), one vote per user per feature (409).
func (s *Service) Vote(ctx context.Context, featureID, userID uuid.UUID) (VoteResponse, error) {
	authorID, err := s.features.GetAuthorID(ctx, featureID)
	if err != nil {
		if errors.Is(err, feature.ErrFeatureNotFD) {
			return VoteResponse{}, apperr.NotFound("Feature request not found.")
		}
		return VoteResponse{}, apperr.Internal(err)
	}

	if authorID == userID {
		return VoteResponse{}, apperr.SelfVoteForbidden()
	}

	v := &Vote{
		ID:               uuid.New(),
		FeatureRequestID: featureID,
		UserID:           userID,
		CreatedAt:        s.now().UTC(),
	}
	if err := s.repo.Create(ctx, v); err != nil {
		if errors.Is(err, ErrAlreadyVoted) {
			return VoteResponse{}, apperr.AlreadyVoted()
		}
		return VoteResponse{}, apperr.Internal(err)
	}

	total, err := s.repo.CountByFeature(ctx, featureID)
	if err != nil {
		return VoteResponse{}, apperr.Internal(err)
	}
	return VoteResponse{FeatureID: featureID, TotalVotes: total, HasVoted: true}, nil
}
