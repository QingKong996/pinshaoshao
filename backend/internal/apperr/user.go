package apperr

import "errors"

var (
	ErrUserNotFound       = errors.New(`user not found`)
	ErrInvalidCredentials = errors.New(`invalid credentials`)
	ErrUsernameTaken      = errors.New(`username already exists`)
	ErrEmailTaken         = errors.New(`email already exists`)
)
