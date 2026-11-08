package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"

	"pinshaoshao/internal/model"
	"pinshaoshao/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	testUserUsername = "alice01"
	testUserPassword = "hashed-password"
	testUserEmail    = "alice01@example.com"

	testShopUsername = "shop01"
	testShopPassword = "hashed-password"
	testShopEmail    = "shop01@example.com"
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
	return goose.UpContext(ctx, db, filepath.Join("..", "..", "migrations"))
}

func resetDatabase(t *testing.T) {
	t.Helper()

	const query = `TRUNCATE users, sessions, shops, shop_sessions, products, product_skus RESTART IDENTITY CASCADE`
	if _, err := testPool.Exec(context.Background(), query); err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

func requireErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want %v", err, target)
	}
}

func insertUser(t *testing.T) *model.User {
	t.Helper()

	user := &model.User{
		Username:     testUserUsername,
		PasswordHash: testUserPassword,
		Email:        testUserEmail,
	}
	if err := repository.NewUserRepository(testPool).CreateUser(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func insertShop(t *testing.T) *model.Shop {
	t.Helper()

	shop := &model.Shop{
		Username:     testShopUsername,
		PasswordHash: testShopPassword,
		Email:        testShopEmail,
	}
	if err := repository.NewShopRepository(testPool).CreateShop(context.Background(), shop); err != nil {
		t.Fatalf("create shop: %v", err)
	}
	return shop
}

func insertProduct(t *testing.T, shopID int) *model.Product {
	t.Helper()

	product := &model.Product{
		ShopID: shopID,
		Name:   "T-shirt",
	}
	if err := repository.NewProductRepository(testPool).CreateProduct(context.Background(), product); err != nil {
		t.Fatalf("create product: %v", err)
	}
	return product
}
