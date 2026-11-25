package mockpay

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
	"time"

	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/model"
	"pinshaoshao/internal/http/sessionclient"
)

const (
	loginPath              = "/mockpay/api/v1/auth/login"
	createPaymentPath      = "/mockpay/api/v1/payments"
	closePaymentPathPrefix = "/mockpay/api/v1/payments/"
	maxResponseBytes       = 64 << 10
)

type Client struct {
	baseURL *url.URL
	session *sessionclient.Client
}

func NewClient(
	baseURL string,
	username string,
	password string,
	httpClient *http.Client,
) (*Client, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Host == "" {
		return nil, fmt.Errorf("invalid mock payment gateway URL")
	}
	session, err := sessionclient.New(
		parsed.String(),
		loginPath,
		username,
		password,
		httpClient,
	)
	if err != nil {
		return nil, fmt.Errorf("create mock payment session: %w", err)
	}
	return &Client{baseURL: parsed, session: session}, nil
}

type createPaymentRequest struct {
	MerchantOrderNo  string `json:"merchant_order_no"`
	AmountCents      int64  `json:"amount_cents"`
	Currency         string `json:"currency"`
	Subject          string `json:"subject"`
	NotifyURL        string `json:"notify_url"`
	ReturnURL        string `json:"return_url,omitempty"`
	ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
}

type createPaymentResponse struct {
	PaymentNo       string    `json:"payment_no"`
	MerchantOrderNo string    `json:"merchant_order_no"`
	Status          string    `json:"status"`
	AmountCents     int64     `json:"amount_cents"`
	Currency        string    `json:"currency"`
	CheckoutURL     string    `json:"checkout_url"`
	ExpiresAt       time.Time `json:"expires_at"`
}

func (c *Client) CreatePayment(
	ctx context.Context,
	input model.ProviderPaymentRequest,
) (*model.ProviderPaymentResult, error) {
	payload, err := json.Marshal(createPaymentRequest{
		MerchantOrderNo:  input.MerchantOrderNo,
		AmountCents:      input.AmountCents,
		Currency:         input.Currency,
		Subject:          input.Subject,
		NotifyURL:        input.NotifyURL,
		ReturnURL:        input.ReturnURL,
		ExpiresInSeconds: int(input.ExpiresIn.Seconds()),
	})
	if err != nil {
		return nil, fmt.Errorf("encode mock payment request: %w", err)
	}

	endpoint := c.baseURL.ResolveReference(&url.URL{Path: createPaymentPath})
	response, err := c.session.Do(ctx, func() (*http.Request, error) {
		request, err := http.NewRequest(
			http.MethodPost,
			endpoint.String(),
			bytes.NewReader(payload),
		)
		if err != nil {
			return nil, fmt.Errorf("create mock payment request: %w", err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json")
		return request, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrPaymentProviderUnavailable, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: read response: %v", apperr.ErrPaymentProviderUnavailable, err)
	}
	if response.StatusCode == http.StatusConflict {
		return nil, apperr.ErrPaymentConflict
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf(
			"%w: status %d: %s",
			apperr.ErrPaymentProviderUnavailable,
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var result createPaymentResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("%w: decode response: %v", apperr.ErrPaymentProviderInvalid, err)
	}
	checkoutURL, err := c.resolveCheckoutURL(result.CheckoutURL)
	if err != nil || checkoutURL == "" {
		return nil, apperr.ErrPaymentProviderInvalid
	}
	return &model.ProviderPaymentResult{
		PaymentNo:   result.PaymentNo,
		MerchantNo:  result.MerchantOrderNo,
		Status:      result.Status,
		AmountCents: result.AmountCents,
		Currency:    result.Currency,
		CheckoutURL: checkoutURL,
		ExpiresAt:   result.ExpiresAt,
	}, nil
}

func (c *Client) ClosePayment(ctx context.Context, providerPaymentNo string) error {
	providerPaymentNo = strings.TrimSpace(providerPaymentNo)
	if providerPaymentNo == "" {
		return apperr.ErrPaymentProviderInvalid
	}

	endpoint := c.baseURL.ResolveReference(&url.URL{
		Path: closePaymentPathPrefix + url.PathEscape(providerPaymentNo) + "/close",
	})
	response, err := c.session.Do(ctx, func() (*http.Request, error) {
		request, err := http.NewRequest(http.MethodPost, endpoint.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("create mock payment close request: %w", err)
		}
		request.Header.Set("Accept", "application/json")
		return request, nil
	})
	if err != nil {
		return fmt.Errorf("%w: %v", apperr.ErrPaymentProviderUnavailable, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("%w: read response: %v", apperr.ErrPaymentProviderUnavailable, err)
	}
	if response.StatusCode == http.StatusConflict {
		return fmt.Errorf(
			"%w: status %d: %s",
			apperr.ErrPaymentConflict,
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"%w: status %d: %s",
			apperr.ErrPaymentProviderUnavailable,
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}
	return nil
}

func (c *Client) resolveCheckoutURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("empty checkout URL")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	if !parsed.IsAbs() {
		parsed = c.baseURL.ResolveReference(parsed)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("unsupported checkout URL scheme")
	}
	if parsed.Host == "" || parsed.User != nil {
		return "", errors.New("invalid checkout URL")
	}
	return parsed.String(), nil
}
