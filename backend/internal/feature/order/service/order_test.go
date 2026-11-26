package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/inventory/model"
	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/mocks"
	ordermodel "pinshaoshao/internal/feature/order/model"
	"pinshaoshao/internal/feature/order/paymentport"
	"pinshaoshao/internal/feature/order/service"
	"pinshaoshao/internal/outbox"

	"go.uber.org/mock/gomock"
)

const testCheckoutKey = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"

func newTestService(t *testing.T) (*service.OrderService, *mocks.MockOrderRepository) {
	t.Helper()

	orderService, repository, _, _, _ := newTestServiceWithDependencies(t)
	return orderService, repository
}

func newTestServiceWithDependencies(
	t *testing.T,
) (
	*service.OrderService,
	*mocks.MockOrderRepository,
	*fakeReservation,
	*fakePaymentCloser,
	*fakeOutboxWriter,
) {
	t.Helper()

	repository := mocks.NewMockOrderRepository(gomock.NewController(t))
	reservation := &fakeReservation{}
	paymentCloser := &fakePaymentCloser{}
	outboxWriter := &fakeOutboxWriter{}
	orderService := service.NewOrderService(
		repository,
		reservation,
		paymentCloser,
		outboxWriter,
		noopTransactor{},
	)
	return orderService, repository, reservation, paymentCloser, outboxWriter
}

func TestOrderServiceCreateNormalizesInput(t *testing.T) {
	order := &ordermodel.Order{ID: 1}
	orderService, repository := newTestService(t)

	repository.EXPECT().
		Create(gomock.Any(), 7, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int, input ordermodel.CreateOrderInput) (*ordermodel.Order, error) {
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
		Items: []ordermodel.CreateOrderItem{
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
				Items:           []ordermodel.CreateOrderItem{{CartItemID: 1, Quantity: 1}},
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
				Items: []ordermodel.CreateOrderItem{
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
				Items:           []ordermodel.CreateOrderItem{{CartItemID: 1, Quantity: 1}},
				ReceiverName:    "  ",
				ReceiverPhone:   "13800138000",
				ShippingAddress: "Shanghai",
			},
		},
		{
			name: "receiver name too long",
			input: service.CreateOrderInput{
				CheckoutKey:     testCheckoutKey,
				Items:           []ordermodel.CreateOrderItem{{CartItemID: 1, Quantity: 1}},
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

func TestOrderServiceCreateDirectNormalizesInput(t *testing.T) {
	order := &ordermodel.Order{ID: 1}
	orderService, repository := newTestService(t)

	repository.EXPECT().
		CreateDirect(gomock.Any(), 7, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int, input ordermodel.CreateDirectOrderInput) (*ordermodel.Order, error) {
			if input.CheckoutKey != testCheckoutKey {
				t.Fatalf("checkout key = %q, want %q", input.CheckoutKey, testCheckoutKey)
			}
			if len(input.RequestHash) != 64 {
				t.Fatalf("request hash length = %d, want 64", len(input.RequestHash))
			}
			if input.SKUID != 5 || input.Quantity != 2 {
				t.Fatalf("item = %d/%d, want 5/2", input.SKUID, input.Quantity)
			}
			if input.ReceiverName != "Alice" ||
				input.ReceiverPhone != "+8613800138000" ||
				input.ShippingAddress != "Shanghai" {
				t.Fatalf("receiver = %q/%q/%q, want trimmed values", input.ReceiverName, input.ReceiverPhone, input.ShippingAddress)
			}
			return order, nil
		})

	got, err := orderService.CreateDirect(context.Background(), 7, service.CreateDirectOrderInput{
		CheckoutKey:     strings.ToUpper(testCheckoutKey),
		SKUID:           5,
		Quantity:        2,
		ReceiverName:    "  Alice  ",
		ReceiverPhone:   "  +8613800138000 ",
		ShippingAddress: "  Shanghai ",
	})
	if err != nil {
		t.Fatalf("create direct order: %v", err)
	}
	if got != order {
		t.Fatalf("order = %p, want %p", got, order)
	}
}

func TestOrderServiceCreateDirectRejectsInvalidInput(t *testing.T) {
	valid := service.CreateDirectOrderInput{
		CheckoutKey:     testCheckoutKey,
		SKUID:           5,
		Quantity:        2,
		ReceiverName:    "Alice",
		ReceiverPhone:   "13800138000",
		ShippingAddress: "Shanghai",
	}
	tests := []struct {
		name   string
		mutate func(*service.CreateDirectOrderInput)
	}{
		{
			name: "invalid checkout key",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.CheckoutKey = "not-a-uuid"
			},
		},
		{
			name: "invalid sku",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.SKUID = 0
			},
		},
		{
			name: "invalid quantity",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.Quantity = 0
			},
		},
		{
			name: "quantity too large",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.Quantity = math.MaxInt32 + 1
			},
		},
		{
			name: "blank receiver",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.ReceiverName = "  "
			},
		},
		{
			name: "address too long",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.ShippingAddress = strings.Repeat("a", 1001)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			tt.mutate(&input)
			orderService, _ := newTestService(t)

			_, err := orderService.CreateDirect(context.Background(), 7, input)
			if !errors.Is(err, apperr.ErrInvalidCheckout) {
				t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidCheckout)
			}
		})
	}
}

