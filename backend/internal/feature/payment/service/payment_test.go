package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"pinshaoshao/internal/feature/order/paymentport"
	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/model"
	"pinshaoshao/internal/feature/payment/service"

	"github.com/google/uuid"
)

const (
	testPaymentNo     = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"
	testProviderNo    = "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98"
	testEventID       = "a1c6cfe8-e77b-4ac2-b8bf-c6a0f56673f1"
	testWebhookTarget = "http://localhost:8080/api/v1/payment/webhook"
)

type repositoryStub struct {
	createForOrder         func(context.Context, int, int, string) (*model.Payment, bool, error)
	getForUser             func(context.Context, int, string) (*model.Payment, error)
	getForPayment          func(context.Context, int, string) (*model.Payment, error)
	setProviderTransaction func(context.Context, string, string) error
	applyPaidCallback      func(context.Context, model.PaidCallback) (*model.Payment, error)
}

func (s *repositoryStub) CreateForOrder(
	ctx context.Context,
	userID int,
	orderID int,
	paymentNo string,
) (*model.Payment, bool, error) {
	return s.createForOrder(ctx, userID, orderID, paymentNo)
}

func (s *repositoryStub) GetForUser(
	ctx context.Context,
	userID int,
	paymentNo string,
) (*model.Payment, error) {
	return s.getForUser(ctx, userID, paymentNo)
}

func (s *repositoryStub) GetForPayment(
	ctx context.Context,
	userID int,
	paymentNo string,
) (*model.Payment, error) {
	return s.getForPayment(ctx, userID, paymentNo)
}

func (s *repositoryStub) SetProviderTransaction(
	ctx context.Context,
	paymentNo string,
	providerTransactionID string,
) error {
	return s.setProviderTransaction(ctx, paymentNo, providerTransactionID)
}

func (s *repositoryStub) ApplyPaidCallback(
	ctx context.Context,
	callback model.PaidCallback,
) (*model.Payment, error) {
	return s.applyPaidCallback(ctx, callback)
}

type gatewayStub struct {
	create       func(context.Context, model.ProviderPaymentRequest) (*model.ProviderPaymentResult, error)
	closePayment func(context.Context, string) error
}

func (s gatewayStub) CreatePayment(
	ctx context.Context,
	input model.ProviderPaymentRequest,
) (*model.ProviderPaymentResult, error) {
	return s.create(ctx, input)
}

func (s gatewayStub) ClosePayment(ctx context.Context, providerPaymentNo string) error {
	return s.closePayment(ctx, providerPaymentNo)
}

type orderCancellerStub struct {
	cancel func(context.Context, int, int) error
}

func (s orderCancellerStub) CancelForPayment(ctx context.Context, userID int, orderID int) error {
	return s.cancel(ctx, userID, orderID)
}

type orderExpirerStub struct {
	expire func(context.Context, int) error
}

func (s orderExpirerStub) Expire(ctx context.Context, orderID int) error {
	return s.expire(ctx, orderID)
}

func TestPaymentServiceCreateUsesSinglePaymentPerOrder(t *testing.T) {
	repository := &repositoryStub{}
	payment := &model.Payment{PaymentNo: testPaymentNo, OrderID: 7}
	repository.createForOrder = func(
		_ context.Context,
		userID int,
		orderID int,
		paymentNo string,
	) (*model.Payment, bool, error) {
		if userID != 3 || orderID != 7 {
			t.Fatalf("arguments = %d/%d, want 3/7", userID, orderID)
		}
		if _, err := uuid.Parse(paymentNo); err != nil {
			t.Fatalf("payment number = %q: %v", paymentNo, err)
		}
		return payment, true, nil
	}
	paymentService := service.NewPaymentService(repository, service.Config{})

	got, created, err := paymentService.Create(context.Background(), 3, 7)
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if !created || got != payment {
		t.Fatalf("result = %p/%t, want %p/true", got, created, payment)
	}
}

