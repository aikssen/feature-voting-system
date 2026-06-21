package feature

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"soundflow/internal/shared/apperr"
)

type mockRepo struct {
	createErr error
	created   *FeatureRequest
	items     []FeatureView
	total     int
}

func (m *mockRepo) Create(_ context.Context, f *FeatureRequest) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.created = f
	return nil
}

func (m *mockRepo) List(_ context.Context, _ *uuid.UUID, _ ListParams) ([]FeatureView, int, error) {
	return m.items, m.total, nil
}

func (m *mockRepo) GetByID(_ context.Context, _ *uuid.UUID, _ uuid.UUID) (*FeatureView, error) {
	return nil, ErrFeatureNotFD
}

func (m *mockRepo) GetAuthorID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, ErrFeatureNotFD
}

func newSvc(repo Repository) *Service {
	return NewServiceWith(repo, func() time.Time { return time.Unix(0, 0).UTC() })
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	ae, ok := apperr.As(err)
	if !ok {
		t.Fatalf("expected apperr, got %v", err)
	}
	return ae.Code
}

func TestCreate_Success(t *testing.T) {
	repo := &mockRepo{}
	author := uuid.New()
	view, err := newSvc(repo).Create(context.Background(), author, CreateRequest{
		Title: "  Offline Downloads  ", Description: "Download playlists for travel.",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created.Title != "Offline Downloads" {
		t.Errorf("title not trimmed: %q", repo.created.Title)
	}
	if repo.created.NormalizedTitle != "offline downloads" {
		t.Errorf("normalized title = %q", repo.created.NormalizedTitle)
	}
	if !view.IsAuthor || view.Author.ID != author {
		t.Errorf("author not set on view: %+v", view)
	}
}

func TestCreate_ValidationErrors(t *testing.T) {
	cases := map[string]CreateRequest{
		"title too short":  {Title: "x", Description: "valid description"},
		"title whitespace": {Title: "   ", Description: "valid description"},
		"desc too short":   {Title: "Valid title", Description: "y"},
		"title too long":   {Title: strings.Repeat("a", 101), Description: "valid description"},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := newSvc(&mockRepo{}).Create(context.Background(), uuid.New(), req)
			if got := codeOf(t, err); got != "VALIDATION_ERROR" {
				t.Errorf("code = %s, want VALIDATION_ERROR", got)
			}
		})
	}
}

func TestCreate_Duplicate(t *testing.T) {
	repo := &mockRepo{createErr: ErrDuplicate}
	_, err := newSvc(repo).Create(context.Background(), uuid.New(), CreateRequest{
		Title: "Offline Downloads", Description: "Download playlists for travel.",
	})
	if got := codeOf(t, err); got != "DUPLICATE_FEATURE" {
		t.Errorf("code = %s, want DUPLICATE_FEATURE", got)
	}
}

func TestList_RankAndPagination(t *testing.T) {
	repo := &mockRepo{
		items: []FeatureView{{Title: "a"}, {Title: "b"}},
		total: 12,
	}
	// page 2, limit 5 -> offset 5 -> ranks 6, 7; total_pages 3; has_next true.
	page, err := newSvc(repo).List(context.Background(), nil, "", "trending", 2, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Items[0].Rank != 6 || page.Items[1].Rank != 7 {
		t.Errorf("ranks = %d,%d want 6,7", page.Items[0].Rank, page.Items[1].Rank)
	}
	if page.TotalPages != 3 {
		t.Errorf("total_pages = %d want 3", page.TotalPages)
	}
	if !page.HasNext {
		t.Error("expected has_next true on page 2 of 3")
	}
}

func TestList_LastPageHasNoNext(t *testing.T) {
	repo := &mockRepo{items: []FeatureView{{Title: "a"}}, total: 11}
	// page 3, limit 5 -> total_pages 3 -> has_next false.
	page, err := newSvc(repo).List(context.Background(), nil, "", "newest", 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.HasNext {
		t.Error("expected has_next false on last page")
	}
	if page.Items[0].Rank != 11 {
		t.Errorf("rank = %d want 11", page.Items[0].Rank)
	}
}
