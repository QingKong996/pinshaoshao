package apperr

import "errors"

var (
	ErrInvalidJsonFormat = errors.New("invalid json format")
	ErrInternalServer    = errors.New("internal server error")
)
