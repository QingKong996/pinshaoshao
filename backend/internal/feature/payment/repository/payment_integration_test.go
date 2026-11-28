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

	"pinshaoshao/internal/feature/order/paymentport"
	orderrepository "pinshaoshao/internal/feature/order/repository"
	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/model"
	paymentrepository "pinshaoshao/internal/feature/payment/repository"
	"pinshaoshao/internal/outbox"

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

func TestPaymentRepositoryCreateForOrderRejectsUnpayableOrder(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	userID := insertUser(t, "buyer01")
	orderID := insertOrder(t, userID, 3998)
	repository := newPaymentRepository()

	if _, err := testPool.Exec(
		ctx,
		`UPDATE orders SET created_at = now() - interval '31 minutes' WHERE id = $1`,
		orderID,
	); err != nil {
		t.Fatalf("expire order: %v", err)
	}
	_, _, err := repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if !errors.Is(err, apperr.ErrOrderNotPayable) {
		t.Fatalf("expired order error = %v, want %v", err, apperr.ErrOrderNotPayable)
	}

	if _, err := testPool.Exec(
		ctx,
		`UPDATE orders SET status = 'paid', paid_at = now() WHERE id = $1`,
		orderID,
	); err != nil {
		t.Fatalf("mark order paid: %v", err)
	}
	_, _, err = repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if !errors.Is(err, apperr.ErrOrderNotPayable) {
		t.Fatalf("paid order error = %v, want %v", err, apperr.ErrOrderNotPayable)
	}
}

func TestPaymentRepositoryCreateForOrderRejectsZeroAmountOrder(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	userID := insertUser(t, "buyer01")
	orderID := insertOrder(t, userID, 0)
	if _, err := testPool.Exec(
		ctx,
		`UPDATE orders SET status = 'paid', paid_at = now() WHERE id = $1`,
		orderID,
	); err != nil {
		t.Fatalf("mark zero-amount order paid: %v", err)
	}

	_, _, err := newPaymentRepository().CreateForOrder(
		ctx,
		userID,
		orderID,
		uuid.NewString(),
	)
	if !errors.Is(err, apperr.ErrOrderNotPayable) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrOrderNotPayable)
	}
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

func TestPaymentRepositoryRefundsPaymentForCancelledOrder(t *testing.T) {
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

	callback := model.PaidCallback{
		EventID:               uuid.NewString(),
		EventType:             "payment.paid",
		ProviderTransactionID: uuid.NewString(),
		PaymentNo:             payment.PaymentNo,
		AmountCents:           3998,
		Currency:              model.CurrencyCNY,
		PaidAt:                time.Now().UTC(),
	}
	paid, err := repository.ApplyPaidCallback(ctx, callback)
	if err != nil {
		t.Fatalf("apply callback: %v", err)
	}
	if paid.Status != model.PaymentStatusSucceeded {
		t.Fatalf("payment status = %q, want succeeded", paid.Status)
	}
	if paid.Refund == nil || paid.Refund.Status != model.RefundStatusPending {
		t.Fatalf("refund = %+v, want pending", paid.Refund)
	}
	requireOrderStatus(t, orderID, "cancelled", "cancelled")
	requireRefundStatus(t, payment.PaymentNo, model.RefundStatusPending)
	requireRefundOutboxEvent(t, payment.PaymentNo)

	replayed, err := repository.ApplyPaidCallback(ctx, callback)
	if err != nil {
		t.Fatalf("replay callback: %v", err)
	}
	if replayed.Refund == nil || replayed.Refund.RefundNo != paid.Refund.RefundNo {
		t.Fatalf("replayed refund = %+v, want %q", replayed.Refund, paid.Refund.RefundNo)
	}
	requireRefundOutboxEvent(t, payment.PaymentNo)

	stored, err := repository.GetForUser(ctx, userID, payment.PaymentNo)
	if err != nil {
		t.Fatalf("get payment: %v", err)
	}
	if stored.Refund == nil || stored.Refund.RefundNo != paid.Refund.RefundNo {
		t.Fatalf("stored refund = %+v, want %q", stored.Refund, paid.Refund.RefundNo)
	}

	byOrder, err := repository.GetForOrder(ctx, userID, orderID)
	if err != nil {
		t.Fatalf("get payment by order: %v", err)
	}
	if byOrder.PaymentNo != payment.PaymentNo ||
		byOrder.Refund == nil ||
		byOrder.Refund.RefundNo != paid.Refund.RefundNo {
		t.Fatalf("payment by order = %+v, want payment %q with refund %q", byOrder, payment.PaymentNo, paid.Refund.RefundNo)
	}

	otherID := insertUser(t, "buyer02")
	_, err = repository.GetForOrder(ctx, otherID, orderID)
	if !errors.Is(err, apperr.ErrPaymentNotFound) {
		t.Fatalf("other user error = %v, want %v", err, apperr.ErrPaymentNotFound)
	}
}

