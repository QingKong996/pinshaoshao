package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pinshaoshao/internal/mock/logistics/model"
	"pinshaoshao/internal/webhooksig"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type notificationTestRepository struct {
	due           []*model.Notification
	failedMessage string
}

func (r *notificationTestRepository) ClaimDue(
	context.Context,
	time.Time,
	time.Duration,
	int,
) ([]*model.Notification, error) {
	return r.due, nil
}

func (r *notificationTestRepository) MarkSucceeded(context.Context, int64, time.Time) error {
	return nil
}

func (r *notificationTestRepository) MarkFailed(
	_ context.Context,
	_ int64,
	_ time.Time,
	_ time.Time,
	lastError string,
) (model.NotificationStatus, error) {
	r.failedMessage = lastError
	return model.NotificationStatusPending, nil
}

func (r *notificationTestRepository) ListByShipmentNo(
	context.Context,
	string,
) ([]*model.Notification, error) {
	return nil, nil
}

func (r *notificationTestRepository) Retry(
	context.Context,
	string,
	int64,
	time.Time,
) (*model.Notification, error) {
	return nil, nil
}

func TestNotificationFailureRedactsResponseBody(t *testing.T) {
	const responseBody = "sensitive callback response"
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, responseBody, http.StatusServiceUnavailable)
	}))
	defer server.Close()

	repository := &notificationTestRepository{
		due: []*model.Notification{{
			ID:           12,
			EventID:      "event-4",
			EventType:    model.EventTypeShipmentDelivered,
			TargetURL:    server.URL,
			Payload:      []byte(`{"id":"event-4"}`),
			AttemptCount: 1,
			MaxAttempts:  3,
		}},
	}
	notifier := NewNotificationService(repository, NotificationConfig{
		Signer: webhooksig.NewSigner("test-signing-key"),
		Policy: allowAllNotifyURLs{},
	})
	notifier.now = func() time.Time { return now }

	if err := notifier.DeliverDue(context.Background(), 10); err != nil {
		t.Fatalf("deliver notification: %v", err)
	}
	if !strings.Contains(repository.failedMessage, "503") {
		t.Fatalf("failure message = %q, want status code", repository.failedMessage)
	}
	if strings.Contains(repository.failedMessage, responseBody) {
		t.Fatalf("failure message leaked response body: %q", repository.failedMessage)
	}
}

func TestNotificationFailureRedactsRequestError(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	repository := &notificationTestRepository{
		due: []*model.Notification{{
			ID:           13,
			EventID:      "event-5",
			EventType:    model.EventTypeShipmentDelivered,
			TargetURL:    "https://allowed.example/webhook?token=sensitive",
			Payload:      []byte(`{"id":"event-5"}`),
			AttemptCount: 1,
			MaxAttempts:  3,
		}},
	}
	notifier := NewNotificationService(repository, NotificationConfig{
		Client: &http.Client{
			Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("request to %s failed", request.URL)
			}),
		},
		Signer: webhooksig.NewSigner("test-signing-key"),
		Policy: allowAllNotifyURLs{},
	})
	notifier.now = func() time.Time { return now }

	if err := notifier.DeliverDue(context.Background(), 10); err != nil {
		t.Fatalf("deliver notification: %v", err)
	}
	if repository.failedMessage != "webhook request failed" {
		t.Fatalf("failure message = %q, want sanitized summary", repository.failedMessage)
	}
	if strings.Contains(repository.failedMessage, "token") {
		t.Fatalf("failure message leaked request URL: %q", repository.failedMessage)
	}
}
