package model_test

import (
	"errors"
	"testing"
	"time"

	"pinshaoshao/internal/mock/logistics/model"
)

func TestShipmentStateMachine(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	shipment := &model.Shipment{Status: model.StatusPendingOutbound}

	if _, err := shipment.Apply(model.ActionAdvance, now); !errors.Is(err, model.ErrInvalidShipmentTransition) {
		t.Fatalf("advance before outbound error = %v, want invalid transition", err)
	}

	transition, err := shipment.Apply(model.ActionOutbound, now)
	if err != nil {
		t.Fatalf("outbound shipment: %v", err)
	}
	if transition.Status != model.StatusPendingPickup || transition.EventType != model.EventTypeShipmentDispatched {
		t.Fatalf("outbound transition = %+v", transition)
	}
	if shipment.DispatchedAt == nil {
		t.Fatal("dispatched_at was not set")
	}

	for _, want := range []model.Status{
		model.StatusPickedUp,
		model.StatusInTransit,
		model.StatusAtTransitHub,
		model.StatusOutForDelivery,
	} {
		if _, err := shipment.Apply(model.ActionAdvance, now); err != nil {
			t.Fatalf("advance to %s: %v", want, err)
		}
		if shipment.Status != want {
			t.Fatalf("status = %s, want %s", shipment.Status, want)
		}
	}
	if shipment.PickedUpAt == nil {
		t.Fatal("picked_up_at was not set")
	}
	if _, err := shipment.Apply(model.ActionAdvance, now); !errors.Is(err, model.ErrInvalidShipmentTransition) {
		t.Fatalf("advance from out for delivery error = %v, want invalid transition", err)
	}

	if _, err := shipment.Apply(model.ActionComplete, now); err != nil {
		t.Fatalf("complete shipment: %v", err)
	}
	if shipment.Status != model.StatusDelivered || shipment.DeliveredAt == nil {
		t.Fatalf("delivered shipment = %+v", shipment)
	}
	if !shipment.Status.IsTerminal() {
		t.Fatal("delivered shipment should be terminal")
	}
	if _, err := shipment.Apply(model.ActionComplete, now); !errors.Is(err, model.ErrInvalidShipmentTransition) {
		t.Fatalf("repeat complete error = %v, want invalid transition", err)
	}
}
