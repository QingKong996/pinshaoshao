package apperr

import "errors"

var (
	ErrInternalServer         = errors.New("internal server error")
	ErrInventoryItemNotFound  = errors.New("inventory item not found")
	ErrInventoryVersion       = errors.New("inventory version conflict")
	ErrInventoryBelowFrozen   = errors.New("total stock must not be below frozen stock")
	ErrInvalidInventoryUpdate = errors.New("invalid inventory update")
	ErrInvalidSKUID           = errors.New("invalid sku id")
)
