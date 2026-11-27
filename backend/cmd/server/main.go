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
	"pinshaoshao/internal/outbox"
	"pinshaoshao/internal/ratelimit"
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

	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fatal("load .env file", "error", err)
	}

	accessLogger := middleware.NewAccessLogger(slog.Default())
	defer func() {
		closeContext, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := accessLogger.Close(closeContext); err != nil {
			slog.Error("close access logger", "error", err)
		}
	}()

	r := gin.New()
	r.Use(accessLogger.Middleware())
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.NoStoreAPI())
	if err := configureTrustedProxies(r); err != nil {
		fatal("configure trusted proxies", "error", err)
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

	rateLimitConfig, err := ratelimit.ConfigFromEnv()
	if err != nil {
		fatal("load rate limit config", "error", err)
	}
	rateLimiter, err := ratelimit.NewRedisLimiter(ctx, os.Getenv("REDIS_URL"), rateLimitConfig)
	if err != nil {
		fatal("open redis rate limiter", "error", err)
	}
	defer func() {
		if err := rateLimiter.Close(); err != nil {
			slog.Error("close redis rate limiter", "error", err)
		}
	}()
	r.Use(middleware.CORS(corsOrigins, ratelimit.PreflightMiddleware(rateLimiter)))
	r.Use(ratelimit.APIEntryMiddleware(rateLimiter, router.IsPublicAPI))
	r.Use(middleware.RequestBodyLimit(middleware.DefaultRequestBodyLimit))

	userModule := usermodule.NewModule(pool)
	shopModule := shopmodule.NewModule(pool)
	inventoryModule := inventorymodule.NewModule(pool)
	productModule := productmodule.NewModule(pool, inventoryModule.Lifecycle())
	cartModule := cartmodule.NewModule(pool)
	feedModule := feedmodule.NewModule(pool)
	outboxWriter := outbox.NewWriter()
	orderModule := ordermodule.NewModule(pool, inventoryModule.Reservation(), outboxWriter)
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
		orderModule.OrderService(),
		paymentmodule.ConfigFromEnv(),
		outboxWriter,
	)
	if err != nil {
		fatal("create payment module", "error", err)
	}
	orderModule.SetPaymentCloser(paymentModule.PaymentCloser())
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

	outboxWorker := outbox.NewWorker(
		outbox.NewRepository(pool),
		paymentModule.OutboxHandlers(),
		outbox.WorkerConfig{},
	)
	outboxWorker.Start(context.Background())
	defer outboxWorker.Stop()
	orderModule.StartBackgroundWorkers(context.Background())
	defer orderModule.StopBackgroundWorkers()

	router.Register(r, router.Modules{
		User:          userModule,
		Shop:          shopModule,
		Product:       productModule,
		Inventory:     inventoryModule,
		Cart:          cartModule,
		Feed:          feedModule,
		Order:         orderModule,
		Payment:       paymentModule,
		Logistics:     logisticsModule,
		MockPayment:   mockPaymentModule,
		MockLogistics: mockLogisticsModule,
		UserAuth: userModule.AuthMiddlewareWithGuards(
			ratelimit.UserMiddleware(rateLimiter),
			ratelimit.UnauthenticatedGuard(rateLimiter),
		),
		ShopAuth: shopModule.AuthMiddlewareWithGuards(
			ratelimit.ShopMiddleware(rateLimiter),
			ratelimit.UnauthenticatedGuard(rateLimiter),
		),
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

func configureTrustedProxies(r *gin.Engine) error {
	value := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))
	if value == "" {
		return r.SetTrustedProxies(nil)
	}

	proxies := strings.Split(value, ",")
	for index := range proxies {
		proxies[index] = strings.TrimSpace(proxies[index])
	}
	return r.SetTrustedProxies(proxies)
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
