package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"pinshaoshao/internal/mockpayment/model"
)

const (
	defaultNotificationBatchSize = 20
	defaultDeliveryLease         = 30 * time.Second
	defaultHTTPTimeout           = 5 * time.Second
	maxErrorBodyLength           = 1000
)

var defaultRetryDelays = []time.Duration{
	time.Second,
	5 * time.Second,
	30 * time.Second,
	2 * time.Minute,
	10 * time.Minute,
	30 * time.Minute,
}

type NotificationRepository interface {
	ClaimDue(ctx context.Context, now time.Time, lease time.Duration, limit int) ([]*model.Notification, error)
	MarkSucceeded(ctx context.Context, id int64, now time.Time) error
	MarkFailed(
		ctx context.Context,
		id int64,
		now time.Time,
		nextAttemptAt time.Time,
		lastError string,
	) (model.NotificationStatus, error)
	ListByPaymentNo(ctx context.Context, paymentNo string) ([]*model.Notification, error)
	Retry(ctx context.Context, paymentNo string, now time.Time) (*model.Notification, error)
}

type WebhookSigner interface {
	Sign(timestamp time.Time, body []byte) string
}

type NotificationConfig struct {
	Client      *http.Client
	Signer      WebhookSigner
	RetryDelays []time.Duration
	Lease       time.Duration
	BatchSize   int
}

type NotificationService struct {
	repository  NotificationRepository
	client      *http.Client
	signer      WebhookSigner
	retryDelays []time.Duration
	lease       time.Duration
	batchSize   int
	now         func() time.Time
}

func NewNotificationService(repository NotificationRepository, config NotificationConfig) *NotificationService {
	if config.Client == nil {
		config.Client = &http.Client{
			Timeout: defaultHTTPTimeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	if len(config.RetryDelays) == 0 {
		config.RetryDelays = defaultRetryDelays
	}
	if config.Lease <= 0 {
		config.Lease = defaultDeliveryLease
	}
	if config.BatchSize <= 0 {
		config.BatchSize = defaultNotificationBatchSize
	}
	return &NotificationService{
		repository:  repository,
		client:      config.Client,
		signer:      config.Signer,
		retryDelays: config.RetryDelays,
		lease:       config.Lease,
		batchSize:   config.BatchSize,
		now:         time.Now,
	}
}

func (s *NotificationService) DeliverDue(ctx context.Context, limit int) error {
	if limit <= 0 {
		limit = s.batchSize
	}
	notifications, err := s.repository.ClaimDue(ctx, s.now().UTC(), s.lease, limit)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if err := s.deliver(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

func (s *NotificationService) ListByPaymentNo(
	ctx context.Context,
	paymentNo string,
) ([]*model.Notification, error) {
	return s.repository.ListByPaymentNo(ctx, paymentNo)
}

func (s *NotificationService) Retry(ctx context.Context, paymentNo string) (*model.Notification, error) {
	return s.repository.Retry(ctx, paymentNo, s.now().UTC())
}

func (s *NotificationService) deliver(ctx context.Context, notification *model.Notification) error {
	now := s.now().UTC()
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		notification.TargetURL,
		bytes.NewReader(notification.Payload),
	)
	if err != nil {
		return s.finishFailure(ctx, notification, now, err)
	}

	timestamp := strconv.FormatInt(now.Unix(), 10)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "pinshaoshao-mockpayment/1.0")
	request.Header.Set("X-MockPay-Event-Id", notification.EventID)
	request.Header.Set("X-MockPay-Event-Type", notification.EventType)
	request.Header.Set("X-MockPay-Timestamp", timestamp)
	request.Header.Set("X-MockPay-Signature", s.signer.Sign(now, notification.Payload))
	request.Header.Set("X-MockPay-Delivery-Attempt", strconv.Itoa(notification.AttemptCount))

	response, err := s.client.Do(request)
	if err != nil {
		return s.finishFailure(ctx, notification, now, err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		if _, err := io.Copy(io.Discard, io.LimitReader(response.Body, maxErrorBodyLength)); err != nil {
			return fmt.Errorf("drain notification response: %w", err)
		}
		if err := s.repository.MarkSucceeded(ctx, notification.ID, now); err != nil {
			return err
		}
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyLength))
	message := fmt.Sprintf("webhook returned %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	return s.finishFailure(ctx, notification, now, errors.New(message))
}

func (s *NotificationService) finishFailure(
	ctx context.Context,
	notification *model.Notification,
	now time.Time,
	deliveryErr error,
) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	nextAttemptAt := now.Add(s.retryDelay(notification.AttemptCount))
	status, err := s.repository.MarkFailed(
		ctx,
		notification.ID,
		now,
		nextAttemptAt,
		truncateError(deliveryErr.Error()),
	)
	if err != nil {
		return err
	}
	slog.WarnContext(
		ctx,
		"mock payment webhook delivery failed",
		"event_id", notification.EventID,
		"attempt", notification.AttemptCount,
		"status", status,
		"error", deliveryErr,
	)
	return nil
}

func (s *NotificationService) retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	index := attempt - 1
	if index >= len(s.retryDelays) {
		index = len(s.retryDelays) - 1
	}
	return s.retryDelays[index]
}

func truncateError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= maxErrorBodyLength {
		return value
	}
	value = value[:maxErrorBodyLength]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}
