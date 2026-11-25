package sessionclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientReloginsAndRetriesOnceAfterUnauthorized(t *testing.T) {
	var loginCalls atomic.Int32
	var requestCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/login" {
			call := loginCalls.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token": "token-" + string(rune('0'+call)),
			})
			return
		}

		call := requestCalls.Add(1)
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"value":"unchanged"}` {
			t.Fatalf("request body = %q", body)
		}
		if call == 1 {
			if got := r.Header.Get("Authorization"); got != "Bearer token-1" {
				t.Fatalf("first authorization = %q", got)
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-2" {
			t.Fatalf("retry authorization = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(
		server.URL,
		"/auth/login",
		"operator",
		"password-long-enough",
		nil,
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	response, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(
			http.MethodPost,
			server.URL+"/protected",
			bytes.NewBufferString(`{"value":"unchanged"}`),
		)
	})
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if loginCalls.Load() != 2 || requestCalls.Load() != 2 {
		t.Fatalf(
			"calls = login %d, request %d",
			loginCalls.Load(),
			requestCalls.Load(),
		)
	}
}
