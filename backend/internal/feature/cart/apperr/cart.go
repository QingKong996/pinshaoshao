package apperr

import "errors"

var (
	ErrInternalServer      = errors.New("internal server error")
	ErrInvalidCartItem     = errors.New("invalid cart item")
	ErrInvalidCartItemID   = errors.New("invalid cart item id")
	ErrInvalidCartQuantity = errors.New("cart quantity must be a positive integer")
	ErrInvalidSKUID        = errors.New("invalid sku id")
	ErrCartItemNotFound    = errors.New("cart item not found")
	ErrProductNotFound     = errors.New("product not found")
	ErrProductUnavailable  = errors.New("product is not available")
)