func TestOrderServiceCreateDirectHashesNormalizedPayload(t *testing.T) {
	input := service.CreateDirectOrderInput{
		CheckoutKey:     testCheckoutKey,
		SKUID:           5,
		Quantity:        2,
		ReceiverName:    "Test Buyer",
		ReceiverPhone:   "1234567890",
		ShippingAddress: "Shanghai",
	}
	baseline := directOrderHash(t, input)

	normalized := input
	normalized.CheckoutKey = strings.ToUpper(normalized.CheckoutKey)
	normalized.ReceiverName = "  Test Buyer  "
	normalized.ReceiverPhone = " 1234567890 "
	normalized.ShippingAddress = " Shanghai "
	if got := directOrderHash(t, normalized); got != baseline {
		t.Fatalf("normalized request hash = %q, want %q", got, baseline)
	}

	tests := []struct {
		name   string
		mutate func(*service.CreateDirectOrderInput)
	}{
		{
			name: "sku",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.SKUID = 6
			},
		},
		{
			name: "quantity",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.Quantity = 3
			},
		},
		{
			name: "receiver name",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.ReceiverName = "Other Buyer"
			},
		},
		{
			name: "receiver phone",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.ReceiverPhone = "1234567891"
			},
		},
		{
			name: "shipping address",
			mutate: func(input *service.CreateDirectOrderInput) {
				input.ShippingAddress = "Beijing"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := input
			tt.mutate(&changed)
			if got := directOrderHash(t, changed); got == baseline {
				t.Fatalf("request hash did not change when %s changed", tt.name)
			}
		})
	}
}

func directOrderHash(t *testing.T, input service.CreateDirectOrderInput) string {
	t.Helper()

	orderService, repository := newTestService(t)
	var requestHash string
	repository.EXPECT().
		CreateDirect(gomock.Any(), 7, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int, stored ordermodel.CreateDirectOrderInput) (*ordermodel.Order, error) {
			requestHash = stored.RequestHash
			return &ordermodel.Order{ID: 1}, nil
		})

	if _, err := orderService.CreateDirect(context.Background(), 7, input); err != nil {
		t.Fatalf("create direct order: %v", err)
	}
	return requestHash
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
		List(gomock.Any(), 7, ordermodel.OrderListFilter{Page: 2, PageSize: 5}).
		Return(&ordermodel.OrderListResult{
			Orders:   []ordermodel.OrderSummary{{ID: 3}},
			Page:     2,
			PageSize: 5,
			Total:    6,
		}, nil)
	orders, err := orderService.List(context.Background(), 7, ordermodel.OrderListFilter{Page: 2, PageSize: 5})
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
		List(gomock.Any(), 7, ordermodel.OrderListFilter{Page: 1, PageSize: 10}).
		Return(&ordermodel.OrderListResult{Page: 1, PageSize: 10}, nil)
	if _, err := orderService.List(context.Background(), 7, ordermodel.OrderListFilter{}); err != nil {
		t.Fatalf("list orders with defaults: %v", err)
	}

	_, err := orderService.List(context.Background(), 7, ordermodel.OrderListFilter{PageSize: 101})
	if !errors.Is(err, apperr.ErrInvalidOrderList) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidOrderList)
	}
}

