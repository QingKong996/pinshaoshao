package logistics

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	logisticshandler "pinshaoshao/internal/mock/logistics/handler"
	"pinshaoshao/internal/mock/logistics/model"
	logisticsrepository "pinshaoshao/internal/mock/logistics/repository"
	logisticsservice "pinshaoshao/internal/mock/logistics/service"
	"pinshaoshao/internal/webhooksig"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultNotificationInterval  = 500 * time.Millisecond
	defaultNotificationBatchSize = 20
)

type Config struct {
	WebhookSecret             string
	NotificationPollInterval  time.Duration
	NotificationBatchSize     int
	NotificationMaxAttempts   int
	NotificationRetryDelays   []time.Duration
	NotificationDeliveryLease time.Duration
	HTTPClient                *http.Client
}

func ConfigFromEnv() Config {
	secret := strings.TrimSpace(os.Getenv("MOCKLOGISTICS_WEBHOOK_SECRET"))
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("MOCKPAY_WEBHOOK_SECRET"))
	}
	return Config{WebhookSecret: secret}
}

func (c Config) withDefaults() Config {
	if c.NotificationPollInterval <= 0 {
		c.NotificationPollInterval = defaultNotificationInterval
	}
	if c.NotificationBatchSize <= 0 {
		c.NotificationBatchSize = defaultNotificationBatchSize
	}
	if c.NotificationMaxAttempts <= 0 {
		c.NotificationMaxAttempts = model.DefaultNotificationMaxAttempts
	}
	return c
}

type Module struct {
	shipments     *logisticshandler.ShipmentHandler
	notifications *logisticshandler.NotificationHandler
	delivery      *logisticsservice.NotificationService
	config        Config

	mu      sync.Mutex
	started bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewModule(pool *pgxpool.Pool, config Config) (*Module, error) {
	config = config.withDefaults()
	if strings.TrimSpace(config.WebhookSecret) == "" {
		return nil, errors.New("MOCKLOGISTICS_WEBHOOK_SECRET is required")
	}

	notificationRepository := logisticsrepository.NewNotificationRepository(pool)
	shipmentRepository := logisticsrepository.NewShipmentRepository(
		pool,
		notificationRepository,
		config.NotificationMaxAttempts,
	)
	shipmentService := logisticsservice.NewShipmentService(shipmentRepository)
	notificationService := logisticsservice.NewNotificationService(
		notificationRepository,
		logisticsservice.NotificationConfig{
			Client:      config.HTTPClient,
			Signer:      webhooksig.NewSigner(config.WebhookSecret),
			RetryDelays: config.NotificationRetryDelays,
			Lease:       config.NotificationDeliveryLease,
			BatchSize:   config.NotificationBatchSize,
		},
	)

	return &Module{
		shipments:     logisticshandler.NewShipmentHandler(shipmentService),
		notifications: logisticshandler.NewNotificationHandler(shipmentService, notificationService),
		delivery:      notificationService,
		config:        config,
	}, nil
}

func (m *Module) Register(r *gin.Engine, _ gin.HandlerFunc) {
	api := r.Group("/mocklogistics/api/v1/shipments")
	api.POST("", m.shipments.Create)
	api.GET("/:shipmentNo", m.shipments.Get)
	api.POST("/:shipmentNo/outbound", m.shipments.Outbound)
	api.POST("/:shipmentNo/advance", m.shipments.Advance)
	api.POST("/:shipmentNo/complete", m.shipments.Complete)
	api.GET("/:shipmentNo/notifications", m.notifications.List)
	api.POST("/:shipmentNo/notifications/:notificationID/retry", m.notifications.Retry)

	r.POST("/mocklogistics/webhook", m.notifications.WebhookSink)
}

func (m *Module) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return
	}

	workerCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.started = true
	m.wg.Add(1)
	go m.runNotificationWorker(workerCtx)
}

func (m *Module) Stop() {
	m.mu.Lock()
	if !m.started {
		m.mu.Unlock()
		return
	}
	cancel := m.cancel
	m.started = false
	m.mu.Unlock()

	cancel()
	m.wg.Wait()
}

func (m *Module) runNotificationWorker(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.config.NotificationPollInterval)
	defer ticker.Stop()

	for {
		if err := m.delivery.DeliverDue(ctx, m.config.NotificationBatchSize); err != nil &&
			ctx.Err() == nil {
			slog.ErrorContext(ctx, "deliver mock logistics notifications", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
