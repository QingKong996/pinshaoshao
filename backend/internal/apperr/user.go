package apperr

import "errors"

var (
	ErrUserNotFound       = errors.New(`user not found`)
	ErrInvalidCredentials = errors.New(`invalid credentials`)
	ErrUserBanned         = errors.New(`user is banned`)
	ErrUnauthorized       = errors.New(`authentication required`)
	ErrInvalidSession     = errors.New(`invalid or expired session`)
	ErrUsernameTaken      = errors.New(`username already exists`)
	ErrEmailTaken         = errors.New(`email already exists`)
)
