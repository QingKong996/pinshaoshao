package apperr

import "errors"

var (
	ErrInvalidFeedQuery = errors.New("invalid feed query")
	ErrInvalidProductID = errors.New("invalid product id")
	ErrProductNotFound  = errors.New("product not found")
	ErrInternalServer   = errors.New("internal server error")
)
