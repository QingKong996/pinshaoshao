package webhookpolicy

import "testing"

func TestAllowlist(t *testing.T) {
	allowlist, err := NewAllowlist("https://example.com, http://localhost:8080")
	if err != nil {
		t.Fatalf("create allowlist: %v", err)
	}

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "configured origin", value: "https://example.com/webhooks/payment", want: true},
		{name: "case and default port", value: "HTTPS://EXAMPLE.COM:443/webhooks/payment", want: true},
		{name: "loopback local development", value: "http://localhost:8080/internal/webhook", want: true},
		{name: "different scheme", value: "http://example.com/webhook", want: false},
		{name: "different port", value: "http://localhost:8081/webhook", want: false},
		{name: "lookalike host", value: "https://example.com.evil.test/webhook", want: false},
		{name: "userinfo", value: "https://user@example.com/webhook", want: false},
		{name: "fragment", value: "https://example.com/webhook#target", want: false},
		{name: "unsupported scheme", value: "ftp://example.com/webhook", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allowlist.Allows(tt.value); got != tt.want {
				t.Fatalf("Allows(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestNewAllowlistRejectsInvalidConfiguration(t *testing.T) {
	for _, value := range []string{
		"",
		"*",
		"https://example.com/path",
		"https://example.com?query=1",
		"https://example.com:bad",
		"https://example.com:",
		"ftp://example.com",
	} {
		t.Run(value, func(t *testing.T) {
			if _, err := NewAllowlist(value); err == nil {
				t.Fatalf("NewAllowlist(%q) returned no error", value)
			}
		})
	}
}