func TestPaymentServicePayCreatesIdempotentProviderSession(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(20 * time.Minute)
	payment := &model.Payment{
		PaymentNo:      testPaymentNo,
		OrderID:        7,
		Status:         model.PaymentStatusPending,
		AmountCents:    3998,
		OrderExpiresAt: expiresAt,
		CreatedAt:      now.Add(-10 * time.Minute),
	}
	updated := &model.Payment{
		PaymentNo:             testPaymentNo,
		OrderID:               7,
		Status:                model.PaymentStatusPending,
		AmountCents:           3998,
		ProviderTransactionID: stringPointer(testProviderNo),
	}
	getCalls := 0
	repository := &repositoryStub{
		getForPayment: func(_ context.Context, userID int, paymentNo string) (*model.Payment, error) {
			if userID != 3 || paymentNo != testPaymentNo {
				t.Fatalf("get payment arguments = %d/%q", userID, paymentNo)
			}
			getCalls++
			if getCalls == 1 {
				return payment, nil
			}
			return updated, nil
		},
		setProviderTransaction: func(_ context.Context, paymentNo string, providerNo string) error {
			if paymentNo != testPaymentNo || providerNo != testProviderNo {
				t.Fatalf("provider transaction = %q/%q", paymentNo, providerNo)
			}
			return nil
		},
	}
	var gatewayInput model.ProviderPaymentRequest
	gateway := gatewayStub{create: func(
		_ context.Context,
		input model.ProviderPaymentRequest,
	) (*model.ProviderPaymentResult, error) {
		gatewayInput = input
		return &model.ProviderPaymentResult{
			PaymentNo:   testProviderNo,
			MerchantNo:  testPaymentNo,
			Status:      "pending",
			AmountCents: 3998,
			Currency:    model.CurrencyCNY,
			CheckoutURL: "http://localhost:8080/mockpay/checkout/" + testProviderNo,
		}, nil
	}}
	paymentService := service.NewPaymentService(repository, service.Config{
		Gateway:       gateway,
		NotifyURL:     testWebhookTarget,
		ReturnURLBase: "http://localhost:5173",
	})

	session, err := paymentService.Pay(context.Background(), 3, testPaymentNo)
	if err != nil {
		t.Fatalf("pay payment: %v", err)
	}
	if gatewayInput.MerchantOrderNo != testPaymentNo ||
		gatewayInput.AmountCents != 3998 ||
		gatewayInput.NotifyURL != testWebhookTarget ||
		gatewayInput.ReturnURL != "http://localhost:5173/orders/7" {
		t.Fatalf("gateway input = %+v", gatewayInput)
	}
	if gatewayInput.ExpiresIn <= 19*time.Minute || gatewayInput.ExpiresIn > 20*time.Minute {
		t.Fatalf("gateway expiry = %s, want remaining order window", gatewayInput.ExpiresIn)
	}
	if session.Payment.Status != model.PaymentStatusPending || session.CheckoutURL == "" {
		t.Fatalf("session = %+v", session)
	}
}

func TestPaymentServicePayRejectsMismatchedProviderResponse(t *testing.T) {
	payment := &model.Payment{
		PaymentNo:      testPaymentNo,
		OrderID:        7,
		Status:         model.PaymentStatusPending,
		AmountCents:    3998,
		OrderExpiresAt: time.Now().Add(20 * time.Minute),
	}
	repository := &repositoryStub{
		getForPayment: func(context.Context, int, string) (*model.Payment, error) {
			return payment, nil
		},
		setProviderTransaction: func(context.Context, string, string) error {
			t.Fatal("provider transaction must not be stored for an invalid response")
			return nil
		},
	}
	gateway := gatewayStub{create: func(
		context.Context,
		model.ProviderPaymentRequest,
	) (*model.ProviderPaymentResult, error) {
		return &model.ProviderPaymentResult{
			PaymentNo:   testProviderNo,
			MerchantNo:  testPaymentNo,
			Status:      "pending",
			AmountCents: 1,
			Currency:    model.CurrencyCNY,
			CheckoutURL: "http://localhost:8080/mockpay/checkout/" + testProviderNo,
		}, nil
	}}
	paymentService := service.NewPaymentService(repository, service.Config{Gateway: gateway})

	_, err := paymentService.Pay(context.Background(), 3, testPaymentNo)
	if !errors.Is(err, apperr.ErrPaymentProviderInvalid) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrPaymentProviderInvalid)
	}
}

