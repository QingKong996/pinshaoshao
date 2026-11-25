package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pinshaoshao/internal/http/authctx"
	"pinshaoshao/internal/middleware"

	"github.com/gin-gonic/gin"
)

type stubLimiter struct {
	decision  Decision
	err       error
	principal Principal
}

type stubEntryLimiter struct {
	totalDecision  TotalDecision
	publicDecision PublicDecision
	err            error
	totalCalls     int
	publicCalls    int
	ip             string
}

func (s *stubEntryLimiter) AcquireAPI(context.Context) (TotalDecision, error) {
	s.totalCalls++
	return s.totalDecision, s.err
}

func (s *stubEntryLimiter) AcquirePublic(
	_ context.Context,
	ip string,
) (PublicDecision, error) {
	s.publicCalls++
	s.ip = ip
	return s.publicDecision, s.err
}

func (s *stubLimiter) Acquire(_ context.Context, principal Principal) (Decision, error) {
	s.principal = principal
	return s.decision, s.err
}

func TestUserMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		decision   Decision
		wantStatus int
		wantLayer  string
	}{
		{
			name: "allowed",
			decision: Decision{
				Allowed:          true,
				APILimit:         60,
				APIRemaining:     59,
				SubjectLimit:     300,
				SubjectRemaining: 299,
			},
			wantStatus: http.StatusOK,
			wantLayer:  "api, subject",
		},
		{
			name: "API layer denied",
			decision: Decision{
				APILimit:         60,
				SubjectLimit:     300,
				SubjectRemaining: 200,
				Layer:            LayerAPI,
				RetryAfter:       1500 * time.Millisecond,
			},
			wantStatus: http.StatusTooManyRequests,
			wantLayer:  "api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := &stubLimiter{decision: tt.decision}
			handlerCalled := false
			r := gin.New()
			r.GET("/api/v1/orders/:orderID", func(c *gin.Context) {
				c.Set(authctx.UserIDKey, 42)
			}, UserMiddleware(limiter), func(c *gin.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/orders/1", nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if got := rec.Header().Get("X-RateLimit-Policy"); got != tt.wantLayer {
				t.Fatalf("X-RateLimit-Policy = %q, want %q", got, tt.wantLayer)
			}
			if limiter.principal.ID != 42 ||
				limiter.principal.Method != http.MethodGet ||
				limiter.principal.Route != "/api/v1/orders/:orderID" {
				t.Fatalf("principal = %+v", limiter.principal)
			}
			if handlerCalled != (tt.wantStatus == http.StatusOK) {
				t.Fatalf("handler called = %v", handlerCalled)
			}
			if tt.wantStatus == http.StatusTooManyRequests {
				if got := rec.Header().Get("Retry-After"); got != "2" {
					t.Fatalf("Retry-After = %q, want 2", got)
				}
			}
		})
	}
}

func TestUserMiddlewareFailsOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handlerCalled := false
	r := gin.New()
	r.GET("/api/v1/orders", func(c *gin.Context) {
		c.Set(authctx.UserIDKey, 42)
	}, UserMiddleware(&stubLimiter{err: context.Canceled}), func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil))

	if rec.Code != http.StatusOK || !handlerCalled {
		t.Fatalf("status = %d, handler called = %v", rec.Code, handlerCalled)
	}
}

func TestAPIEntryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		method          string
		path            string
		route           string
		noRoute         bool
		public          bool
		totalDecision   TotalDecision
		publicDecision  PublicDecision
		wantStatus      int
		wantTotalCalls  int
		wantPublicCalls int
		wantPolicy      string
		wantHandler     bool
	}{
		{
			name:   "protected route uses total limiter",
			method: http.MethodGet,
			path:   "/api/v1/orders",
			route:  "/api/v1/orders",
			totalDecision: TotalDecision{
				Allowed: true, Limit: 3000, Remaining: 2999,
			},
			wantStatus:      http.StatusOK,
			wantTotalCalls:  1,
			wantPublicCalls: 0,
			wantHandler:     true,
		},
		{
			name:   "public route uses combined limiter",
			method: http.MethodPost,
			path:   "/api/v1/login",
			route:  "/api/v1/login",
			public: true,
			publicDecision: PublicDecision{
				Allowed: true, APITotalLimit: 3000, APITotalRemaining: 2999,
				UnauthenticatedLimit: 60, UnauthenticatedRemaining: 59,
			},
			wantStatus:      http.StatusOK,
			wantTotalCalls:  0,
			wantPublicCalls: 1,
			wantHandler:     true,
		},
		{
			name:   "public API total denied",
			method: http.MethodPost,
			path:   "/api/v1/payment/webhook",
			route:  "/api/v1/payment/webhook",
			public: true,
			publicDecision: PublicDecision{
				Layer: PublicLayerAPI, APITotalLimit: 3000,
				UnauthenticatedLimit: 60, UnauthenticatedRemaining: 60,
				RetryAfter: time.Second,
			},
			wantStatus:      http.StatusTooManyRequests,
			wantTotalCalls:  0,
			wantPublicCalls: 1,
			wantPolicy:      "api-total",
		},
		{
			name:   "public unauthenticated layer denied",
			method: http.MethodPost,
			path:   "/api/v1/payment/webhook",
			route:  "/api/v1/payment/webhook",
			public: true,
			publicDecision: PublicDecision{
				Layer: PublicLayerUnauthenticated, APITotalLimit: 3000,
				APITotalRemaining: 2999, UnauthenticatedLimit: 60,
				RetryAfter: time.Second,
			},
			wantStatus:      http.StatusTooManyRequests,
			wantTotalCalls:  0,
			wantPublicCalls: 1,
			wantPolicy:      "unauthenticated-ip",
		},
		{
			name:    "unknown API route uses combined limiter",
			method:  http.MethodGet,
			path:    "/api/v1/unknown",
			noRoute: true,
			publicDecision: PublicDecision{
				Allowed: true, APITotalLimit: 3000, APITotalRemaining: 2999,
				UnauthenticatedLimit: 60, UnauthenticatedRemaining: 59,
			},
			wantStatus:      http.StatusOK,
			wantTotalCalls:  0,
			wantPublicCalls: 1,
			wantHandler:     true,
		},
		{
			name:   "non API route ignored",
			method: http.MethodGet,
			path:   "/health",
			route:  "/health",
			totalDecision: TotalDecision{
				Allowed: true,
			},
			wantStatus:      http.StatusOK,
			wantTotalCalls:  0,
			wantPublicCalls: 0,
			wantHandler:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := &stubEntryLimiter{
				totalDecision:  tt.totalDecision,
				publicDecision: tt.publicDecision,
			}
			handlerCalled := false
			r := gin.New()
			r.Use(APIEntryMiddleware(limiter, func(method, route string) bool {
				return tt.public && method == tt.method && route == tt.route
			}))
			handler := func(c *gin.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			}
			if tt.noRoute {
				r.NoRoute(handler)
			} else {
				r.Handle(tt.method, tt.route, handler)
			}

			rec := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)
			request.RemoteAddr = "192.0.2.10:1234"
			r.ServeHTTP(rec, request)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if limiter.totalCalls != tt.wantTotalCalls || limiter.publicCalls != tt.wantPublicCalls {
				t.Fatalf(
					"limiter calls: total=%d public=%d, want total=%d public=%d",
					limiter.totalCalls,
					limiter.publicCalls,
					tt.wantTotalCalls,
					tt.wantPublicCalls,
				)
			}
			if handlerCalled != tt.wantHandler {
				t.Fatalf("handler called = %v, want %v", handlerCalled, tt.wantHandler)
			}
			if tt.wantPublicCalls > 0 && limiter.ip != "192.0.2.10" {
				t.Fatalf("client IP = %q, want 192.0.2.10", limiter.ip)
			}
			if got := rec.Header().Get("X-RateLimit-Policy"); got != tt.wantPolicy {
				t.Fatalf("X-RateLimit-Policy = %q, want %q", got, tt.wantPolicy)
			}
		})
	}
}

func TestCORSPreflightAppliesRateLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := &stubEntryLimiter{
		publicDecision: PublicDecision{
			Allowed: true, APITotalLimit: 3000, APITotalRemaining: 2999,
			UnauthenticatedLimit: 60, UnauthenticatedRemaining: 59,
		},
	}
	r := gin.New()
	r.Use(middleware.CORS(
		"https://docs.example.com",
		PreflightMiddleware(limiter),
	))
	r.OPTIONS("/api/v1/orders", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/orders", nil)
	req.Header.Set("Origin", "https://docs.example.com")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if limiter.publicCalls != 1 || limiter.totalCalls != 0 {
		t.Fatalf("limiter calls: total=%d public=%d, want total=0 public=1", limiter.totalCalls, limiter.publicCalls)
	}
}
