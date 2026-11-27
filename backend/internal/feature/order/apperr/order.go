package apperr

import "errors"

var (
	ErrInternalServer      = errors.New("internal server error")
	ErrInvalidCheckout     = errors.New("invalid checkout request")
	ErrInvalidOrderID      = errors.New("invalid order id")
	ErrCheckoutConflict    = errors.New("checkout key was already used with a different request")
	ErrCartItemsChanged    = errors.New("selected cart items changed")
	ErrProductUnavailable  = errors.New("product is not available")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderCannotCancel   = errors.New("order cannot be cancelled")
	ErrReceiverUnavailable = errors.New("receiver information is unavailable for this order")
	ErrInvalidOrderList    = errors.New("invalid order list query")
	ErrOrderStateInvalid   = errors.New("order state is inconsistent")
	ErrOrderAmountInvalid  = errors.New("order amount is invalid")
)
