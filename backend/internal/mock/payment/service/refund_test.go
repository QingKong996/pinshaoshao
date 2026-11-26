package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"pinshaoshao/internal/mock/payment/model"
)

type refundRepositoryStub struct {
	refund *model.Refund
	err    error
	input  model.RefundInput
}

func (r *refundRepositoryStub) Refund(
	_ context.Context,
	_ string,
	input model.RefundInput,
	_ time.Time,
) (*model.Refund, error) {
	r.input = input
	return r.refund, r.err
}

func TestRefundServiceNormalizesRequest(t *testing.T) {
	refund := &model.Refund{RefundNo: "refund-1"}
	repository := &refundRepositoryStub{refund: refund}
	service := NewRefundService(repository)

	got, err := service.Refund(
		context.Background(),
		"4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
		model.RefundInput{
			RefundNo:    "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98",
			AmountCents: 3998,
			Reason:      " order_cancelled_after_payment ",
		},
	)
	if err != nil {
		t.Fatalf("refund payment: %v", err)
	}
	if got != refund ||
		repository.input.RefundNo != "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98" ||
		repository.input.Reason != "order_cancelled_after_payment" {
		t.Fatalf("refund/input = %+v/%+v", got, repository.input)
	}
}

func TestRefundServiceRejectsInvalidRequest(t *testing.T) {
	service := NewRefundService(&refundRepositoryStub{})

	tests := []model.RefundInput{
		{RefundNo: "invalid", AmountCents: 3998, Reason: "refund"},
		{RefundNo: "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98", AmountCents: 0, Reason: "refund"},
		{RefundNo: "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98", AmountCents: 3998, Reason: " "},
	}
	for _, input := range tests {
		if _, err := service.Refund(
			context.Background(),
			"4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
			input,
		); !errors.Is(err, model.ErrInvalidRefundInput) {
			t.Fatalf("input %+v: error = %v, want invalid refund", input, err)
		}
	}
}
