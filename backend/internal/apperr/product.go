package apperr

import "errors"

var (
	ErrProductNotFound         = errors.New(`product not found`)
	ErrProductAccessDenied     = errors.New(`product access denied`)
	ErrInvalidProductID        = errors.New(`invalid product id`)
	ErrProductSKUAlreadyExists = errors.New(`product sku with the same specs already exists`)
	ErrProductSKUSpecsMismatch = errors.New(`product sku specs must have the same keys as the product's existing skus`)
	ErrCategoryNotFound        = errors.New(`category not found`)
	ErrCategoryNameTaken       = errors.New(`category name already exists`)
	ErrInvalidCategoryID       = errors.New(`invalid category`)
)
