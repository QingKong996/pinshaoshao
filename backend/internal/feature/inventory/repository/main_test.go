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

	productmodel "pinshaoshao/internal/feature/product/model"
	productrepository "pinshaoshao/internal/feature/product/repository"
	shopmodel "pinshaoshao/internal/feature/shop/model"
	shoprepository "pinshaoshao/internal/feature/shop/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
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
	return goose.UpContext(ctx, db, filepath.Join("..", "..", "..", "..", "migrations"))
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

// testTokenHash returns a deterministic fixture token hash, unique per subject.
func testTokenHash(subject string) []byte {
	return []byte("test-session-" + subject)
}

func insertShop(t *testing.T) *shopmodel.Shop {
	t.Helper()

	return insertShopWithUsername(t, context.Background(), testShopUsername, testShopEmail)
}

func insertShopWithUsername(t *testing.T, ctx context.Context, username, email string) *shopmodel.Shop {
	t.Helper()

	shop := &shopmodel.Shop{
		Username:     username,
		PasswordHash: testShopPassword,
		Email:        email,
	}
	if err := shoprepository.NewShopRepository(testPool).CreateShopWithSession(ctx, shop, testTokenHash(shop.Username)); err != nil {
		t.Fatalf("create shop: %v", err)
	}
	return shop
}

func createTestSKU(t *testing.T, ctx context.Context, skuRepository *productrepository.ProductSKURepository, productID int, specs map[string]string, priceCents int64) *productmodel.ProductSKU {
	t.Helper()

	sku := &productmodel.ProductSKU{
		ProductID:  productID,
		Specs:      specs,
		PriceCents: priceCents,
		IsEnabled:  true,
	}
	if err := skuRepository.CreateProductSKU(ctx, testPool, sku); err != nil {
		t.Fatalf("create product sku: %v", err)
	}
	return sku
}

func setTestInventory(t *testing.T, ctx context.Context, skuID int, totalStock int, frozenStock int) {
	t.Helper()

	if _, err := testPool.Exec(ctx, `UPDATE inventory_items SET total_stock = $1, frozen_stock = $2 WHERE sku_id = $3`, totalStock, frozenStock, skuID); err != nil {
		t.Fatalf("set inventory: %v", err)
	}
}
