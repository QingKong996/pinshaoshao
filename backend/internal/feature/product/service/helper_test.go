package service_test

import (
	"context"
	"errors"
	"testing"

	"pinshaoshao/internal/database"
)

type testTransactor struct{}

func (testTransactor) WithinTx(ctx context.Context, fn func(context.Context, database.DBTX) error) error {
	return fn(ctx, nil)
}

type testInventoryLifecycle struct {
	createErr        error
	deleteSKUErr     error
	deleteProductErr error
}

func (s testInventoryLifecycle) CreateEmptyForSKU(context.Context, database.DBTX, int) error {
	return s.createErr
}

func (s testInventoryLifecycle) DeleteEmptyForSKU(context.Context, database.DBTX, int, int) error {
	return s.deleteSKUErr
}

func (s testInventoryLifecycle) DeleteEmptyForProduct(context.Context, database.DBTX, int) error {
	return s.deleteProductErr
}

func requireErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want %v", err, target)
	}
}
