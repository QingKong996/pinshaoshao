package service_test

import (
	"context"
	"errors"
	"testing"

	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/model"

	"go.uber.org/mock/gomock"
)

func TestOrderServiceListMerchantNormalizesFilter(t *testing.T) {
	orderService, repository := newTestService(t)

	repository.EXPECT().
		ListMerchant(gomock.Any(), 9, model.MerchantOrderListFilter{
			Status:   model.SubOrderStatusAwaitingShipment,
			Page:     2,
			PageSize: 5,
		}).
		Return(&model.MerchantOrderListResult{
			Orders:   []model.MerchantOrderSummary{{ID: 4}},
			Page:     2,
			PageSize: 5,
			Total:    6,
		}, nil)

	orders, err := orderService.ListMerchant(context.Background(), 9, model.MerchantOrderListFilter{
		Status:   model.SubOrderStatusAwaitingShipment,
		Page:     2,
		PageSize: 5,
	})
	if err != nil {
		t.Fatalf("list merchant orders: %v", err)
	}
	if len(orders.Orders) != 1 || orders.Orders[0].ID != 4 || orders.Total != 6 {
		t.Fatalf("merchant orders = %+v, want order 4", orders)
	}
}

func TestOrderServiceListMerchantDefaultsPagination(t *testing.T) {
	orderService, repository := newTestService(t)

	repository.EXPECT().
		ListMerchant(gomock.Any(), 9, model.MerchantOrderListFilter{
			Page:     1,
			PageSize: 10,
		}).
		Return(&model.MerchantOrderListResult{
			Orders:   []model.MerchantOrderSummary{},
			Page:     1,
			PageSize: 10,
		}, nil)

	if _, err := orderService.ListMerchant(
		context.Background(),
		9,
		model.MerchantOrderListFilter{},
	); err != nil {
		t.Fatalf("list merchant orders with defaults: %v", err)
	}
}

func TestOrderServiceListMerchantRejectsInvalidFilter(t *testing.T) {
	orderService, _ := newTestService(t)

	_, err := orderService.ListMerchant(context.Background(), 9, model.MerchantOrderListFilter{
		Status: "unknown",
	})
	if !errors.Is(err, apperr.ErrInvalidOrderList) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidOrderList)
	}
}

func TestOrderServiceGetMerchant(t *testing.T) {
	orderService, repository := newTestService(t)
	detail := &model.MerchantOrderDetail{
		ID:      4,
		OrderID: 12,
		Status:  model.SubOrderStatusAwaitingShipment,
		Items:   []model.OrderItem{{ID: 8, ProductName: "T-shirt", Quantity: 2}},
	}

	repository.EXPECT().
		GetMerchant(gomock.Any(), 9, 4).
		Return(detail, nil)

	got, err := orderService.GetMerchant(context.Background(), 9, 4)
	if err != nil {
		t.Fatalf("get merchant order: %v", err)
	}
	if got != detail {
		t.Fatalf("merchant order = %+v, want %+v", got, detail)
	}
}

func TestOrderServiceGetMerchantRejectsInvalidID(t *testing.T) {
	orderService, _ := newTestService(t)

	_, err := orderService.GetMerchant(context.Background(), 9, 0)
	if !errors.Is(err, apperr.ErrInvalidOrderID) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidOrderID)
	}
}

func TestOrderServiceGetMerchantReceiver(t *testing.T) {
	orderService, repository := newTestService(t)
	receiver := &model.MerchantOrderReceiver{
		ReceiverName:    "Alice",
		ReceiverPhone:   "1234567890",
		ShippingAddress: "Shanghai",
	}

	repository.EXPECT().
		GetMerchantReceiver(gomock.Any(), 9, 4).
		Return(receiver, nil)

	got, err := orderService.GetMerchantReceiver(context.Background(), 9, 4)
	if err != nil {
		t.Fatalf("get merchant receiver: %v", err)
	}
	if got != receiver {
		t.Fatalf("receiver = %+v, want %+v", got, receiver)
	}
}

func TestOrderServiceGetMerchantReceiverRejectsInvalidID(t *testing.T) {
	orderService, _ := newTestService(t)

	_, err := orderService.GetMerchantReceiver(context.Background(), 9, 0)
	if !errors.Is(err, apperr.ErrInvalidOrderID) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidOrderID)
	}
}