func TestPaymentRepositoryRejectsExpiredPendingOrderSettlement(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	userID := insertUser(t, "buyer01")
	orderID := insertOrder(t, userID, 3998)
	skuID := insertOrderItem(t, orderID)
	repository := newPaymentRepository()

	payment, _, err := repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if _, err := testPool.Exec(
		ctx,
		`UPDATE orders SET created_at = now() - interval '31 minutes' WHERE id = $1`,
		orderID,
	); err != nil {
		t.Fatalf("expire order: %v", err)
	}

	_, err = repository.ApplyPaidCallback(ctx, model.PaidCallback{
		EventID:               uuid.NewString(),
		EventType:             "payment.paid",
		ProviderTransactionID: uuid.NewString(),
		PaymentNo:             payment.PaymentNo,
		AmountCents:           3998,
		Currency:              model.CurrencyCNY,
		PaidAt:                time.Now().UTC(),
	})
	var expired *paymentport.OrderExpiredError
	if !errors.As(err, &expired) {
		t.Fatalf("error = %v, want %T", err, expired)
	}
	if expired.OrderID != orderID {
		t.Fatalf("expired order = %d, want %d", expired.OrderID, orderID)
	}
	requireOrderStatus(t, orderID, "pending_payment", "pending_payment")
	requirePaymentFrozenStock(t, skuID, 2)

	stored, err := repository.GetForUser(ctx, userID, payment.PaymentNo)
	if err != nil {
		t.Fatalf("get payment: %v", err)
	}
	if stored.Status != model.PaymentStatusPending {
		t.Fatalf("payment status = %q, want pending", stored.Status)
	}
}

func TestPaymentRepositoryRecordsLatePaymentAfterLocalClose(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	userID := insertUser(t, "buyer01")
	orderID := insertOrder(t, userID, 3998)
	repository := newPaymentRepository()

	payment, _, err := repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	closed, err := repository.CloseForOrder(ctx, testPool, orderID)
	if err != nil {
		t.Fatalf("close payment: %v", err)
	}
	if closed == nil || !closed.Closed {
		t.Fatalf("close result = %+v, want closed", closed)
	}
	if _, err := testPool.Exec(
		ctx,
		`UPDATE orders SET status = 'cancelled' WHERE id = $1`,
		orderID,
	); err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if _, err := testPool.Exec(
		ctx,
		`UPDATE sub_orders SET status = 'cancelled' WHERE order_id = $1`,
		orderID,
	); err != nil {
		t.Fatalf("cancel sub-order: %v", err)
	}

	providerNo := uuid.NewString()
	paid, err := repository.ApplyPaidCallback(ctx, model.PaidCallback{
		EventID:               uuid.NewString(),
		EventType:             "payment.paid",
		ProviderTransactionID: providerNo,
		PaymentNo:             payment.PaymentNo,
		AmountCents:           3998,
		Currency:              model.CurrencyCNY,
		PaidAt:                time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("apply late callback: %v", err)
	}
	if paid.Status != model.PaymentStatusSucceeded ||
		paid.ProviderTransactionID == nil ||
		*paid.ProviderTransactionID != providerNo {
		t.Fatalf("paid = %+v, want succeeded with provider transaction", paid)
	}
	if paid.Refund == nil || paid.Refund.Status != model.RefundStatusPending {
		t.Fatalf("refund = %+v, want pending", paid.Refund)
	}
	requireOrderStatus(t, orderID, "cancelled", "cancelled")
	requireRefundStatus(t, payment.PaymentNo, model.RefundStatusPending)
}

func TestPaymentRepositoryCloseForOrderClosesPendingPayment(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	userID := insertUser(t, "buyer01")
	orderID := insertOrder(t, userID, 3998)
	repository := newPaymentRepository()

	payment, _, err := repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	closed, err := repository.CloseForOrder(ctx, testPool, orderID)
	if err != nil {
		t.Fatalf("close payment: %v", err)
	}
	if closed == nil ||
		!closed.Closed ||
		closed.PaymentID != payment.ID ||
		closed.PaymentNo != payment.PaymentNo {
		t.Fatalf("close result = %+v, want closed payment %+v", closed, payment)
	}
	requireOrderStatus(t, orderID, "pending_payment", "pending_payment")

	again, err := repository.CloseForOrder(ctx, testPool, orderID)
	if err != nil {
		t.Fatalf("close payment again: %v", err)
	}
	if again == nil || !again.Closed {
		t.Fatalf("second close result = %+v, want closed", again)
	}
}

func TestPaymentRepositoryCloseForOrderReportsSettledPayment(t *testing.T) {
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
		AmountCents:           3998,
		Currency:              model.CurrencyCNY,
		PaidAt:                time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("apply callback: %v", err)
	}

	closed, err := repository.CloseForOrder(ctx, testPool, orderID)
	if err != nil {
		t.Fatalf("close payment: %v", err)
	}
	if closed == nil || !closed.AlreadySettled || closed.Closed {
		t.Fatalf("close result = %+v, want already settled", closed)
	}
	requireOrderStatus(t, orderID, "paid", "awaiting_shipment")
}