func TestOrderServiceCancel(t *testing.T) {
	orderService, repository, reservation, paymentCloser, outboxWriter :=
		newTestServiceWithDependencies(t)
	order := &ordermodel.Order{ID: 3, Status: ordermodel.OrderStatusCancelled}
	providerTransactionID := "provider-1"
	paymentCloser.result = &paymentport.CloseResult{
		PaymentID:             9,
		PaymentNo:             "payment-1",
		ProviderTransactionID: &providerTransactionID,
		Closed:                true,
	}

	repository.EXPECT().
		LockForCancel(gomock.Any(), gomock.Any(), 7, 3).
		Return(&ordermodel.CancelTarget{
			OrderID: 3,
			Status:  ordermodel.OrderStatusPendingPayment,
			Items: []ordermodel.ReservationItem{
				{SKUID: 5, Quantity: 2},
			},
		}, nil)
	repository.EXPECT().MarkCancelled(gomock.Any(), gomock.Any(), 3).Return(nil)
	repository.EXPECT().Load(gomock.Any(), gomock.Any(), 7, 3).Return(order, nil)

	got, err := orderService.Cancel(context.Background(), 7, 3)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if got != order {
		t.Fatalf("order = %p, want %p", got, order)
	}
	if len(reservation.released) != 1 ||
		len(reservation.released[0]) != 1 ||
		reservation.released[0][0].SKUID != 5 ||
		reservation.released[0][0].Quantity != 2 {
		t.Fatalf("released reservations = %+v", reservation.released)
	}
	if len(paymentCloser.orders) != 1 || paymentCloser.orders[0] != 3 {
		t.Fatalf("payment close order IDs = %v, want [3]", paymentCloser.orders)
	}
	if len(outboxWriter.events) != 1 {
		t.Fatalf("outbox events = %d, want 1", len(outboxWriter.events))
	}
	event := outboxWriter.events[0]
	if event.Key != "payment.close:payment-1" ||
		event.Topic != outbox.TopicPaymentCloseProvider ||
		event.AggregateType != "payment" ||
		event.AggregateID != 9 {
		t.Fatalf("outbox event = %+v", event)
	}
	var payload map[string]string
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		t.Fatalf("decode outbox payload: %v", err)
	}
	if payload["payment_no"] != "payment-1" ||
		payload["provider_transaction_id"] != providerTransactionID {
		t.Fatalf("outbox payload = %+v", payload)
	}
}

func TestOrderServiceCancelAlreadyCancelledClosesRemainingPayment(t *testing.T) {
	orderService, repository, reservation, paymentCloser, outboxWriter :=
		newTestServiceWithDependencies(t)
	providerTransactionID := "provider-1"
	paymentCloser.result = &paymentport.CloseResult{
		PaymentID:             9,
		PaymentNo:             "payment-1",
		ProviderTransactionID: &providerTransactionID,
		Closed:                true,
	}
	order := &ordermodel.Order{ID: 3, Status: ordermodel.OrderStatusCancelled}

	repository.EXPECT().
		LockForCancel(gomock.Any(), gomock.Any(), 7, 3).
		Return(&ordermodel.CancelTarget{
			OrderID: 3,
			Status:  ordermodel.OrderStatusCancelled,
			Items:   []ordermodel.ReservationItem{{SKUID: 5, Quantity: 2}},
		}, nil)
	repository.EXPECT().Load(gomock.Any(), gomock.Any(), 7, 3).Return(order, nil)

	got, err := orderService.Cancel(context.Background(), 7, 3)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if got != order {
		t.Fatalf("order = %p, want %p", got, order)
	}
	if len(reservation.released) != 0 {
		t.Fatalf("released reservations = %+v, want none", reservation.released)
	}
	if len(paymentCloser.orders) != 1 || len(outboxWriter.events) != 1 {
		t.Fatalf(
			"payment close calls/outbox events = %d/%d, want 1/1",
			len(paymentCloser.orders),
			len(outboxWriter.events),
		)
	}
}

func TestOrderServiceCancelRejectsSettledPayment(t *testing.T) {
	orderService, repository, _, paymentCloser, outboxWriter :=
		newTestServiceWithDependencies(t)
	paymentCloser.result = &paymentport.CloseResult{
		PaymentID:      9,
		PaymentNo:      "payment-1",
		AlreadySettled: true,
	}

	repository.EXPECT().
		LockForCancel(gomock.Any(), gomock.Any(), 7, 3).
		Return(&ordermodel.CancelTarget{
			OrderID: 3,
			Status:  ordermodel.OrderStatusPendingPayment,
		}, nil)

	_, err := orderService.Cancel(context.Background(), 7, 3)
	if !errors.Is(err, apperr.ErrOrderCannotCancel) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrOrderCannotCancel)
	}
	if len(outboxWriter.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxWriter.events))
	}
}

func TestOrderServiceCancelStopsWhenStockReleaseFails(t *testing.T) {
	orderService, repository, reservation, paymentCloser, outboxWriter :=
		newTestServiceWithDependencies(t)
	reservation.releaseErr = errors.New("release failed")

	repository.EXPECT().
		LockForCancel(gomock.Any(), gomock.Any(), 7, 3).
		Return(&ordermodel.CancelTarget{
			OrderID: 3,
			Status:  ordermodel.OrderStatusPendingPayment,
			Items:   []ordermodel.ReservationItem{{SKUID: 5, Quantity: 2}},
		}, nil)

	_, err := orderService.Cancel(context.Background(), 7, 3)
	if err == nil || !strings.Contains(err.Error(), "release failed") {
		t.Fatalf("error = %v, want release failure", err)
	}
	if len(paymentCloser.orders) != 0 || len(outboxWriter.events) != 0 {
		t.Fatalf(
			"payment close calls/outbox events = %d/%d, want 0/0",
			len(paymentCloser.orders),
			len(outboxWriter.events),
		)
	}
}

