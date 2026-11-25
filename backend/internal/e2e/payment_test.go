//go:build integration

package e2e_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	inventorymodule "pinshaoshao/internal/feature/inventory"
	logisticsmodule "pinshaoshao/internal/feature/logistics"
	ordermodule "pinshaoshao/internal/feature/order"
	orderdto "pinshaoshao/internal/feature/order/dto"
	paymentmodule "pinshaoshao/internal/feature/payment"
	paymentdto "pinshaoshao/internal/feature/payment/dto"
	"pinshaoshao/internal/http/authctx"
	mockauth "pinshaoshao/internal/mock/auth"
	mocklogisticsmodule "pinshaoshao/internal/mock/logistics"
	mockpaymentmodule "pinshaoshao/internal/mock/payment"
	mockpaymentdto "pinshaoshao/internal/mock/payment/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const testWebhookSecret = "payment-e2e-webhook-secret"

const (
	testMockUsername = "mock-operator"
	testMockPassword = "mockpay-test-password"
)

var (
	testPool              *pgxpool.Pool
	testMockPayPool       *pgxpool.Pool
	testMockLogisticsPool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("pinshaoshao_e2e"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		slog.Error("start postgres container", "error", err)
		os.Exit(1)
	}

	dsn := container.MustConnectionString(ctx, "sslmode=disable")
	databaseNames := []string{
		"pinshaoshao_mockpay_e2e",
		"pinshaoshao_mocklogistics_e2e",
	}
	for _, databaseName := range databaseNames {
		if err := createDatabase(ctx, dsn, databaseName); err != nil {
			slog.Error("create test database", "database", databaseName, "error", err)
			os.Exit(1)
		}
	}

	mockPayDSN := databaseDSN(dsn, "pinshaoshao_mockpay_e2e")
	mockLogisticsDSN := databaseDSN(dsn, "pinshaoshao_mocklogistics_e2e")
	for position, migration := range []struct {
		dsn  string
		path string
	}{
		{dsn: dsn, path: filepath.Join("..", "..", "migrations")},
		{dsn: mockPayDSN, path: filepath.Join("..", "..", "migrations", "mockpay")},
		{dsn: mockLogisticsDSN, path: filepath.Join("..", "..", "migrations", "mocklogistics")},
	} {
		if err := migrate(ctx, migration.dsn, migration.path); err != nil {
			slog.Error("migrate test database", "position", position, "error", err)
			os.Exit(1)
		}
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("connect test database", "error", err)
		os.Exit(1)
	}
	testMockPayPool, err = pgxpool.New(ctx, mockPayDSN)
	if err != nil {
		slog.Error("connect mock payment test database", "error", err)
		os.Exit(1)
	}
	testMockLogisticsPool, err = pgxpool.New(ctx, mockLogisticsDSN)
	if err != nil {
		slog.Error("connect mock logistics test database", "error", err)
		os.Exit(1)
	}

	code := m.Run()

	testPool.Close()
	testMockPayPool.Close()
	testMockLogisticsPool.Close()
	if err := testcontainers.TerminateContainer(container); err != nil {
		slog.Error("terminate postgres container", "error", err)
	}
	os.Exit(code)
}

func migrate(ctx context.Context, dsn string, migrationPath string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetLogger(goose.NopLogger())
	return goose.UpContext(ctx, db, migrationPath)
}

func createDatabase(ctx context.Context, dsn string, databaseName string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.ExecContext(
		ctx,
		fmt.Sprintf(`CREATE DATABASE "%s"`, databaseName),
	)
	return err
}

func databaseDSN(dsn string, databaseName string) string {
	parsed, err := url.Parse(dsn)
	if err != nil {
		panic(err)
	}
	parsed.Path = "/" + databaseName
	return parsed.String()
}