func TestPaymentServicePayClosesProviderSessionWhenLocalPaymentWasCancelled(t *testing.T) {
	payment := &model.Payment{
		PaymentNo:      testPaymentNo,
		OrderID:        7,
		Status:         model.PaymentStatusPending,
		AmountCents:    3998,
		OrderExpiresAt: time.Now().Add(20 * time.Minute),
	}
	repository := &repositoryStub{
		getForPayment: func(context.Context, int, string) (*model.Payment, error) {
			return payment, nil
		},
		setProviderTransaction: func(context.Context, string, string) error {
			return apperr.ErrPaymentClosed
		},
	}
	var closedProviderNo string
	gateway := gatewayStub{
		create: func(
			context.Context,
			model.ProviderPaymentRequest,
		) (*model.ProviderPaymentResult, error) {
			return &model.ProviderPaymentResult{
				PaymentNo:   testProviderNo,
				MerchantNo:  testPaymentNo,
				Status:      "pending",
				AmountCents: 3998,
				Currency:    model.CurrencyCNY,
				CheckoutURL: "http://localhost:8080/mockpay/checkout/" + testProviderNo,
			}, nil
		},
		closePayment: func(_ context.Context, providerNo string) error {
			closedProviderNo = providerNo
			return nil
		},
	}
	paymentService := service.NewPaymentService(repository, service.Config{Gateway: gateway})

	_, err := paymentService.Pay(context.Background(), 3, testPaymentNo)
	if !errors.Is(err, apperr.ErrPaymentClosed) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrPaymentClosed)
	}
	if closedProviderNo != testProviderNo {
		t.Fatalf("closed provider payment = %q, want %q", closedProviderNo, testProviderNo)
	}
}

func TestPaymentServicePayRejectsOrderWithinProviderMinimumTTL(t *testing.T) {
	payment := &model.Payment{
		PaymentNo:      testPaymentNo,
		OrderID:        7,
		Status:         model.PaymentStatusPending,
		AmountCents:    3998,
		OrderExpiresAt: time.Now().Add(30 * time.Second),
	}
	repository := &repositoryStub{
		getForPayment: func(context.Context, int, string) (*model.Payment, error) {
			return payment, nil
		},
	}
	gateway := gatewayStub{create: func(
		context.Context,
		model.ProviderPaymentRequest,
	) (*model.ProviderPaymentResult, error) {
		t.Fatal("provider must not be called for an order too close to expiry")
		return nil, nil
	}}
	paymentService := service.NewPaymentService(repository, service.Config{Gateway: gateway})

	_, err := paymentService.Pay(context.Background(), 3, testPaymentNo)
	if !errors.Is(err, apperr.ErrOrderNotPayable) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrOrderNotPayable)
	}
}

func TestPaymentServicePayAcceptsPaidProviderWithoutCheckoutURL(t *testing.T) {
	payment := &model.Payment{
		PaymentNo:      testPaymentNo,
		OrderID:        7,
		Status:         model.PaymentStatusPending,
		AmountCents:    3998,
		OrderExpiresAt: time.Now().Add(20 * time.Minute),
	}
	paid := &model.Payment{
		PaymentNo:             testPaymentNo,
		OrderID:               7,
		Status:                model.PaymentStatusSucceeded,
		AmountCents:           3998,
		ProviderTransactionID: stringPointer(testProviderNo),
	}
	getCalls := 0
	repository := &repositoryStub{
		getForPayment: func(context.Context, int, string) (*model.Payment, error) {
			getCalls++
			if getCalls == 1 {
				return payment, nil
			}
			return paid, nil
		},
		setProviderTransaction: func(context.Context, string, string) error {
			return nil
		},
	}
	gateway := gatewayStub{create: func(
		context.Context,
		model.ProviderPaymentRequest,
	) (*model.ProviderPaymentResult, error) {
		return &model.ProviderPaymentResult{
			PaymentNo:   testProviderNo,
			MerchantNo:  testPaymentNo,
			Status:      "paid",
			AmountCents: 3998,
			Currency:    model.CurrencyCNY,
		}, nil
	}}
	paymentService := service.NewPaymentService(repository, service.Config{Gateway: gateway})

	session, err := paymentService.Pay(context.Background(), 3, testPaymentNo)
	if err != nil {
		t.Fatalf("pay payment: %v", err)
	}
	if session.Payment != paid || session.CheckoutURL != "" {
		t.Fatalf("session = %+v, want paid payment with no checkout URL", session)
	}
}

