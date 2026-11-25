package sessionclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const maxLoginResponseBytes = 64 << 10

var ErrLoginFailed = errors.New("mock service login failed")

type Client struct {
	baseURL    *url.URL
	loginPath  string
	username   string
	password   string
	httpClient *http.Client

	mu    sync.Mutex
	token string
}

func New(
	baseURL string,
	loginPath string,
	username string,
	password string,
	httpClient *http.Client,
) (*Client, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Host == "" {
		return nil, errors.New("invalid mock service URL")
	}
	if strings.TrimSpace(loginPath) == "" {
		return nil, errors.New("mock service login path is required")
	}
	if strings.TrimSpace(username) == "" || password == "" {
		return nil, errors.New("mock service credentials are required")
	}
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	return &Client{
		baseURL:    parsed,
		loginPath:  loginPath,
		username:   strings.TrimSpace(username),
		password:   password,
		httpClient: httpClient,
	}, nil
}

func (c *Client) Do(
	ctx context.Context,
	newRequest func() (*http.Request, error),
) (*http.Response, error) {
	token, err := c.tokenFor(ctx, "")
	if err != nil {
		return nil, err
	}

	response, err := c.do(ctx, newRequest, token)
	if err != nil || response.StatusCode != http.StatusUnauthorized {
		return response, err
	}
	_ = response.Body.Close()

	token, err = c.tokenFor(ctx, token)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, newRequest, token)
}

func (c *Client) do(
	ctx context.Context,
	newRequest func() (*http.Request, error),
	token string,
) (*http.Response, error) {
	request, err := newRequest()
	if err != nil {
		return nil, err
	}
	request = request.WithContext(ctx)
	request.Header.Set("Authorization", "Bearer "+token)
	return c.httpClient.Do(request)
}

func (c *Client) tokenFor(ctx context.Context, staleToken string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && c.token != staleToken {
		return c.token, nil
	}
	c.token = ""

	token, err := c.login(ctx)
	if err != nil {
		return "", err
	}
	c.token = token
	return token, nil
}

func (c *Client) login(ctx context.Context) (string, error) {
	payload, err := json.Marshal(map[string]string{
		"username": c.username,
		"password": c.password,
	})
	if err != nil {
		return "", fmt.Errorf("%w: encode credentials", ErrLoginFailed)
	}

	endpoint := c.baseURL.ResolveReference(&url.URL{Path: c.loginPath})
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint.String(),
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("%w: create login request: %v", ErrLoginFailed, err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrLoginFailed, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxLoginResponseBytes))
	if err != nil {
		return "", fmt.Errorf("%w: read login response: %v", ErrLoginFailed, err)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"%w: status %d: %s",
			ErrLoginFailed,
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &result); err != nil ||
		strings.TrimSpace(result.Token) == "" {
		return "", fmt.Errorf("%w: invalid login response", ErrLoginFailed)
	}
	return result.Token, nil
}
