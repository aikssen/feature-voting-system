// Package apperr defines the application's typed error envelope.
//
// Services return *apperr.Error to signal a specific HTTP status + stable
// machine code. Handlers hand these to httpx.WriteError, which renders the
// frozen JSON envelope from DECISIONS.md:
//
//	{ "error": { "code": "...", "message": "...", "details": [...] } }
package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// FieldError is an optional per-field validation detail.
type FieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// Error is a typed application error carrying the HTTP status and machine code.
type Error struct {
	Status  int
	Code    string
	Message string
	Details []FieldError
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

// Unwrap exposes the wrapped cause for errors.Is / errors.As.
func (e *Error) Unwrap() error { return e.cause }

// WithCause attaches an underlying error for logging without changing the
// client-facing message.
func (e *Error) WithCause(err error) *Error {
	clone := *e
	clone.cause = err
	return &clone
}

// As reports whether err is an *Error and returns it.
func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// Constructors — one per code in the frozen contract.

func Validation(message string, details ...FieldError) *Error {
	if message == "" {
		message = "Validation failed."
	}
	return &Error{Status: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: message, Details: details}
}

func BadRequest(message string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: message}
}

func Unauthenticated(message string) *Error {
	if message == "" {
		message = "Authentication required."
	}
	return &Error{Status: http.StatusUnauthorized, Code: "UNAUTHENTICATED", Message: message}
}

func InvalidCredentials() *Error {
	return &Error{Status: http.StatusUnauthorized, Code: "INVALID_CREDENTIALS", Message: "Invalid email or password."}
}

func SelfVoteForbidden() *Error {
	return &Error{Status: http.StatusForbidden, Code: "SELF_VOTE_FORBIDDEN", Message: "You cannot vote on your own feature request."}
}

func NotFound(message string) *Error {
	if message == "" {
		message = "Resource not found."
	}
	return &Error{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: message}
}

func AlreadyVoted() *Error {
	return &Error{Status: http.StatusConflict, Code: "ALREADY_VOTED", Message: "You have already voted on this feature request."}
}

func DuplicateFeature() *Error {
	return &Error{Status: http.StatusConflict, Code: "DUPLICATE_FEATURE", Message: "A feature request with this title already exists."}
}

func Internal(cause error) *Error {
	return &Error{Status: http.StatusInternalServerError, Code: "INTERNAL", Message: "Something went wrong.", cause: cause}
}
