package outbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

type workerRepositoryStub struct {
	events      []Event
	claimErr    error
	published   []int64
	failures    []failure
	publishErr  error
	markFailErr error
}

type failure struct {
	eventID     int64
	attempts    int
	maxAttempts int
	availableAt time.Time
	cause       string
	permanent   bool
}

func (s *workerRepositoryStub) Claim(
	context.Context,
	int,
	time.Duration,
) ([]Event, error) {
	return s.events, s.claimErr
}

func (s *workerRepositoryStub) MarkPublished(
	_ context.Context,
	eventID int64,
	attempts int,
) error {
	_ = attempts
	s.published = append(s.published, eventID)
	return s.publishErr
}

func (s *workerRepositoryStub) MarkFailed(
	_ context.Context,
	eventID int64,
	attempts int,
	maxAttempts int,
	availableAt time.Time,
	cause string,
	permanent bool,
) error {
	s.failures = append(s.failures, failure{
		eventID:     eventID,
		attempts:    attempts,
		maxAttempts: maxAttempts,
		availableAt: availableAt,
		cause:       cause,
		permanent:   permanent,
	})
	return s.markFailErr
}

func TestWorkerPublishesSuccessfulEvent(t *testing.T) {
	repository := &workerRepositoryStub{
		events: []Event{{ID: 7, Topic: "known", Attempts: 1}},
	}
	worker := NewWorker(repository, map[string]Handler{
		"known": HandlerFunc(func(context.Context, Event) error {
			return nil
		}),
	}, WorkerConfig{})

	worker.process(context.Background())

	if len(repository.published) != 1 || repository.published[0] != 7 {
		t.Fatalf("published events = %v, want [7]", repository.published)
	}
	if len(repository.failures) != 0 {
		t.Fatalf("failures = %+v, want none", repository.failures)
	}
}

func TestWorkerMarksUnknownTopicAsDead(t *testing.T) {
	repository := &workerRepositoryStub{
		events: []Event{{ID: 7, Topic: "unknown", Attempts: 1}},
	}
	worker := NewWorker(repository, nil, WorkerConfig{MaxAttempts: 3})

	worker.process(context.Background())

	if len(repository.failures) != 1 {
		t.Fatalf("failures = %+v, want one", repository.failures)
	}
	got := repository.failures[0]
	if !got.permanent || got.maxAttempts != 3 || got.eventID != 7 {
		t.Fatalf("failure = %+v", got)
	}
}

func TestWorkerRetriesTransientFailureAndStopsOnPermanentFailure(t *testing.T) {
	transientErr := errors.New("provider unavailable")
	repository := &workerRepositoryStub{
		events: []Event{
			{ID: 7, Topic: "transient", Attempts: 2},
			{ID: 8, Topic: "permanent", Attempts: 1},
		},
	}
	worker := NewWorker(repository, map[string]Handler{
		"transient": HandlerFunc(func(context.Context, Event) error {
			return transientErr
		}),
		"permanent": HandlerFunc(func(context.Context, Event) error {
			return Permanent(errors.New("invalid payload"))
		}),
	}, WorkerConfig{MaxAttempts: 5})

	worker.process(context.Background())

	if len(repository.failures) != 2 {
		t.Fatalf("failures = %+v, want two", repository.failures)
	}
	if repository.failures[0].permanent {
		t.Fatalf("transient failure marked permanent: %+v", repository.failures[0])
	}
	if !repository.failures[1].permanent {
		t.Fatalf("permanent failure marked transient: %+v", repository.failures[1])
	}
	if repository.failures[0].cause != transientErr.Error() {
		t.Fatalf("failure cause = %q", repository.failures[0].cause)
	}
}

func TestWorkerBackoffIsBounded(t *testing.T) {
	if got := backoff(1); got != time.Second {
		t.Fatalf("backoff(1) = %s, want 1s", got)
	}
	if got := backoff(2); got != 2*time.Second {
		t.Fatalf("backoff(2) = %s, want 2s", got)
	}
	if got := backoff(100); got != maxBackoff {
		t.Fatalf("backoff(100) = %s, want %s", got, maxBackoff)
	}
}
