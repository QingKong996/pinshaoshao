package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubAuthService struct {
	authenticateErr error
}

func (s stubAuthService) Login(context.Context, string, string) (string, error) {
	return "mockpay-session-token", nil
}

func (s stubAuthService) Authenticate(context.Context, string) (*Account, error) {
	return &Account{ID: 1, Username: "operator"}, s.authenticateErr
}

func (s stubAuthService) Logout(context.Context, string) error {
	return nil
}

func TestMiddlewareRequiresBearerSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, test := range []struct {
		name          string
		authorization string
		service       authService
		wantError     string
	}{
		{
			name:      "missing bearer",
			service:   stubAuthService{},
			wantError: ErrUnauthorized.Error(),
		},
		{
			name:          "invalid session",
			authorization: "Bearer stale-token",
			service:       stubAuthService{authenticateErr: ErrInvalidSession},
			wantError:     ErrInvalidSession.Error(),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(test.service)
			router := gin.New()
			router.GET("/protected", handler.Middleware(), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if test.authorization != "" {
				request.Header.Set("Authorization", test.authorization)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d", response.Code)
			}
			var body errorResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Error != test.wantError {
				t.Fatalf("error = %q, want %q", body.Error, test.wantError)
			}
		})
	}
}

func TestLoginRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(stubAuthService{})
	router := gin.New()
	router.POST("/login", handler.Login)

	request := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"username":`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestLocalIPOnlyBlocksRemoteMockLogins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", LocalIPOnly(), NewHandler(stubAuthService{}).Login)

	for _, test := range []struct {
		name       string
		remoteAddr string
		forwarded  string
		wantStatus int
	}{
		{name: "IPv4 loopback", remoteAddr: "127.0.0.1:1234", wantStatus: http.StatusOK},
		{name: "IPv6 loopback", remoteAddr: "[::1]:1234", wantStatus: http.StatusOK},
		{name: "remote address", remoteAddr: "192.0.2.10:1234", wantStatus: http.StatusForbidden},
		{name: "spoofed forwarded address", remoteAddr: "192.0.2.10:1234", forwarded: "127.0.0.1", wantStatus: http.StatusForbidden},
		{name: "loopback proxy", remoteAddr: "127.0.0.1:1234", forwarded: "192.0.2.10", wantStatus: http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"operator"}`))
			request.RemoteAddr = test.remoteAddr
			request.Header.Set("Content-Type", "application/json")
			if test.forwarded != "" {
				request.Header.Set("X-Forwarded-For", test.forwarded)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}
