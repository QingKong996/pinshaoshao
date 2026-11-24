package service

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"pinshaoshao/internal/mock/logistics/model"
)

type fakeShipmentRepository struct {
	shipment        *model.Shipment
	transitionCalls int
}

func (r *fakeShipmentRepository) Create(
	_ context.Context,
	shipment *model.Shipment,
) (*model.Shipment, bool, error) {
	if r.shipment != nil {
		if r.shipment.MerchantOrderNo != shipment.MerchantOrderNo {
			return nil, false, model.ErrShipmentNotFound
		}
		if r.shipment.RequestHash != shipment.RequestHash {
			return nil, false, model.ErrShipmentConflict
		}
		return r.shipment, false, nil
	}
	r.shipment = shipment
	return shipment, true, nil
}

func (r *fakeShipmentRepository) Get(context.Context, string) (*model.Shipment, error) {
	if r.shipment == nil {
		return nil, model.ErrShipmentNotFound
	}
	return r.shipment, nil
}

func (r *fakeShipmentRepository) Transition(
	_ context.Context,
	_ string,
	action model.Action,
	now time.Time,
	details model.EventDetails,
) (*model.Shipment, error) {
	r.transitionCalls++
	transition, err := r.shipment.Apply(action, now)
	if err != nil {
		return nil, err
	}
	r.shipment.CurrentLocation = details.Location
	r.shipment.Events = append(r.shipment.Events, model.Event{
		Sequence:    len(r.shipment.Events) + 1,
		Status:      transition.Status,
		EventType:   transition.EventType,
		Description: details.Description,
		Location:    details.Location,
		OccurredAt:  now,
		CreatedAt:   now,
	})
	return r.shipment, nil
}

func TestCreateShipmentGeneratesMockData(t *testing.T) {
	repository := &fakeShipmentRepository{}
	service := NewShipmentService(repository)
	input := CreateShipmentInput{
		MerchantOrderNo: "order-1",
		NotifyURL:       "http://localhost:8080/mocklogistics/webhook",
	}

	created, isNew, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("create shipment: %v", err)
	}
	if !isNew {
		t.Fatal("first create should create a shipment")
	}
	if created.Status != model.StatusPendingOutbound {
		t.Fatalf("status = %s, want pending_outbound", created.Status)
	}
	if !regexp.MustCompile(`^[A-Z]{2,4}[0-9]{12}$`).MatchString(created.TrackingNo) {
		t.Fatalf("tracking_no = %q", created.TrackingNo)
	}
	if created.Carrier == "" || created.ItemSummary == "" ||
		created.Sender.Name == "" || created.Receiver.Name == "" ||
		created.OriginCity == created.DestinationCity {
		t.Fatalf("generated shipment = %+v", created)
	}

	replayed, isNew, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("replay shipment: %v", err)
	}
	if isNew || replayed.ShipmentNo != created.ShipmentNo {
		t.Fatalf("replay = %s/%v, want %s/false", replayed.ShipmentNo, isNew, created.ShipmentNo)
	}

	_, _, err = service.Create(context.Background(), CreateShipmentInput{
		MerchantOrderNo: "order-1",
		ItemSummary:     "different",
		NotifyURL:       input.NotifyURL,
	})
	if !errors.Is(err, model.ErrShipmentConflict) {
		t.Fatalf("conflicting replay error = %v, want shipment conflict", err)
	}
}

func TestShipmentActionsRespectMerchantOutbound(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	repository := &fakeShipmentRepository{
		shipment: &model.Shipment{
			ShipmentNo:      "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
			MerchantOrderNo: "order-1",
			Status:          model.StatusPendingOutbound,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}
	service := NewShipmentService(repository)
	service.now = func() time.Time { return now }

	if _, err := service.Advance(context.Background(), repository.shipment.ShipmentNo); !errors.Is(err, model.ErrInvalidShipmentTransition) {
		t.Fatalf("advance before outbound error = %v, want invalid transition", err)
	}
	if repository.transitionCalls != 0 {
		t.Fatalf("transition calls = %d, want 0", repository.transitionCalls)
	}

	dispatched, err := service.Outbound(context.Background(), repository.shipment.ShipmentNo)
	if err != nil {
		t.Fatalf("outbound shipment: %v", err)
	}
	if dispatched.Status != model.StatusPendingPickup {
		t.Fatalf("status = %s, want pending_pickup", dispatched.Status)
	}

	if _, err := service.Outbound(context.Background(), repository.shipment.ShipmentNo); err != nil {
		t.Fatalf("repeat outbound: %v", err)
	}
	if repository.transitionCalls != 1 {
		t.Fatalf("transition calls = %d, want idempotent outbound", repository.transitionCalls)
	}
}

func TestCreateShipmentRejectsInvalidURL(t *testing.T) {
	service := NewShipmentService(&fakeShipmentRepository{})
	_, _, err := service.Create(context.Background(), CreateShipmentInput{
		MerchantOrderNo: "order-1",
		NotifyURL:       "ftp://localhost/webhook",
	})
	if !errors.Is(err, model.ErrInvalidShipmentInput) {
		t.Fatalf("error = %v, want invalid shipment input", err)
	}
}
