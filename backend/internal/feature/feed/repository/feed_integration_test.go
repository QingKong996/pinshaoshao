//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"pinshaoshao/internal/feature/feed/apperr"
	"pinshaoshao/internal/feature/feed/model"
	"pinshaoshao/internal/feature/feed/repository"

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

func TestFeedRepositoryFiltersSortsAndPaginates(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	shopID := insertShop(t, "feed-shop")

	aggregatedID := insertProduct(t, shopID, "Aggregated", "on_shelf", []skuFixture{
		{spec: "black", priceCents: 2000, totalStock: 4, frozenStock: 1, enabled: true},
		{spec: "white", priceCents: 1000, totalStock: 3, enabled: true},
	})
	insertProduct(t, shopID, "Off shelf", "off_shelf", []skuFixture{
		{spec: "default", priceCents: 1200, totalStock: 10, enabled: true},
	})
	insertProduct(t, shopID, "Out of stock", "on_shelf", []skuFixture{
		{spec: "default", priceCents: 1200, totalStock: 0, enabled: true},
	})
	insertProduct(t, shopID, "Disabled sku", "on_shelf", []skuFixture{
		{spec: "default", priceCents: 1200, totalStock: 10, enabled: false},
	})
	insertProduct(t, shopID, "Only frozen", "on_shelf", []skuFixture{
		{spec: "default", priceCents: 1200, totalStock: 3, frozenStock: 3, enabled: true},
	})

	regularIDs := make([]int, 0, 8)
	for index := range 8 {
		regularIDs = append(regularIDs, insertProduct(
			t,
			shopID,
			"Regular product",
			"on_shelf",
			[]skuFixture{{
				spec: "default", priceCents: int64(3000 + index), totalStock: index + 1, enabled: true,
			}},
		))
	}

	feedRepository := repository.NewFeedRepository(testPool)
	firstPage, err := feedRepository.List(ctx, 7, 1, 5)
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if firstPage.Total != 9 {
		t.Fatalf("total = %d, want 9", firstPage.Total)
	}
	if len(firstPage.Items) != 5 {
		t.Fatalf("first page item count = %d, want 5", len(firstPage.Items))
	}

	secondPage, err := feedRepository.List(ctx, 7, 2, 5)
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage.Items) != 4 {
		t.Fatalf("second page item count = %d, want 4", len(secondPage.Items))
	}

	allItems := append(slices.Clone(firstPage.Items), secondPage.Items...)
	if ids := feedItemIDs(allItems); !containsAll(ids, append(regularIDs, aggregatedID)) {
		t.Fatalf("feed IDs = %v, want eligible products", ids)
	}

	aggregated := firstFeedItemByID(t, allItems, aggregatedID)
	if aggregated.MinPriceCents != 1000 || aggregated.MaxPriceCents != 2000 {
		t.Fatalf("aggregated price = %d/%d, want 1000/2000", aggregated.MinPriceCents, aggregated.MaxPriceCents)
	}

	repeated, err := feedRepository.List(ctx, 7, 1, 5)
	if err != nil {
		t.Fatalf("repeat first page: %v", err)
	}
	if !slices.Equal(feedItemIDs(firstPage.Items), feedItemIDs(repeated.Items)) {
		t.Fatalf("first page order changed for the same seed")
	}

	baseline := feedItemIDs(allItems)
	for seed := 8; seed <= 16; seed++ {
		other, err := feedRepository.List(ctx, seed, 1, 20)
		if err != nil {
			t.Fatalf("list with seed %d: %v", seed, err)
		}
		if !slices.Equal(baseline, feedItemIDs(other.Items)) {
			return
		}
	}
	t.Fatal("expected at least one user seed to produce a different feed order")
}

func TestFeedRepositoryGetReturnsEverySKU(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	shopID := insertShop(t, "feed-shop")
	productID := insertProduct(t, shopID, "Coffee", "on_shelf", []skuFixture{
		{spec: "500g", priceCents: 3499, totalStock: 3, enabled: true},
		{spec: "250g", priceCents: 1999, totalStock: 4, frozenStock: 1, enabled: true},
		{spec: "1kg", priceCents: 5999, totalStock: 5, enabled: false},
	})

	detail, err := repository.NewFeedRepository(testPool).Get(ctx, productID)
	if err != nil {
		t.Fatalf("get product detail: %v", err)
	}
	if detail.ID != productID ||
		detail.ShopID != shopID ||
		detail.ShopName != "Feed shop" ||
		detail.Name != "Coffee" {
		t.Fatalf("product detail = %+v", detail)
	}
	if len(detail.SKUs) != 3 {
		t.Fatalf("sku count = %d, want 3", len(detail.SKUs))
	}

	first := detail.SKUs[0]
	if first.Specs["variant"] != "250g" ||
		first.PriceCents != 1999 ||
		first.AvailableStock != 3 ||
		!first.IsEnabled {
		t.Fatalf("first sku = %+v", first)
	}
	if last := detail.SKUs[2]; last.Specs["variant"] != "1kg" || last.IsEnabled {
		t.Fatalf("last sku = %+v, want disabled 1kg", last)
	}
}

