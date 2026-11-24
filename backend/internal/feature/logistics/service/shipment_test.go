package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"pinshaoshao/internal/feature/logistics/apperr"
	"pinshaoshao/internal/feature/logistics/model"
	"pinshaoshao/internal/feature/order/shippingport"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestShipmentServiceShipPersistsProviderCarrier(t *testing.T) {
	orders := &fakeOrderLifecycle{
		target: &shippingport.Target{
			SubOrderID:      12,
			OrderID:         8,
			UserID:          3,
			ShopID:          4,
			ShopName:        "Test shop",
			Status:          "awaiting_shipment",
			ItemSummary:     "Keyboard x1",
			ReceiverName:    "Receiver",
			ReceiverPhone:   "1234567890",
			ShippingAddress: "Shanghai",
		},
	}
	repository := &fakeShipmentRepository{}
	gateway := &fakeShippingGateway{}
	service := NewShipmentService(repository, orders, Config{
		Gateway:   gateway,
		NotifyURL: "http://example.com/api/v1/logistics/webhook",
	})

	created, isNew, err := service.Ship(context.Background(), 4, 12, "顺丰速运")
	if err != nil {
		t.Fatalf("ship sub-order: %v", err)
	}
	if !isNew {
		t.Fatal("isNew = false, want true")
	}
	if created.Carrier != "顺丰速运" {
		t.Fatalf("carrier = %q, want provider carrier", created.Carrier)
	}
	if created.ProviderTrackingNo != "SF123456789" {
		t.Fatalf("tracking number = %q", created.ProviderTrackingNo)
	}
	if gateway.request.NotifyURL != "http://example.com/api/v1/logistics/webhook" {
		t.Fatalf("notify URL = %q", gateway.request.NotifyURL)
	}
	if repository.createdCarrier != "顺丰速运" {
		t.Fatalf("persisted carrier = %q", repository.createdCarrier)
	}
}

func TestShipmentServiceRejectsMismatchedEventAndStatus(t *testing.T) {
	repository := &fakeShipmentRepository{}
	service := NewShipmentService(
		repository,
		&fakeOrderLifecycle{},
		Config{Gateway: &fakeShippingGateway{}},
	)

	_, err := service.HandleCallback(context.Background(), model.Callback{
		EventID:            uuid.NewString(),
		EventType:          model.EventTypeShipmentPickedUp,
		ShipmentNo:         uuid.NewString(),
		ProviderShipmentNo: uuid.NewString(),
		Status:             model.StatusDelivered,
		Sequence:           3,
		Description:        "Delivered",
		Location:           "Shanghai",
		OccurredAt:         time.Now().UTC(),
	})
	if !errors.Is(err, apperr.ErrInvalidCallback) {
		t.Fatalf("error = %v, want invalid callback", err)
	}
	if repository.applyEventCalls != 0 {
		t.Fatalf("ApplyEvent calls = %d, want 0", repository.applyEventCalls)
	}
}

type fakeShippingGateway struct {
	request model.ProviderShipmentRequest
}

func (g *fakeShippingGateway) Dispatch(
	_ context.Context,
	input model.ProviderShipmentRequest,
) (*model.ProviderShipmentResult, error) {
	g.request = input
	return &model.ProviderShipmentResult{
		ShipmentNo:      uuid.NewString(),
		MerchantNo:      input.MerchantOrderNo,
		TrackingNo:      "SF123456789",
		Carrier:         "顺丰速运",
		Status:          model.StatusPendingPickup,
		OriginCity:      "Shanghai",
		DestinationCity: "Hangzhou",
		CurrentLocation: "Shanghai merchant warehouse",
	}, nil
}

type fakeOrderLifecycle struct {
	target *shippingport.Target
}

func (o *fakeOrderLifecycle) GetShipmentTarget(
	_ context.Context,
	_ int,
	_ int,
) (*shippingport.Target, error) {
	return o.target, nil
}

func (o *fakeOrderLifecycle) LockOrder(
	_ context.Context,
	_ pgx.Tx,
	_ int,
) error {
	return nil
}

func (o *fakeOrderLifecycle) ShipSubOrder(
	_ context.Context,
	_ pgx.Tx,
	_ int,
	_ int,
	_ time.Time,
) (*shippingport.Target, bool, error) {
	return o.target, true, nil
}

func (o *fakeOrderLifecycle) CompleteSubOrder(
	_ context.Context,
	_ pgx.Tx,
	_ int,
	_ time.Time,
) error {
	return nil
}

type fakeShipmentRepository struct {
	createdCarrier  string
	applyEventCalls int
}

func (r *fakeShipmentRepository) GetBySubOrder(
	_ context.Context,
	_ int,
	_ int,
) (*model.Shipment, bool, error) {
	return nil, false, nil
}

func (r *fakeShipmentRepository) Create(
	_ context.Context,
	shipment *model.Shipment,
	_ *shippingport.Target,
) (*model.Shipment, bool, error) {
	r.createdCarrier = shipment.Carrier
	return shipment, true, nil
}

func (r *fakeShipmentRepository) GetForMerchant(
	_ context.Context,
	_ int,
	_ string,
) (*model.Shipment, error) {
	return nil, apperr.ErrShipmentNotFound
}

func (r *fakeShipmentRepository) ListForMerchant(
	_ context.Context,
	_ int,
	_ model.ListFilter,
) (*model.ListResult, error) {
	return &model.ListResult{}, nil
}

func (r *fakeShipmentRepository) GetForBuyer(
	_ context.Context,
	_ int,
	_ string,
) (*model.Shipment, error) {
	return nil, apperr.ErrShipmentNotFound
}

func (r *fakeShipmentRepository) ListForBuyerOrder(
	_ context.Context,
	_ int,
	_ int,
) ([]model.Summary, error) {
	return nil, nil
}

func (r *fakeShipmentRepository) ApplyEvent(
	_ context.Context,
	_ model.Callback,
) (*model.Shipment, error) {
	r.applyEventCalls++
	return nil, apperr.ErrShipmentConflict
}
