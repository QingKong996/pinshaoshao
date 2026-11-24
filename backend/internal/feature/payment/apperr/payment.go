package apperr

import "errors"

var (
	ErrInternalServer             = errors.New("internal server error")
	ErrInvalidOrderID             = errors.New("invalid order id")
	ErrInvalidPaymentNo           = errors.New("invalid payment number")
	ErrInvalidCallback            = errors.New("invalid payment callback")
	ErrInvalidWebhookSignature    = errors.New("invalid webhook signature")
	ErrStaleWebhookSignature      = errors.New("webhook signature timestamp is outside the allowed window")
	ErrPaymentNotFound            = errors.New("payment not found")
	ErrOrderNotFound              = errors.New("order not found")
	ErrOrderNotPayable            = errors.New("order is not payable")
	ErrPaymentClosed              = errors.New("payment is closed")
	ErrPaymentConflict            = errors.New("payment conflicts with the existing transaction")
	ErrPaymentProviderUnavailable = errors.New("payment provider is unavailable")
	ErrPaymentProviderInvalid     = errors.New("payment provider returned an invalid response")
)
