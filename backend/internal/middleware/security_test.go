package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	want := map[string]string{
		"Content-Security-Policy":   contentSecurityPolicy,
		"Permissions-Policy":        "camera=(), geolocation=(), microphone=()",
		"Referrer-Policy":           "no-referrer",
		"Strict-Transport-Security": strictTransportSecurity,
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
	}
	for header, value := range want {
		if got := rec.Header().Get(header); got != value {
			t.Errorf("%s = %q, want %q", header, got, value)
		}
	}
}

func TestNoStoreAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NoStoreAPI())
	for _, path := range []string{
		"/api/v1/orders",
		"/mockpay/api/v1/payments/1",
		"/mockpay/checkout/1",
		"/mocklogistics/api/v1/shipments/1",
	} {
		r.GET(path, func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
	}
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for _, path := range []string{
		"/api/v1/orders",
		"/mockpay/api/v1/payments/1",
		"/mockpay/checkout/1",
		"/mocklogistics/api/v1/shipments/1",
	} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if got := rec.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			if got := rec.Header().Get("Pragma"); got != "no-cache" {
				t.Errorf("Pragma = %q, want no-cache", got)
			}
		})
	}

	otherResponse := httptest.NewRecorder()
	r.ServeHTTP(otherResponse, httptest.NewRequest(http.MethodGet, "/health", nil))
	if got := otherResponse.Header().Get("Cache-Control"); got != "" {
		t.Errorf("health Cache-Control = %q, want empty", got)
	}
	if got := otherResponse.Header().Get("Pragma"); got != "" {
		t.Errorf("health Pragma = %q, want empty", got)
	}
}
