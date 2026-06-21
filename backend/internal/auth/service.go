package auth

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"soundflow/internal/shared/apperr"
)

// Password policy (DECISIONS.md D-AUTH): 4–12 chars, at least one special char.
const (
	passwordMin = 4
	passwordMax = 12
	nameMin     = 2
	nameMax     = 40
)

const specialChars = `!@#$%^&*()_+-=[]{};:'",.<>/?\|` + "`~"

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Hasher abstracts password hashing so the service is unit-testable without
// paying bcrypt's cost in every test.
type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// TokenIssuer issues signed auth tokens (satisfied by shared/token.Manager).
type TokenIssuer interface {
	Generate(userID uuid.UUID, name string) (string, error)
}

// Service implements the authentication business rules.
type Service struct {
	repo   Repository
	hasher Hasher
	tokens TokenIssuer
	now    func() time.Time
}

// NewService wires the auth service with the production bcrypt hasher.
func NewService(repo Repository, tokens TokenIssuer) *Service {
	return &Service{repo: repo, hasher: BcryptHasher{}, tokens: tokens, now: time.Now}
}

// NewServiceWith allows injecting a custom hasher/clock (used by tests).
func NewServiceWith(repo Repository, hasher Hasher, tokens TokenIssuer, now func() time.Time) *Service {
	return &Service{repo: repo, hasher: hasher, tokens: tokens, now: now}
}

// Signup validates the request, creates the user, and auto-logs-in by returning
// a token (DECISIONS.md M1).
func (s *Service) Signup(ctx context.Context, req SignupRequest) (AuthResponse, error) {
	name := strings.TrimSpace(req.Name)
	email := normalizeEmail(req.Email)

	if err := validateSignup(name, email, req.Password); err != nil {
		return AuthResponse{}, err
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return AuthResponse{}, apperr.Internal(err)
	}

	user := &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    s.now().UTC(),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, ErrEmailTaken) {
			return AuthResponse{}, apperr.Validation("Email is already registered.",
				apperr.FieldError{Field: "email", Issue: "already registered"})
		}
		return AuthResponse{}, apperr.Internal(err)
	}
	return s.issue(user)
}

// Login verifies credentials and returns a token. Failures are indistinguishable
// (no user enumeration): unknown email and wrong password both yield 401.
func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	email := normalizeEmail(req.Email)
	if email == "" || req.Password == "" {
		return AuthResponse{}, apperr.InvalidCredentials()
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return AuthResponse{}, apperr.InvalidCredentials()
		}
		return AuthResponse{}, apperr.Internal(err)
	}
	if err := s.hasher.Compare(user.PasswordHash, req.Password); err != nil {
		return AuthResponse{}, apperr.InvalidCredentials()
	}
	return s.issue(user)
}

func (s *Service) issue(user *User) (AuthResponse, error) {
	tok, err := s.tokens.Generate(user.ID, user.Name)
	if err != nil {
		return AuthResponse{}, apperr.Internal(err)
	}
	return AuthResponse{Token: tok, User: user.View()}, nil
}

func validateSignup(name, email, password string) error {
	var details []apperr.FieldError

	if n := len([]rune(name)); n < nameMin || n > nameMax {
		details = append(details, apperr.FieldError{Field: "name", Issue: "must be 2–40 characters"})
	}
	if !emailRegex.MatchString(email) {
		details = append(details, apperr.FieldError{Field: "email", Issue: "must be a valid email address"})
	}
	if n := len([]rune(password)); n < passwordMin || n > passwordMax {
		details = append(details, apperr.FieldError{Field: "password", Issue: "must be 4–12 characters"})
	} else if !strings.ContainsAny(password, specialChars) {
		details = append(details, apperr.FieldError{Field: "password", Issue: "must contain at least one special character"})
	}

	if len(details) > 0 {
		return apperr.Validation("One or more fields are invalid.", details...)
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// BcryptHasher is the production password hasher.
type BcryptHasher struct{}

func (BcryptHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func (BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