func TestPaymentServiceCloseCancelsOrderAndReloadsPayment(t *testing.T) {
	payment := &model.Payment{
		PaymentNo: testPaymentNo,
		OrderID:   7,
		Status:    model.PaymentStatusPending,
	}
	closed := &model.Payment{
		PaymentNo: testPaymentNo,
		OrderID:   7,
		Status:    model.PaymentStatusClosed,
	}
	getCalls := 0
	repository := &repositoryStub{
		getForUser: func(_ context.Context, userID int, paymentNo string) (*model.Payment, error) {
			if userID != 3 || paymentNo != testPaymentNo {
				t.Fatalf("get payment arguments = %d/%q", userID, paymentNo)
			}
			getCalls++
			if getCalls == 1 {
				return payment, nil
			}
			return closed, nil
		},
	}
	var cancelledUserID, cancelledOrderID int
	paymentService := service.NewPaymentService(repository, service.Config{
		OrderCanceller: orderCancellerStub{cancel: func(
			_ context.Context,
			userID int,
			orderID int,
		) error {
			cancelledUserID = userID
			cancelledOrderID = orderID
			return nil
		}},
	})

	got, err := paymentService.Close(context.Background(), 3, testPaymentNo)
	if err != nil {
		t.Fatalf("close payment: %v", err)
	}
	if got != closed || cancelledUserID != 3 || cancelledOrderID != 7 {
		t.Fatalf(
			"result/order = %+v/%d/%d, want %+v/3/7",
			got,
			cancelledUserID,
			cancelledOrderID,
			closed,
		)
	}
}

func TestPaymentServiceCloseRequiresOrderCanceller(t *testing.T) {
	repository := &repositoryStub{
		getForUser: func(context.Context, int, string) (*model.Payment, error) {
			return &model.Payment{
				PaymentNo: testPaymentNo,
				OrderID:   7,
				Status:    model.PaymentStatusPending,
			}, nil
		},
	}
	paymentService := service.NewPaymentService(repository, service.Config{})

	_, err := paymentService.Close(context.Background(), 3, testPaymentNo)
	if !errors.Is(err, apperr.ErrInternalServer) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrInternalServer)
	}
}

func TestPaymentServiceCloseIsIdempotentWhenAlreadyClosed(t *testing.T) {
	closed := &model.Payment{PaymentNo: testPaymentNo, Status: model.PaymentStatusClosed}
	repository := &repositoryStub{
		getForUser: func(context.Context, int, string) (*model.Payment, error) {
			return closed, nil
		},
	}
	paymentService := service.NewPaymentService(repository, service.Config{
		OrderCanceller: orderCancellerStub{cancel: func(context.Context, int, int) error {
			t.Fatal("already closed payment must not cancel its order")
			return nil
		}},
	})

	got, err := paymentService.Close(context.Background(), 3, testPaymentNo)
	if err != nil {
		t.Fatalf("close payment: %v", err)
	}
	if got != closed {
		t.Fatalf("result = %+v, want %+v", got, closed)
	}
}

func TestPaymentServiceCloseRejectsPaidPayment(t *testing.T) {
	repository := &repositoryStub{
		getForUser: func(context.Context, int, string) (*model.Payment, error) {
			return &model.Payment{
				PaymentNo: testPaymentNo,
				Status:    model.PaymentStatusSucceeded,
			}, nil
		},
	}
	paymentService := service.NewPaymentService(repository, service.Config{
		OrderCanceller: orderCancellerStub{cancel: func(context.Context, int, int) error {
			t.Fatal("paid payment must not cancel its order")
			return nil
		}},
	})

	_, err := paymentService.Close(context.Background(), 3, testPaymentNo)
	if !errors.Is(err, apperr.ErrPaymentConflict) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrPaymentConflict)
	}
}

func TestPaymentServiceCloseReturnsOrderCancellationError(t *testing.T) {
	providerNo := testProviderNo
	repository := &repositoryStub{
		getForUser: func(context.Context, int, string) (*model.Payment, error) {
			return &model.Payment{
				PaymentNo:             testPaymentNo,
				Status:                model.PaymentStatusPending,
				ProviderTransactionID: &providerNo,
			}, nil
		},
	}
	paymentService := service.NewPaymentService(repository, service.Config{
		OrderCanceller: orderCancellerStub{cancel: func(context.Context, int, int) error {
			return apperr.ErrPaymentConflict
		}},
	})

	_, err := paymentService.Close(context.Background(), 3, testPaymentNo)
	if !errors.Is(err, apperr.ErrPaymentConflict) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrPaymentConflict)
	}
}

