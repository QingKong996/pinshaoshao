//go:build integration

package logistics_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pinshaoshao/internal/mock/logistics/model"
	logisticsrepository "pinshaoshao/internal/mock/logistics/repository"
	logisticsservice "pinshaoshao/internal/mock/logistics/service"
	"pinshaoshao/internal/webhooksig"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const testSigningKey = "logistics-integration-test-signing-key"

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("pinshaoshao_logistics_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		slog.Error("start postgres container", "error", err)
		os.Exit(1)
	}

	dsn := container.MustConnectionString(ctx, "sslmode=disable")
	if err := migrate(ctx, dsn); err != nil {
		slog.Error("migrate test database", "error", err)
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("connect test database", "error", err)
		os.Exit(1)
	}

	code := m.Run()

	testPool.Close()
	if err := testcontainers.TerminateContainer(container); err != nil {
		slog.Error("terminate postgres container", "error", err)
	}
	os.Exit(code)
}

func migrate(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetLogger(goose.NopLogger())
	return goose.UpContext(
		ctx,
		db,
		filepath.Join("..", "..", "..", "migrations", "mocklogistics"),
	)
}

func resetTables(t *testing.T) {
	t.Helper()
	if _, err := testPool.Exec(
		context.Background(),
		`TRUNCATE mock_logistics_notifications, mock_logistics_events, mock_shipments RESTART IDENTITY CASCADE`,
	); err != nil {
		t.Fatalf("reset mock logistics tables: %v", err)
	}
}

func newTestServices() (
	*logisticsservice.ShipmentService,
	*logisticsservice.NotificationService,
	*logisticsrepository.NotificationRepository,
) {
	notificationRepository := logisticsrepository.NewNotificationRepository(testPool)
	shipmentRepository := logisticsrepository.NewShipmentRepository(
		testPool,
		notificationRepository,
		3,
	)
	shipmentService := logisticsservice.NewShipmentService(shipmentRepository)
	notificationService := logisticsservice.NewNotificationService(
		notificationRepository,
		logisticsservice.NotificationConfig{
			Signer:      webhooksig.NewSigner(testSigningKey),
			RetryDelays: []time.Duration{time.Second},
			BatchSize:   10,
		},
	)
	return shipmentService, notificationService, notificationRepository
}

type capturedCallback struct {
	body      []byte
	eventID   string
	eventType string
	timestamp string
	signature string
	attempt   string
}

