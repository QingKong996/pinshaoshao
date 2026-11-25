package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAccessLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	accessLogger := newAccessLogger(logger, 4)
	r := gin.New()
	r.Use(accessLogger.Middleware())
	r.GET("/orders/:orderID", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	request.RemoteAddr = "192.0.2.10:1234"
	request.Header.Set("User-Agent", "access-log-test")
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, request)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := accessLogger.Close(ctx); err != nil {
		t.Fatalf("close access logger: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode access log: %v", err)
	}
	if entry["msg"] != "http request" ||
		entry["method"] != http.MethodGet ||
		entry["route"] != "/orders/:orderID" ||
		entry["status"] != float64(http.StatusCreated) ||
		entry["client_ip"] != "192.0.2.10" ||
		entry["user_agent"] != "access-log-test" {
		t.Fatalf("access log = %v", entry)
	}
}
