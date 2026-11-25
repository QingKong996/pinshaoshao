package mocklogistics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pinshaoshao/internal/feature/logistics/apperr"
	"pinshaoshao/internal/feature/logistics/model"
	"pinshaoshao/internal/http/sessionclient"
)

const (
	loginPath          = "/mocklogistics/api/v1/auth/login"
	createShipmentPath = "/mocklogistics/api/v1/shipments"
	maxResponseBytes   = 64 << 10
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
		return nil, fmt.Errorf("invalid mock logistics gateway URL")
	}
	session, err := sessionclient.New(
		parsed.String(),
		loginPath,
		username,
		password,
		httpClient,
	)
	if err != nil {
		return nil, fmt.Errorf("create mock logistics session: %w", err)
	}
	return &Client{baseURL: parsed, session: session}, nil
}

type createShipmentRequest struct {
	MerchantOrderNo string `json:"merchant_order_no"`
	ItemSummary     string `json:"item_summary,omitempty"`
	Carrier         string `json:"carrier,omitempty"`
	ReceiverName    string `json:"receiver_name,omitempty"`
	ReceiverPhone   string `json:"receiver_phone,omitempty"`
	ReceiverAddress string `json:"receiver_address,omitempty"`
	NotifyURL       string `json:"notify_url"`
}

type shipmentResponse struct {
	ShipmentNo      string     `json:"shipment_no"`
	MerchantOrderNo string     `json:"merchant_order_no"`
	TrackingNo      string     `json:"tracking_no"`
	Carrier         string     `json:"carrier"`
	Status          string     `json:"status"`
	OriginCity      string     `json:"origin_city"`
	DestinationCity string     `json:"destination_city"`
	CurrentLocation string     `json:"current_location"`
	DispatchedAt    *time.Time `json:"dispatched_at"`
}

func (c *Client) Dispatch(
	ctx context.Context,
	input model.ProviderShipmentRequest,
) (*model.ProviderShipmentResult, error) {
	var payload bytes.Buffer
	err := json.NewEncoder(&payload).Encode(createShipmentRequest{
		MerchantOrderNo: input.MerchantOrderNo,
		ItemSummary:     input.ItemSummary,
		Carrier:         input.Carrier,
		ReceiverName:    input.ReceiverName,
		ReceiverPhone:   input.ReceiverPhone,
		ReceiverAddress: input.ReceiverAddress,
		NotifyURL:       input.NotifyURL,
	})
	if err != nil {
		return nil, fmt.Errorf("encode mock logistics request: %w", err)
	}

	created, err := c.post(ctx, c.baseURL.ResolveReference(&url.URL{Path: createShipmentPath}), payload.Bytes())
	if err != nil {
		return nil, err
	}
	if created.MerchantOrderNo != input.MerchantOrderNo {
		return nil, apperr.ErrProviderInvalid
	}

	outboundPath := createShipmentPath + "/" + url.PathEscape(created.ShipmentNo) + "/outbound"
	outbound, err := c.post(ctx, c.baseURL.ResolveReference(&url.URL{Path: outboundPath}), nil)
	if err != nil {
		return nil, err
	}
	if outbound.ShipmentNo != created.ShipmentNo ||
		outbound.MerchantOrderNo != input.MerchantOrderNo ||
		outbound.Status == "" ||
		outbound.Status == "pending_outbound" {
		return nil, apperr.ErrProviderInvalid
	}

	dispatchedAt := outbound.DispatchedAt
	return &model.ProviderShipmentResult{
		ShipmentNo:      outbound.ShipmentNo,
		MerchantNo:      outbound.MerchantOrderNo,
		TrackingNo:      outbound.TrackingNo,
		Carrier:         outbound.Carrier,
		Status:          model.Status(outbound.Status),
		OriginCity:      outbound.OriginCity,
		DestinationCity: outbound.DestinationCity,
		CurrentLocation: outbound.CurrentLocation,
		DispatchedAt:    dispatchedAt,
	}, nil
}

func (c *Client) post(
	ctx context.Context,
	endpoint *url.URL,
	payload []byte,
) (*shipmentResponse, error) {
	response, err := c.session.Do(ctx, func() (*http.Request, error) {
		var body io.Reader
		if payload != nil {
			body = bytes.NewReader(payload)
		}
		request, err := http.NewRequest(http.MethodPost, endpoint.String(), body)
		if err != nil {
			return nil, fmt.Errorf("create mock logistics request: %w", err)
		}
		request.Header.Set("Accept", "application/json")
		if payload != nil {
			request.Header.Set("Content-Type", "application/json")
		}
		return request, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrProviderUnavailable, err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: read response: %v", apperr.ErrProviderUnavailable, err)
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		if response.StatusCode >= http.StatusInternalServerError {
			return nil, fmt.Errorf(
				"%w: status %d: %s",
				apperr.ErrProviderUnavailable,
				response.StatusCode,
				strings.TrimSpace(string(responseBody)),
			)
		}
		return nil, fmt.Errorf(
			"%w: status %d: %s",
			apperr.ErrProviderInvalid,
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}

	var result shipmentResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("%w: decode response: %v", apperr.ErrProviderInvalid, err)
	}
	if strings.TrimSpace(result.ShipmentNo) == "" {
		return nil, apperr.ErrProviderInvalid
	}
	return &result, nil
}
