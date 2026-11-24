package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"pinshaoshao/internal/feature/payment/model"
	"pinshaoshao/internal/webhooksig"

	"github.com/gin-gonic/gin"
)

const (
	webhookTestSecret = "webhook-test-secret"
	webhookEventID    = "a1c6cfe8-e77b-4ac2-b8bf-c6a0f56673f1"
	webhookPaymentNo  = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"
	webhookProviderNo = "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98"
)

type callbackServiceStub struct {
	handle func(context.Context, model.PaidCallback) (*model.Payment, error)
}

func (s callbackServiceStub) HandlePaidCallback(
	ctx context.Context,
	request model.PaidCallback,
) (*model.Payment, error) {
	return s.handle(ctx, request)
}

func TestWebhookHandlerProcessesSignedPayment(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	body := paidWebhookBody(t, now)
	var received model.PaidCallback
	handler := NewWebhookHandler(callbackServiceStub{
		handle: func(_ context.Context, callback model.PaidCallback) (*model.Payment, error) {
			received = callback
			return &model.Payment{PaymentNo: callback.PaymentNo}, nil
		},
	}, webhookTestSecret, 5*time.Minute)
	handler.now = func() time.Time { return now }

	rec := doWebhookRequest(handler, body, now, webhooksig.NewSigner(webhookTestSecret).Sign(now, body))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204, body = %s", rec.Code, rec.Body.String())
	}
	if received.EventID != webhookEventID ||
		received.EventType != "payment.paid" ||
		received.ProviderTransactionID != webhookProviderNo ||
		received.PaymentNo != webhookPaymentNo ||
		received.AmountCents != 3998 ||
		received.Currency != model.CurrencyCNY {
		t.Fatalf("callback = %+v", received)
	}
}

func TestWebhookHandlerRejectsInvalidSignature(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	body := paidWebhookBody(t, now)
	handler := NewWebhookHandler(callbackServiceStub{
		handle: func(context.Context, model.PaidCallback) (*model.Payment, error) {
			t.Fatal("invalid signature must not reach the service")
			return nil, nil
		},
	}, webhookTestSecret, 5*time.Minute)
	handler.now = func() time.Time { return now }

	rec := doWebhookRequest(handler, body, now, "v1=invalid")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body = %s", rec.Code, rec.Body.String())
	}
}

func TestWebhookHandlerRejectsStaleSignature(t *testing.T) {
	signedAt := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	body := paidWebhookBody(t, signedAt)
	handler := NewWebhookHandler(callbackServiceStub{
		handle: func(context.Context, model.PaidCallback) (*model.Payment, error) {
			t.Fatal("stale signature must not reach the service")
			return nil, nil
		},
	}, webhookTestSecret, 5*time.Minute)
	handler.now = func() time.Time { return signedAt.Add(6 * time.Minute) }

	rec := doWebhookRequest(handler, body, signedAt, webhooksig.NewSigner(webhookTestSecret).Sign(signedAt, body))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body = %s", rec.Code, rec.Body.String())
	}
}

func doWebhookRequest(
	handler *WebhookHandler,
	body []byte,
	timestamp time.Time,
	signature string,
) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/payment/webhook", handler.Handle)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/payment/webhook", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(headerEventID, webhookEventID)
	request.Header.Set(headerEventType, "payment.paid")
	request.Header.Set(headerTimestamp, strconv.FormatInt(timestamp.Unix(), 10))
	request.Header.Set(headerSignature, signature)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, request)
	return rec
}

func paidWebhookBody(t *testing.T, paidAt time.Time) []byte {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"id":          webhookEventID,
		"type":        "payment.paid",
		"occurred_at": paidAt.Format(time.RFC3339Nano),
		"data": map[string]any{
			"payment_no":        webhookProviderNo,
			"merchant_order_no": webhookPaymentNo,
			"status":            "paid",
			"amount_cents":      3998,
			"currency":          "CNY",
			"subject":           "Order #7",
			"paid_at":           paidAt.Format(time.RFC3339Nano),
		},
	})
	if err != nil {
		t.Fatalf("encode callback: %v", err)
	}
	return body
}