func TestFeedRepositoryGetRejectsUnavailableProduct(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	shopID := insertShop(t, "feed-shop")
	offShelfID := insertProduct(t, shopID, "Hidden", "off_shelf", []skuFixture{{
		spec: "default", priceCents: 1000, totalStock: 1, enabled: true,
	}})
	feedRepository := repository.NewFeedRepository(testPool)

	_, err := feedRepository.Get(ctx, offShelfID)
	if !errors.Is(err, apperr.ErrProductNotFound) {
		t.Fatalf("off-shelf error = %v, want %v", err, apperr.ErrProductNotFound)
	}
	_, err = feedRepository.Get(ctx, 999999)
	if !errors.Is(err, apperr.ErrProductNotFound) {
		t.Fatalf("missing product error = %v, want %v", err, apperr.ErrProductNotFound)
	}
}

func TestFeedRepositoryHidesBannedShopProducts(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	activeShopID := insertShop(t, "active-shop")
	bannedShopID := insertShop(t, "banned-shop")
	if _, err := testPool.Exec(
		ctx,
		`UPDATE shops SET status = 'banned' WHERE id = $1`,
		bannedShopID,
	); err != nil {
		t.Fatalf("ban shop: %v", err)
	}

	visibleID := insertProduct(t, activeShopID, "Visible", "on_shelf", []skuFixture{{
		spec: "default", priceCents: 1000, totalStock: 1, enabled: true,
	}})
	hiddenID := insertProduct(t, bannedShopID, "Hidden", "on_shelf", []skuFixture{{
		spec: "default", priceCents: 1000, totalStock: 1, enabled: true,
	}})
	feedRepository := repository.NewFeedRepository(testPool)

	feed, err := feedRepository.List(ctx, 7, 1, 20)
	if err != nil {
		t.Fatalf("list feed: %v", err)
	}
	if feed.Total != 1 || len(feed.Items) != 1 || feed.Items[0].ID != visibleID {
		t.Fatalf("feed = %+v, want only product %d", feed, visibleID)
	}

	_, err = feedRepository.Get(ctx, hiddenID)
	if !errors.Is(err, apperr.ErrProductNotFound) {
		t.Fatalf("banned shop detail error = %v, want %v", err, apperr.ErrProductNotFound)
	}
}

type skuFixture struct {
	spec        string
	priceCents  int64
	totalStock  int
	frozenStock int
	enabled     bool
}

func resetDatabase(t *testing.T) {
	t.Helper()

	if _, err := testPool.Exec(
		context.Background(),
		`TRUNCATE shops, products, product_skus, inventory_items RESTART IDENTITY CASCADE`,
	); err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

func insertShop(t *testing.T, username string) int {
	t.Helper()

	var shopID int
	err := testPool.QueryRow(
		context.Background(),
		`INSERT INTO shops (username, name, password_hash, email)
VALUES ($1, $2, $3, $4)
RETURNING id`,
		username,
		"Feed shop",
		"password-hash",
		username+"@example.com",
	).Scan(&shopID)
	if err != nil {
		t.Fatalf("insert shop: %v", err)
	}
	return shopID
}

func insertProduct(t *testing.T, shopID int, name string, status string, skus []skuFixture) int {
	t.Helper()
	ctx := context.Background()

	var productID int
	if err := testPool.QueryRow(
		ctx,
		`INSERT INTO products (shop_id, name, description, status)
VALUES ($1, $2, $3, $4)
RETURNING id`,
		shopID,
		name,
		name+" description",
		status,
	).Scan(&productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}

	for _, sku := range skus {
		var skuID int
		if err := testPool.QueryRow(
			ctx,
			`INSERT INTO product_skus (product_id, specs, price_cents, is_enabled)
VALUES ($1, jsonb_build_object('variant', $2::text), $3, $4)
RETURNING id`,
			productID,
			sku.spec,
			sku.priceCents,
			sku.enabled,
		).Scan(&skuID); err != nil {
			t.Fatalf("insert sku: %v", err)
		}

		if _, err := testPool.Exec(
			ctx,
			`INSERT INTO inventory_items (sku_id, total_stock, frozen_stock)
VALUES ($1, $2, $3)`,
			skuID,
			sku.totalStock,
			sku.frozenStock,
		); err != nil {
			t.Fatalf("insert inventory: %v", err)
		}
	}
	return productID
}

func feedItemIDs(items []model.FeedItem) []int {
	ids := make([]int, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func containsAll(ids []int, wanted []int) bool {
	for _, id := range wanted {
		if !slices.Contains(ids, id) {
			return false
		}
	}
	return true
}

func firstFeedItemByID(t *testing.T, items []model.FeedItem, id int) model.FeedItem {
	t.Helper()

	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("feed item %d not found", id)
	return model.FeedItem{}
}
