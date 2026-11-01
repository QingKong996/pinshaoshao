package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCORSRouter(allowedOrigins string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS(allowedOrigins))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	return r
}

func doCORSRequest(r http.Handler, method, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/ping", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins string
		origin         string
		wantOrigin     string
	}{
		{
			name:           "all origins",
			allowedOrigins: "*",
			origin:         "https://docs.example.com",
			wantOrigin:     "*",
		},
		{
			name:           "listed origin",
			allowedOrigins: "https://a.example.com, https://b.example.com",
			origin:         "https://b.example.com",
			wantOrigin:     "https://b.example.com",
		},
		{
			name:           "unlisted origin",
			allowedOrigins: "https://a.example.com",
			origin:         "https://docs.example.com",
		},
		{
			name:           "no origin",
			allowedOrigins: "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doCORSRequest(newCORSRouter(tt.allowedOrigins), http.MethodGet, tt.origin)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantOrigin)
			}
		})
	}
}

func TestCORSPreflight(t *testing.T) {
	rec := doCORSRequest(newCORSRouter("https://docs.example.com"), http.MethodOptions, "https://docs.example.com")

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://docs.example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want configured origin", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != corsAllowMethods {
		t.Fatalf("Access-Control-Allow-Methods = %q, want %q", got, corsAllowMethods)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != corsAllowHeaders {
		t.Fatalf("Access-Control-Allow-Headers = %q, want %q", got, corsAllowHeaders)
	}
}
