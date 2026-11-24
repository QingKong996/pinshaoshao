package apperr

import "errors"

var (
	ErrInternalServer          = errors.New("internal server error")
	ErrInvalidSubOrderID       = errors.New("invalid sub-order id")
	ErrInvalidShipmentNo       = errors.New("invalid shipment number")
	ErrInvalidShipmentList     = errors.New("invalid shipment list query")
	ErrInvalidShipmentInput    = errors.New("invalid shipment request")
	ErrShipmentNotFound        = errors.New("shipment not found")
	ErrShipmentConflict        = errors.New("shipment conflicts with the existing record")
	ErrOrderNotShippable       = errors.New("sub-order is not shippable")
	ErrInvalidCallback         = errors.New("invalid logistics callback")
	ErrInvalidWebhookSignature = errors.New("invalid logistics webhook signature")
	ErrProviderUnavailable     = errors.New("logistics provider is unavailable")
	ErrProviderInvalid         = errors.New("logistics provider returned an invalid response")
)
