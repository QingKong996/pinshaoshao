package webhookpolicy

import "testing"

func TestAllowlist(t *testing.T) {
	allowlist, err := NewAllowlist(
		"https://example.com/webhooks/payment, http://localhost:8080/internal/webhook",
	)
	if err != nil {
		t.Fatalf("create allowlist: %v", err)
	}

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "exact URL",
			value: "https://example.com/webhooks/payment",
			want:  true,
		},
		{
			name:  "case and default port",
			value: "HTTPS://EXAMPLE.COM:443/webhooks/payment",
			want:  true,
		},
		{
			name:  "loopback local development",
			value: "http://localhost:8080/internal/webhook",
			want:  true,
		},
		{
			name:  "same origin different path",
			value: "https://example.com/webhooks/feed",
			want:  false,
		},
		{
			name:  "path prefix",
			value: "https://example.com/webhooks/payment/extra",
			want:  false,
		},
		{
			name:  "trailing slash",
			value: "https://example.com/webhooks/payment/",
			want:  false,
		},
		{
			name:  "encoded path",
			value: "https://example.com/webhooks/%70ayment",
			want:  false,
		},
		{
			name:  "dot segment",
			value: "https://example.com/webhooks/payment/../feed",
			want:  false,
		},
		{
			name:  "query",
			value: "https://example.com/webhooks/payment?token=1",
			want:  false,
		},
		{
			name:  "fragment",
			value: "https://example.com/webhooks/payment#target",
			want:  false,
		},
		{
			name:  "userinfo",
			value: "https://user@example.com/webhooks/payment",
			want:  false,
		},
		{
			name:  "origin only",
			value: "https://example.com",
			want:  false,
		},
		{
			name:  "different scheme",
			value: "http://example.com/webhooks/payment",
			want:  false,
		},
		{
			name:  "different port",
			value: "http://localhost:8081/internal/webhook",
			want:  false,
		},
		{
			name:  "lookalike host",
			value: "https://example.com.evil.test/webhooks/payment",
			want:  false,
		},
		{
			name:  "unsupported scheme",
			value: "ftp://example.com/webhooks/payment",
			want:  false,
		},
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
		"https://example.com",
		"https://example.com?query=1",
		"https://example.com/path?query=1",
		"https://example.com/path#fragment",
		"https://example.com:bad/path",
		"https://example.com:/path",
		"ftp://example.com/path",
		"https://user@example.com/path",
		"https://example.com/api/../feed",
		"https://example.com/api/%2e%2e/feed",
		"https://example.com/api/%2Ffeed",
	} {
		t.Run(value, func(t *testing.T) {
			if _, err := NewAllowlist(value); err == nil {
				t.Fatalf("NewAllowlist(%q) returned no error", value)
			}
		})
	}
}
