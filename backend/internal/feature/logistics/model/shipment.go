package model

import (
	"errors"
	"time"
)

const (
	ProviderMockLogistics = "mocklogistics"

	EventTypeShipmentDispatched     = "shipment.dispatched"
	EventTypeShipmentPickedUp       = "shipment.picked_up"
	EventTypeShipmentInTransit      = "shipment.in_transit"
	EventTypeShipmentAtHub          = "shipment.at_transit_hub"
	EventTypeShipmentOutForDelivery = "shipment.out_for_delivery"
	EventTypeShipmentDelivered      = "shipment.delivered"
)

var ErrInvalidStatus = errors.New("invalid shipment status")

type Status string

const (
	StatusPendingPickup  Status = "pending_pickup"
	StatusPickedUp       Status = "picked_up"
	StatusInTransit      Status = "in_transit"
	StatusAtTransitHub   Status = "at_transit_hub"
	StatusOutForDelivery Status = "out_for_delivery"
	StatusDelivered      Status = "delivered"
)

var statusRank = map[Status]int{
	StatusPendingPickup:  1,
	StatusPickedUp:       2,
	StatusInTransit:      3,
	StatusAtTransitHub:   4,
	StatusOutForDelivery: 5,
	StatusDelivered:      6,
}

var eventTypes = map[string]bool{
	EventTypeShipmentDispatched:     true,
	EventTypeShipmentPickedUp:       true,
	EventTypeShipmentInTransit:      true,
	EventTypeShipmentAtHub:          true,
	EventTypeShipmentOutForDelivery: true,
	EventTypeShipmentDelivered:      true,
}

var statusEventTypes = map[Status]string{
	StatusPendingPickup:  EventTypeShipmentDispatched,
	StatusPickedUp:       EventTypeShipmentPickedUp,
	StatusInTransit:      EventTypeShipmentInTransit,
	StatusAtTransitHub:   EventTypeShipmentAtHub,
	StatusOutForDelivery: EventTypeShipmentOutForDelivery,
	StatusDelivered:      EventTypeShipmentDelivered,
}

func (s Status) Valid() bool {
	_, ok := statusRank[s]
	return ok
}

func (s Status) Rank() int {
	return statusRank[s]
}

func (s Status) IsDelivered() bool {
	return s == StatusDelivered
}

func ValidEventType(value string) bool {
	return eventTypes[value]
}

func EventTypeForStatus(status Status) string {
	return statusEventTypes[status]
}

type Shipment struct {
	ID                 int64
	ShipmentNo         string
	SubOrderID         int
	OrderID            int
	UserID             int
	ShopID             int
	Provider           string
	ProviderShipmentNo string
	ProviderTrackingNo string
	Carrier            string
	Status             Status
	ItemSummary        string
	ReceiverName       string
	ReceiverPhone      string
	ReceiverAddress    string
	OriginCity         string
	DestinationCity    string
	CurrentLocation    string
	ShippedAt          time.Time
	DeliveredAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Events             []Event
}

type Event struct {
	ID          int64
	ShipmentID  int64
	EventID     string
	Sequence    int
	Status      Status
	EventType   string
	Description string
	Location    string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

type Callback struct {
	EventID            string
	EventType          string
	ShipmentNo         string
	ProviderShipmentNo string
	ProviderTrackingNo string
	Status             Status
	Sequence           int
	Description        string
	Location           string
	OccurredAt         time.Time
}

type ProviderShipmentRequest struct {
	MerchantOrderNo string
	ItemSummary     string
	Carrier         string
	ReceiverName    string
	ReceiverPhone   string
	ReceiverAddress string
	NotifyURL       string
}

type ProviderShipmentResult struct {
	ShipmentNo      string
	MerchantNo      string
	TrackingNo      string
	Carrier         string
	Status          Status
	OriginCity      string
	DestinationCity string
	CurrentLocation string
	DispatchedAt    *time.Time
}

type ListFilter struct {
	Status   Status
	Page     int
	PageSize int
}

type Summary struct {
	ID                 int64
	ShipmentNo         string
	SubOrderID         int
	OrderID            int
	Status             Status
	Carrier            string
	ProviderTrackingNo string
	ItemSummary        string
	ReceiverName       string
	ReceiverPhone      string
	CurrentLocation    string
	ShippedAt          time.Time
	DeliveredAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ListResult struct {
	Shipments []Summary
	Page      int
	PageSize  int
	Total     int64
}
