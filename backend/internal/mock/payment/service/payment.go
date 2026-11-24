package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"pinshaoshao/internal/mock/payment/model"

	"github.com/google/uuid"
)

const (
	defaultPaymentTTL = 30 * time.Minute
	minPaymentTTL     = time.Minute
	maxPaymentTTL     = 24 * time.Hour
	maxMerchantNoLen  = 64
	maxSubjectLen     = 200
	maxURLLen         = 2048
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) (*model.Payment, bool, error)
	Get(ctx context.Context, paymentNo string) (*model.Payment, error)
	Pay(ctx context.Context, paymentNo string, now time.Time) (*model.Payment, error)
	Close(ctx context.Context, paymentNo string, now time.Time) (*model.Payment, error)
	CloseExpired(ctx context.Context, now time.Time, limit int) ([]model.Payment, error)
}

type CreatePaymentInput struct {
	MerchantOrderNo string
	AmountCents     int64
	Currency        string
	Subject         string
	NotifyURL       string
	ReturnURL       string
	ExpiresIn       time.Duration
}

type PaymentService struct {
	repository PaymentRepository
	now        func() time.Time
}

func NewPaymentService(repository PaymentRepository) *PaymentService {
	return &PaymentService{
		repository: repository,
		now:        time.Now,
	}
}

func (s *PaymentService) Create(ctx context.Context, input CreatePaymentInput) (*model.Payment, bool, error) {
	normalized, requestHash, err := normalizeCreateInput(input)
	if err != nil {
		return nil, false, err
	}

	now := s.now().UTC()
	payment := &model.Payment{
		PaymentNo:       uuid.NewString(),
		MerchantOrderNo: normalized.MerchantOrderNo,
		RequestHash:     requestHash,
		Status:          model.PaymentStatusPending,
		AmountCents:     normalized.AmountCents,
		Currency:        normalized.Currency,
		Subject:         normalized.Subject,
		NotifyURL:       normalized.NotifyURL,
		ReturnURL:       normalized.ReturnURL,
		ExpiresAt:       now.Add(normalized.ExpiresIn),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	created, isNew, err := s.repository.Create(ctx, payment)
	if err != nil {
		return nil, false, fmt.Errorf("service.CreatePayment: %w", err)
	}
	return created, isNew, nil
}

func (s *PaymentService) Get(ctx context.Context, paymentNo string) (*model.Payment, error) {
	if _, err := parsePaymentNo(paymentNo); err != nil {
		return nil, err
	}

	payment, err := s.repository.Get(ctx, paymentNo)
	if err != nil {
		return nil, fmt.Errorf("service.GetPayment: %w", err)
	}

	now := s.now().UTC()
	if payment.Status != model.PaymentStatusPending || now.Before(payment.ExpiresAt) {
		return payment, nil
	}

	closed, err := s.repository.Close(ctx, paymentNo, now)
	if err == nil {
		return closed, nil
	}
	if errors.Is(err, model.ErrPaymentNotClosable) {
		payment, getErr := s.repository.Get(ctx, paymentNo)
		if getErr != nil {
			return nil, fmt.Errorf("service.GetPayment: %w", getErr)
		}
		return payment, nil
	}
	return nil, fmt.Errorf("service.GetPayment: %w", err)
}

func (s *PaymentService) Pay(ctx context.Context, paymentNo string) (*model.Payment, error) {
	if _, err := parsePaymentNo(paymentNo); err != nil {
		return nil, err
	}

	payment, err := s.repository.Pay(ctx, paymentNo, s.now().UTC())
	if err != nil {
		return nil, fmt.Errorf("service.PayPayment: %w", err)
	}
	return payment, nil
}

func (s *PaymentService) Close(ctx context.Context, paymentNo string) (*model.Payment, error) {
	if _, err := parsePaymentNo(paymentNo); err != nil {
		return nil, err
	}

	payment, err := s.repository.Close(ctx, paymentNo, s.now().UTC())
	if err != nil {
		return nil, fmt.Errorf("service.ClosePayment: %w", err)
	}
	return payment, nil
}

func (s *PaymentService) CloseExpired(ctx context.Context, limit int) (int, error) {
	closed, err := s.repository.CloseExpired(ctx, s.now().UTC(), limit)
	if err != nil {
		return 0, fmt.Errorf("service.CloseExpiredPayments: %w", err)
	}
	return len(closed), nil
}

func normalizeCreateInput(input CreatePaymentInput) (CreatePaymentInput, string, error) {
	input.MerchantOrderNo = strings.TrimSpace(input.MerchantOrderNo)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.Subject = strings.TrimSpace(input.Subject)
	input.NotifyURL = strings.TrimSpace(input.NotifyURL)
	input.ReturnURL = strings.TrimSpace(input.ReturnURL)

	if input.MerchantOrderNo == "" || utf8.RuneCountInString(input.MerchantOrderNo) > maxMerchantNoLen {
		return CreatePaymentInput{}, "", model.ErrInvalidPaymentInput
	}
	if input.AmountCents <= 0 {
		return CreatePaymentInput{}, "", model.ErrInvalidPaymentInput
	}
	if input.Currency == "" {
		input.Currency = "CNY"
	}
	if !currencyPattern.MatchString(input.Currency) {
		return CreatePaymentInput{}, "", model.ErrInvalidPaymentInput
	}
	if input.Subject == "" || utf8.RuneCountInString(input.Subject) > maxSubjectLen {
		return CreatePaymentInput{}, "", model.ErrInvalidPaymentInput
	}
	if err := validateCallbackURL(input.NotifyURL, true); err != nil {
		return CreatePaymentInput{}, "", err
	}
	if err := validateCallbackURL(input.ReturnURL, false); err != nil {
		return CreatePaymentInput{}, "", err
	}
	if input.ExpiresIn == 0 {
		input.ExpiresIn = defaultPaymentTTL
	}
	if input.ExpiresIn < minPaymentTTL || input.ExpiresIn > maxPaymentTTL {
		return CreatePaymentInput{}, "", model.ErrInvalidPaymentInput
	}

	payload := struct {
		MerchantOrderNo string
		AmountCents     int64
		Currency        string
		Subject         string
		NotifyURL       string
		ReturnURL       string
		ExpiresIn       int64
	}{
		MerchantOrderNo: input.MerchantOrderNo,
		AmountCents:     input.AmountCents,
		Currency:        input.Currency,
		Subject:         input.Subject,
		NotifyURL:       input.NotifyURL,
		ReturnURL:       input.ReturnURL,
		ExpiresIn:       int64(input.ExpiresIn),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return CreatePaymentInput{}, "", err
	}
	sum := sha256.Sum256(encoded)
	return input, hex.EncodeToString(sum[:]), nil
}

func validateCallbackURL(value string, required bool) error {
	if value == "" {
		if required {
			return model.ErrInvalidPaymentInput
		}
		return nil
	}
	if len(value) > maxURLLen {
		return model.ErrInvalidPaymentInput
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return model.ErrInvalidPaymentInput
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return model.ErrInvalidPaymentInput
	}
	return nil
}

func parsePaymentNo(value string) (string, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil {
		return "", model.ErrInvalidPaymentNumber
	}
	return parsed.String(), nil
}
