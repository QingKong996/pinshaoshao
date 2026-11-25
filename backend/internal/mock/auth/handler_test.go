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
