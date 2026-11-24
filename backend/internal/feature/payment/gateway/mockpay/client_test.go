package mockpay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pinshaoshao/internal/feature/payment/model"
)

func TestClientCreatePayment(t *testing.T) {
	const merchantNo = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"
	const providerNo = "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98"

	var received createPaymentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != createPaymentPath {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"payment_no":        providerNo,
			"merchant_order_no": merchantNo,
			"status":            "pending",
			"amount_cents":      3998,
			"currency":          "CNY",
			"checkout_url":      "/mockpay/checkout/" + providerNo,
			"expires_at":        time.Now().Add(30 * time.Minute),
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	result, err := client.CreatePayment(context.Background(), model.ProviderPaymentRequest{
		MerchantOrderNo: merchantNo,
		AmountCents:     3998,
		Currency:        model.CurrencyCNY,
		Subject:         "Order #7",
		NotifyURL:       "http://localhost:8080/api/v1/payment/webhook",
		ReturnURL:       "http://localhost:5173/orders/7",
		ExpiresIn:       20 * time.Minute,
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if received.MerchantOrderNo != merchantNo ||
		received.AmountCents != 3998 ||
		received.NotifyURL != "http://localhost:8080/api/v1/payment/webhook" ||
		received.ExpiresInSeconds != 1200 {
		t.Fatalf("request = %+v", received)
	}
	if result.PaymentNo != providerNo ||
		result.MerchantNo != merchantNo ||
		result.CheckoutURL != server.URL+"/mockpay/checkout/"+providerNo {
		t.Fatalf("result = %+v", result)
	}
}
