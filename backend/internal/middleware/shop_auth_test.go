package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubShopAuthenticator struct {
	shop  *model.Shop
	err   error
	calls int
	token string
}

func (s *stubShopAuthenticator) Authenticate(_ context.Context, token string) (*model.Shop, error) {
	s.calls++
	s.token = token
	return s.shop, s.err
}

func TestShopAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	activeShop := &model.Shop{ID: 1, Username: "shop01", Status: model.ShopStatusActive}

	tests := []struct {
		name           string
		authorization  string
		authenticator  *stubShopAuthenticator
		wantStatus     int
		wantError      string
		wantHandler    bool
		wantCalls      int
		wantTokenSent  string
		wantTokenInCtx string
		wantUsername   string
	}{
		{
			name:          "missing header",
			authenticator: &stubShopAuthenticator{},
			wantStatus:    http.StatusUnauthorized,
			wantError:     apperr.ErrUnauthorized.Error(),
		},
		{
			name:          "wrong scheme",
			authorization: "Basic abc",
			authenticator: &stubShopAuthenticator{},
			wantStatus:    http.StatusUnauthorized,
			wantError:     apperr.ErrUnauthorized.Error(),
		},
		{
			name:          "blank token",
			authorization: "Bearer    ",
			authenticator: &stubShopAuthenticator{},
			wantStatus:    http.StatusUnauthorized,
			wantError:     apperr.ErrUnauthorized.Error(),
		},
		{
			name:          "invalid session",
			authorization: "Bearer expired-token",
			authenticator: &stubShopAuthenticator{err: apperr.ErrInvalidShopSession},
			wantStatus:    http.StatusUnauthorized,
			wantError:     apperr.ErrInvalidShopSession.Error(),
			wantCalls:     1,
			wantTokenSent: "expired-token",
		},
		{
			name:          "authenticator requires authentication",
			authorization: "Bearer empty-token",
			authenticator: &stubShopAuthenticator{err: apperr.ErrUnauthorized},
			wantStatus:    http.StatusUnauthorized,
			wantError:     apperr.ErrUnauthorized.Error(),
			wantCalls:     1,
			wantTokenSent: "empty-token",
		},
		{
			name:          "banned shop",
			authorization: "Bearer banned-token",
			authenticator: &stubShopAuthenticator{err: apperr.ErrShopBanned},
			wantStatus:    http.StatusForbidden,
			wantError:     apperr.ErrShopBanned.Error(),
			wantCalls:     1,
			wantTokenSent: "banned-token",
		},
		{
			name:          "authenticator failure",
			authorization: "Bearer some-token",
			authenticator: &stubShopAuthenticator{err: errors.New("database is down")},
			wantStatus:    http.StatusInternalServerError,
			wantError:     apperr.ErrInternalServer.Error(),
			wantCalls:     1,
			wantTokenSent: "some-token",
		},
		{
			name:           "valid token",
			authorization:  "Bearer good-token",
			authenticator:  &stubShopAuthenticator{shop: activeShop},
			wantStatus:     http.StatusOK,
			wantHandler:    true,
			wantCalls:      1,
			wantTokenSent:  "good-token",
			wantTokenInCtx: "good-token",
			wantUsername:   "shop01",
		},
		{
			name:           "scheme is case insensitive",
			authorization:  "bearer good-token",
			authenticator:  &stubShopAuthenticator{shop: activeShop},
			wantStatus:     http.StatusOK,
			wantHandler:    true,
			wantCalls:      1,
			wantTokenSent:  "good-token",
			wantTokenInCtx: "good-token",
			wantUsername:   "shop01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false

			r := gin.New()
			r.GET("/protected", ShopAuth(tt.authenticator), func(c *gin.Context) {
				handlerCalled = true

				shop, ok := CurrentShop(c)
				if !ok {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "shop missing from context"})
					return
				}
				token := MustShopSessionToken(c)
				c.JSON(http.StatusOK, gin.H{
					"username": shop.Username,
					"token":    token,
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if handlerCalled != tt.wantHandler {
				t.Fatalf("handler called = %v, want %v", handlerCalled, tt.wantHandler)
			}
			if tt.authenticator.calls != tt.wantCalls {
				t.Fatalf("authenticator calls = %d, want %d", tt.authenticator.calls, tt.wantCalls)
			}
			if tt.authenticator.token != tt.wantTokenSent {
				t.Fatalf("authenticator token = %q, want %q", tt.authenticator.token, tt.wantTokenSent)
			}
			if tt.wantStatus == http.StatusUnauthorized {
				if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
					t.Fatalf("WWW-Authenticate = %q, want %q", got, "Bearer")
				}
			} else if got := rec.Header().Get("WWW-Authenticate"); got != "" {
				t.Fatalf("WWW-Authenticate = %q, want no challenge", got)
			}

			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if tt.wantError == "" {
				if _, ok := body["error"]; ok {
					t.Fatalf("unexpected error field: %q", body["error"])
				}
			} else if body["error"] != tt.wantError {
				t.Fatalf("error = %q, want %q", body["error"], tt.wantError)
			}
			if tt.wantHandler {
				if body["username"] != tt.wantUsername {
					t.Fatalf("username = %q, want %q", body["username"], tt.wantUsername)
				}
				if body["token"] != tt.wantTokenInCtx {
					t.Fatalf("token = %q, want %q", body["token"], tt.wantTokenInCtx)
				}
			}
		})
	}
}

func TestMustShopSessionTokenPanicsWithoutAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/unprotected", func(c *gin.Context) {
		_ = MustShopSessionToken(c)
	})

	defer func() {
		if recover() == nil {
			t.Fatal("expected MustShopSessionToken to panic outside ShopAuth")
		}
	}()

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/unprotected", nil))
}
