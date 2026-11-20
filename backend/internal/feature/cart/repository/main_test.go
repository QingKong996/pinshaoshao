//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"

	cartrepository "pinshaoshao/internal/feature/cart/repository"
	inventoryrepository "pinshaoshao/internal/feature/inventory/repository"
	productmodel "pinshaoshao/internal/feature/product/model"
	productrepository "pinshaoshao/internal/feature/product/repository"
	shopmodel "pinshaoshao/internal/feature/shop/model"
	shoprepository "pinshaoshao/internal/feature/shop/repository"
	usermodel "pinshaoshao/internal/feature/user/model"
	userrepository "pinshaoshao/internal/feature/user/repository"

	"github.com/jackc/pgx/v5"
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

func resetDatabase(t *testing.T) {
	t.Helper()

	const query = `TRUNCATE users, sessions, shops, shop_sessions, cart_items,
products, product_skus, inventory_items RESTART IDENTITY CASCADE`
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

func insertUser(t *testing.T, username string) *usermodel.User {
	t.Helper()

	user := &usermodel.User{
		Username:     username,
		PasswordHash: "test-password-hash",
	}
	if err := userrepository.NewUserRepository(testPool).CreateUserWithSession(
		context.Background(),
		user,
		[]byte("test-session-"+username),
	); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func insertCatalogSKU(t *testing.T, status productmodel.ProductStatus, enabled bool, totalStock int) (*productmodel.Product, *productmodel.ProductSKU) {
	t.Helper()
	ctx := context.Background()

	shop := &shopmodel.Shop{
		Username:     "shop01",
		PasswordHash: "test-password-hash",
		Email:        "shop01@example.com",
	}
	if err := shoprepository.NewShopRepository(testPool).CreateShopWithSession(ctx, shop, []byte("test-shop-session")); err != nil {
		t.Fatalf("create shop: %v", err)
	}

	productRepository := productrepository.NewProductRepository(testPool)
	product := &productmodel.Product{
		ShopID: shop.ID,
		Name:   "T-shirt",
		Status: status,
	}
	if err := productRepository.CreateProduct(ctx, product); err != nil {
		t.Fatalf("create product: %v", err)
	}

	skuRepository := productrepository.NewProductSKURepository(testPool)
	sku := &productmodel.ProductSKU{
		ProductID:  product.ID,
		Specs:      map[string]string{"color": "black"},
		PriceCents: 1999,
		IsEnabled:  enabled,
	}
	if err := pgx.BeginFunc(ctx, testPool, func(tx pgx.Tx) error {
		if err := skuRepository.CreateProductSKU(ctx, tx, sku); err != nil {
			return err
		}
		return inventoryrepository.NewInventoryRepository(testPool).CreateEmptyForSKU(ctx, tx, sku.ID)
	}); err != nil {
		t.Fatalf("create sku: %v", err)
	}
	if _, err := testPool.Exec(ctx, `UPDATE inventory_items SET total_stock = $1 WHERE sku_id = $2`, totalStock, sku.ID); err != nil {
		t.Fatalf("set inventory: %v", err)
	}
	return product, sku
}

func newTestCartRepository() *cartrepository.CartRepository {
	return cartrepository.NewCartRepository(testPool)
}
