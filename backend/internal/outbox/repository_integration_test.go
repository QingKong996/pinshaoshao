//go:build integration

package outbox_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pinshaoshao/internal/outbox"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("pinshaoshao_outbox_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}

	dsn := container.MustConnectionString(ctx, "sslmode=disable")
	if err := migrate(ctx, dsn); err != nil {
		log.Fatalf("migrate outbox test database: %v", err)
	}
	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect outbox test database: %v", err)
	}

	code := m.Run()

	testPool.Close()
	if err := testcontainers.TerminateContainer(container); err != nil {
		log.Printf("terminate postgres container: %v", err)
	}
	os.Exit(code)
}

func migrate(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetLogger(goose.NopLogger())
	return goose.UpContext(ctx, db, filepath.Join("..", "..", "migrations"))
}

func resetOutbox(t *testing.T) {
	t.Helper()

	if _, err := testPool.Exec(
		context.Background(),
		`TRUNCATE outbox_events RESTART IDENTITY`,
	); err != nil {
		t.Fatalf("reset outbox: %v", err)
	}
}

func TestWriterEnqueueDeduplicatesByEventKey(t *testing.T) {
	resetOutbox(t)
	ctx := context.Background()
	writer := outbox.NewWriter()
	event := outbox.Event{
		Key:           "payment.close:1",
		Topic:         outbox.TopicPaymentCloseProvider,
		AggregateType: "payment",
		AggregateID:   1,
		Payload:       []byte(`{"provider_transaction_id":"provider-1"}`),
	}

	if err := writer.Enqueue(ctx, testPool, event); err != nil {
		t.Fatalf("enqueue first event: %v", err)
	}
	event.Topic = "other"
	event.AggregateID = 2
	event.Payload = []byte(`{"provider_transaction_id":"provider-2"}`)
	if err := writer.Enqueue(ctx, testPool, event); err != nil {
		t.Fatalf("enqueue duplicate event: %v", err)
	}

	var count int
	var topic string
	var aggregateID int64
	if err := testPool.QueryRow(
		ctx,
		`SELECT COUNT(*), MIN(topic), MIN(aggregate_id) FROM outbox_events`,
	).Scan(&count, &topic, &aggregateID); err != nil {
		t.Fatalf("read outbox events: %v", err)
	}
	if count != 1 || topic != outbox.TopicPaymentCloseProvider || aggregateID != 1 {
		t.Fatalf("stored count/topic/aggregate = %d/%q/%d", count, topic, aggregateID)
	}
}

func TestRepositoryClaimSkipsLockedEvents(t *testing.T) {
	resetOutbox(t)
	ctx := context.Background()
	writer := outbox.NewWriter()
	for index, key := range []string{"event-1", "event-2"} {
		if err := writer.Enqueue(ctx, testPool, outbox.Event{
			Key:           key,
			Topic:         outbox.TopicPaymentCloseProvider,
			AggregateType: "payment",
			AggregateID:   int64(index + 1),
			Payload:       []byte(`{"provider_transaction_id":"provider"}`),
		}); err != nil {
			t.Fatalf("enqueue %s: %v", key, err)
		}
	}

	tx, err := testPool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin lock transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	var lockedID int64
	if err := tx.QueryRow(
		ctx,
		`SELECT id FROM outbox_events WHERE event_key = 'event-1' FOR UPDATE`,
	).Scan(&lockedID); err != nil {
		t.Fatalf("lock first event: %v", err)
	}

	events, err := outbox.NewRepository(testPool).Claim(ctx, 10, time.Minute)
	if err != nil {
		t.Fatalf("claim events: %v", err)
	}
	if len(events) != 1 || events[0].Key != "event-2" {
		t.Fatalf("claimed events = %+v, want event-2 only", events)
	}
}

