package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/model"

	"github.com/google/uuid"
)

const (
	eventTypePaymentPaid  = "payment.paid"
	providerPaidStatus    = "paid"
	minProviderPaymentTTL = 60 * time.Second
)

type PaymentRepository interface {
	CreateForOrder(ctx context.Context, userID int, orderID int, paymentNo string) (*model.Payment, bool, error)
	GetForUser(ctx context.Context, userID int, paymentNo string) (*model.Payment, error)
	GetForPayment(ctx context.Context, userID int, paymentNo string) (*model.Payment, error)
	SetProviderTransaction(ctx context.Context, paymentNo string, providerTransactionID string) error
	ApplyPaidCallback(ctx context.Context, callback model.PaidCallback) (*model.Payment, error)
}

type PaymentGateway interface {
	CreatePayment(ctx context.Context, input model.ProviderPaymentRequest) (*model.ProviderPaymentResult, error)
	ClosePayment(ctx context.Context, providerPaymentNo string) error
}

type OrderCanceller interface {
	CancelForPayment(ctx context.Context, userID int, orderID int) error
}

type Config struct {
	NotifyURL      string
	ReturnURLBase  string
	Gateway        PaymentGateway
	OrderCanceller OrderCanceller
}

type PaymentService struct {
	repository     PaymentRepository
	gateway        PaymentGateway
	orderCanceller OrderCanceller
	notifyURL      string
	returnURLBase  string
}

type PaymentSession struct {
	Payment     *model.Payment
	CheckoutURL string
}

func NewPaymentService(repository PaymentRepository, config Config) *PaymentService {
	return &PaymentService{
		repository:     repository,
		gateway:        config.Gateway,
		orderCanceller: config.OrderCanceller,
		notifyURL:      config.NotifyURL,
		returnURLBase:  strings.TrimRight(config.ReturnURLBase, "/"),
	}
}

func (s *PaymentService) Create(
	ctx context.Context,
	userID int,
	orderID int,
) (*model.Payment, bool, error) {
	if orderID <= 0 {
		return nil, false, apperr.ErrInvalidOrderID
	}

	payment, created, err := s.repository.CreateForOrder(ctx, userID, orderID, uuid.NewString())
	if err != nil {
		return nil, false, fmt.Errorf("service.CreatePayment: %w", err)
	}
	return payment, created, nil
}

func (s *PaymentService) Get(
	ctx context.Context,
	userID int,
	paymentNo string,
) (*model.Payment, error) {
	normalized, err := parsePaymentNo(paymentNo)
	if err != nil {
		return nil, err
	}

	payment, err := s.repository.GetForUser(ctx, userID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.GetPayment: %w", err)
	}
	return payment, nil
}

func (s *PaymentService) Pay(
	ctx context.Context,
	userID int,
	paymentNo string,
) (*PaymentSession, error) {
	normalized, err := parsePaymentNo(paymentNo)
	if err != nil {
		return nil, err
	}

	payment, err := s.repository.GetForPayment(ctx, userID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.PayPayment: %w", err)
	}
	if payment.Status == model.PaymentStatusSucceeded {
		return &PaymentSession{Payment: payment}, nil
	}

	if time.Until(payment.OrderExpiresAt) < minProviderPaymentTTL {
		return nil, fmt.Errorf("service.PayPayment: %w", apperr.ErrOrderNotPayable)
	}
	providerExpiresIn := time.Until(payment.OrderExpiresAt)
	result, err := s.gateway.CreatePayment(ctx, model.ProviderPaymentRequest{
		MerchantOrderNo: payment.PaymentNo,
		AmountCents:     payment.AmountCents,
		Currency:        model.CurrencyCNY,
		Subject:         fmt.Sprintf("Order #%d", payment.OrderID),
		NotifyURL:       s.notifyURL,
		ReturnURL:       s.returnURL(payment.OrderID),
		ExpiresIn:       providerExpiresIn,
	})
	if err != nil {
		return nil, fmt.Errorf("service.PayPayment: %w", err)
	}
	if err := validateProviderResult(payment, result); err != nil {
		return nil, fmt.Errorf("service.PayPayment: %w", err)
	}

	if err := s.repository.SetProviderTransaction(ctx, normalized, result.PaymentNo); err != nil {
		if errors.Is(err, apperr.ErrPaymentConflict) ||
			errors.Is(err, apperr.ErrPaymentClosed) {
			if closeErr := s.gateway.ClosePayment(ctx, result.PaymentNo); closeErr != nil {
				slog.ErrorContext(
					ctx,
					"close provider payment after local conflict",
					"payment_no",
					normalized,
					"provider_transaction_id",
					result.PaymentNo,
					"error",
					closeErr,
				)
			}
		}
		return nil, fmt.Errorf("service.PayPayment: %w", err)
	}
	updated, err := s.repository.GetForPayment(ctx, userID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.PayPayment: %w", err)
	}
	return &PaymentSession{
		Payment:     updated,
		CheckoutURL: result.CheckoutURL,
	}, nil
}

