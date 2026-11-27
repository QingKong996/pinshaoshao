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

	addressrepository "pinshaoshao/internal/feature/address/repository"
	usermodel "pinshaoshao/internal/feature/user/model"
	userrepository "pinshaoshao/internal/feature/user/repository"

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
		postgres.WithDatabase("pinshaoshao_address_test"),
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

func resetAddressDatabase(t *testing.T) {
	t.Helper()

	if _, err := testPool.Exec(
		context.Background(),
		`TRUNCATE users, sessions, user_addresses RESTART IDENTITY CASCADE`,
	); err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

func insertAddressUser(t *testing.T, username string) *usermodel.User {
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

func newTestAddressRepository() *addressrepository.AddressRepository {
	return addressrepository.NewAddressRepository(testPool)
}

func requireAddressErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want %v", err, target)
	}
}
