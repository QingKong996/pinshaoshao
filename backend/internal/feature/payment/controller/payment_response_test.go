package controller

import (
	stdjson "encoding/json"
	"strings"
	"testing"

	"pinshaoshao/internal/feature/payment/model"
)

func TestOrderPaymentResponseOmitsRefundFailureDetails(t *testing.T) {
	lastError := "provider response contained internal data"
	response := newOrderPaymentResponse(&model.Payment{
		Refund: &model.Refund{LastError: &lastError},
	}, "")

	payload, err := stdjson.Marshal(response)
	if err != nil {
		t.Fatalf("encode payment response: %v", err)
	}
	if strings.Contains(string(payload), "last_error") ||
		strings.Contains(string(payload), lastError) {
		t.Fatalf("payment response leaked refund failure details: %s", payload)
	}
}