func TestPaymentRepositoryCloseForOrderReturnsNilWhenPaymentDoesNotExist(t *testing.T) {
	resetPaymentDatabase(t)
	ctx := context.Background()
	userID := insertUser(t, "buyer01")
	orderID := insertOrder(t, userID, 3998)

	result, err := newPaymentRepository().CloseForOrder(ctx, testPool, orderID)
	if err != nil {
		t.Fatalf("close payment without payment: %v", err)
	}
	if result != nil {
		t.Fatalf("close result = %+v, want nil", result)
	}
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
	orderLifecycle := orderrepository.NewPaymentLifecycle()
	refundRepository := paymentrepository.NewRefundRepository(testPool, outbox.NewWriter())
	return paymentrepository.NewPaymentRepository(testPool, orderLifecycle, refundRepository)
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

func insertOrderItem(t *testing.T, orderID int) int {
	t.Helper()
	ctx := context.Background()

	var shopID int
	var subOrderID int
	if err := testPool.QueryRow(
		ctx,
		`SELECT id, shop_id FROM sub_orders WHERE order_id = $1`,
		orderID,
	).Scan(&subOrderID, &shopID); err != nil {
		t.Fatalf("read sub-order: %v", err)
	}

	var productID int
	if err := testPool.QueryRow(ctx, `
INSERT INTO products (shop_id, name, status)
VALUES ($1, 'T-shirt', 'on_shelf')
RETURNING id`, shopID).Scan(&productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}

	var skuID int
	if err := testPool.QueryRow(ctx, `
INSERT INTO product_skus (product_id, specs, price_cents, is_enabled)
VALUES ($1, '{"color":"black"}'::jsonb, 1999, true)
RETURNING id`, productID).Scan(&skuID); err != nil {
		t.Fatalf("insert SKU: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO inventory_items (sku_id, total_stock, frozen_stock)
VALUES ($1, 4, 2)`, skuID); err != nil {
		t.Fatalf("insert inventory: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO order_item_snapshots
    (sub_order_id, product_id, sku_id, product_name, sku_specs, unit_price_cents, quantity)
VALUES ($1, $2, $3, 'T-shirt', '{"color":"black"}'::jsonb, 1999, 2)`,
		subOrderID, productID, skuID,
	); err != nil {
		t.Fatalf("insert order item: %v", err)
	}
	return skuID
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

func requirePaymentFrozenStock(t *testing.T, skuID int, want int) {
	t.Helper()

	var frozen int
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT frozen_stock FROM inventory_items WHERE sku_id = $1`,
		skuID,
	).Scan(&frozen); err != nil {
		t.Fatalf("read frozen stock: %v", err)
	}
	if frozen != want {
		t.Fatalf("frozen stock = %d, want %d", frozen, want)
	}
}

func resetPaymentDatabase(t *testing.T) {
	t.Helper()

	const query = `TRUNCATE users, sessions, shops, shop_sessions, products, product_skus,
inventory_items, orders, sub_orders, order_item_snapshots, payments, refunds,
outbox_events RESTART IDENTITY CASCADE`
	if _, err := testPool.Exec(context.Background(), query); err != nil {
		t.Fatalf("reset payment database: %v", err)
	}
}

func requireRefundStatus(
	t *testing.T,
	paymentNo string,
	want model.RefundStatus,
) {
	t.Helper()

	var status string
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT r.status::text
FROM refunds r
JOIN payments p ON p.id = r.payment_id
WHERE p.payment_no = $1::uuid`,
		paymentNo,
	).Scan(&status); err != nil {
		t.Fatalf("read refund status: %v", err)
	}
	if model.RefundStatus(status) != want {
		t.Fatalf("refund status = %q, want %q", status, want)
	}
}

func requireRefundOutboxEvent(t *testing.T, paymentNo string) {
	t.Helper()

	var count int
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT COUNT(*)::int FROM outbox_events
WHERE event_key = $1 AND topic = $2`,
		"payment.refund:"+paymentNo,
		outbox.TopicPaymentRefundProvider,
	).Scan(&count); err != nil {
		t.Fatalf("read refund outbox event: %v", err)
	}
	if count != 1 {
		t.Fatalf("refund outbox events = %d, want 1", count)
	}
}

var _ paymentport.Lifecycle = (*orderrepository.PaymentLifecycle)(nil)
