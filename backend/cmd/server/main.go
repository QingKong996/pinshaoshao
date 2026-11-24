package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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

	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fatal("load .env file", "error", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}
	r.Use(middleware.CORS(corsOrigins))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fatal("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fatal("create database pool", "error", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fatal("connect database", "error", err)
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
	mockPaymentModule, err := mockpaymentmodule.NewModule(pool, mockpaymentmodule.ConfigFromEnv())
	if err != nil {
		fatal("create mock payment module", "error", err)
	}
	mockLogisticsModule, err := mocklogisticsmodule.NewModule(pool, mocklogisticsmodule.ConfigFromEnv())
	if err != nil {
		fatal("create mock logistics module", "error", err)
	}
	mockPaymentModule.Start(context.Background())
	defer mockPaymentModule.Stop()
	mockLogisticsModule.Start(context.Background())
	defer mockLogisticsModule.Stop()

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
		UserAuth:      userModule.AuthMiddleware(),
		ShopAuth:      shopModule.AuthMiddleware(),
	})

	if err := r.Run(":" + port); err != nil {
		fatal("run server", "error", err)
	}
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
