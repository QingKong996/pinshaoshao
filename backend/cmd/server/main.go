package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	inventorymodule "pinshaoshao/internal/feature/inventory"
	productmodule "pinshaoshao/internal/feature/product"
	shopmodule "pinshaoshao/internal/feature/shop"
	usermodule "pinshaoshao/internal/feature/user"
	"pinshaoshao/internal/middleware"
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
	productModule := productmodule.NewModule(pool)
	inventoryModule := inventorymodule.NewModule(pool)

	router.Register(r, router.Modules{
		User:      userModule,
		Shop:      shopModule,
		Product:   productModule,
		Inventory: inventoryModule,
		UserAuth:  middleware.UserAuth(userModule.Authenticator()),
		ShopAuth:  middleware.ShopAuth(shopModule.Authenticator()),
	})

	if err := r.Run(":" + port); err != nil {
		fatal("run server", "error", err)
	}
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
