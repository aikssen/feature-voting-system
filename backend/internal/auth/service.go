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
	"soundflow/internal/shared/logging"
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
	// Email is an identifier, not a credential — safe to log. The password and
	// its hash never appear in any line below.
	log := logging.FromContext(ctx)
	name := strings.TrimSpace(req.Name)
	email := normalizeEmail(req.Email)
	log.DebugContext(ctx, "signup requested", "email", email)

	if err := validateSignup(name, email, req.Password); err != nil {
		log.InfoContext(ctx, "signup rejected: validation failed", "email", email)
		return AuthResponse{}, err
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		log.ErrorContext(ctx, "password hashing failed", "error", err)
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
			log.InfoContext(ctx, "signup rejected: email already registered", "email", email)
			return AuthResponse{}, apperr.Validation("Email is already registered.",
				apperr.FieldError{Field: "email", Issue: "already registered"})
		}
		log.ErrorContext(ctx, "signup failed: could not persist user", "email", email, "error", err)
		return AuthResponse{}, apperr.Internal(err)
	}

	log.InfoContext(ctx, "user signed up", "user_id", user.ID.String(), "email", email)
	return s.issue(ctx, user)
}

// Login verifies credentials and returns a token. Failures are indistinguishable
// (no user enumeration): unknown email and wrong password both yield 401.
func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	log := logging.FromContext(ctx)
	email := normalizeEmail(req.Email)
	log.DebugContext(ctx, "login requested", "email", email)

	if email == "" || req.Password == "" {
		log.InfoContext(ctx, "login failed", "email", email, "reason", "missing_credentials")
		return AuthResponse{}, apperr.InvalidCredentials()
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// The client gets an indistinguishable 401; the log keeps the real
			// reason so a support request can be answered.
			log.InfoContext(ctx, "login failed", "email", email, "reason", "unknown_email")
			return AuthResponse{}, apperr.InvalidCredentials()
		}
		log.ErrorContext(ctx, "login failed: user lookup errored", "email", email, "error", err)
		return AuthResponse{}, apperr.Internal(err)
	}
	if err := s.hasher.Compare(user.PasswordHash, req.Password); err != nil {
		log.InfoContext(ctx, "login failed", "email", email, "reason", "wrong_password", "user_id", user.ID.String())
		return AuthResponse{}, apperr.InvalidCredentials()
	}

	log.InfoContext(ctx, "user logged in", "user_id", user.ID.String(), "email", email)
	return s.issue(ctx, user)
}

func (s *Service) issue(ctx context.Context, user *User) (AuthResponse, error) {
	tok, err := s.tokens.Generate(user.ID, user.Name)
	if err != nil {
		logging.FromContext(ctx).ErrorContext(ctx, "token generation failed", "user_id", user.ID.String(), "error", err)
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
