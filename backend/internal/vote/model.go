package vote

import (
	"time"

	"github.com/google/uuid"
)

// Vote is the domain model.
type Vote struct {
	ID               uuid.UUID
	FeatureRequestID uuid.UUID
	UserID           uuid.UUID
	CreatedAt        time.Time
}

// VoteResponse is the POST /features/{id}/vote response body.
type VoteResponse struct {
	FeatureID  uuid.UUID `json:"feature_id"`
	TotalVotes int       `json:"total_votes"`
	HasVoted   bool      `json:"has_voted"`
}
