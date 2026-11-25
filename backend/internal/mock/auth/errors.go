package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("authentication required")
	ErrInvalidSession     = errors.New("invalid or expired session")
	ErrInvalidJSON        = errors.New("invalid json format")
	ErrInternalServer     = errors.New("internal server error")

	errAccountNotFound = errors.New("mock auth account not found")
)
