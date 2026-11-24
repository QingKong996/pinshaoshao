//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	inventoryrepository "pinshaoshao/internal/feature/inventory/repository"
	"pinshaoshao/internal/feature/order/paymentport"
	orderrepository "pinshaoshao/internal/feature/order/repository"
	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/model"
	paymentrepository "pinshaoshao/internal/feature/payment/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

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
		log.Fatalf("start postgres container: %v", err)
	}

	dsn := container.MustConnectionString(ctx, "sslmode=disable")
	if err := migrate(ctx, dsn); err != nil {
		log.Fatalf("migrate test database: %v", err)
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect test database: %v", err)
	}

	code := m.Run()

	testPool.Close()
	if err := testcontainers.TerminateContainer(container); err != nil {
		log.Printf("terminate postgres container: %v", err)
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
	return goose.UpContext(ctx, db, filepath.Join("..", "..", "..", "..", "migrations"))
}

func TestPaymentRepositoryCreateAndSettle(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	userID := insertUser(t, "buyer01")
	orderID := insertOrder(t, userID, 3998)
	repository := newPaymentRepository()

	payment, created, err := repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if !created || payment.Status != model.PaymentStatusPending ||
		payment.OrderID != orderID || payment.AmountCents != 3998 {
		t.Fatalf("payment = %+v, created = %t", payment, created)
	}

	replayed, created, err := repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if err != nil {
		t.Fatalf("replay payment: %v", err)
	}
	if created || replayed.PaymentNo != payment.PaymentNo {
		t.Fatalf("replayed payment = %+v, created = %t", replayed, created)
	}

	providerNo := uuid.NewString()
	callback := model.PaidCallback{
		EventID:               uuid.NewString(),
		EventType:             "payment.paid",
		ProviderTransactionID: providerNo,
		PaymentNo:             payment.PaymentNo,
		AmountCents:           3998,
		Currency:              model.CurrencyCNY,
		PaidAt:                time.Now().UTC(),
	}
	paid, err := repository.ApplyPaidCallback(ctx, callback)
	if err != nil {
		t.Fatalf("apply callback: %v", err)
	}
	if paid.Status != model.PaymentStatusSucceeded || paid.PaidAt == nil {
		t.Fatalf("paid = %+v", paid)
	}
	requireOrderStatus(t, orderID, "paid", "awaiting_shipment")

	again, err := repository.ApplyPaidCallback(ctx, callback)
	if err != nil {
		t.Fatalf("replay callback: %v", err)
	}
	if again.Status != model.PaymentStatusSucceeded {
		t.Fatalf("replayed payment status = %q", again.Status)
	}
	requireOrderStatus(t, orderID, "paid", "awaiting_shipment")
}

func TestPaymentRepositoryRejectsCallbackAmountMismatch(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	userID := insertUser(t, "buyer01")
	orderID := insertOrder(t, userID, 3998)
	repository := newPaymentRepository()

	payment, _, err := repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	_, err = repository.ApplyPaidCallback(ctx, model.PaidCallback{
		EventID:               uuid.NewString(),
		EventType:             "payment.paid",
		ProviderTransactionID: uuid.NewString(),
		PaymentNo:             payment.PaymentNo,
		AmountCents:           1,
		Currency:              model.CurrencyCNY,
		PaidAt:                time.Now().UTC(),
	})
	if !errors.Is(err, apperr.ErrPaymentConflict) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrPaymentConflict)
	}
	requireOrderStatus(t, orderID, "pending_payment", "pending_payment")

	stored, err := repository.GetForUser(ctx, userID, payment.PaymentNo)
	if err != nil {
		t.Fatalf("get payment: %v", err)
	}
	if stored.Status != model.PaymentStatusPending {
		t.Fatalf("payment status = %q, want pending", stored.Status)
	}
}

