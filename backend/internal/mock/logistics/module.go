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

	mockauth "pinshaoshao/internal/mock/auth"
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
	defaultDeliveryDuration      = 4 * time.Minute
	maxDeliveryDuration          = 5 * time.Minute
	defaultShipmentPollInterval  = 500 * time.Millisecond
	defaultShipmentBatchSize     = 100
)

type Config struct {
	WebhookSecret             string
	DeliveryDuration          time.Duration
	ShipmentPollInterval      time.Duration
	ShipmentBatchSize         int
	NotificationPollInterval  time.Duration
	NotificationBatchSize     int
	NotificationMaxAttempts   int
	NotificationRetryDelays   []time.Duration
	NotificationDeliveryLease time.Duration
	HTTPClient                *http.Client
	Auth                      mockauth.Config
}

func ConfigFromEnv() Config {
	secret := strings.TrimSpace(os.Getenv("MOCKLOGISTICS_WEBHOOK_SECRET"))
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("MOCKPAY_WEBHOOK_SECRET"))
	}
	config := Config{
		WebhookSecret: secret,
		Auth:          mockauth.ConfigFromEnv("MOCKLOGISTICS"),
	}

	value := strings.TrimSpace(os.Getenv("MOCKLOGISTICS_DELIVERY_DURATION"))
	if value == "" {
		return config
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		slog.Warn(
			"ignore invalid MOCKLOGISTICS_DELIVERY_DURATION",
			"value",
			value,
			"error",
			err,
		)
		return config
	}
	if duration <= 0 {
		slog.Warn(
			"ignore non-positive MOCKLOGISTICS_DELIVERY_DURATION",
			"value",
			value,
		)
		return config
	}
	if duration > maxDeliveryDuration {
		slog.Warn(
			"clamp MOCKLOGISTICS_DELIVERY_DURATION",
			"value",
			value,
			"max",
			maxDeliveryDuration,
		)
		duration = maxDeliveryDuration
	}
	config.DeliveryDuration = duration
	return config
}

func (c Config) withDefaults() Config {
	if c.DeliveryDuration <= 0 {
		c.DeliveryDuration = defaultDeliveryDuration
	}
	if c.DeliveryDuration > maxDeliveryDuration {
		c.DeliveryDuration = maxDeliveryDuration
	}
	if c.ShipmentPollInterval <= 0 {
		c.ShipmentPollInterval = defaultShipmentPollInterval
	}
	if c.ShipmentBatchSize <= 0 {
		c.ShipmentBatchSize = defaultShipmentBatchSize
	}
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
	authHandler   *mockauth.Handler
	delivery      *logisticsservice.NotificationService
	progress      *logisticsservice.ShipmentService
	authService   *mockauth.Service
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
	authService, err := mockauth.NewService(pool, config.Auth)
	if err != nil {
		return nil, err
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
		authHandler:   mockauth.NewHandler(authService),
		delivery:      notificationService,
		progress:      shipmentService,
		authService:   authService,
		config:        config,
	}, nil
}

func (m *Module) BootstrapAuth(ctx context.Context) error {
	return m.authService.Bootstrap(ctx)
}

func (m *Module) AuthMiddleware() gin.HandlerFunc {
	return m.authHandler.Middleware()
}

func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	r.POST("/mocklogistics/api/v1/auth/login", m.authHandler.Login)
	r.POST("/mocklogistics/api/v1/auth/logout", authMiddleware, m.authHandler.Logout)

	api := r.Group("/mocklogistics/api/v1/shipments")
	api.Use(authMiddleware)
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
	m.wg.Add(2)
	go m.runNotificationWorker(workerCtx)
	go m.runShipmentWorker(workerCtx)
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

func (m *Module) runShipmentWorker(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.config.ShipmentPollInterval)
	defer ticker.Stop()

	for {
		if _, err := m.progress.AdvanceDue(
			ctx,
			m.config.DeliveryDuration,
			m.config.ShipmentBatchSize,
		); err != nil && ctx.Err() == nil {
			slog.ErrorContext(ctx, "advance mock logistics shipments", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
