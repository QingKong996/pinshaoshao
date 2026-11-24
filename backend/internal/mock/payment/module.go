package payment

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	mockpayhandler "pinshaoshao/internal/mock/payment/handler"
	"pinshaoshao/internal/mock/payment/model"
	mockpayrepository "pinshaoshao/internal/mock/payment/repository"
	mockpayservice "pinshaoshao/internal/mock/payment/service"
	"pinshaoshao/internal/webhooksig"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultNotificationInterval   = 500 * time.Millisecond
	defaultExpiryInterval         = 10 * time.Second
	defaultNotificationBatchSize  = 20
	defaultPaymentExpiryBatchSize = 100
)

type Config struct {
	WebhookSecret             string
	PublicBaseURL             string
	NotificationPollInterval  time.Duration
	ExpiryPollInterval        time.Duration
	NotificationBatchSize     int
	PaymentExpiryBatchSize    int
	NotificationMaxAttempts   int
	NotificationRetryDelays   []time.Duration
	NotificationDeliveryLease time.Duration
	HTTPClient                *http.Client
}

func ConfigFromEnv() Config {
	return Config{
		WebhookSecret: strings.TrimSpace(os.Getenv("MOCKPAY_WEBHOOK_SECRET")),
		PublicBaseURL: strings.TrimSpace(os.Getenv("MOCKPAY_PUBLIC_BASE_URL")),
	}
}

func (c Config) withDefaults() Config {
	if c.NotificationPollInterval <= 0 {
		c.NotificationPollInterval = defaultNotificationInterval
	}
	if c.ExpiryPollInterval <= 0 {
		c.ExpiryPollInterval = defaultExpiryInterval
	}
	if c.NotificationBatchSize <= 0 {
		c.NotificationBatchSize = defaultNotificationBatchSize
	}
	if c.PaymentExpiryBatchSize <= 0 {
		c.PaymentExpiryBatchSize = defaultPaymentExpiryBatchSize
	}
	if c.NotificationMaxAttempts <= 0 {
		c.NotificationMaxAttempts = model.DefaultNotificationMaxAttempts
	}
	return c
}

type Module struct {
	paymentHandler  *mockpayhandler.PaymentHandler
	checkoutHandler *mockpayhandler.CheckoutHandler
	webhookHandler  *mockpayhandler.WebhookHandler
	payments        *mockpayservice.PaymentService
	notifications   *mockpayservice.NotificationService
	config          Config

	mu      sync.Mutex
	started bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewModule(pool *pgxpool.Pool, config Config) (*Module, error) {
	config = config.withDefaults()
	if strings.TrimSpace(config.WebhookSecret) == "" {
		return nil, errors.New("MOCKPAY_WEBHOOK_SECRET is required")
	}
	notificationRepository := mockpayrepository.NewNotificationRepository(pool)
	paymentRepository := mockpayrepository.NewPaymentRepository(
		pool,
		notificationRepository,
		config.NotificationMaxAttempts,
	)
	paymentService := mockpayservice.NewPaymentService(paymentRepository)
	notificationService := mockpayservice.NewNotificationService(
		notificationRepository,
		mockpayservice.NotificationConfig{
			Client:      config.HTTPClient,
			Signer:      webhooksig.NewSigner(config.WebhookSecret),
			RetryDelays: config.NotificationRetryDelays,
			Lease:       config.NotificationDeliveryLease,
			BatchSize:   config.NotificationBatchSize,
		},
	)

	return &Module{
		paymentHandler:  mockpayhandler.NewPaymentHandler(paymentService, config.PublicBaseURL),
		checkoutHandler: mockpayhandler.NewCheckoutHandler(paymentService),
		webhookHandler:  mockpayhandler.NewWebhookHandler(paymentService, notificationService),
		payments:        paymentService,
		notifications:   notificationService,
		config:          config,
	}, nil
}

func (m *Module) Register(r *gin.Engine, _ gin.HandlerFunc) {
	api := r.Group("/mockpay/api/v1/payments")
	api.POST("", m.paymentHandler.Create)
	api.GET("/:paymentNo", m.paymentHandler.Get)
	api.POST("/:paymentNo/close", m.paymentHandler.Close)
	api.GET("/:paymentNo/notifications", m.webhookHandler.List)
	api.POST("/:paymentNo/notifications/retry", m.webhookHandler.Retry)

	checkout := r.Group("/mockpay/checkout")
	checkout.GET("/:paymentNo", m.checkoutHandler.Show)
	checkout.POST("/:paymentNo/pay", m.checkoutHandler.Pay)
	checkout.POST("/:paymentNo/close", m.checkoutHandler.Close)
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
	go m.runExpiryWorker(workerCtx)
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
		if err := m.notifications.DeliverDue(ctx, m.config.NotificationBatchSize); err != nil &&
			ctx.Err() == nil {
			slog.ErrorContext(ctx, "deliver mock payment notifications", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (m *Module) runExpiryWorker(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.config.ExpiryPollInterval)
	defer ticker.Stop()

	for {
		closed, err := m.payments.CloseExpired(ctx, m.config.PaymentExpiryBatchSize)
		if err != nil && ctx.Err() == nil {
			slog.ErrorContext(ctx, "close expired mock payments", "error", err)
		} else if closed > 0 {
			slog.InfoContext(ctx, "closed expired mock payments", "count", closed)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
