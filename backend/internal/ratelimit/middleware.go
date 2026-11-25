package ratelimit

import (
	"errors"
	"log/slog"
	"math"
	"net/http"
	"pinshaoshao/internal/http/authctx"
	"pinshaoshao/internal/http/httpx"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var ErrTooManyRequests = errors.New("too many requests")

type EntryLimiter interface {
	TotalLimiter
	PublicLimiter
}

func APIEntryMiddleware(
	limiter EntryLimiter,
	isPublicRoute func(method, route string) bool,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			return
		}

		route := c.FullPath()
		if route == "" || isPublicRoute(c.Request.Method, route) {
			applyPublicAccess(c, limiter)
			return
		}
		applyAPITotal(c, limiter)
	}
}

func UnauthenticatedGuard(limiter UnauthenticatedLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		applyUnauthenticated(c, limiter)
	}
}

func PreflightMiddleware(limiter PublicLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		applyPublicAccess(c, limiter)
	}
}

func UserMiddleware(limiter Limiter) gin.HandlerFunc {
	return principalMiddleware(limiter, RoleUser, authctx.UserID)
}

func ShopMiddleware(limiter Limiter) gin.HandlerFunc {
	return principalMiddleware(limiter, RoleShop, authctx.ShopID)
}

func principalMiddleware(
	limiter Limiter,
	role Role,
	subjectID func(*gin.Context) (int, bool),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := subjectID(c)
		if !ok {
			slog.ErrorContext(c.Request.Context(), "resolve rate limit principal", "role", role)
			return
		}

		route := c.FullPath()
		if route == "" {
			slog.ErrorContext(c.Request.Context(), "resolve rate limit route")
			return
		}

		decision, err := limiter.Acquire(c.Request.Context(), Principal{
			Role:   role,
			ID:     id,
			Method: c.Request.Method,
			Route:  route,
		})
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "apply rate limit", "role", role, "subject_id", id, "error", err)
			return
		}

		c.Header("X-RateLimit-Api-Limit", strconv.Itoa(decision.APILimit))
		c.Header("X-RateLimit-Api-Remaining", strconv.Itoa(decision.APIRemaining))
		c.Header("X-RateLimit-Subject-Limit", strconv.Itoa(decision.SubjectLimit))
		c.Header("X-RateLimit-Subject-Remaining", strconv.Itoa(decision.SubjectRemaining))
		if decision.Allowed {
			c.Header("X-RateLimit-Policy", "api, subject")
			return
		}

		layer := "subject"
		if decision.Layer == LayerAPI {
			layer = "api"
		}
		retryAfterSeconds := max(1, int64(math.Ceil(decision.RetryAfter.Seconds())))
		c.Header("X-RateLimit-Policy", layer)
		c.Header("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
		c.AbortWithStatusJSON(http.StatusTooManyRequests, httpx.ErrorResponse{
			Error: ErrTooManyRequests.Error(),
		})
	}
}

func applyAPITotal(c *gin.Context, limiter TotalLimiter) {
	if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
		return
	}

	decision, err := limiter.AcquireAPI(c.Request.Context())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "apply API total rate limit", "error", err)
		return
	}

	c.Header("X-RateLimit-Api-Total-Limit", strconv.Itoa(decision.Limit))
	c.Header("X-RateLimit-Api-Total-Remaining", strconv.Itoa(decision.Remaining))
	if decision.Allowed {
		return
	}

	abortTooManyRequests(c, decision.RetryAfter, "api-total")
}

func applyUnauthenticated(c *gin.Context, limiter UnauthenticatedLimiter) {
	decision, err := limiter.AcquireUnauthenticated(c.Request.Context(), c.ClientIP())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "apply unauthenticated rate limit", "client_ip", c.ClientIP(), "error", err)
		return
	}

	c.Header("X-RateLimit-Unauthenticated-Limit", strconv.Itoa(decision.Limit))
	c.Header("X-RateLimit-Unauthenticated-Remaining", strconv.Itoa(decision.Remaining))
	if decision.Allowed {
		return
	}

	abortTooManyRequests(c, decision.RetryAfter, "unauthenticated-ip")
}

func applyPublicAccess(c *gin.Context, limiter PublicLimiter) {
	if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
		return
	}

	decision, err := limiter.AcquirePublic(c.Request.Context(), c.ClientIP())
	if err != nil {
		slog.ErrorContext(
			c.Request.Context(),
			"apply public rate limit",
			"client_ip",
			c.ClientIP(),
			"error",
			err,
		)
		return
	}

	c.Header("X-RateLimit-Api-Total-Limit", strconv.Itoa(decision.APITotalLimit))
	c.Header("X-RateLimit-Api-Total-Remaining", strconv.Itoa(decision.APITotalRemaining))
	if decision.Layer == PublicLayerAPI {
		abortTooManyRequests(c, decision.RetryAfter, "api-total")
		return
	}

	c.Header(
		"X-RateLimit-Unauthenticated-Limit",
		strconv.Itoa(decision.UnauthenticatedLimit),
	)
	c.Header(
		"X-RateLimit-Unauthenticated-Remaining",
		strconv.Itoa(decision.UnauthenticatedRemaining),
	)
	if decision.Allowed {
		return
	}

	abortTooManyRequests(c, decision.RetryAfter, "unauthenticated-ip")
}

func abortTooManyRequests(c *gin.Context, retryAfter time.Duration, policy string) {
	retryAfterSeconds := max(1, int64(math.Ceil(retryAfter.Seconds())))
	c.Header("X-RateLimit-Policy", policy)
	c.Header("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
	c.AbortWithStatusJSON(http.StatusTooManyRequests, httpx.ErrorResponse{
		Error: ErrTooManyRequests.Error(),
	})
}
