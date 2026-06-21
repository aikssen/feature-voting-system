// Package token signs and parses the application's JWTs (HS256), per the
// frozen contract: claims { sub, name, iat, exp }, configurable TTL, no refresh.
package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken is returned for any malformed, mis-signed, or expired token.
var ErrInvalidToken = errors.New("invalid token")

// Claims carries the authenticated identity extracted from a token.
type Claims struct {
	UserID uuid.UUID
	Name   string
}

// Manager signs and verifies tokens with a fixed secret + TTL.
type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewManager builds a Manager. ttlHours <= 0 falls back to 24h.
func NewManager(secret string, ttlHours int) *Manager {
	if ttlHours <= 0 {
		ttlHours = 24
	}
	return &Manager{
		secret: []byte(secret),
		ttl:    time.Duration(ttlHours) * time.Hour,
		now:    time.Now,
	}
}

// Generate issues a signed token for the given user.
func (m *Manager) Generate(userID uuid.UUID, name string) (string, error) {
	now := m.now()
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"name": name,
		"iat":  now.Unix(),
		"exp":  now.Add(m.ttl).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

// Parse validates the token signature/expiry and extracts the claims.
func (m *Manager) Parse(raw string) (Claims, error) {
	parsed, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !parsed.Valid {
		return Claims{}, ErrInvalidToken
	}

	mc, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, ErrInvalidToken
	}
	sub, _ := mc["sub"].(string)
	id, err := uuid.Parse(sub)
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	name, _ := mc["name"].(string)
	return Claims{UserID: id, Name: name}, nil
}
