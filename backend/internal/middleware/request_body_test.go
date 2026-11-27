package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pinshaoshao/internal/http/httpx"

	"github.com/gin-gonic/gin"
)

func TestRequestBodyLimitAllowsBoundedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestBodyLimit(4))

	var body string
	r.POST("/", func(c *gin.Context) {
		raw, err := io.ReadAll(c.Request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		body = string(raw)
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test")))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if body != "test" {
		t.Fatalf("body = %q, want test", body)
	}
}

func TestRequestBodyLimitRejectsOversizedBody(t *testing.T) {
	tests := []struct {
		name          string
		contentLength int64
	}{
		{name: "known length", contentLength: 5},
		{name: "unknown length", contentLength: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(RequestBodyLimit(4))

			called := false
			r.POST("/", func(c *gin.Context) {
				called = true
				c.Status(http.StatusNoContent)
			})

			request := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader("12345"),
			)
			request.ContentLength = tt.contentLength
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, request)

			if rec.Code != http.StatusRequestEntityTooLarge {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
			}
			if called {
				t.Fatal("handler was called for an oversized request body")
			}

			var body httpx.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if body.Error != requestBodyTooLargeError {
				t.Fatalf("error = %q, want %q", body.Error, requestBodyTooLargeError)
			}
		})
	}
}