func TestPaymentHTTPFlow(t *testing.T) {
	resetDatabase(t)
	buyerID := insertBuyer(t)
	skuID, unitPriceCents := insertSellableSKU(t)
	app := startTestApp(t, buyerID)

	var order orderdto.OrderResponse
	app.doJSON(t, http.MethodPost, "/api/v1/orders/direct", map[string]any{
		"checkout_key":     uuid.NewString(),
		"sku_id":           skuID,
		"quantity":         2,
		"receiver_name":    "E2E Buyer",
		"receiver_phone":   "13800000000",
		"shipping_address": "E2E Street 1",
	}, http.StatusCreated, &order)
	if order.Status != "pending_payment" ||
		order.TotalAmountCents != unitPriceCents*2 ||
		len(order.SubOrders) != 1 ||
		order.SubOrders[0].Status != "pending_payment" {
		t.Fatalf("created order = %+v", order)
	}
	requireFrozenStock(t, skuID, 2)

	var payment paymentdto.OrderPaymentResponse
	app.doJSON(t, http.MethodPost, "/api/v1/payments", map[string]int{
		"order_id": order.ID,
	}, http.StatusCreated, &payment)
	if payment.Status != "pending" ||
		payment.OrderID != order.ID ||
		payment.AmountCents != order.TotalAmountCents {
		t.Fatalf("created payment = %+v", payment)
	}

	var session paymentdto.OrderPaymentResponse
	app.doJSON(
		t,
		http.MethodPost,
		"/api/v1/payments/"+payment.PaymentNo+"/pay",
		nil,
		http.StatusOK,
		&session,
	)
	if session.PaymentNo != payment.PaymentNo ||
		session.Status != "pending" ||
		session.ProviderTransactionID == nil ||
		*session.ProviderTransactionID == "" ||
		session.CheckoutURL == "" {
		t.Fatalf("payment session = %+v", session)
	}
	providerNo := *session.ProviderTransactionID
	checkoutURL, err := url.Parse(session.CheckoutURL)
	if err != nil {
		t.Fatalf("parse checkout URL: %v", err)
	}
	wantCheckoutPath := "/mockpay/checkout/" + providerNo
	if checkoutURL.Path != wantCheckoutPath {
		t.Fatalf("checkout path = %q, want %q", checkoutURL.Path, wantCheckoutPath)
	}

	response, body := app.do(t, http.MethodGet, session.CheckoutURL, nil, "")
	if response.StatusCode != http.StatusOK ||
		!strings.Contains(string(body), "Pay now") {
		t.Fatalf("checkout response = %d/%s", response.StatusCode, body)
	}

	response, _ = app.do(t, http.MethodPost, session.CheckoutURL+"/pay", nil, "")
	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("checkout pay status = %d, want 303", response.StatusCode)
	}
	if got := response.Header.Get("Location"); got != wantCheckoutPath {
		t.Fatalf("checkout redirect = %q, want %q", got, wantCheckoutPath)
	}

	var settled paymentdto.OrderPaymentResponse
	waitFor(t, 10*time.Second, 50*time.Millisecond, func() bool {
		app.doJSON(
			t,
			http.MethodGet,
			"/api/v1/payments/"+payment.PaymentNo,
			nil,
			http.StatusOK,
			&settled,
		)
		return settled.Status == "succeeded"
	})
	if settled.ProviderTransactionID == nil ||
		*settled.ProviderTransactionID != providerNo ||
		settled.PaidAt == nil {
		t.Fatalf("settled payment = %+v", settled)
	}

	var paidOrder orderdto.OrderResponse
	waitFor(t, 5*time.Second, 50*time.Millisecond, func() bool {
		app.doJSON(
			t,
			http.MethodGet,
			fmt.Sprintf("/api/v1/orders/%d", order.ID),
			nil,
			http.StatusOK,
			&paidOrder,
		)
		return paidOrder.Status == "paid" &&
			len(paidOrder.SubOrders) == 1 &&
			paidOrder.SubOrders[0].Status == "awaiting_shipment"
	})
	if paidOrder.PaidAt == nil {
		t.Fatalf("paid order has no paid_at: %+v", paidOrder)
	}
	requireFrozenStock(t, skuID, 2)

	var notifications mockpaymentdto.NotificationListResponse
	waitFor(t, 5*time.Second, 50*time.Millisecond, func() bool {
		app.doJSON(
			t,
			http.MethodGet,
			"/mockpay/api/v1/payments/"+providerNo+"/notifications",
			nil,
			http.StatusOK,
			&notifications,
		)
		for _, notification := range notifications.Items {
			if notification.Status == "succeeded" &&
				notification.EventType == "payment.paid" {
				return true
			}
		}
		return false
	})

	var replayed paymentdto.OrderPaymentResponse
	app.doJSON(
		t,
		http.MethodPost,
		"/api/v1/payments/"+payment.PaymentNo+"/pay",
		nil,
		http.StatusOK,
		&replayed,
	)
	if replayed.Status != "succeeded" || replayed.CheckoutURL != "" {
		t.Fatalf("replayed payment session = %+v", replayed)
	}
}

type testApp struct {
	serverURL          string
	mockPayToken       string
	mockLogisticsToken string
	client             *http.Client
}

