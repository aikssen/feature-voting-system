package vote

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"soundflow/internal/feature"
	"soundflow/internal/shared/apperr"
)

type mockRepo struct {
	createErr error
	count     int
	created   *Vote
}

func (m *mockRepo) Create(_ context.Context, v *Vote) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.created = v
	return nil
}

func (m *mockRepo) CountByFeature(_ context.Context, _ uuid.UUID) (int, error) {
	return m.count, nil
}

// mockAuthorReader returns a fixed author, or a not-found error.
type mockAuthorReader struct {
	authorID uuid.UUID
	notFound bool
}

func (m mockAuthorReader) GetAuthorID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	if m.notFound {
		return uuid.Nil, feature.ErrFeatureNotFD
	}
	return m.authorID, nil
}

func newSvc(repo Repository, reader AuthorReader) *Service {
	return NewServiceWith(repo, reader, func() time.Time { return time.Unix(0, 0).UTC() })
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	ae, ok := apperr.As(err)
	if !ok {
		t.Fatalf("expected apperr, got %v", err)
	}
	return ae.Code
}

func TestVote_Success(t *testing.T) {
	author := uuid.New()
	voter := uuid.New()
	feat := uuid.New()
	repo := &mockRepo{count: 7}

	res, err := newSvc(repo, mockAuthorReader{authorID: author}).Vote(context.Background(), feat, voter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalVotes != 7 || !res.HasVoted || res.FeatureID != feat {
		t.Errorf("unexpected response: %+v", res)
	}
	if repo.created.UserID != voter || repo.created.FeatureRequestID != feat {
		t.Errorf("vote persisted with wrong ids: %+v", repo.created)
	}
}

func TestVote_SelfVoteForbidden(t *testing.T) {
	me := uuid.New()
	_, err := newSvc(&mockRepo{}, mockAuthorReader{authorID: me}).Vote(context.Background(), uuid.New(), me)
	if got := codeOf(t, err); got != "SELF_VOTE_FORBIDDEN" {
		t.Errorf("code = %s, want SELF_VOTE_FORBIDDEN", got)
	}
}

func TestVote_DuplicateRejected(t *testing.T) {
	repo := &mockRepo{createErr: ErrAlreadyVoted}
	_, err := newSvc(repo, mockAuthorReader{authorID: uuid.New()}).Vote(context.Background(), uuid.New(), uuid.New())
	if got := codeOf(t, err); got != "ALREADY_VOTED" {
		t.Errorf("code = %s, want ALREADY_VOTED", got)
	}
}

func TestVote_FeatureNotFound(t *testing.T) {
	_, err := newSvc(&mockRepo{}, mockAuthorReader{notFound: true}).Vote(context.Background(), uuid.New(), uuid.New())
	if got := codeOf(t, err); got != "NOT_FOUND" {
		t.Errorf("code = %s, want NOT_FOUND", got)
	}
}
