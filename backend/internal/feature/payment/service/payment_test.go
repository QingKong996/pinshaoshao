package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
	closeForUser           func(context.Context, int, string) (*model.Payment, error)
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

func (s *repositoryStub) CloseForUser(
	ctx context.Context,
	userID int,
	paymentNo string,
) (*model.Payment, error) {
	return s.closeForUser(ctx, userID, paymentNo)
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
	paid := &model.Payment{
		PaymentNo:             testPaymentNo,
		OrderID:               7,
		Status:                model.PaymentStatusSucceeded,
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
			return paid, nil
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
	if gatewayInput.ExpiresIn != 30*time.Minute {
		t.Fatalf("gateway expiry = %s, want stable payment window", gatewayInput.ExpiresIn)
	}
	if session.Payment.Status != model.PaymentStatusSucceeded || session.CheckoutURL == "" {
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

func TestPaymentServiceCloseClosesProviderAndOrder(t *testing.T) {
	providerNo := testProviderNo
	payment := &model.Payment{
		PaymentNo:             testPaymentNo,
		OrderID:               7,
		Status:                model.PaymentStatusPending,
		ProviderTransactionID: &providerNo,
	}
	closed := &model.Payment{
		PaymentNo: testPaymentNo,
		OrderID:   7,
		Status:    model.PaymentStatusClosed,
	}
	var providerClosed string
	var repositoryClosed bool
	repository := &repositoryStub{
		getForUser: func(_ context.Context, userID int, paymentNo string) (*model.Payment, error) {
			if userID != 3 || paymentNo != testPaymentNo {
				t.Fatalf("get payment arguments = %d/%q", userID, paymentNo)
			}
			return payment, nil
		},
		closeForUser: func(_ context.Context, userID int, paymentNo string) (*model.Payment, error) {
			if userID != 3 || paymentNo != testPaymentNo {
				t.Fatalf("close payment arguments = %d/%q", userID, paymentNo)
			}
			repositoryClosed = true
			return closed, nil
		},
	}
	gateway := gatewayStub{closePayment: func(_ context.Context, providerPaymentNo string) error {
		providerClosed = providerPaymentNo
		return nil
	}}
	paymentService := service.NewPaymentService(repository, service.Config{Gateway: gateway})

	got, err := paymentService.Close(context.Background(), 3, testPaymentNo)
	if err != nil {
		t.Fatalf("close payment: %v", err)
	}
	if got != closed || providerClosed != testProviderNo || !repositoryClosed {
		t.Fatalf("result/provider/repository = %+v/%q/%t", got, providerClosed, repositoryClosed)
	}
}

func TestPaymentServiceCloseSkipsProviderWithoutTransaction(t *testing.T) {
	payment := &model.Payment{
		PaymentNo: testPaymentNo,
		Status:    model.PaymentStatusPending,
	}
	repository := &repositoryStub{
		getForUser: func(context.Context, int, string) (*model.Payment, error) {
			return payment, nil
		},
		closeForUser: func(context.Context, int, string) (*model.Payment, error) {
			return &model.Payment{
				PaymentNo: testPaymentNo,
				Status:    model.PaymentStatusClosed,
			}, nil
		},
	}
	gateway := gatewayStub{closePayment: func(context.Context, string) error {
		t.Fatal("payment without a provider transaction must not call the provider")
		return nil
	}}
	paymentService := service.NewPaymentService(repository, service.Config{Gateway: gateway})

	closed, err := paymentService.Close(context.Background(), 3, testPaymentNo)
	if err != nil {
		t.Fatalf("close payment: %v", err)
	}
	if closed.Status != model.PaymentStatusClosed {
		t.Fatalf("status = %q, want closed", closed.Status)
	}
}

func TestPaymentServiceCloseIsIdempotentWhenAlreadyClosed(t *testing.T) {
	closed := &model.Payment{PaymentNo: testPaymentNo, Status: model.PaymentStatusClosed}
	repository := &repositoryStub{
		getForUser: func(context.Context, int, string) (*model.Payment, error) {
			return closed, nil
		},
		closeForUser: func(context.Context, int, string) (*model.Payment, error) {
			t.Fatal("already closed payment must not be closed again")
			return nil, nil
		},
	}
	gateway := gatewayStub{closePayment: func(context.Context, string) error {
		t.Fatal("already closed payment must not call the provider")
		return nil
	}}
	paymentService := service.NewPaymentService(repository, service.Config{Gateway: gateway})

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
		closeForUser: func(context.Context, int, string) (*model.Payment, error) {
			t.Fatal("paid payment must not be closed")
			return nil, nil
		},
	}
	gateway := gatewayStub{closePayment: func(context.Context, string) error {
		t.Fatal("paid payment must not call the provider")
		return nil
	}}
	paymentService := service.NewPaymentService(repository, service.Config{Gateway: gateway})

	_, err := paymentService.Close(context.Background(), 3, testPaymentNo)
	if !errors.Is(err, apperr.ErrPaymentConflict) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrPaymentConflict)
	}
}

func TestPaymentServiceCloseStopsWhenProviderRejects(t *testing.T) {
	providerNo := testProviderNo
	repository := &repositoryStub{
		getForUser: func(context.Context, int, string) (*model.Payment, error) {
			return &model.Payment{
				PaymentNo:             testPaymentNo,
				Status:                model.PaymentStatusPending,
				ProviderTransactionID: &providerNo,
			}, nil
		},
		closeForUser: func(context.Context, int, string) (*model.Payment, error) {
			t.Fatal("local payment must not close after a provider conflict")
			return nil, nil
		},
	}
	gateway := gatewayStub{closePayment: func(context.Context, string) error {
		return apperr.ErrPaymentConflict
	}}
	paymentService := service.NewPaymentService(repository, service.Config{Gateway: gateway})

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

func stringPointer(value string) *string {
	return &value
}