func startTestApp(t *testing.T, userID int) *testApp {
	t.Helper()

	server := httptest.NewUnstartedServer(nil)
	t.Cleanup(server.Close)
	baseURL := "http://" + server.Listener.Addr().String()

	inventoryModule := inventorymodule.NewModule(testPool)
	orderModule := ordermodule.NewModule(testPool, inventoryModule.Reservation())
	paymentModule, err := paymentmodule.NewModule(
		testPool,
		orderModule.PaymentLifecycle(),
		paymentmodule.Config{
			GatewayBaseURL:  baseURL,
			WebhookSecret:   testWebhookSecret,
			NotifyURL:       baseURL + "/api/v1/payment/webhook",
			ReturnURLBase:   "http://127.0.0.1:4173",
			HTTPClient:      &http.Client{Timeout: 5 * time.Second},
			GatewayUsername: testMockUsername,
			GatewayPassword: testMockPassword,
		},
	)
	if err != nil {
		t.Fatalf("create payment module: %v", err)
	}
	logisticsModule, err := logisticsmodule.NewModule(
		testPool,
		orderModule.ShippingLifecycle(),
		logisticsmodule.Config{
			GatewayBaseURL:  baseURL,
			WebhookSecret:   testWebhookSecret,
			NotifyURL:       baseURL + "/api/v1/logistics/webhook",
			HTTPClient:      &http.Client{Timeout: 5 * time.Second},
			GatewayUsername: testMockUsername,
			GatewayPassword: testMockPassword,
		},
	)
	if err != nil {
		t.Fatalf("create logistics module: %v", err)
	}
	mockPaymentModule, err := mockpaymentmodule.NewModule(testMockPayPool, mockpaymentmodule.Config{
		WebhookSecret:            testWebhookSecret,
		PublicBaseURL:            baseURL,
		NotificationPollInterval: 25 * time.Millisecond,
		ExpiryPollInterval:       time.Hour,
		NotificationRetryDelays:  []time.Duration{50 * time.Millisecond},
		Auth: mockauth.Config{
			Username: testMockUsername,
			Password: testMockPassword,
		},
	})
	if err != nil {
		t.Fatalf("create mock payment module: %v", err)
	}
	mockLogisticsModule, err := mocklogisticsmodule.NewModule(testMockLogisticsPool, mocklogisticsmodule.Config{
		WebhookSecret:            testWebhookSecret,
		NotificationPollInterval: 25 * time.Millisecond,
		NotificationRetryDelays:  []time.Duration{50 * time.Millisecond},
		HTTPClient:               &http.Client{Timeout: 5 * time.Second},
		Auth: mockauth.Config{
			Username: testMockUsername,
			Password: testMockPassword,
		},
	})
	if err != nil {
		t.Fatalf("create mock logistics module: %v", err)
	}
	if err := mockPaymentModule.BootstrapAuth(context.Background()); err != nil {
		t.Fatalf("bootstrap mock payment auth: %v", err)
	}
	if err := mockLogisticsModule.BootstrapAuth(context.Background()); err != nil {
		t.Fatalf("bootstrap mock logistics auth: %v", err)
	}
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	mockPaymentModule.Start(workerCtx)
	mockLogisticsModule.Start(workerCtx)
	t.Cleanup(func() {
		stopWorkers()
		mockPaymentModule.Stop()
		mockLogisticsModule.Stop()
	})

	var shopID int
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT id FROM shops ORDER BY id LIMIT 1`,
	).Scan(&shopID); err != nil {
		t.Fatalf("read test shop: %v", err)
	}
	router := gin.New()
	auth := func(ctx *gin.Context) {
		ctx.Set(authctx.UserIDKey, userID)
		ctx.Set(authctx.ShopIDKey, shopID)
		ctx.Next()
	}
	orderModule.Register(router, auth)
	paymentModule.Register(router, auth)
	logisticsModule.Register(router, auth)
	logisticsModule.RegisterMerchant(router, auth)
	mockPaymentModule.Register(router, mockPaymentModule.AuthMiddleware())
	mockLogisticsModule.Register(router, mockLogisticsModule.AuthMiddleware())
	server.Config.Handler = router
	server.Start()
	mockPayToken := loginMockService(t, baseURL, "/mockpay/api/v1/auth/login")
	mockLogisticsToken := loginMockService(
		t,
		baseURL,
		"/mocklogistics/api/v1/auth/login",
	)

	return &testApp{
		serverURL:          baseURL,
		mockPayToken:       mockPayToken,
		mockLogisticsToken: mockLogisticsToken,
		client: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func loginMockService(t *testing.T, baseURL string, loginPath string) string {
	t.Helper()
	body, err := json.Marshal(map[string]string{
		"username": testMockUsername,
		"password": testMockPassword,
	})
	if err != nil {
		t.Fatalf("encode mock login: %v", err)
	}
	request, err := http.NewRequest(
		http.MethodPost,
		baseURL+loginPath,
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create mock login request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("login mock service: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 1<<16))
		t.Fatalf("mock login status = %d, body = %s", response.StatusCode, responseBody)
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode mock login: %v", err)
	}
	return result.Token
}

func (a *testApp) doJSON(
	t *testing.T,
	method string,
	target string,
	input any,
	wantStatus int,
	output any,
) {
	t.Helper()

	var body []byte
	var err error
	contentType := ""
	if input != nil {
		body, err = json.Marshal(input)
		if err != nil {
			t.Fatalf("encode request: %v", err)
		}
		contentType = "application/json"
	}
	response, responseBody := a.do(t, method, target, body, contentType)
	if response.StatusCode != wantStatus {
		t.Fatalf(
			"%s %s status = %d, want %d, body = %s",
			method,
			target,
			response.StatusCode,
			wantStatus,
			responseBody,
		)
	}
	if output == nil {
		return
	}
	if err := json.Unmarshal(responseBody, output); err != nil {
		t.Fatalf("decode %s %s response: %v", method, target, err)
	}
}

func (a *testApp) do(
	t *testing.T,
	method string,
	target string,
	body []byte,
	contentType string,
) (*http.Response, []byte) {
	t.Helper()

	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = a.serverURL + target
	}
	request, err := http.NewRequestWithContext(
		context.Background(),
		method,
		target,
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create %s %s request: %v", method, target, err)
	}
	token := "e2e-session"
	switch {
	case strings.Contains(target, "/mockpay/api/"):
		token = a.mockPayToken
	case strings.Contains(target, "/mocklogistics/api/"):
		token = a.mockLogisticsToken
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	response, err := a.client.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		t.Fatalf("read %s %s response: %v", method, target, err)
	}
	return response, responseBody
}

func waitFor(t *testing.T, timeout time.Duration, interval time.Duration, check func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(interval)
	}
	t.Fatal("condition was not met before timeout")
}

func insertBuyer(t *testing.T) int {
	t.Helper()

	var userID int
	if err := testPool.QueryRow(
		context.Background(),
		`INSERT INTO users (username, password_hash, email)
VALUES ('e2e-buyer', 'hash', 'e2e-buyer@example.com')
RETURNING id`,
	).Scan(&userID); err != nil {
		t.Fatalf("insert buyer: %v", err)
	}
	return userID
}

func insertSellableSKU(t *testing.T) (int, int64) {
	t.Helper()
	ctx := context.Background()
	const priceCents = int64(1999)

	var shopID int
	if err := testPool.QueryRow(ctx, `INSERT INTO shops (username, password_hash, email, name)
VALUES ('e2e-shop', 'hash', 'e2e-shop@example.com', 'E2E Shop')
RETURNING id`).Scan(&shopID); err != nil {
		t.Fatalf("insert shop: %v", err)
	}

	var productID int
	if err := testPool.QueryRow(ctx, `INSERT INTO products (shop_id, name, status)
VALUES ($1, 'E2E Product', 'on_shelf')
RETURNING id`, shopID).Scan(&productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}

	var skuID int
	if err := testPool.QueryRow(ctx, `INSERT INTO product_skus
    (product_id, specs, price_cents, is_enabled)
VALUES ($1, '{"Size":"M"}'::jsonb, $2, true)
RETURNING id`, productID, priceCents).Scan(&skuID); err != nil {
		t.Fatalf("insert SKU: %v", err)
	}
	if _, err := testPool.Exec(
		ctx,
		`INSERT INTO inventory_items (sku_id, total_stock) VALUES ($1, 10)`,
		skuID,
	); err != nil {
		t.Fatalf("insert inventory: %v", err)
	}
	return skuID, priceCents
}

func requireFrozenStock(t *testing.T, skuID int, want int) {
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

func resetDatabase(t *testing.T) {
	t.Helper()

	const query = `TRUNCATE shipment_events, shipments, payments,
order_item_snapshots, sub_orders, orders, inventory_items, product_skus, products,
product_categories, shop_sessions, shops, sessions, users RESTART IDENTITY CASCADE`
	if _, err := testPool.Exec(context.Background(), query); err != nil {
		t.Fatalf("reset database: %v", err)
	}
	if _, err := testMockPayPool.Exec(
		context.Background(),
		`TRUNCATE mock_payment_notifications, mock_payment_orders RESTART IDENTITY CASCADE`,
	); err != nil {
		t.Fatalf("reset mock payment database: %v", err)
	}
	if _, err := testMockLogisticsPool.Exec(
		context.Background(),
		`TRUNCATE mock_logistics_notifications, mock_logistics_events, mock_shipments RESTART IDENTITY CASCADE`,
	); err != nil {
		t.Fatalf("reset mock logistics database: %v", err)
	}
}
