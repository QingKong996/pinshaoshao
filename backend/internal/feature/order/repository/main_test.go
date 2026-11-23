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
	orderrepository "pinshaoshao/internal/feature/order/repository"
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
products, product_skus, inventory_items, orders, sub_orders, order_item_snapshots
RESTART IDENTITY CASCADE`
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

func insertBuyer(t *testing.T, username string) *usermodel.User {
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

func insertCatalogSKU(
	t *testing.T,
	shopUsername string,
	shopName string,
	status productmodel.ProductStatus,
	enabled bool,
	priceCents int64,
	totalStock int,
) (*shopmodel.Shop, *productmodel.Product, *productmodel.ProductSKU) {
	t.Helper()
	ctx := context.Background()

	shop := &shopmodel.Shop{
		Username:     shopUsername,
		Name:         shopName,
		PasswordHash: "test-password-hash",
		Email:        shopUsername + "@example.com",
	}
	if err := shoprepository.NewShopRepository(testPool).CreateShopWithSession(ctx, shop, []byte("test-shop-session-"+shopUsername)); err != nil {
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

	sku := &productmodel.ProductSKU{
		ProductID:  product.ID,
		Specs:      map[string]string{"color": "black"},
		PriceCents: priceCents,
		IsEnabled:  enabled,
	}
	if err := pgx.BeginFunc(ctx, testPool, func(tx pgx.Tx) error {
		if err := productrepository.NewProductSKURepository(testPool).CreateProductSKU(ctx, tx, sku); err != nil {
			return err
		}
		return inventoryrepository.NewInventoryRepository(testPool).CreateEmptyForSKU(ctx, tx, sku.ID)
	}); err != nil {
		t.Fatalf("create sku: %v", err)
	}
	if _, err := testPool.Exec(ctx, `UPDATE inventory_items SET total_stock = $1 WHERE sku_id = $2`, totalStock, sku.ID); err != nil {
		t.Fatalf("set inventory: %v", err)
	}
	return shop, product, sku
}

func newTestOrderRepository() *orderrepository.OrderRepository {
	inventoryRepository := inventoryrepository.NewInventoryRepository(testPool)
	return orderrepository.NewOrderRepository(testPool, inventoryRepository)
}

func addCartItem(t *testing.T, userID int, skuID int, quantity int) *cartItem {
	t.Helper()

	item, err := cartrepository.NewCartRepository(testPool).AddItem(context.Background(), userID, skuID, quantity)
	if err != nil {
		t.Fatalf("add cart item: %v", err)
	}
	return &cartItem{ID: item.ID}
}

type cartItem struct {
	ID int
}
