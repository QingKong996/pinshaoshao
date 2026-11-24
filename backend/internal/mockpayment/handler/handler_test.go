package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pinshaoshao/internal/mockpayment/dto"
	"pinshaoshao/internal/mockpayment/handler"
	"pinshaoshao/internal/mockpayment/model"
	"pinshaoshao/internal/mockpayment/service"

	"github.com/gin-gonic/gin"
)

type stubPaymentService struct {
	payment     *model.Payment
	created     bool
	createInput service.CreatePaymentInput
}

func (s *stubPaymentService) Create(_ context.Context, input service.CreatePaymentInput) (*model.Payment, bool, error) {
	s.createInput = input
	return s.payment, s.created, nil
}

func (s *stubPaymentService) Get(context.Context, string) (*model.Payment, error) {
	return s.payment, nil
}

func (s *stubPaymentService) Pay(context.Context, string) (*model.Payment, error) {
	paidAt := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	payment := *s.payment
	payment.Status = model.PaymentStatusPaid
	payment.PaidAt = &paidAt
	s.payment = &payment
	return &payment, nil
}

func (s *stubPaymentService) Close(context.Context, string) (*model.Payment, error) {
	payment := *s.payment
	payment.Status = model.PaymentStatusClosed
	s.payment = &payment
	return &payment, nil
}

func TestPaymentHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const paymentNo = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"
	paymentService := &stubPaymentService{
		created: true,
		payment: &model.Payment{
			PaymentNo:       paymentNo,
			MerchantOrderNo: "order-1",
			Status:          model.PaymentStatusPending,
			AmountCents:     3998,
			Currency:        "CNY",
			Subject:         "T-shirt",
			ExpiresAt:       time.Now().Add(time.Hour),
		},
	}
	paymentHandler := handler.NewPaymentHandler(paymentService, "http://localhost:8080")
	router := gin.New()
	router.POST("/mockpay/api/v1/payments", paymentHandler.Create)

	createRequest := func() *http.Request {
		request := httptest.NewRequest(
			http.MethodPost,
			"/mockpay/api/v1/payments",
			strings.NewReader(`{
			"merchant_order_no":"order-1",
			"amount_cents":3998,
			"subject":"T-shirt",
			"notify_url":"http://localhost:8080/webhook",
			"expires_in_seconds":1800
		}`),
		)
		request.Header.Set("Content-Type", "application/json")
		return request
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, createRequest())

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body = %s", response.Code, response.Body.String())
	}
	var body dto.PaymentResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	wantURL := "http://localhost:8080/mockpay/checkout/" + paymentNo
	if body.CheckoutURL != wantURL {
		t.Fatalf("checkout URL = %q, want %q", body.CheckoutURL, wantURL)
	}
	if paymentService.createInput.ExpiresIn != 30*time.Minute {
		t.Fatalf("expires in = %s, want 30m", paymentService.createInput.ExpiresIn)
	}

	paymentService.created = false
	replayResponse := httptest.NewRecorder()
	router.ServeHTTP(replayResponse, createRequest())
	if replayResponse.Code != http.StatusOK {
		t.Fatalf("replay status = %d, want 200", replayResponse.Code)
	}
}

func TestPaymentHandlerCreateRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	paymentHandler := handler.NewPaymentHandler(&stubPaymentService{}, "")
	router := gin.New()
	router.POST("/mockpay/api/v1/payments", paymentHandler.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/mockpay/api/v1/payments",
		strings.NewReader(`{"merchant_order_no":`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestCheckoutHandlerPayAndShow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const paymentNo = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"
	paymentService := &stubPaymentService{
		payment: &model.Payment{
			PaymentNo:       paymentNo,
			MerchantOrderNo: "order-1",
			Status:          model.PaymentStatusPending,
			AmountCents:     3998,
			Currency:        "CNY",
			Subject:         "T-shirt",
			ExpiresAt:       time.Now().Add(time.Hour),
		},
	}
	checkoutHandler := handler.NewCheckoutHandler(paymentService)
	router := gin.New()
	router.GET("/mockpay/checkout/:paymentNo", checkoutHandler.Show)
	router.POST("/mockpay/checkout/:paymentNo/pay", checkoutHandler.Pay)

	showRequest := httptest.NewRequest(http.MethodGet, "/mockpay/checkout/"+paymentNo, nil)
	showResponse := httptest.NewRecorder()
	router.ServeHTTP(showResponse, showRequest)
	if showResponse.Code != http.StatusOK || !strings.Contains(showResponse.Body.String(), "T-shirt") {
		t.Fatalf("checkout response = %d/%s", showResponse.Code, showResponse.Body.String())
	}

	payRequest := httptest.NewRequest(http.MethodPost, "/mockpay/checkout/"+paymentNo+"/pay", nil)
	payResponse := httptest.NewRecorder()
	router.ServeHTTP(payResponse, payRequest)
	if payResponse.Code != http.StatusSeeOther {
		t.Fatalf("pay status = %d, want 303", payResponse.Code)
	}
	if got := payResponse.Header().Get("Location"); got != "/mockpay/checkout/"+paymentNo {
		t.Fatalf("pay location = %q", got)
	}
}
