package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecoveryDoesNotLogSensitiveRequestPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const paymentNo = "sensitive-payment-capability"
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	router := gin.New()
	router.Use(Recovery(logger))
	router.GET("/mockpay/checkout/:paymentNo", func(ctx *gin.Context) {
		panic(ctx.Param("paymentNo"))
	})

	request := httptest.NewRequest(http.MethodGet, "/mockpay/checkout/"+paymentNo, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(output.String(), paymentNo) {
		t.Fatalf("recovery log exposed payment capability: %s", output.String())
	}
	if !strings.Contains(output.String(), "/mockpay/checkout/:paymentNo") {
		t.Fatalf("recovery log omitted route template: %s", output.String())
	}
}
