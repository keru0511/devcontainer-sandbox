package domain

import "errors"

// Sentinel errors used across the application.
var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned on duplicate creation attempts.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when an operation is not permitted.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrConflict is returned when a state transition is not allowed.
	ErrConflict = errors.New("conflict")
)

// IsNotFound reports whether err wraps ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists reports whether err wraps ErrAlreadyExists.
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}
