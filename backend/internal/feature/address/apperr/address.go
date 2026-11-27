package apperr

import "errors"

var (
	ErrAddressNotFound  = errors.New("address not found")
	ErrInvalidAddress   = errors.New("invalid address")
	ErrInvalidAddressID = errors.New("invalid address id")
	ErrInternalServer   = errors.New("internal server error")
)
