package token

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateAndParse_RoundTrip(t *testing.T) {
	m := NewManager("super-secret", 24)
	id := uuid.New()

	tok, err := m.Generate(id, "Ever")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != id {
		t.Errorf("user id = %v want %v", claims.UserID, id)
	}
	if claims.Name != "Ever" {
		t.Errorf("name = %q want Ever", claims.Name)
	}
}

func TestParse_RejectsWrongSecret(t *testing.T) {
	tok, _ := NewManager("secret-a", 24).Generate(uuid.New(), "Ever")
	if _, err := NewManager("secret-b", 24).Parse(tok); err == nil {
		t.Error("expected error parsing token signed with a different secret")
	}
}

func TestParse_RejectsGarbage(t *testing.T) {
	if _, err := NewManager("secret", 24).Parse("not.a.jwt"); err == nil {
		t.Error("expected error for malformed token")
	}
}
