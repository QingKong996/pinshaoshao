package payment

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"pinshaoshao/internal/feature/order/paymentport"
	paymentcontroller "pinshaoshao/internal/feature/payment/controller"
	"pinshaoshao/internal/feature/payment/gateway/mockpay"
	paymentrepository "pinshaoshao/internal/feature/payment/repository"
	paymentservice "pinshaoshao/internal/feature/payment/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPublicBaseURL    = "http://localhost:8080"
	defaultWebhookTolerance = 5 * time.Minute
)

var (
	errMissingWebhookSecret = errors.New("MOCKPAY_WEBHOOK_SECRET is required")
	errInvalidNotifyURL     = errors.New("PAYMENT_NOTIFY_URL must be an absolute HTTP URL")
	errInvalidReturnURL     = errors.New("PAYMENT_RETURN_URL_BASE must be an absolute HTTP URL")
)

type Config struct {
	GatewayBaseURL   string
	WebhookSecret    string
	NotifyURL        string
	ReturnURLBase    string
	WebhookTolerance time.Duration
	HTTPClient       *http.Client
}

func ConfigFromEnv() Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	publicBaseURL := strings.TrimSpace(os.Getenv("PAYMENT_PUBLIC_BASE_URL"))
	if publicBaseURL == "" {
		publicBaseURL = "http://localhost:" + port
	}
	notifyURL := strings.TrimSpace(os.Getenv("PAYMENT_NOTIFY_URL"))
	if notifyURL == "" {
		notifyURL = strings.TrimRight(publicBaseURL, "/") + "/api/v1/payment/webhook"
	}
	gatewayBaseURL := strings.TrimSpace(os.Getenv("PAYMENT_GATEWAY_BASE_URL"))
	if gatewayBaseURL == "" {
		gatewayBaseURL = publicBaseURL
	}
	returnURLBase := strings.TrimSpace(os.Getenv("PAYMENT_RETURN_URL_BASE"))
	if returnURLBase == "" {
		returnURLBase = "http://localhost:5173"
	}
	return Config{
		GatewayBaseURL: gatewayBaseURL,
		WebhookSecret:  strings.TrimSpace(os.Getenv("MOCKPAY_WEBHOOK_SECRET")),
		NotifyURL:      notifyURL,
		ReturnURLBase:  returnURLBase,
	}
}

type Module struct {
	paymentController *paymentcontroller.PaymentController
	webhookHandler    *paymentcontroller.WebhookHandler
}

func NewModule(
	pool *pgxpool.Pool,
	orders paymentport.Lifecycle,
	config Config,
) (*Module, error) {
	config = config.withDefaults()
	if strings.TrimSpace(config.WebhookSecret) == "" {
		return nil, errMissingWebhookSecret
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	gateway, err := mockpay.NewClient(config.GatewayBaseURL, config.HTTPClient)
	if err != nil {
		return nil, err
	}

	paymentRepository := paymentrepository.NewPaymentRepository(pool, orders)
	paymentService := paymentservice.NewPaymentService(paymentRepository, paymentservice.Config{
		NotifyURL:     config.NotifyURL,
		ReturnURLBase: config.ReturnURLBase,
		Gateway:       gateway,
	})
	return &Module{
		paymentController: paymentcontroller.NewPaymentController(paymentService),
		webhookHandler: paymentcontroller.NewWebhookHandler(
			paymentService,
			config.WebhookSecret,
			config.WebhookTolerance,
		),
	}, nil
}

func (c Config) withDefaults() Config {
	c.GatewayBaseURL = strings.TrimSpace(c.GatewayBaseURL)
	if c.GatewayBaseURL == "" {
		c.GatewayBaseURL = defaultPublicBaseURL
	}
	c.WebhookSecret = strings.TrimSpace(c.WebhookSecret)
	c.NotifyURL = strings.TrimSpace(c.NotifyURL)
	if c.NotifyURL == "" {
		c.NotifyURL = defaultPublicBaseURL + "/api/v1/payment/webhook"
	}
	c.ReturnURLBase = strings.TrimRight(strings.TrimSpace(c.ReturnURLBase), "/")
	if c.WebhookTolerance <= 0 {
		c.WebhookTolerance = defaultWebhookTolerance
	}
	return c
}

func (c Config) validate() error {
	if !isAbsoluteHTTPURL(c.NotifyURL) {
		return errInvalidNotifyURL
	}
	if c.ReturnURLBase != "" && !isAbsoluteHTTPURL(c.ReturnURLBase) {
		return errInvalidReturnURL
	}
	return nil
}

func isAbsoluteHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil &&
		parsed.Host != "" &&
		(parsed.Scheme == "http" || parsed.Scheme == "https")
}

func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	protected := r.Group("/api/v1")
	protected.Use(authMiddleware)
	protected.POST("/payments", m.paymentController.Create)
	protected.GET("/payments/:paymentNo", m.paymentController.Get)
	protected.POST("/payments/:paymentNo/pay", m.paymentController.Pay)
	protected.POST("/payments/:paymentNo/close", m.paymentController.Close)

	r.POST("/api/v1/payment/webhook", m.webhookHandler.Handle)
}