func TestPaymentServiceHandlePaidCallbackNormalizesProviderEvent(t *testing.T) {
	paidAt := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	paid := &model.Payment{PaymentNo: testPaymentNo, Status: model.PaymentStatusSucceeded}
	var received model.PaidCallback
	repository := &repositoryStub{
		applyPaidCallback: func(_ context.Context, callback model.PaidCallback) (*model.Payment, error) {
			received = callback
			return paid, nil
		},
	}
	paymentService := service.NewPaymentService(repository, service.Config{})

	got, err := paymentService.HandlePaidCallback(context.Background(), model.PaidCallback{
		EventID:               testEventID,
		EventType:             "payment.paid",
		ProviderTransactionID: testProviderNo,
		PaymentNo:             testPaymentNo,
		AmountCents:           3998,
		Currency:              "cny",
		PaidAt:                paidAt,
	})
	if err != nil {
		t.Fatalf("handle callback: %v", err)
	}
	if got != paid ||
		received.PaymentNo != testPaymentNo ||
		received.ProviderTransactionID != testProviderNo ||
		!received.PaidAt.Equal(paidAt) {
		t.Fatalf("payment/callback = %+v/%+v", got, received)
	}
}

func TestPaymentServiceHandlePaidCallbackRejectsInvalidEvent(t *testing.T) {
	repository := &repositoryStub{
		applyPaidCallback: func(context.Context, model.PaidCallback) (*model.Payment, error) {
			t.Fatal("invalid callback must not reach the repository")
			return nil, nil
		},
	}
	paymentService := service.NewPaymentService(repository, service.Config{})

	_, err := paymentService.HandlePaidCallback(context.Background(), model.PaidCallback{
		EventID:               testEventID,
		EventType:             "payment.closed",
		ProviderTransactionID: testProviderNo,
		PaymentNo:             testPaymentNo,
		AmountCents:           3998,
		Currency:              model.CurrencyCNY,
		PaidAt:                time.Now(),
	})
	if !errors.Is(err, apperr.ErrInvalidCallback) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidCallback)
	}
}

func TestPaymentServiceHandlePaidCallbackExpiresOrderBeforeRecordingLatePayment(t *testing.T) {
	paid := &model.Payment{PaymentNo: testPaymentNo, Status: model.PaymentStatusSucceeded}
	applyCalls := 0
	repository := &repositoryStub{
		applyPaidCallback: func(context.Context, model.PaidCallback) (*model.Payment, error) {
			applyCalls++
			if applyCalls == 1 {
				return nil, &paymentport.OrderExpiredError{OrderID: 7}
			}
			return paid, nil
		},
	}
	expiredOrderID := 0
	paymentService := service.NewPaymentService(repository, service.Config{
		OrderExpirer: orderExpirerStub{expire: func(_ context.Context, orderID int) error {
			expiredOrderID = orderID
			return nil
		}},
	})

	got, err := paymentService.HandlePaidCallback(context.Background(), validPaidCallback())
	if err != nil {
		t.Fatalf("handle late callback: %v", err)
	}
	if got != paid || applyCalls != 2 || expiredOrderID != 7 {
		t.Fatalf(
			"payment/calls/order = %+v/%d/%d, want %+v/2/7",
			got,
			applyCalls,
			expiredOrderID,
			paid,
		)
	}
}

func TestPaymentServiceHandlePaidCallbackStopsWhenOrderExpiryFails(t *testing.T) {
	applyCalls := 0
	repository := &repositoryStub{
		applyPaidCallback: func(context.Context, model.PaidCallback) (*model.Payment, error) {
			applyCalls++
			return nil, &paymentport.OrderExpiredError{OrderID: 7}
		},
	}
	expireCalls := 0
	paymentService := service.NewPaymentService(repository, service.Config{
		OrderExpirer: orderExpirerStub{expire: func(context.Context, int) error {
			expireCalls++
			return apperr.ErrOrderNotPayable
		}},
	})

	_, err := paymentService.HandlePaidCallback(context.Background(), validPaidCallback())
	if !errors.Is(err, apperr.ErrOrderNotPayable) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrOrderNotPayable)
	}
	if applyCalls != 1 || expireCalls != 1 {
		t.Fatalf("apply/expire calls = %d/%d, want 1/1", applyCalls, expireCalls)
	}
}

func validPaidCallback() model.PaidCallback {
	return model.PaidCallback{
		EventID:               testEventID,
		EventType:             "payment.paid",
		ProviderTransactionID: testProviderNo,
		PaymentNo:             testPaymentNo,
		AmountCents:           3998,
		Currency:              model.CurrencyCNY,
		PaidAt:                time.Now(),
	}
}

func stringPointer(value string) *string {
	return &value
}
