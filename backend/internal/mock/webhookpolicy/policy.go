package webhookpolicy

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Allowlist accepts callback URLs whose scheme, host, and port match a configured origin.
type Allowlist struct {
	origins map[string]struct{}
}

func NewAllowlist(value string) (Allowlist, error) {
	origins := make(map[string]struct{})
	for _, value := range strings.Split(value, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if strings.Contains(value, "#") {
			return Allowlist{}, fmt.Errorf("callback origin %q must not contain a fragment", value)
		}

		parsed, err := url.ParseRequestURI(value)
		if err != nil {
			return Allowlist{}, fmt.Errorf("parse callback origin %q: %w", value, err)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return Allowlist{}, fmt.Errorf("callback origin %q must use http or https", value)
		}
		if parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" ||
			parsed.RawQuery != "" || parsed.ForceQuery ||
			(parsed.Path != "" && parsed.Path != "/") {
			return Allowlist{}, fmt.Errorf("callback origin %q must contain only scheme, host, and port", value)
		}

		origin, err := normalizeOrigin(parsed)
		if err != nil {
			return Allowlist{}, fmt.Errorf("parse callback origin %q: %w", value, err)
		}
		origins[origin] = struct{}{}
	}
	if len(origins) == 0 {
		return Allowlist{}, errors.New("callback origin allowlist is empty")
	}
	return Allowlist{origins: origins}, nil
}

// Allows reports whether value is an HTTP(S) URL with an allowed origin.
func (a Allowlist) Allows(value string) bool {
	if strings.Contains(value, "#") {
		return false
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	origin, err := normalizeOrigin(parsed)
	if err != nil {
		return false
	}
	_, ok := a.origins[origin]
	return ok
}

func normalizeOrigin(parsed *url.URL) (string, error) {
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", errors.New("scheme must be http or https")
	}

	hostname := strings.ToLower(parsed.Hostname())
	if hostname == "" {
		return "", errors.New("host is required")
	}
	if strings.HasSuffix(parsed.Host, ":") {
		return "", errors.New("port is empty")
	}

	port := parsed.Port()
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		port = ""
	}
	if strings.Contains(hostname, ":") {
		hostname = "[" + hostname + "]"
	}
	if port != "" {
		hostname = net.JoinHostPort(strings.Trim(hostname, "[]"), port)
	}
	return scheme + "://" + hostname, nil
}