func TestRepositoryRetriesFailedEventThenDeadLetters(t *testing.T) {
	resetOutbox(t)
	ctx := context.Background()
	writer := outbox.NewWriter()
	if err := writer.Enqueue(ctx, testPool, outbox.Event{
		Key:           "event-1",
		Topic:         outbox.TopicPaymentCloseProvider,
		AggregateType: "payment",
		AggregateID:   1,
		Payload:       []byte(`{}`),
	}); err != nil {
		t.Fatalf("enqueue event: %v", err)
	}

	repository := outbox.NewRepository(testPool)
	first, err := repository.Claim(ctx, 1, time.Minute)
	if err != nil {
		t.Fatalf("claim first attempt: %v", err)
	}
	if len(first) != 1 || first[0].Attempts != 1 {
		t.Fatalf("first claim = %+v, want one attempt", first)
	}
	if err := repository.MarkFailed(
		ctx,
		first[0].ID,
		first[0].Attempts,
		2,
		time.Now(),
		"temporary",
		false,
	); err != nil {
		t.Fatalf("requeue event: %v", err)
	}

	second, err := repository.Claim(ctx, 1, time.Minute)
	if err != nil {
		t.Fatalf("claim second attempt: %v", err)
	}
	if len(second) != 1 || second[0].Attempts != 2 {
		t.Fatalf("second claim = %+v, want two attempts", second)
	}
	if err := repository.MarkFailed(
		ctx,
		second[0].ID,
		second[0].Attempts,
		2,
		time.Now(),
		"permanent",
		false,
	); err != nil {
		t.Fatalf("dead-letter event: %v", err)
	}

	third, err := repository.Claim(ctx, 1, time.Minute)
	if err != nil {
		t.Fatalf("claim dead event: %v", err)
	}
	if len(third) != 0 {
		t.Fatalf("dead event was claimed: %+v", third)
	}

	var status string
	if err := testPool.QueryRow(
		ctx,
		`SELECT status FROM outbox_events WHERE id = $1`,
		first[0].ID,
	).Scan(&status); err != nil {
		t.Fatalf("read event status: %v", err)
	}
	if status != "dead" {
		t.Fatalf("status = %q, want dead", status)
	}
}

func TestRepositoryMarkPublishedCompletesEvent(t *testing.T) {
	resetOutbox(t)
	ctx := context.Background()
	if err := outbox.NewWriter().Enqueue(ctx, testPool, outbox.Event{
		Key:           "event-1",
		Topic:         outbox.TopicPaymentCloseProvider,
		AggregateType: "payment",
		AggregateID:   1,
		Payload:       []byte(`{}`),
	}); err != nil {
		t.Fatalf("enqueue event: %v", err)
	}

	repository := outbox.NewRepository(testPool)
	events, err := repository.Claim(ctx, 1, time.Minute)
	if err != nil {
		t.Fatalf("claim event: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("claimed events = %d, want 1", len(events))
	}
	if err := repository.MarkPublished(ctx, events[0].ID, events[0].Attempts); err != nil {
		t.Fatalf("mark published: %v", err)
	}

	var status string
	var publishedAt *time.Time
	if err := testPool.QueryRow(
		ctx,
		`SELECT status, published_at FROM outbox_events WHERE id = $1`,
		events[0].ID,
	).Scan(&status, &publishedAt); err != nil {
		t.Fatalf("read published event: %v", err)
	}
	if status != "published" || publishedAt == nil {
		t.Fatalf("status/published_at = %q/%v, want published/non-nil", status, publishedAt)
	}

	var remaining int
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM outbox_events`).Scan(&remaining); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if remaining != 1 {
		t.Fatalf("event count = %d, want 1", remaining)
	}
}

func TestRepositoryIgnoresStaleClaimResult(t *testing.T) {
	resetOutbox(t)
	ctx := context.Background()
	if err := outbox.NewWriter().Enqueue(ctx, testPool, outbox.Event{
		Key:           "event-1",
		Topic:         outbox.TopicPaymentCloseProvider,
		AggregateType: "payment",
		AggregateID:   1,
		Payload:       []byte(`{}`),
	}); err != nil {
		t.Fatalf("enqueue event: %v", err)
	}

	repository := outbox.NewRepository(testPool)
	first, err := repository.Claim(ctx, 1, time.Minute)
	if err != nil || len(first) != 1 {
		t.Fatalf("claim first attempt = %+v/%v", first, err)
	}
	if _, err := testPool.Exec(
		ctx,
		`UPDATE outbox_events
SET attempts = 2, status = 'pending', available_at = now()
WHERE id = $1`,
		first[0].ID,
	); err != nil {
		t.Fatalf("simulate stale claim: %v", err)
	}

	if err := repository.MarkPublished(ctx, first[0].ID, first[0].Attempts); err != nil {
		t.Fatalf("mark stale claim published: %v", err)
	}
	if err := repository.MarkFailed(
		ctx,
		first[0].ID,
		first[0].Attempts,
		5,
		time.Now(),
		"stale",
		false,
	); err != nil {
		t.Fatalf("mark stale claim failed: %v", err)
	}

	var status string
	var attempts int
	if err := testPool.QueryRow(
		ctx,
		`SELECT status, attempts FROM outbox_events WHERE id = $1`,
		first[0].ID,
	).Scan(&status, &attempts); err != nil {
		t.Fatalf("read event: %v", err)
	}
	if status != "pending" || attempts != 2 {
		t.Fatalf("status/attempts = %q/%d, want pending/2", status, attempts)
	}
}
