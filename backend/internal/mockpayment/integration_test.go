//go:build integration

package mockpayment_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"pinshaoshao/internal/mockpayment/model"
	mockpayrepository "pinshaoshao/internal/mockpayment/repository"
	mockpayservice "pinshaoshao/internal/mockpayment/service"
	"pinshaoshao/internal/webhooksig"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const testSigningKey = "integration-test-signing-key"

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("pinshaoshao_test"),
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
	return goose.UpContext(ctx, db, filepath.Join("..", "..", "migrations"))
}

func resetMockPaymentTables(t *testing.T) {
	t.Helper()
	if _, err := testPool.Exec(
		context.Background(),
		`TRUNCATE mock_payment_notifications, mock_payment_orders RESTART IDENTITY CASCADE`,
	); err != nil {
		t.Fatalf("reset mock payment tables: %v", err)
	}
}

func newTestServices() (*mockpayservice.PaymentService, *mockpayservice.NotificationService, *mockpayrepository.NotificationRepository) {
	notificationRepository := mockpayrepository.NewNotificationRepository(testPool)
	paymentRepository := mockpayrepository.NewPaymentRepository(testPool, notificationRepository, 3)
	paymentService := mockpayservice.NewPaymentService(paymentRepository)
	notificationService := mockpayservice.NewNotificationService(
		notificationRepository,
		mockpayservice.NotificationConfig{
			Signer:      webhooksig.NewSigner(testSigningKey),
			RetryDelays: []time.Duration{time.Second},
			BatchSize:   10,
		},
	)
	return paymentService, notificationService, notificationRepository
}

type capturedCallback struct {
	body      []byte
	eventID   string
	timestamp string
	signature string
	attempt   string
}

