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
	r.GET("/api/v1/orders", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	apiResponse := httptest.NewRecorder()
	r.ServeHTTP(apiResponse, httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil))
	if got := apiResponse.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("API Cache-Control = %q, want no-store", got)
	}
	if got := apiResponse.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("API Pragma = %q, want no-cache", got)
	}

	otherResponse := httptest.NewRecorder()
	r.ServeHTTP(otherResponse, httptest.NewRequest(http.MethodGet, "/health", nil))
	if got := otherResponse.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("non-API Cache-Control = %q, want empty", got)
	}
}
