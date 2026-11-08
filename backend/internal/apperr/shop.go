package apperr

import "errors"

var (
	ErrShopNotFound       = errors.New(`shop not found`)
	ErrShopBanned         = errors.New(`shop is banned`)
	ErrInvalidShopSession = errors.New(`invalid or expired shop session`)
	ErrShopUsernameTaken  = errors.New(`shop username already exists`)
	ErrShopEmailTaken     = errors.New(`shop email already exists`)
)
