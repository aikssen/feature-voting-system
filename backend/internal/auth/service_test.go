package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"soundflow/internal/shared/apperr"
)

// --- test doubles ---

type mockRepo struct {
	created   *User
	byEmail   map[string]*User
	createErr error
}

func (m *mockRepo) Create(_ context.Context, u *User) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.created = u
	return nil
}

func (m *mockRepo) GetByEmail(_ context.Context, email string) (*User, error) {
	if u, ok := m.byEmail[email]; ok {
		return u, nil
	}
	return nil, ErrUserNotFound
}

func (m *mockRepo) GetByID(_ context.Context, _ uuid.UUID) (*User, error) {
	return nil, ErrUserNotFound
}

// plainHasher is a deterministic, fast stand-in for bcrypt.
type plainHasher struct{}

func (plainHasher) Hash(p string) (string, error) { return "hashed:" + p, nil }
func (plainHasher) Compare(hash, p string) error {
	if hash == "hashed:"+p {
		return nil
	}
	return errors.New("mismatch")
}

type stubTokens struct{}

func (stubTokens) Generate(uuid.UUID, string) (string, error) { return "test-token", nil }

func newSvc(repo Repository) *Service {
	return NewServiceWith(repo, plainHasher{}, stubTokens{}, func() time.Time { return time.Unix(0, 0).UTC() })
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	ae, ok := apperr.As(err)
	if !ok {
		t.Fatalf("expected apperr, got %v", err)
	}
	return ae.Code
}

// --- signup ---

func TestSignup_Success(t *testing.T) {
	repo := &mockRepo{}
	res, err := newSvc(repo).Signup(context.Background(), SignupRequest{
		Name: "Ever", Email: "Ever@Example.com ", Password: "Pa$s",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token != "test-token" {
		t.Errorf("token = %q", res.Token)
	}
	if repo.created.Email != "ever@example.com" {
		t.Errorf("email not normalized: %q", repo.created.Email)
	}
	if repo.created.PasswordHash != "hashed:Pa$s" {
		t.Errorf("password not hashed: %q", repo.created.PasswordHash)
	}
}

func TestSignup_ValidationErrors(t *testing.T) {
	cases := map[string]SignupRequest{
		"short name":        {Name: "E", Email: "e@x.com", Password: "Pa$s"},
		"bad email":         {Name: "Ever", Email: "not-an-email", Password: "Pa$s"},
		"no special char":   {Name: "Ever", Email: "e@x.com", Password: "passw"},
		"password too long": {Name: "Ever", Email: "e@x.com", Password: "Password123!!"},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := newSvc(&mockRepo{}).Signup(context.Background(), req)
			if got := codeOf(t, err); got != "VALIDATION_ERROR" {
				t.Errorf("code = %s, want VALIDATION_ERROR", got)
			}
		})
	}
}

func TestSignup_DuplicateEmail(t *testing.T) {
	repo := &mockRepo{createErr: ErrEmailTaken}
	_, err := newSvc(repo).Signup(context.Background(), SignupRequest{
		Name: "Ever", Email: "e@x.com", Password: "Pa$s",
	})
	if got := codeOf(t, err); got != "VALIDATION_ERROR" {
		t.Errorf("code = %s, want VALIDATION_ERROR", got)
	}
}

// --- login ---

func TestLogin_Success(t *testing.T) {
	repo := &mockRepo{byEmail: map[string]*User{
		"e@x.com": {ID: uuid.New(), Name: "Ever", Email: "e@x.com", PasswordHash: "hashed:Pa$s"},
	}}
	res, err := newSvc(repo).Login(context.Background(), LoginRequest{Email: "e@x.com", Password: "Pa$s"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token == "" {
		t.Error("expected token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &mockRepo{byEmail: map[string]*User{
		"e@x.com": {ID: uuid.New(), Email: "e@x.com", PasswordHash: "hashed:Pa$s"},
	}}
	_, err := newSvc(repo).Login(context.Background(), LoginRequest{Email: "e@x.com", Password: "nope!"})
	if got := codeOf(t, err); got != "INVALID_CREDENTIALS" {
		t.Errorf("code = %s, want INVALID_CREDENTIALS", got)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	_, err := newSvc(&mockRepo{}).Login(context.Background(), LoginRequest{Email: "ghost@x.com", Password: "Pa$s"})
	if got := codeOf(t, err); got != "INVALID_CREDENTIALS" {
		t.Errorf("code = %s, want INVALID_CREDENTIALS", got)
	}
}
