package apperr

import "errors"

var (
	ErrInvalidFeedQuery = errors.New("invalid feed query")
	ErrInternalServer   = errors.New("internal server error")
)
