package auth

import (
	"time"

	"github.com/google/uuid"
)

// User is the domain model. password_hash never leaves this package in a view.
type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// SignupRequest is the POST /auth/signup body.
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the POST /auth/login body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserView is the safe, public projection of a user.
type UserView struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse is returned by both signup and login.
type AuthResponse struct {
	Token string   `json:"token"`
	User  UserView `json:"user"`
}

func (u User) View() UserView {
	return UserView{ID: u.ID, Name: u.Name, Email: u.Email, CreatedAt: u.CreatedAt}
}
