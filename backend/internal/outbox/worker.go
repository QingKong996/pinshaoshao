package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const (
	defaultPollInterval = time.Second
	defaultBatchSize    = 50
	defaultLease        = 30 * time.Second
	defaultMaxAttempts  = 10
	maxBackoff          = 5 * time.Minute
)

type WorkerConfig struct {
	PollInterval time.Duration
	BatchSize    int
	Lease        time.Duration
	MaxAttempts  int
}

type WorkerRepository interface {
	Claim(ctx context.Context, limit int, lease time.Duration) ([]Event, error)
	MarkPublished(ctx context.Context, eventID int64, attempts int) error
	MarkFailed(
		ctx context.Context,
		eventID int64,
		attempts int,
		maxAttempts int,
		availableAt time.Time,
		cause string,
		permanent bool,
	) error
}

type Worker struct {
	repository   WorkerRepository
	handlers     map[string]Handler
	pollInterval time.Duration
	batchSize    int
	lease        time.Duration
	maxAttempts  int

	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

func NewWorker(repository WorkerRepository, handlers map[string]Handler, config WorkerConfig) *Worker {
	if config.PollInterval <= 0 {
		config.PollInterval = defaultPollInterval
	}
	if config.BatchSize <= 0 {
		config.BatchSize = defaultBatchSize
	}
	if config.Lease <= 0 {
		config.Lease = defaultLease
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = defaultMaxAttempts
	}
	return &Worker{
		repository:   repository,
		handlers:     handlers,
		pollInterval: config.PollInterval,
		batchSize:    config.BatchSize,
		lease:        config.Lease,
		maxAttempts:  config.MaxAttempts,
	}
}

func (w *Worker) Start(parent context.Context) {
	if w == nil || w.repository == nil {
		return
	}
	if w.cancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(parent)
	w.cancel = cancel
	w.done = make(chan struct{})
	go func() {
		defer close(w.done)
		w.run(ctx)
	}()
}

func (w *Worker) Stop() {
	if w == nil || w.cancel == nil {
		return
	}
	w.once.Do(func() {
		w.cancel()
		<-w.done
	})
}

func (w *Worker) run(ctx context.Context) {
	w.process(ctx)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *Worker) process(ctx context.Context) {
	events, err := w.repository.Claim(ctx, w.batchSize, w.lease)
	if err != nil {
		slog.ErrorContext(ctx, "claim outbox events", "error", err)
		return
	}

	for _, event := range events {
		w.handle(ctx, event)
	}
}

func (w *Worker) handle(ctx context.Context, event Event) {
	handler, ok := w.handlers[event.Topic]
	if !ok {
		err := fmt.Errorf("outbox topic %q has no handler", event.Topic)
		w.fail(ctx, event, err, true)
		return
	}

	if err := handler.Handle(ctx, event); err != nil {
		w.fail(ctx, event, err, IsPermanent(err))
		return
	}
	if err := w.repository.MarkPublished(ctx, event.ID, event.Attempts); err != nil {
		slog.ErrorContext(ctx, "mark outbox event published", "event_id", event.ID, "error", err)
	}
}

func (w *Worker) fail(ctx context.Context, event Event, err error, permanent bool) {
	availableAt := time.Now().Add(backoff(event.Attempts))
	if markErr := w.repository.MarkFailed(
		ctx,
		event.ID,
		event.Attempts,
		w.maxAttempts,
		availableAt,
		err.Error(),
		permanent,
	); markErr != nil {
		slog.ErrorContext(ctx, "mark outbox event failed", "event_id", event.ID, "error", markErr)
	}
	slog.ErrorContext(
		ctx,
		"handle outbox event",
		"event_id",
		event.ID,
		"topic",
		event.Topic,
		"attempts",
		event.Attempts,
		"permanent",
		permanent,
		"error",
		err,
	)
}

func backoff(attempts int) time.Duration {
	if attempts <= 1 {
		return time.Second
	}
	if attempts > 10 {
		return maxBackoff
	}
	delay := time.Second << (attempts - 1)
	if delay > maxBackoff {
		return maxBackoff
	}
	return delay
}
