package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/repository"
	"pinshaoshao/internal/router"
	"pinshaoshao/internal/service"
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

	userRepository := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepository)
	userController := controller.NewUserController(userService)
	userAuthMiddleware := middleware.UserAuth(userService)

	shopRepository := repository.NewShopRepository(pool)
	shopService := service.NewShopService(shopRepository)
	shopController := controller.NewShopController(shopService)
	shopAuthMiddleware := middleware.ShopAuth(shopService)

	productRepository := repository.NewProductRepository(pool)
	productCategoryRepository := repository.NewProductCategoryRepository(pool)
	productSKURepository := repository.NewProductSKURepository(pool)
	inventoryRepository := repository.NewInventoryRepository(pool)
	dashboardRepository := repository.NewDashboardRepository(pool)
	productSKUService := service.NewProductSKUService(productRepository, productSKURepository)
	productCategoryService := service.NewProductCategoryService(productCategoryRepository)
	productService := service.NewProductService(productRepository, productCategoryRepository)
	inventoryService := service.NewInventoryService(inventoryRepository)
	dashboardService := service.NewDashboardService(dashboardRepository)
	productSKUController := controller.NewProductSKUController(productSKUService)
	productCategoryController := controller.NewProductCategoryController(productCategoryService)
	productController := controller.NewProductController(productService)
	inventoryController := controller.NewInventoryController(inventoryService)
	dashboardController := controller.NewDashboardController(dashboardService)

	router.UserRegister(r, userController, userAuthMiddleware)
	router.ShopRegister(r, shopController, shopAuthMiddleware)
	router.MerchantRegister(
		r,
		shopController,
		productCategoryController,
		productController,
		productSKUController,
		inventoryController,
		dashboardController,
		shopAuthMiddleware,
	)

	if err := r.Run(":" + port); err != nil {
		fatal("run server", "error", err)
	}
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
