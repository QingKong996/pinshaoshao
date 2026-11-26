package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"pinshaoshao/internal/mock/payment/model"

	"github.com/google/uuid"
)

const maxRefundReasonLen = 200

type RefundRepository interface {
	Refund(
		ctx context.Context,
		paymentNo string,
		input model.RefundInput,
		now time.Time,
	) (*model.Refund, error)
}

type RefundService struct {
	repository RefundRepository
	now        func() time.Time
}

func NewRefundService(repository RefundRepository) *RefundService {
	return &RefundService{
		repository: repository,
		now:        time.Now,
	}
}

func (s *RefundService) Refund(
	ctx context.Context,
	paymentNo string,
	input model.RefundInput,
) (*model.Refund, error) {
	paymentNo, err := parsePaymentNo(paymentNo)
	if err != nil {
		return nil, err
	}
	refundNo, err := uuid.Parse(strings.TrimSpace(input.RefundNo))
	if err != nil || refundNo == uuid.Nil {
		return nil, model.ErrInvalidRefundInput
	}
	input.RefundNo = refundNo.String()
	input.Reason = strings.TrimSpace(input.Reason)
	if input.AmountCents <= 0 ||
		input.Reason == "" ||
		utf8.RuneCountInString(input.Reason) > maxRefundReasonLen {
		return nil, model.ErrInvalidRefundInput
	}

	refund, err := s.repository.Refund(ctx, paymentNo, input, s.now().UTC())
	if err != nil {
		return nil, fmt.Errorf("service.RefundPayment: %w", err)
	}
	return refund, nil
}
