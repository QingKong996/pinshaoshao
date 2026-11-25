package webhookpolicy

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Allowlist accepts exact callback URLs after normalizing scheme, host, port, and path.
type Allowlist struct {
	urls map[string]struct{}
}

func NewAllowlist(value string) (Allowlist, error) {
	urls := make(map[string]struct{})
	for _, value := range strings.Split(value, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized, err := normalizeCallbackURL(value)
		if err != nil {
			return Allowlist{}, fmt.Errorf("parse callback URL %q: %w", value, err)
		}
		urls[normalized] = struct{}{}
	}
	if len(urls) == 0 {
		return Allowlist{}, errors.New("callback URL allowlist is empty")
	}
	return Allowlist{urls: urls}, nil
}

// Allows reports whether value exactly matches an allowed callback URL.
func (a Allowlist) Allows(value string) bool {
	normalized, err := normalizeCallbackURL(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	_, ok := a.urls[normalized]
	return ok
}

func normalizeCallbackURL(value string) (string, error) {
	if strings.Contains(value, "#") {
		return "", errors.New("fragment is not allowed")
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return "", err
	}
	if parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" ||
		parsed.RawQuery != "" || parsed.ForceQuery {
		return "", errors.New("URL must contain only scheme, host, port, and path")
	}
	if err := validateCallbackPath(parsed); err != nil {
		return "", err
	}

	origin, err := normalizeOrigin(parsed)
	if err != nil {
		return "", err
	}
	return origin + parsed.EscapedPath(), nil
}

func validateCallbackPath(parsed *url.URL) error {
	if parsed.Path == "" {
		return errors.New("path is required")
	}
	if strings.Contains(parsed.Path, "\\") || strings.Contains(parsed.Path, "//") {
		return errors.New("path must not contain empty segments or backslashes")
	}
	for _, segment := range strings.Split(parsed.Path, "/") {
		if segment == "." || segment == ".." {
			return errors.New("path must not contain dot segments")
		}
	}

	escapedPath := strings.ToLower(parsed.EscapedPath())
	if strings.Contains(escapedPath, "%2f") || strings.Contains(escapedPath, "%5c") {
		return errors.New("path must not contain encoded separators")
	}
	return nil
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
