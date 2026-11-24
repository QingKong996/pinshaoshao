package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	DefaultNotificationMaxAttempts  = 7
	EventTypeShipmentCreated        = "shipment.created"
	EventTypeShipmentDispatched     = "shipment.dispatched"
	EventTypeShipmentPickedUp       = "shipment.picked_up"
	EventTypeShipmentInTransit      = "shipment.in_transit"
	EventTypeShipmentAtHub          = "shipment.at_transit_hub"
	EventTypeShipmentOutForDelivery = "shipment.out_for_delivery"
	EventTypeShipmentDelivered      = "shipment.delivered"
)

var (
	ErrInvalidShipmentInput      = errors.New("invalid shipment request")
	ErrInvalidShipmentNumber     = errors.New("invalid shipment number")
	ErrInvalidNotificationID     = errors.New("invalid notification id")
	ErrShipmentNotFound          = errors.New("shipment not found")
	ErrShipmentConflict          = errors.New("merchant order number was already used with different data")
	ErrInvalidShipmentTransition = errors.New("invalid shipment state transition")
	ErrNotificationNotFound      = errors.New("shipment notification not found")
	ErrNotificationNotRetryable  = errors.New("shipment notification cannot be retried")
)

type Status string

const (
	StatusPendingOutbound Status = "pending_outbound"
	StatusPendingPickup   Status = "pending_pickup"
	StatusPickedUp        Status = "picked_up"
	StatusInTransit       Status = "in_transit"
	StatusAtTransitHub    Status = "at_transit_hub"
	StatusOutForDelivery  Status = "out_for_delivery"
	StatusDelivered       Status = "delivered"
)

type Action string

const (
	ActionOutbound Action = "outbound"
	ActionAdvance  Action = "advance"
	ActionComplete Action = "complete"
)

type Transition struct {
	Status    Status
	EventType string
}

var shipmentTransitions = map[Action]map[Status]Transition{
	ActionOutbound: {
		StatusPendingOutbound: {
			Status:    StatusPendingPickup,
			EventType: EventTypeShipmentDispatched,
		},
	},
	ActionAdvance: {
		StatusPendingPickup: {
			Status:    StatusPickedUp,
			EventType: EventTypeShipmentPickedUp,
		},
		StatusPickedUp: {
			Status:    StatusInTransit,
			EventType: EventTypeShipmentInTransit,
		},
		StatusInTransit: {
			Status:    StatusAtTransitHub,
			EventType: EventTypeShipmentAtHub,
		},
		StatusAtTransitHub: {
			Status:    StatusOutForDelivery,
			EventType: EventTypeShipmentOutForDelivery,
		},
	},
	ActionComplete: {
		StatusOutForDelivery: {
			Status:    StatusDelivered,
			EventType: EventTypeShipmentDelivered,
		},
	},
}

var dispatchedStatuses = map[Status]bool{
	StatusPendingPickup:  true,
	StatusPickedUp:       true,
	StatusInTransit:      true,
	StatusAtTransitHub:   true,
	StatusOutForDelivery: true,
	StatusDelivered:      true,
}

func (s Status) IsTerminal() bool {
	return s == StatusDelivered
}

type Party struct {
	Name    string
	Phone   string
	Address string
}

type Shipment struct {
	ID              int64
	ShipmentNo      string
	MerchantOrderNo string
	RequestHash     string
	TrackingNo      string
	Carrier         string
	Status          Status
	ItemSummary     string
	Sender          Party
	Receiver        Party
	OriginCity      string
	DestinationCity string
	NotifyURL       string
	CurrentLocation string
	DispatchedAt    *time.Time
	PickedUpAt      *time.Time
	DeliveredAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Events          []Event
}

type Event struct {
	ID          int64
	ShipmentID  int64
	Sequence    int
	Status      Status
	EventType   string
	Description string
	Location    string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

type EventDetails struct {
	Description string
	Location    string
}

func (s *Shipment) Next(action Action) (Transition, error) {
	transition, ok := shipmentTransitions[action][s.Status]
	if !ok {
		return Transition{}, fmt.Errorf(
			"%w: cannot %s a %s shipment",
			ErrInvalidShipmentTransition,
			action,
			s.Status,
		)
	}
	return transition, nil
}

func (s *Shipment) Apply(action Action, now time.Time) (Transition, error) {
	transition, err := s.Next(action)
	if err != nil {
		return Transition{}, err
	}

	now = now.UTC()
	s.Status = transition.Status
	s.UpdatedAt = now
	switch s.Status {
	case StatusPendingPickup:
		dispatchedAt := now
		s.DispatchedAt = &dispatchedAt
	case StatusPickedUp:
		pickedUpAt := now
		s.PickedUpAt = &pickedUpAt
	case StatusDelivered:
		deliveredAt := now
		s.DeliveredAt = &deliveredAt
	}
	return transition, nil
}

func (s *Shipment) IsDispatched() bool {
	return dispatchedStatuses[s.Status]
}

type NotificationStatus string

const (
	NotificationStatusPending    NotificationStatus = "pending"
	NotificationStatusDelivering NotificationStatus = "delivering"
	NotificationStatusSucceeded  NotificationStatus = "succeeded"
	NotificationStatusExhausted  NotificationStatus = "exhausted"
)

type Notification struct {
	ID            int64
	EventID       string
	ShipmentID    int64
	EventType     string
	TargetURL     string
	Payload       []byte
	Status        NotificationStatus
	AttemptCount  int
	MaxAttempts   int
	NextAttemptAt time.Time
	LockedUntil   *time.Time
	LastAttemptAt *time.Time
	DeliveredAt   *time.Time
	LastError     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ShipmentEventPayload struct {
	ID         string                   `json:"id"`
	Type       string                   `json:"type"`
	OccurredAt time.Time                `json:"occurred_at"`
	Data       ShipmentEventPayloadData `json:"data"`
}

type ShipmentEventPayloadData struct {
	ShipmentNo      string    `json:"shipment_no"`
	TrackingNo      string    `json:"tracking_no"`
	MerchantOrderNo string    `json:"merchant_order_no"`
	Status          Status    `json:"status"`
	Carrier         string    `json:"carrier"`
	Sequence        int       `json:"sequence"`
	Description     string    `json:"description"`
	Location        string    `json:"location"`
	OccurredAt      time.Time `json:"occurred_at"`
}

func NewEventPayload(
	eventID string,
	shipment *Shipment,
	event *Event,
) ([]byte, error) {
	occurredAt := event.OccurredAt.UTC()
	return json.Marshal(ShipmentEventPayload{
		ID:         eventID,
		Type:       event.EventType,
		OccurredAt: occurredAt,
		Data: ShipmentEventPayloadData{
			ShipmentNo:      shipment.ShipmentNo,
			TrackingNo:      shipment.TrackingNo,
			MerchantOrderNo: shipment.MerchantOrderNo,
			Status:          shipment.Status,
			Carrier:         shipment.Carrier,
			Sequence:        event.Sequence,
			Description:     event.Description,
			Location:        event.Location,
			OccurredAt:      occurredAt,
		},
	})
}

func NewNotification(
	eventID string,
	shipment *Shipment,
	event *Event,
	payload []byte,
	maxAttempts int,
	now time.Time,
) *Notification {
	now = now.UTC()
	return &Notification{
		EventID:       eventID,
		ShipmentID:    shipment.ID,
		EventType:     event.EventType,
		TargetURL:     shipment.NotifyURL,
		Payload:       payload,
		Status:        NotificationStatusPending,
		MaxAttempts:   maxAttempts,
		NextAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
