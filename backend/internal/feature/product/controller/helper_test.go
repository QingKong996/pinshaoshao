package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pinshaoshao/internal/feature/product/dto"
	shopmodel "pinshaoshao/internal/feature/shop/model"
)

const testShopUsername = "shop01"

type stubShopAuthenticator struct {
	shop *shopmodel.Shop
}

func (s stubShopAuthenticator) Authenticate(context.Context, string) (*shopmodel.Shop, error) {
	return s.shop, nil
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

	if got := decodeError(t, rec); got != want.Error() {
		t.Fatalf("error = %q, want %q", got, want.Error())
	}
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body dto.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	return body.Error
}
