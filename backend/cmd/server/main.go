package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	cartmodule "pinshaoshao/internal/feature/cart"
	feedmodule "pinshaoshao/internal/feature/feed"
	inventorymodule "pinshaoshao/internal/feature/inventory"
	logisticsmodule "pinshaoshao/internal/feature/logistics"
	ordermodule "pinshaoshao/internal/feature/order"
	paymentmodule "pinshaoshao/internal/feature/payment"
	productmodule "pinshaoshao/internal/feature/product"
	shopmodule "pinshaoshao/internal/feature/shop"
	usermodule "pinshaoshao/internal/feature/user"
	"pinshaoshao/internal/middleware"
	mocklogisticsmodule "pinshaoshao/internal/mock/logistics"
	mockpaymentmodule "pinshaoshao/internal/mock/payment"
	"pinshaoshao/internal/router"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	serverReadHeaderTimeout = 5 * time.Second
	serverReadTimeout       = 15 * time.Second
	serverWriteTimeout      = 30 * time.Second
	serverIdleTimeout       = 60 * time.Second
)

// @title                      pinshaoshao API
// @version                    0.1.0
// @description                E-commerce backend API.
// @BasePath                   /api/v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Session token in the form "Bearer <token>".
func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders())

	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fatal("load .env file", "error", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	corsOrigins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if corsOrigins == "" {
		fatal("CORS_ALLOWED_ORIGINS is required")
	}
	if strings.Contains(corsOrigins, "*") {
		fatal("CORS_ALLOWED_ORIGINS must contain explicit origins")
	}
	r.Use(middleware.CORS(corsOrigins))

	pool, err := openDatabasePool(ctx, "DATABASE_URL")
	if err != nil {
		fatal("open main database", "error", err)
	}
	defer pool.Close()

	mockPayPool, err := openDatabasePool(ctx, "MOCKPAY_DATABASE_URL")
	if err != nil {
		fatal("open mock payment database", "error", err)
	}
	defer mockPayPool.Close()

	mockLogisticsPool, err := openDatabasePool(ctx, "MOCKLOGISTICS_DATABASE_URL")
	if err != nil {
		fatal("open mock logistics database", "error", err)
	}
	defer mockLogisticsPool.Close()

	if err := validateDatabaseIsolation(ctx, map[string]*pgxpool.Pool{
		"DATABASE_URL":               pool,
		"MOCKPAY_DATABASE_URL":       mockPayPool,
		"MOCKLOGISTICS_DATABASE_URL": mockLogisticsPool,
	}); err != nil {
		fatal("validate database isolation", "error", err)
	}

	userModule := usermodule.NewModule(pool)
	shopModule := shopmodule.NewModule(pool)
	inventoryModule := inventorymodule.NewModule(pool)
	productModule := productmodule.NewModule(pool, inventoryModule.Lifecycle())
	cartModule := cartmodule.NewModule(pool)
	feedModule := feedmodule.NewModule(pool)
	orderModule := ordermodule.NewModule(pool, inventoryModule.Reservation())
	logisticsModule, err := logisticsmodule.NewModule(
		pool,
		orderModule.ShippingLifecycle(),
		logisticsmodule.ConfigFromEnv(),
	)
	if err != nil {
		fatal("create logistics module", "error", err)
	}
	paymentModule, err := paymentmodule.NewModule(
		pool,
		orderModule.PaymentLifecycle(),
		paymentmodule.ConfigFromEnv(),
	)
	if err != nil {
		fatal("create payment module", "error", err)
	}
	mockPaymentModule, err := mockpaymentmodule.NewModule(
		mockPayPool,
		mockpaymentmodule.ConfigFromEnv(),
	)
	if err != nil {
		fatal("create mock payment module", "error", err)
	}
	if err := mockPaymentModule.BootstrapAuth(ctx); err != nil {
		fatal("bootstrap mock payment auth", "error", err)
	}
	mockLogisticsModule, err := mocklogisticsmodule.NewModule(
		mockLogisticsPool,
		mocklogisticsmodule.ConfigFromEnv(),
	)
	if err != nil {
		fatal("create mock logistics module", "error", err)
	}
	if err := mockLogisticsModule.BootstrapAuth(ctx); err != nil {
		fatal("bootstrap mock logistics auth", "error", err)
	}
	mockPaymentModule.Start(context.Background())
	defer mockPaymentModule.Stop()
	mockLogisticsModule.Start(context.Background())
	defer mockLogisticsModule.Stop()

	router.Register(r, router.Modules{
		User:              userModule,
		Shop:              shopModule,
		Product:           productModule,
		Inventory:         inventoryModule,
		Cart:              cartModule,
		Feed:              feedModule,
		Order:             orderModule,
		Payment:           paymentModule,
		Logistics:         logisticsModule,
		MockPayment:       mockPaymentModule,
		MockLogistics:     mockLogisticsModule,
		UserAuth:          userModule.AuthMiddleware(),
		ShopAuth:          shopModule.AuthMiddleware(),
		MockPaymentAuth:   mockPaymentModule.AuthMiddleware(),
		MockLogisticsAuth: mockLogisticsModule.AuthMiddleware(),
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fatal("run server", "error", err)
	}
}

func openDatabasePool(ctx context.Context, environment string) (*pgxpool.Pool, error) {
	dsn := os.Getenv(environment)
	if dsn == "" {
		return nil, fmt.Errorf("%s is not set", environment)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func validateDatabaseIsolation(
	ctx context.Context,
	pools map[string]*pgxpool.Pool,
) error {
	environmentByDatabase := make(map[string]string, len(pools))
	for environment, pool := range pools {
		var databaseName string
		if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
			return fmt.Errorf("read current database for %s: %w", environment, err)
		}
		if previous, exists := environmentByDatabase[databaseName]; exists {
			return fmt.Errorf(
				"%s and %s use the same database %q",
				previous,
				environment,
				databaseName,
			)
		}
		environmentByDatabase[databaseName] = environment
	}
	return nil
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
