package apperr

import "errors"

var (
	ErrProductNotFound         = errors.New(`product not found`)
	ErrInvalidProductID        = errors.New(`invalid product id`)
	ErrProductSKUAlreadyExists = errors.New(`product sku with the same specs already exists`)
	ErrProductSKUSpecsMismatch = errors.New(`product sku specs must have the same keys as the product's existing skus`)
	ErrProductSKUNotFound      = errors.New(`product sku not found`)
	ErrInvalidProductSKUID     = errors.New(`invalid product sku id`)
	ErrInvalidProductSKUSpecs  = errors.New(`invalid product sku specs`)
	ErrInvalidProductSKUUpdate = errors.New(`product sku update must contain at least one field`)
	ErrCategoryNotFound        = errors.New(`category not found`)
	ErrCategoryNameTaken       = errors.New(`category name already exists`)
	ErrInvalidCategoryID       = errors.New(`invalid category`)
)
