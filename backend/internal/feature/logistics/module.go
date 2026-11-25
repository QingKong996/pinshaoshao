package logistics

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"pinshaoshao/internal/feature/logistics/controller"
	"pinshaoshao/internal/feature/logistics/gateway/mocklogistics"
	"pinshaoshao/internal/feature/logistics/repository"
	"pinshaoshao/internal/feature/logistics/service"
	"pinshaoshao/internal/feature/order/shippingport"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPublicBaseURL    = "http://localhost:8080"
	defaultWebhookTolerance = 5 * time.Minute
)

var (
	errMissingWebhookSecret = errors.New("MOCKLOGISTICS_WEBHOOK_SECRET is required")
	errInvalidGatewayURL    = errors.New("LOGISTICS_GATEWAY_BASE_URL must be an absolute HTTP URL")
	errInvalidNotifyURL     = errors.New("LOGISTICS_NOTIFY_URL must be an absolute HTTP URL")
)

type Config struct {
	GatewayBaseURL   string
	WebhookSecret    string
	NotifyURL        string
	WebhookTolerance time.Duration
	HTTPClient       *http.Client
	GatewayUsername  string
	GatewayPassword  string
}

func ConfigFromEnv() Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	publicBaseURL := strings.TrimSpace(os.Getenv("LOGISTICS_PUBLIC_BASE_URL"))
	if publicBaseURL == "" {
		publicBaseURL = strings.TrimSpace(os.Getenv("PAYMENT_PUBLIC_BASE_URL"))
	}
	if publicBaseURL == "" {
		publicBaseURL = "http://localhost:" + port
	}
	gatewayBaseURL := strings.TrimSpace(os.Getenv("LOGISTICS_GATEWAY_BASE_URL"))
	if gatewayBaseURL == "" {
		gatewayBaseURL = publicBaseURL
	}
	notifyURL := strings.TrimSpace(os.Getenv("LOGISTICS_NOTIFY_URL"))
	if notifyURL == "" {
		notifyURL = strings.TrimRight(publicBaseURL, "/") + "/api/v1/logistics/webhook"
	}
	return Config{
		GatewayBaseURL:  gatewayBaseURL,
		WebhookSecret:   strings.TrimSpace(os.Getenv("MOCKLOGISTICS_WEBHOOK_SECRET")),
		NotifyURL:       notifyURL,
		GatewayUsername: strings.TrimSpace(os.Getenv("MOCKLOGISTICS_AUTH_USERNAME")),
		GatewayPassword: os.Getenv("MOCKLOGISTICS_AUTH_PASSWORD"),
	}
}

type Module struct {
	shipments *controller.ShipmentController
	webhook   *controller.WebhookHandler
}

func NewModule(
	pool *pgxpool.Pool,
	orders shippingport.Lifecycle,
	config Config,
) (*Module, error) {
	config = config.withDefaults()
	if config.WebhookSecret == "" {
		return nil, errMissingWebhookSecret
	}
	if !isAbsoluteHTTPURL(config.GatewayBaseURL) {
		return nil, errInvalidGatewayURL
	}
	if !isAbsoluteHTTPURL(config.NotifyURL) {
		return nil, errInvalidNotifyURL
	}

	gateway, err := mocklogistics.NewClient(
		config.GatewayBaseURL,
		config.GatewayUsername,
		config.GatewayPassword,
		config.HTTPClient,
	)
	if err != nil {
		return nil, err
	}
	shipmentRepository := repository.NewShipmentRepository(pool, orders)
	shipmentService := service.NewShipmentService(
		shipmentRepository,
		orders,
		service.Config{
			NotifyURL: config.NotifyURL,
			Gateway:   gateway,
		},
	)
	return &Module{
		shipments: controller.NewShipmentController(shipmentService),
		webhook: controller.NewWebhookHandler(
			shipmentService,
			config.WebhookSecret,
			config.WebhookTolerance,
		),
	}, nil
}

func (c Config) withDefaults() Config {
	c.GatewayBaseURL = strings.TrimRight(strings.TrimSpace(c.GatewayBaseURL), "/")
	if c.GatewayBaseURL == "" {
		c.GatewayBaseURL = defaultPublicBaseURL
	}
	c.WebhookSecret = strings.TrimSpace(c.WebhookSecret)
	c.NotifyURL = strings.TrimSpace(c.NotifyURL)
	if c.NotifyURL == "" {
		c.NotifyURL = defaultPublicBaseURL + "/api/v1/logistics/webhook"
	}
	if c.WebhookTolerance <= 0 {
		c.WebhookTolerance = defaultWebhookTolerance
	}
	return c
}

func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	protected := r.Group("/api/v1")
	protected.Use(authMiddleware)
	protected.GET("/shipments", m.shipments.ListForBuyer)
	protected.GET("/shipments/:shipmentNo", m.shipments.GetForBuyer)

	r.POST("/api/v1/logistics/webhook", m.webhook.Handle)
}

func (m *Module) RegisterMerchant(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	protected := r.Group("/api/v1/merchant")
	protected.Use(authMiddleware)
	protected.POST("/shipments", m.shipments.Ship)
	protected.GET("/shipments", m.shipments.ListForMerchant)
	protected.GET("/shipments/:shipmentNo", m.shipments.GetForMerchant)
}

func isAbsoluteHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil &&
		parsed.Host != "" &&
		(parsed.Scheme == "http" || parsed.Scheme == "https")
}
