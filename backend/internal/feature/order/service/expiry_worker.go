package service

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	defaultExpiryInterval = time.Minute
	defaultExpiryBatch    = 100
)

type ExpiryWorker struct {
	service  ExpiryService
	interval time.Duration
	batch    int

	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

type ExpiryService interface {
	DueOrderIDs(ctx context.Context, limit int) ([]int, error)
	Expire(ctx context.Context, orderID int) error
}

func NewExpiryWorker(service ExpiryService, interval time.Duration, batch int) *ExpiryWorker {
	if interval <= 0 {
		interval = defaultExpiryInterval
	}
	if batch <= 0 {
		batch = defaultExpiryBatch
	}
	return &ExpiryWorker{
		service:  service,
		interval: interval,
		batch:    batch,
	}
}

func (w *ExpiryWorker) Start(parent context.Context) {
	if w == nil || w.service == nil || w.cancel != nil {
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

func (w *ExpiryWorker) Stop() {
	if w == nil || w.cancel == nil {
		return
	}
	w.once.Do(func() {
		w.cancel()
		<-w.done
	})
}

func (w *ExpiryWorker) run(ctx context.Context) {
	w.process(ctx)

	ticker := time.NewTicker(w.interval)
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

func (w *ExpiryWorker) process(ctx context.Context) {
	orderIDs, err := w.service.DueOrderIDs(ctx, w.batch)
	if err != nil {
		slog.ErrorContext(ctx, "list due orders", "error", err)
		return
	}
	for _, orderID := range orderIDs {
		if err := w.service.Expire(ctx, orderID); err != nil {
			slog.ErrorContext(ctx, "expire order", "order_id", orderID, "error", err)
		}
	}
}
