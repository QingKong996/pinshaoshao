package service_test

import (
	"context"
	"errors"
	"testing"

	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/service"

	"go.uber.org/mock/gomock"
)

func newTestDashboardService(t *testing.T) (*service.DashboardService, *mocks.MockDashboardRepository) {
	t.Helper()

	repository := mocks.NewMockDashboardRepository(gomock.NewController(t))
	return service.NewDashboardService(repository), repository
}

func TestDashboardOverview(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the stats", func(t *testing.T) {
		dashboardService, repository := newTestDashboardService(t)
		stats := &model.DashboardStats{
			ProductCount:        12,
			OnShelfProductCount: 8,
			SKUCount:            34,
			CategoryCount:       4,
			TotalStock:          500,
			InventoryValueCents: 123456,
			OutOfStockSKUCount:  1,
		}

		repository.EXPECT().GetDashboardStats(ctx, 1).Return(stats, nil)

		got, err := dashboardService.Overview(ctx, 1)
		if err != nil {
			t.Fatalf("get dashboard: %v", err)
		}
		if got.ProductCount != 12 || got.TotalStock != 500 || got.OutOfStockSKUCount != 1 {
			t.Fatalf("stats = %+v, want repository counters", got)
		}
	})

	t.Run("repository failure", func(t *testing.T) {
		dashboardService, repository := newTestDashboardService(t)
		dbErr := errors.New("database is down")

		repository.EXPECT().GetDashboardStats(ctx, 1).Return(nil, dbErr)

		_, err := dashboardService.Overview(ctx, 1)
		requireErrorIs(t, err, dbErr)
	})
}