func (s *PaymentService) Close(
	ctx context.Context,
	userID int,
	paymentNo string,
) (*model.Payment, error) {
	normalized, err := parsePaymentNo(paymentNo)
	if err != nil {
		return nil, err
	}

	payment, err := s.repository.GetForUser(ctx, userID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.ClosePayment: %w", err)
	}
	switch payment.Status {
	case model.PaymentStatusClosed:
		return payment, nil
	case model.PaymentStatusSucceeded:
		return nil, fmt.Errorf("service.ClosePayment: %w", apperr.ErrPaymentConflict)
	}

	if s.orderCanceller == nil {
		return nil, fmt.Errorf("service.ClosePayment: %w", apperr.ErrInternalServer)
	}
	if err := s.orderCanceller.CancelForPayment(ctx, userID, payment.OrderID); err != nil {
		return nil, fmt.Errorf("service.ClosePayment: %w", err)
	}

	closed, err := s.repository.GetForUser(ctx, userID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.ClosePayment: %w", err)
	}
	return closed, nil
}

func (s *PaymentService) HandlePaidCallback(
	ctx context.Context,
	request model.PaidCallback,
) (*model.Payment, error) {
	if _, err := parseUUID(request.EventID); err != nil {
		return nil, fmt.Errorf("service.HandlePaidCallback: %w", apperr.ErrInvalidCallback)
	}
	if request.EventType != eventTypePaymentPaid {
		return nil, fmt.Errorf("service.HandlePaidCallback: %w", apperr.ErrInvalidCallback)
	}
	if _, err := parseUUID(request.ProviderTransactionID); err != nil {
		return nil, fmt.Errorf("service.HandlePaidCallback: %w", apperr.ErrInvalidCallback)
	}
	paymentNo, err := parseUUID(request.PaymentNo)
	if err != nil {
		return nil, fmt.Errorf("service.HandlePaidCallback: %w", apperr.ErrInvalidCallback)
	}
	if request.AmountCents <= 0 ||
		strings.ToUpper(strings.TrimSpace(request.Currency)) != model.CurrencyCNY ||
		request.PaidAt.IsZero() {
		return nil, fmt.Errorf("service.HandlePaidCallback: %w", apperr.ErrInvalidCallback)
	}
	request.PaymentNo = paymentNo
	request.ProviderTransactionID = strings.ToLower(strings.TrimSpace(request.ProviderTransactionID))
	request.PaidAt = request.PaidAt.UTC()

	payment, err := s.repository.ApplyPaidCallback(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("service.HandlePaidCallback: %w", err)
	}
	return payment, nil
}

func (s *PaymentService) returnURL(orderID int) string {
	if s.returnURLBase == "" {
		return ""
	}
	return fmt.Sprintf("%s/orders/%d", s.returnURLBase, orderID)
}

func parsePaymentNo(value string) (string, error) {
	parsed, err := parseUUID(value)
	if err != nil {
		return "", apperr.ErrInvalidPaymentNo
	}
	return parsed, nil
}

func parseUUID(value string) (string, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", err
	}
	if parsed == uuid.Nil {
		return "", errors.New("nil uuid")
	}
	return parsed.String(), nil
}

func validateProviderResult(
	payment *model.Payment,
	result *model.ProviderPaymentResult,
) error {
	if result == nil {
		return apperr.ErrPaymentProviderInvalid
	}
	providerPaymentNo, err := parseUUID(result.PaymentNo)
	if err != nil ||
		result.MerchantNo != payment.PaymentNo ||
		result.AmountCents != payment.AmountCents ||
		strings.ToUpper(strings.TrimSpace(result.Currency)) != model.CurrencyCNY ||
		(result.Status != "pending" && result.Status != providerPaidStatus) ||
		(result.Status != providerPaidStatus && strings.TrimSpace(result.CheckoutURL) == "") {
		return apperr.ErrPaymentProviderInvalid
	}
	result.PaymentNo = providerPaymentNo
	result.Currency = model.CurrencyCNY
	return nil
}