func TestMockLogisticsFlow(t *testing.T) {
	resetTables(t)
	shipmentService, notificationService, notificationRepository := newTestServices()
	ctx := context.Background()
	callbacks := make(chan capturedCallback, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		callbacks <- capturedCallback{
			body:      body,
			eventID:   request.Header.Get("X-MockLogistics-Event-Id"),
			eventType: request.Header.Get("X-MockLogistics-Event-Type"),
			timestamp: request.Header.Get("X-MockLogistics-Timestamp"),
			signature: request.Header.Get("X-MockLogistics-Signature"),
			attempt:   request.Header.Get("X-MockLogistics-Delivery-Attempt"),
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	input := logisticsservice.CreateShipmentInput{
		MerchantOrderNo: "logistics-order-1",
		NotifyURL:       server.URL,
	}
	shipment, created, err := shipmentService.Create(ctx, input)
	if err != nil {
		t.Fatalf("create shipment: %v", err)
	}
	if !created || len(shipment.Events) != 1 {
		t.Fatalf("created = %v, events = %d", created, len(shipment.Events))
	}
	if shipment.Status != model.StatusPendingOutbound {
		t.Fatalf("status = %s, want pending_outbound", shipment.Status)
	}

	replayed, created, err := shipmentService.Create(ctx, input)
	if err != nil {
		t.Fatalf("replay shipment: %v", err)
	}
	if created || replayed.ShipmentNo != shipment.ShipmentNo {
		t.Fatalf("replay = %s/%v, want %s/false", replayed.ShipmentNo, created, shipment.ShipmentNo)
	}
	_, _, err = shipmentService.Create(ctx, logisticsservice.CreateShipmentInput{
		MerchantOrderNo: input.MerchantOrderNo,
		ItemSummary:     "different",
		NotifyURL:       input.NotifyURL,
	})
	if !errors.Is(err, model.ErrShipmentConflict) {
		t.Fatalf("conflicting replay error = %v, want shipment conflict", err)
	}

	if _, err := shipmentService.Advance(ctx, shipment.ShipmentNo); !errors.Is(err, model.ErrInvalidShipmentTransition) {
		t.Fatalf("advance before outbound error = %v, want invalid transition", err)
	}

	shipment, err = shipmentService.Outbound(ctx, shipment.ShipmentNo)
	if err != nil {
		t.Fatalf("outbound shipment: %v", err)
	}
	assertAndDeliverCallback(t, notificationService, callbacks, model.StatusPendingPickup, "shipment.dispatched")

	for _, want := range []model.Status{
		model.StatusPickedUp,
		model.StatusInTransit,
		model.StatusAtTransitHub,
		model.StatusOutForDelivery,
	} {
		shipment, err = shipmentService.Advance(ctx, shipment.ShipmentNo)
		if err != nil {
			t.Fatalf("advance to %s: %v", want, err)
		}
		if shipment.Status != want {
			t.Fatalf("status = %s, want %s", shipment.Status, want)
		}
		assertAndDeliverCallback(t, notificationService, callbacks, want, eventTypeForStatus(want))
	}

	if _, err := shipmentService.Complete(ctx, shipment.ShipmentNo); err != nil {
		t.Fatalf("complete shipment: %v", err)
	}
	callback := assertAndDeliverCallback(
		t,
		notificationService,
		callbacks,
		model.StatusDelivered,
		model.EventTypeShipmentDelivered,
	)
	var payload model.ShipmentEventPayload
	if err := json.Unmarshal(callback.body, &payload); err != nil {
		t.Fatalf("decode delivered callback: %v", err)
	}
	if payload.Data.Status != model.StatusDelivered || payload.Data.TrackingNo == "" {
		t.Fatalf("delivered callback = %+v", payload)
	}

	shipment, err = shipmentService.Complete(ctx, shipment.ShipmentNo)
	if err != nil {
		t.Fatalf("repeat complete: %v", err)
	}
	if len(shipment.Events) != 7 {
		t.Fatalf("event count = %d, want 7", len(shipment.Events))
	}
	notifications, err := notificationRepository.ListByShipmentNo(ctx, shipment.ShipmentNo)
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(notifications) != 6 {
		t.Fatalf("notification count = %d, want 6", len(notifications))
	}
	for _, notification := range notifications {
		if notification.Status != model.NotificationStatusSucceeded {
			t.Fatalf("notification = %+v, want succeeded", notification)
		}
	}
}

func TestMockLogisticsAutoDelivery(t *testing.T) {
	resetTables(t)
	shipmentService, _, _ := newTestServices()
	ctx := context.Background()

	shipment, _, err := shipmentService.Create(ctx, logisticsservice.CreateShipmentInput{
		MerchantOrderNo: "logistics-auto-order",
		NotifyURL:       "http://localhost:8080/mocklogistics/webhook",
	})
	if err != nil {
		t.Fatalf("create shipment: %v", err)
	}
	if _, err := shipmentService.Outbound(ctx, shipment.ShipmentNo); err != nil {
		t.Fatalf("outbound shipment: %v", err)
	}

	for range 5 {
		advanced, err := shipmentService.AdvanceDue(
			ctx,
			time.Nanosecond,
			10,
		)
		if err != nil {
			t.Fatalf("advance due shipment: %v", err)
		}
		if advanced != 1 {
			t.Fatalf("advanced = %d, want 1", advanced)
		}
	}

	shipment, err = shipmentService.Get(ctx, shipment.ShipmentNo)
	if err != nil {
		t.Fatalf("get delivered shipment: %v", err)
	}
	if shipment.Status != model.StatusDelivered || shipment.DeliveredAt == nil {
		t.Fatalf("delivered shipment = %+v", shipment)
	}
	if len(shipment.Events) != 7 {
		t.Fatalf("event count = %d, want 7", len(shipment.Events))
	}
}

func assertAndDeliverCallback(
	t *testing.T,
	notificationService *logisticsservice.NotificationService,
	callbacks <-chan capturedCallback,
	status model.Status,
	eventType string,
) capturedCallback {
	t.Helper()
	if err := notificationService.DeliverDue(context.Background(), 10); err != nil {
		t.Fatalf("deliver callback for %s: %v", status, err)
	}
	callback := receiveCallback(t, callbacks)
	if callback.eventType != eventType {
		t.Fatalf("callback event type = %q, want %q", callback.eventType, eventType)
	}
	if callback.attempt != "1" {
		t.Fatalf("callback attempt = %q, want 1", callback.attempt)
	}
	if err := webhooksig.Verify(
		testSigningKey,
		callback.timestamp,
		callback.signature,
		callback.body,
		time.Now(),
		time.Minute,
	); err != nil {
		t.Fatalf("verify callback: %v", err)
	}
	return callback
}

func eventTypeForStatus(status model.Status) string {
	switch status {
	case model.StatusPickedUp:
		return model.EventTypeShipmentPickedUp
	case model.StatusInTransit:
		return model.EventTypeShipmentInTransit
	case model.StatusAtTransitHub:
		return model.EventTypeShipmentAtHub
	case model.StatusOutForDelivery:
		return model.EventTypeShipmentOutForDelivery
	default:
		return ""
	}
}

func receiveCallback(t *testing.T, callbacks <-chan capturedCallback) capturedCallback {
	t.Helper()
	select {
	case callback := <-callbacks:
		return callback
	case <-time.After(5 * time.Second):
		t.Fatal("callback was not received")
		return capturedCallback{}
	}
}