func TestPaymentRepositoryRecordsLatePaymentWithoutRestoringCancelledOrder(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	userID := insertUser(t, "buyer01")
	orderID := insertOrder(t, userID, 3998)
	repository := newPaymentRepository()

	payment, _, err := repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if _, err := testPool.Exec(ctx, `UPDATE orders SET status = 'cancelled' WHERE id = $1`, orderID); err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if _, err := testPool.Exec(ctx, `UPDATE sub_orders SET status = 'cancelled' WHERE order_id = $1`, orderID); err != nil {
		t.Fatalf("cancel sub-order: %v", err)
	}

	paid, err := repository.ApplyPaidCallback(ctx, model.PaidCallback{
		EventID:               uuid.NewString(),
		EventType:             "payment.paid",
		ProviderTransactionID: uuid.NewString(),
		PaymentNo:             payment.PaymentNo,
		AmountCents:           3998,
		Currency:              model.CurrencyCNY,
		PaidAt:                time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("apply callback: %v", err)
	}
	if paid.Status != model.PaymentStatusSucceeded {
		t.Fatalf("payment status = %q, want succeeded", paid.Status)
	}
	requireOrderStatus(t, orderID, "cancelled", "cancelled")
}

func TestPaymentRepositoryRejectsOtherUser(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	ownerID := insertUser(t, "buyer01")
	otherID := insertUser(t, "buyer02")
	orderID := insertOrder(t, ownerID, 3998)

	_, _, err := newPaymentRepository().CreateForOrder(ctx, otherID, orderID, uuid.NewString())
	if !errors.Is(err, apperr.ErrOrderNotFound) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrOrderNotFound)
	}
}

func newPaymentRepository() *paymentrepository.PaymentRepository {
	orderRepository := orderrepository.NewOrderRepository(
		testPool,
		inventoryrepository.NewInventoryRepository(testPool),
	)
	return paymentrepository.NewPaymentRepository(testPool, orderRepository)
}

func insertUser(t *testing.T, username string) int {
	t.Helper()

	var userID int
	if err := testPool.QueryRow(
		context.Background(),
		`INSERT INTO users (username, password_hash) VALUES ($1, 'hash') RETURNING id`,
		username,
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return userID
}

func insertOrder(t *testing.T, userID int, amountCents int64) int {
	t.Helper()
	ctx := context.Background()

	var shopID int
	if err := testPool.QueryRow(
		ctx,
		`INSERT INTO shops (username, password_hash, email)
VALUES ('shop01', 'hash', 'shop01@example.com')
ON CONFLICT (username) DO UPDATE SET updated_at = now()
RETURNING id`,
	).Scan(&shopID); err != nil {
		t.Fatalf("insert shop: %v", err)
	}

	var orderID int
	if err := testPool.QueryRow(ctx, `
INSERT INTO orders
    (user_id, checkout_key, request_hash, total_amount_cents, receiver_name, receiver_phone, shipping_address)
VALUES ($1, $2::uuid, $3, $4, 'Buyer', '13800000000', 'Shanghai')
RETURNING id`,
		userID,
		uuid.NewString(),
		strings.Repeat("a", 64),
		amountCents,
	).Scan(&orderID); err != nil {
		t.Fatalf("insert order: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO sub_orders (order_id, shop_id, shop_name, total_amount_cents)
VALUES ($1, $2, 'Shop', $3)`, orderID, shopID, amountCents); err != nil {
		t.Fatalf("insert sub-order: %v", err)
	}
	return orderID
}

func requireOrderStatus(t *testing.T, orderID int, wantOrder string, wantSubOrder string) {
	t.Helper()

	var orderStatus string
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT status::text FROM orders WHERE id = $1`,
		orderID,
	).Scan(&orderStatus); err != nil {
		t.Fatalf("read order status: %v", err)
	}
	if orderStatus != wantOrder {
		t.Fatalf("order status = %q, want %q", orderStatus, wantOrder)
	}

	var subOrderStatus string
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT status::text FROM sub_orders WHERE order_id = $1`,
		orderID,
	).Scan(&subOrderStatus); err != nil {
		t.Fatalf("read sub-order status: %v", err)
	}
	if subOrderStatus != wantSubOrder {
		t.Fatalf("sub-order status = %q, want %q", subOrderStatus, wantSubOrder)
	}
}

func resetPaymentDatabase(t *testing.T) {
	t.Helper()

	const query = `TRUNCATE users, sessions, shops, shop_sessions, orders, sub_orders,
order_item_snapshots, payments RESTART IDENTITY CASCADE`
	if _, err := testPool.Exec(context.Background(), query); err != nil {
		t.Fatalf("reset payment database: %v", err)
	}
}

var _ paymentport.Lifecycle = (*orderrepository.OrderRepository)(nil)
