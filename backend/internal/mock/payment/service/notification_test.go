package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"pinshaoshao/internal/mock/payment/model"
	"pinshaoshao/internal/webhooksig"
)

type notificationTestRepository struct {
	due           []*model.Notification
	succeededID   int64
	failedID      int64
	failedNextAt  time.Time
	failedMessage string
}

func (r *notificationTestRepository) ClaimDue(context.Context, time.Time, time.Duration, int) ([]*model.Notification, error) {
	return r.due, nil
}

func (r *notificationTestRepository) MarkSucceeded(_ context.Context, id int64, _ time.Time) error {
	r.succeededID = id
	return nil
}

func (r *notificationTestRepository) MarkFailed(
	_ context.Context,
	id int64,
	_ time.Time,
	nextAttemptAt time.Time,
	lastError string,
) (model.NotificationStatus, error) {
	r.failedID = id
	r.failedNextAt = nextAttemptAt
	r.failedMessage = lastError
	return model.NotificationStatusPending, nil
}

func (r *notificationTestRepository) ListByPaymentNo(context.Context, string) ([]*model.Notification, error) {
	return nil, nil
}

func (r *notificationTestRepository) Retry(context.Context, string, time.Time) (*model.Notification, error) {
	return nil, nil
}

func TestNotificationDeliveryIsSigned(t *testing.T) {
	const signingKey = "integration-test-signing-key"
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	payload := []byte(`{"id":"event-1","type":"payment.paid"}`)
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var err error
		receivedBody, err = io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read callback body: %v", err)
		}
		if got := request.Header.Get("X-MockPay-Event-Id"); got != "event-1" {
			t.Fatalf("event id header = %q", got)
		}
		if got := request.Header.Get("X-MockPay-Delivery-Attempt"); got != "1" {
			t.Fatalf("attempt header = %q, want 1", got)
		}
		timestamp := request.Header.Get("X-MockPay-Timestamp")
		if _, err := strconv.ParseInt(timestamp, 10, 64); err != nil {
			t.Fatalf("timestamp header = %q", timestamp)
		}
		if err := webhooksig.Verify(
			signingKey,
			timestamp,
			request.Header.Get("X-MockPay-Signature"),
			receivedBody,
			now,
			time.Minute,
		); err != nil {
			t.Fatalf("verify callback signature: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	repository := &notificationTestRepository{
		due: []*model.Notification{{
			ID:           9,
			EventID:      "event-1",
			EventType:    model.EventTypePaymentPaid,
			TargetURL:    server.URL,
			Payload:      payload,
			AttemptCount: 1,
			MaxAttempts:  3,
		}},
	}
	notifier := NewNotificationService(repository, NotificationConfig{
		Signer: webhooksig.NewSigner(signingKey),
		Policy: allowAllNotifyURLs{},
	})
	notifier.now = func() time.Time { return now }

	if err := notifier.DeliverDue(context.Background(), 10); err != nil {
		t.Fatalf("deliver notification: %v", err)
	}
	if !strings.EqualFold(string(receivedBody), string(payload)) {
		t.Fatalf("callback body = %s, want %s", receivedBody, payload)
	}
	if repository.succeededID != 9 {
		t.Fatalf("succeeded id = %d, want 9", repository.succeededID)
	}
}

func TestNotificationFailureSchedulesRetry(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "temporarily unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	repository := &notificationTestRepository{
		due: []*model.Notification{{
			ID:           10,
			EventID:      "event-2",
			EventType:    model.EventTypePaymentPaid,
			TargetURL:    server.URL,
			Payload:      []byte(`{"id":"event-2"}`),
			AttemptCount: 2,
			MaxAttempts:  3,
		}},
	}
	notifier := NewNotificationService(repository, NotificationConfig{
		Signer:      webhooksig.NewSigner("test-signing-key"),
		Policy:      allowAllNotifyURLs{},
		RetryDelays: []time.Duration{time.Second, 5 * time.Second},
	})
	notifier.now = func() time.Time { return now }

	if err := notifier.DeliverDue(context.Background(), 10); err != nil {
		t.Fatalf("deliver notification: %v", err)
	}
	if repository.failedID != 10 {
		t.Fatalf("failed id = %d, want 10", repository.failedID)
	}
	if !repository.failedNextAt.Equal(now.Add(5 * time.Second)) {
		t.Fatalf("next attempt = %s, want %s", repository.failedNextAt, now.Add(5*time.Second))
	}
	if !strings.Contains(repository.failedMessage, "503") {
		t.Fatalf("failure message = %q, want status code", repository.failedMessage)
	}
}

func TestTruncateErrorKeepsValidUTF8(t *testing.T) {
	value := strings.Repeat("a", maxErrorBodyLength-1) +
		"汉" +
		strings.Repeat("b", 10)

	truncated := truncateError(value)

	if !utf8.ValidString(truncated) {
		t.Fatalf("truncated error is not valid UTF-8: %q", truncated)
	}
	if len(truncated) > maxErrorBodyLength {
		t.Fatalf("truncated length = %d, want <= %d", len(truncated), maxErrorBodyLength)
	}
}