func TestOrderServiceCancelStopsWhenOutboxWriteFails(t *testing.T) {
	orderService, repository, _, paymentCloser, outboxWriter :=
		newTestServiceWithDependencies(t)
	providerTransactionID := "provider-1"
	paymentCloser.result = &paymentport.CloseResult{
		PaymentID:             9,
		PaymentNo:             "payment-1",
		ProviderTransactionID: &providerTransactionID,
		Closed:                true,
	}
	outboxWriter.err = errors.New("outbox unavailable")

	repository.EXPECT().
		LockForCancel(gomock.Any(), gomock.Any(), 7, 3).
		Return(&ordermodel.CancelTarget{
			OrderID: 3,
			Status:  ordermodel.OrderStatusPendingPayment,
			Items:   []ordermodel.ReservationItem{{SKUID: 5, Quantity: 2}},
		}, nil)

	_, err := orderService.Cancel(context.Background(), 7, 3)
	if err == nil || !strings.Contains(err.Error(), "outbox unavailable") {
		t.Fatalf("error = %v, want outbox failure", err)
	}
}

func TestOrderServiceExpireUsesDueOrderGuard(t *testing.T) {
	orderService, repository, _, _, _ := newTestServiceWithDependencies(t)

	repository.EXPECT().
		LockDueForCancel(gomock.Any(), gomock.Any(), 3).
		Return(nil, nil)
	if err := orderService.Expire(context.Background(), 3); err != nil {
		t.Fatalf("expire already settled order: %v", err)
	}

	repository.EXPECT().
		LockDueForCancel(gomock.Any(), gomock.Any(), 4).
		Return(&ordermodel.CancelTarget{
			OrderID: 4,
			Status:  ordermodel.OrderStatusPendingPayment,
			Items:   []ordermodel.ReservationItem{{SKUID: 5, Quantity: 1}},
		}, nil)
	repository.EXPECT().MarkCancelled(gomock.Any(), gomock.Any(), 4).Return(nil)
	if err := orderService.Expire(context.Background(), 4); err != nil {
		t.Fatalf("expire order: %v", err)
	}
}

func TestOrderServiceCancelForPaymentDelegates(t *testing.T) {
	orderService, repository, _, _, _ := newTestServiceWithDependencies(t)
	repository.EXPECT().
		LockForCancel(gomock.Any(), gomock.Any(), 7, 3).
		Return(&ordermodel.CancelTarget{
			OrderID: 3,
			Status:  ordermodel.OrderStatusCancelled,
		}, nil)
	repository.EXPECT().Load(gomock.Any(), gomock.Any(), 7, 3).
		Return(&ordermodel.Order{ID: 3, Status: ordermodel.OrderStatusCancelled}, nil)

	if err := orderService.CancelForPayment(context.Background(), 7, 3); err != nil {
		t.Fatalf("cancel for payment: %v", err)
	}
}

func TestOrderServiceCancelRejectsInvalidOrderID(t *testing.T) {
	orderService, _ := newTestService(t)

	_, err := orderService.Cancel(context.Background(), 7, 0)
	if !errors.Is(err, apperr.ErrInvalidOrderID) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidOrderID)
	}
}

type noopTransactor struct{}

func (noopTransactor) WithinTx(ctx context.Context, fn func(context.Context, database.DBTX) error) error {
	return fn(ctx, nil)
}

type fakeReservation struct {
	releaseErr error
	released   [][]model.ReservationItem
}

func (f *fakeReservation) Freeze(context.Context, database.DBTX, []model.ReservationItem) error {
	return nil
}

func (f *fakeReservation) Release(
	_ context.Context,
	_ database.DBTX,
	items []model.ReservationItem,
) error {
	f.released = append(f.released, items)
	return f.releaseErr
}

func (f *fakeReservation) Commit(context.Context, database.DBTX, []model.ReservationItem) error {
	return nil
}

type fakePaymentCloser struct {
	result *paymentport.CloseResult
	err    error
	orders []int
}

func (f *fakePaymentCloser) CloseForOrder(
	_ context.Context,
	_ database.DBTX,
	orderID int,
) (*paymentport.CloseResult, error) {
	f.orders = append(f.orders, orderID)
	return f.result, f.err
}

type fakeOutboxWriter struct {
	events []outbox.Event
	err    error
}

func (f *fakeOutboxWriter) Enqueue(_ context.Context, _ database.DBTX, event outbox.Event) error {
	if f.err != nil {
		return f.err
	}
	f.events = append(f.events, event)
	return nil
}
