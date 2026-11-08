package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/model"

	"github.com/gin-gonic/gin"
)

const testPassword = "Password123"

type stubUserAuthenticator struct {
	user *model.User
}

func (s stubUserAuthenticator) Authenticate(context.Context, string) (*model.User, error) {
	return s.user, nil
}

type stubShopAuthenticator struct {
	shop *model.Shop
}

func (s stubShopAuthenticator) Authenticate(context.Context, string) (*model.Shop, error) {
	return s.shop, nil
}

func newControllerContext(t *testing.T, method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		ctx.Request.Header.Set("Content-Type", "application/json")
	}
	return ctx, rec
}

func doRequest(r http.Handler, method, path, authorization, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func requireStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
	}
}

func requireError(t *testing.T, rec *httptest.ResponseRecorder, want error) {
	t.Helper()

	var body dto.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error != want.Error() {
		t.Fatalf("error = %q, want %q", body.Error, want.Error())
	}
}

func requireToken(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body dto.TokenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	if body.Token == "" {
		t.Fatalf("response has no token: %s", rec.Body.String())
	}
	return body.Token
}