func TestMockPaymentFlow(t *testing.T) {
	resetMockPaymentTables(t)
	paymentService, notificationService, notificationRepository := newTestServices()
	callbacks := make(chan capturedCallback, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		callbacks <- capturedCallback{
			body:      body,
			eventID:   request.Header.Get("X-MockPay-Event-Id"),
			timestamp: request.Header.Get("X-MockPay-Timestamp"),
			signature: request.Header.Get("X-MockPay-Signature"),
			attempt:   request.Header.Get("X-MockPay-Delivery-Attempt"),
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	input := mockpayservice.CreatePaymentInput{
		MerchantOrderNo: "order-1",
		AmountCents:     3998,
		Subject:         "T-shirt",
		NotifyURL:       server.URL,
		ExpiresIn:       time.Hour,
	}
	payment, created, err := paymentService.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if !created {
		t.Fatal("first create should create a payment")
	}
	replayed, created, err := paymentService.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("replay payment: %v", err)
	}
	if created || replayed.PaymentNo != payment.PaymentNo {
		t.Fatalf("replay = %s/%v, want %s/false", replayed.PaymentNo, created, payment.PaymentNo)
	}
	_, _, err = paymentService.Create(context.Background(), mockpayservice.CreatePaymentInput{
		MerchantOrderNo: "order-1",
		AmountCents:     3999,
		Subject:         "T-shirt",
		NotifyURL:       server.URL,
		ExpiresIn:       time.Hour,
	})
	if !errors.Is(err, model.ErrPaymentConflict) {
		t.Fatalf("conflicting replay error = %v, want payment conflict", err)
	}

	paid, err := paymentService.Pay(context.Background(), payment.PaymentNo)
	if err != nil {
		t.Fatalf("pay payment: %v", err)
	}
	if paid.Status != model.PaymentStatusPaid {
		t.Fatalf("status = %q, want paid", paid.Status)
	}
	if _, err := paymentService.Pay(context.Background(), payment.PaymentNo); err != nil {
		t.Fatalf("repeat payment: %v", err)
	}

	if err := notificationService.DeliverDue(context.Background(), 10); err != nil {
		t.Fatalf("deliver callback: %v", err)
	}
	callback := receiveCallback(t, callbacks)
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
	if callback.attempt != "1" {
		t.Fatalf("callback attempt = %q, want 1", callback.attempt)
	}

	notifications, err := notificationRepository.ListByPaymentNo(context.Background(), payment.PaymentNo)
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(notifications) != 1 || notifications[0].Status != model.NotificationStatusSucceeded {
		t.Fatalf("notifications = %+v, want one succeeded", notifications)
	}
	if callback.eventID != notifications[0].EventID {
		t.Fatalf("callback event id = %q, want %q", callback.eventID, notifications[0].EventID)
	}
}

func TestMockPaymentClosesAndExpires(t *testing.T) {
	resetMockPaymentTables(t)
	paymentService, _, _ := newTestServices()
	ctx := context.Background()

	closable, _, err := paymentService.Create(ctx, mockpayservice.CreatePaymentInput{
		MerchantOrderNo: "close-1",
		AmountCents:     100,
		Subject:         "Close test",
		NotifyURL:       "http://localhost:8080/webhook",
		ExpiresIn:       time.Hour,
	})
	if err != nil {
		t.Fatalf("create closable payment: %v", err)
	}
	closed, err := paymentService.Close(ctx, closable.PaymentNo)
	if err != nil {
		t.Fatalf("close payment: %v", err)
	}
	if closed.Status != model.PaymentStatusClosed {
		t.Fatalf("status = %q, want closed", closed.Status)
	}
	if _, err := paymentService.Pay(ctx, closable.PaymentNo); !errors.Is(err, model.ErrPaymentClosed) {
		t.Fatalf("pay closed error = %v, want payment closed", err)
	}

	expiring, _, err := paymentService.Create(ctx, mockpayservice.CreatePaymentInput{
		MerchantOrderNo: "expire-1",
		AmountCents:     200,
		Subject:         "Expiry test",
		NotifyURL:       "http://localhost:8080/webhook",
		ExpiresIn:       time.Hour,
	})
	if err != nil {
		t.Fatalf("create expiring payment: %v", err)
	}
	if _, err := testPool.Exec(
		ctx,
		`UPDATE mock_payment_orders
SET created_at = now() - interval '2 hours', expires_at = now() - interval '1 hour'
WHERE payment_no = $1::uuid`,
		expiring.PaymentNo,
	); err != nil {
		t.Fatalf("expire payment: %v", err)
	}
	expired, err := paymentService.Get(ctx, expiring.PaymentNo)
	if err != nil {
		t.Fatalf("get expired payment: %v", err)
	}
	if expired.Status != model.PaymentStatusClosed {
		t.Fatalf("expired status = %q, want closed", expired.Status)
	}

	workerExpiring, _, err := paymentService.Create(ctx, mockpayservice.CreatePaymentInput{
		MerchantOrderNo: "expire-2",
		AmountCents:     300,
		Subject:         "Expiry worker test",
		NotifyURL:       "http://localhost:8080/webhook",
		ExpiresIn:       time.Hour,
	})
	if err != nil {
		t.Fatalf("create worker expiring payment: %v", err)
	}
	if _, err := testPool.Exec(
		ctx,
		`UPDATE mock_payment_orders
SET created_at = now() - interval '2 hours', expires_at = now() - interval '1 hour'
WHERE payment_no = $1::uuid`,
		workerExpiring.PaymentNo,
	); err != nil {
		t.Fatalf("expire worker payment: %v", err)
	}
	closedCount, err := paymentService.CloseExpired(ctx, 10)
	if err != nil {
		t.Fatalf("close expired payments: %v", err)
	}
	if closedCount != 1 {
		t.Fatalf("closed count = %d, want 1", closedCount)
	}
}

func TestMockPaymentRetriesCallback(t *testing.T) {
	resetMockPaymentTables(t)
	paymentService, notificationService, notificationRepository := newTestServices()
	ctx := context.Background()
	var failing atomic.Bool
	failing.Store(true)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if failing.Load() {
			http.Error(w, "try later", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	payment, _, err := paymentService.Create(ctx, mockpayservice.CreatePaymentInput{
		MerchantOrderNo: "retry-1",
		AmountCents:     300,
		Subject:         "Retry test",
		NotifyURL:       server.URL,
		ExpiresIn:       time.Hour,
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if _, err := paymentService.Pay(ctx, payment.PaymentNo); err != nil {
		t.Fatalf("pay payment: %v", err)
	}
	if err := notificationService.DeliverDue(ctx, 10); err != nil {
		t.Fatalf("first callback: %v", err)
	}

	notifications, err := notificationRepository.ListByPaymentNo(ctx, payment.PaymentNo)
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("notification count = %d, want 1", len(notifications))
	}
	first := notifications[0]
	if first.Status != model.NotificationStatusPending || first.AttemptCount != 1 || first.LastError == nil {
		t.Fatalf("first attempt = %+v, want pending with one failed attempt", first)
	}

	if _, err := testPool.Exec(
		ctx,
		`UPDATE mock_payment_notifications SET next_attempt_at = now() - interval '1 second' WHERE id = $1`,
		first.ID,
	); err != nil {
		t.Fatalf("make retry due: %v", err)
	}
	failing.Store(false)
	if err := notificationService.DeliverDue(ctx, 10); err != nil {
		t.Fatalf("retry callback: %v", err)
	}

	notifications, err = notificationRepository.ListByPaymentNo(ctx, payment.PaymentNo)
	if err != nil {
		t.Fatalf("list retried notifications: %v", err)
	}
	if notifications[0].Status != model.NotificationStatusSucceeded || notifications[0].AttemptCount != 2 {
		t.Fatalf("retried notification = %+v, want succeeded with two attempts", notifications[0])
	}

	if _, err := testPool.Exec(
		ctx,
		`UPDATE mock_payment_notifications
SET status = 'delivering', attempt_count = max_attempts,
    delivered_at = NULL, locked_until = now() - interval '1 second'
WHERE id = $1`,
		notifications[0].ID,
	); err != nil {
		t.Fatalf("simulate interrupted final delivery: %v", err)
	}
	if err := notificationService.DeliverDue(ctx, 10); err != nil {
		t.Fatalf("recover interrupted final delivery: %v", err)
	}
	notifications, err = notificationRepository.ListByPaymentNo(ctx, payment.PaymentNo)
	if err != nil {
		t.Fatalf("list recovered notifications: %v", err)
	}
	if notifications[0].Status != model.NotificationStatusExhausted {
		t.Fatalf("interrupted notification = %+v, want exhausted", notifications[0])
	}
}

func TestMockPaymentManualCallbackRetry(t *testing.T) {
	resetMockPaymentTables(t)
	paymentService, notificationService, notificationRepository := newTestServices()
	ctx := context.Background()
	var failing atomic.Bool
	failing.Store(true)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if failing.Load() {
			http.Error(w, "try later", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	payment, _, err := paymentService.Create(ctx, mockpayservice.CreatePaymentInput{
		MerchantOrderNo: "manual-retry-1",
		AmountCents:     300,
		Subject:         "Manual retry test",
		NotifyURL:       server.URL,
		ExpiresIn:       time.Hour,
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if _, err := paymentService.Pay(ctx, payment.PaymentNo); err != nil {
		t.Fatalf("pay payment: %v", err)
	}
	if err := notificationService.DeliverDue(ctx, 10); err != nil {
		t.Fatalf("first callback: %v", err)
	}

	notifications, err := notificationRepository.ListByPaymentNo(ctx, payment.PaymentNo)
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(notifications) != 1 || notifications[0].Status != model.NotificationStatusPending {
		t.Fatalf("notifications = %+v, want one pending", notifications)
	}
	if _, err := notificationRepository.Retry(ctx, payment.PaymentNo, time.Now()); !errors.Is(err, model.ErrNotificationNotRetryable) {
		t.Fatalf("retry pending notification error = %v, want not retryable", err)
	}

	for attempt := 2; attempt <= 3; attempt++ {
		makeNotificationDue(t, notifications[0].ID)
		if err := notificationService.DeliverDue(ctx, 10); err != nil {
			t.Fatalf("callback attempt %d: %v", attempt, err)
		}
		notifications, err = notificationRepository.ListByPaymentNo(ctx, payment.PaymentNo)
		if err != nil {
			t.Fatalf("list notifications after attempt %d: %v", attempt, err)
		}
	}
	if notifications[0].Status != model.NotificationStatusExhausted || notifications[0].AttemptCount != 3 {
		t.Fatalf("exhausted notification = %+v, want three attempts", notifications[0])
	}

	eventID := notifications[0].EventID
	retried, err := notificationRepository.Retry(ctx, payment.PaymentNo, time.Now())
	if err != nil {
		t.Fatalf("retry exhausted notification: %v", err)
	}
	if retried.Status != model.NotificationStatusPending ||
		retried.AttemptCount != 0 ||
		retried.EventID != eventID {
		t.Fatalf("retried notification = %+v, want pending attempt zero with original event id", retried)
	}

	failing.Store(false)
	if err := notificationService.DeliverDue(ctx, 10); err != nil {
		t.Fatalf("deliver retried notification: %v", err)
	}
	notifications, err = notificationRepository.ListByPaymentNo(ctx, payment.PaymentNo)
	if err != nil {
		t.Fatalf("list final notification: %v", err)
	}
	if notifications[0].Status != model.NotificationStatusSucceeded ||
		notifications[0].AttemptCount != 1 ||
		notifications[0].EventID != eventID {
		t.Fatalf("final notification = %+v, want succeeded with original event id", notifications[0])
	}
}

func makeNotificationDue(t *testing.T, id int64) {
	t.Helper()
	if _, err := testPool.Exec(
		context.Background(),
		`UPDATE mock_payment_notifications
SET next_attempt_at = now() - interval '1 second'
WHERE id = $1`,
		id,
	); err != nil {
		t.Fatalf("make notification due: %v", err)
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
