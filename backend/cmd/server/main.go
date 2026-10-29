package main

import (
	"context"
	"log"
	"os"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/repository"
	"pinshaoshao/internal/router"
	"pinshaoshao/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	r := gin.Default()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	userRepository := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepository)
	userController := controller.NewUserController(userService)

	router.UserRegister(r, userController)

	err = r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}

}
