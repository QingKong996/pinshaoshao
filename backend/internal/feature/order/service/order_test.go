package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/mocks"
	"pinshaoshao/internal/feature/order/model"
	"pinshaoshao/internal/feature/order/service"

	"go.uber.org/mock/gomock"
)

const testCheckoutKey = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"

func newTestService(t *testing.T) (*service.OrderService, *mocks.MockOrderRepository) {
	t.Helper()

	repository := mocks.NewMockOrderRepository(gomock.NewController(t))
	return service.NewOrderService(repository), repository
}

func TestOrderServiceCreateNormalizesInput(t *testing.T) {
	order := &model.Order{ID: 1}
	orderService, repository := newTestService(t)

	repository.EXPECT().
		Create(gomock.Any(), 7, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int, input model.CreateOrderInput) (*model.Order, error) {
			if input.CheckoutKey != testCheckoutKey {
				t.Fatalf("checkout key = %q, want %q", input.CheckoutKey, testCheckoutKey)
			}
			if len(input.RequestHash) != 64 {
				t.Fatalf("request hash length = %d, want 64", len(input.RequestHash))
			}
			if input.Items[0].CartItemID != 1 || input.Items[1].CartItemID != 2 {
				t.Fatalf("items = %+v, want sorted cart item IDs", input.Items)
			}
			if input.ReceiverName != "Alice" || input.ReceiverPhone != "+8613800138000" || input.ShippingAddress != "Shanghai" {
				t.Fatalf("receiver = %q/%q/%q, want trimmed values", input.ReceiverName, input.ReceiverPhone, input.ShippingAddress)
			}
			return order, nil
		})

	got, err := orderService.Create(context.Background(), 7, service.CreateOrderInput{
		CheckoutKey: strings.ToUpper(testCheckoutKey),
		Items: []model.CreateOrderItem{
			{CartItemID: 2, Quantity: 1},
			{CartItemID: 1, Quantity: 2},
		},
		ReceiverName:    "  Alice  ",
		ReceiverPhone:   "  +8613800138000 ",
		ShippingAddress: "  Shanghai ",
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if got != order {
		t.Fatalf("order = %p, want %p", got, order)
	}
}

func TestOrderServiceCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input service.CreateOrderInput
	}{
		{
			name: "invalid checkout key",
			input: service.CreateOrderInput{
				CheckoutKey:     "not-a-uuid",
				Items:           []model.CreateOrderItem{{CartItemID: 1, Quantity: 1}},
				ReceiverName:    "Alice",
				ReceiverPhone:   "13800138000",
				ShippingAddress: "Shanghai",
			},
		},
		{
			name: "no items",
			input: service.CreateOrderInput{
				CheckoutKey:     testCheckoutKey,
				ReceiverName:    "Alice",
				ReceiverPhone:   "13800138000",
				ShippingAddress: "Shanghai",
			},
		},
		{
			name: "duplicate cart item",
			input: service.CreateOrderInput{
				CheckoutKey: testCheckoutKey,
				Items: []model.CreateOrderItem{
					{CartItemID: 1, Quantity: 1},
					{CartItemID: 1, Quantity: 2},
				},
				ReceiverName:    "Alice",
				ReceiverPhone:   "13800138000",
				ShippingAddress: "Shanghai",
			},
		},
		{
			name: "blank receiver",
			input: service.CreateOrderInput{
				CheckoutKey:     testCheckoutKey,
				Items:           []model.CreateOrderItem{{CartItemID: 1, Quantity: 1}},
				ReceiverName:    "  ",
				ReceiverPhone:   "13800138000",
				ShippingAddress: "Shanghai",
			},
		},
		{
			name: "receiver name too long",
			input: service.CreateOrderInput{
				CheckoutKey:     testCheckoutKey,
				Items:           []model.CreateOrderItem{{CartItemID: 1, Quantity: 1}},
				ReceiverName:    strings.Repeat("a", 51),
				ReceiverPhone:   "13800138000",
				ShippingAddress: "Shanghai",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderService, _ := newTestService(t)
			_, err := orderService.Create(context.Background(), 7, tt.input)
			if !errors.Is(err, apperr.ErrInvalidCheckout) {
				t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidCheckout)
			}
		})
	}
}

func TestOrderServiceGetRejectsInvalidID(t *testing.T) {
	orderService, _ := newTestService(t)

	_, err := orderService.Get(context.Background(), 7, 0)
	if !errors.Is(err, apperr.ErrInvalidOrderID) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidOrderID)
	}
}

func TestOrderServiceListNormalizesPagination(t *testing.T) {
	orderService, repository := newTestService(t)

	repository.EXPECT().
		List(gomock.Any(), 7, model.OrderListFilter{Page: 2, PageSize: 5}).
		Return(&model.OrderListResult{
			Orders:   []model.OrderSummary{{ID: 3}},
			Page:     2,
			PageSize: 5,
			Total:    6,
		}, nil)
	orders, err := orderService.List(context.Background(), 7, model.OrderListFilter{Page: 2, PageSize: 5})
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if len(orders.Orders) != 1 || orders.Orders[0].ID != 3 || orders.Total != 6 {
		t.Fatalf("orders = %+v, want order 3", orders)
	}
}

func TestOrderServiceListDefaultsAndRejectsInvalidPagination(t *testing.T) {
	orderService, repository := newTestService(t)

	repository.EXPECT().
		List(gomock.Any(), 7, model.OrderListFilter{Page: 1, PageSize: 10}).
		Return(&model.OrderListResult{Page: 1, PageSize: 10}, nil)
	if _, err := orderService.List(context.Background(), 7, model.OrderListFilter{}); err != nil {
		t.Fatalf("list orders with defaults: %v", err)
	}

	_, err := orderService.List(context.Background(), 7, model.OrderListFilter{PageSize: 101})
	if !errors.Is(err, apperr.ErrInvalidOrderList) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidOrderList)
	}
}

func TestOrderServiceCancel(t *testing.T) {
	orderService, repository := newTestService(t)
	order := &model.Order{ID: 3}

	repository.EXPECT().Cancel(gomock.Any(), 7, 3).Return(order, nil)
	got, err := orderService.Cancel(context.Background(), 7, 3)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if got != order {
		t.Fatalf("order = %p, want %p", got, order)
	}
}
