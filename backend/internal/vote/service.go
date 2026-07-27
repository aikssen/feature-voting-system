package vote

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"soundflow/internal/feature"
	"soundflow/internal/shared/apperr"
	"soundflow/internal/shared/logging"
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
	log := logging.FromContext(ctx).With("feature_id", featureID.String())
	log.DebugContext(ctx, "vote requested")

	authorID, err := s.features.GetAuthorID(ctx, featureID)
	if err != nil {
		if errors.Is(err, feature.ErrFeatureNotFD) {
			log.InfoContext(ctx, "vote rejected", "reason", "feature_not_found")
			return VoteResponse{}, apperr.NotFound("Feature request not found.")
		}
		log.ErrorContext(ctx, "vote failed: author lookup errored", "error", err)
		return VoteResponse{}, apperr.Internal(err)
	}

	if authorID == userID {
		log.InfoContext(ctx, "vote rejected", "reason", "self_vote")
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
			// Raised by the DB unique constraint, which is the real invariant —
			// this line is how a lost race shows up in the trace.
			log.InfoContext(ctx, "vote rejected", "reason", "already_voted")
			return VoteResponse{}, apperr.AlreadyVoted()
		}
		log.ErrorContext(ctx, "vote failed: could not persist vote", "error", err)
		return VoteResponse{}, apperr.Internal(err)
	}

	total, err := s.repo.CountByFeature(ctx, featureID)
	if err != nil {
		log.ErrorContext(ctx, "vote recorded but count failed", "vote_id", v.ID.String(), "error", err)
		return VoteResponse{}, apperr.Internal(err)
	}

	log.InfoContext(ctx, "vote recorded", "vote_id", v.ID.String(), "total_votes", total)
	return VoteResponse{FeatureID: featureID, TotalVotes: total, HasVoted: true}, nil
}
