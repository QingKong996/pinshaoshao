package ratelimit

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Role string

const (
	RoleUser Role = "user"
	RoleShop Role = "shop"
)

type Layer int

const (
	LayerNone    Layer = 0
	LayerAPI     Layer = 1
	LayerSubject Layer = 2
)

type PublicLayer int

const (
	PublicLayerNone            PublicLayer = 0
	PublicLayerAPI             PublicLayer = 1
	PublicLayerUnauthenticated PublicLayer = 2
)

type Principal struct {
	Role   Role
	ID     int
	Method string
	Route  string
}

type Decision struct {
	Allowed          bool
	Layer            Layer
	APILimit         int
	APIRemaining     int
	SubjectLimit     int
	SubjectRemaining int
	RetryAfter       time.Duration
}

type TotalDecision struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter time.Duration
}

type UnauthenticatedDecision struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter time.Duration
}

type PublicDecision struct {
	Allowed                  bool
	Layer                    PublicLayer
	APITotalLimit            int
	APITotalRemaining        int
	UnauthenticatedLimit     int
	UnauthenticatedRemaining int
	RetryAfter               time.Duration
}

type Limiter interface {
	Acquire(ctx context.Context, principal Principal) (Decision, error)
}

type TotalLimiter interface {
	AcquireAPI(ctx context.Context) (TotalDecision, error)
}

type UnauthenticatedLimiter interface {
	AcquireUnauthenticated(ctx context.Context, ip string) (UnauthenticatedDecision, error)
}

type PublicLimiter interface {
	AcquirePublic(ctx context.Context, ip string) (PublicDecision, error)
}

type RedisLimiter struct {
	client *redis.Client
	config Config
}

func NewRedisLimiter(ctx context.Context, redisURL string, config Config) (*RedisLimiter, error) {
	if strings.TrimSpace(redisURL) == "" {
		return nil, fmt.Errorf("REDIS_URL is not set")
	}
	if err := config.validate(); err != nil {
		return nil, err
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &RedisLimiter{client: client, config: config}, nil
}

func newRedisLimiter(client *redis.Client, config Config) *RedisLimiter {
	return &RedisLimiter{client: client, config: config}
}

func (l *RedisLimiter) Close() error {
	return l.client.Close()
}

func (l *RedisLimiter) Acquire(ctx context.Context, principal Principal) (Decision, error) {
	if principal.ID <= 0 {
		return Decision{}, fmt.Errorf("rate limit principal ID must be positive")
	}
	if principal.Role != RoleUser && principal.Role != RoleShop {
		return Decision{}, fmt.Errorf("unsupported rate limit role %q", principal.Role)
	}
	if principal.Method == "" || principal.Route == "" {
		return Decision{}, fmt.Errorf("rate limit method and route are required")
	}

	apiKey := fmt.Sprintf(
		"rate_limit:{%s:%d}:api:%s:%s",
		principal.Role,
		principal.ID,
		principal.Method,
		principal.Route,
	)
	subjectKey := fmt.Sprintf(
		"rate_limit:{%s:%d}:subject",
		principal.Role,
		principal.ID,
	)

	values, err := slidingWindowScript.Run(
		ctx,
		l.client,
		[]string{apiKey, subjectKey},
		l.config.Window.Milliseconds(),
		l.config.APILimit,
		l.config.SubjectLimit,
		uuid.NewString(),
	).Int64Slice()
	if err != nil {
		return Decision{}, fmt.Errorf("run sliding window: %w", err)
	}
	if len(values) != 5 {
		return Decision{}, fmt.Errorf("sliding window returned %d values", len(values))
	}

	return Decision{
		Allowed:          values[0] == 1,
		Layer:            Layer(values[1]),
		APILimit:         l.config.APILimit,
		APIRemaining:     max(0, l.config.APILimit-int(values[2])),
		SubjectLimit:     l.config.SubjectLimit,
		SubjectRemaining: max(0, l.config.SubjectLimit-int(values[3])),
		RetryAfter:       time.Duration(values[4]) * time.Millisecond,
	}, nil
}

func (l *RedisLimiter) AcquireAPI(ctx context.Context) (TotalDecision, error) {
	decision, err := l.acquireWindow(ctx, "rate_limit:api:total", l.config.APITotalLimit)
	if err != nil {
		return TotalDecision{}, fmt.Errorf("run API total sliding window: %w", err)
	}

	return TotalDecision{
		Allowed:    decision.Allowed,
		Limit:      decision.Limit,
		Remaining:  decision.Remaining,
		RetryAfter: decision.RetryAfter,
	}, nil
}

func (l *RedisLimiter) AcquireUnauthenticated(
	ctx context.Context,
	ip string,
) (UnauthenticatedDecision, error) {
	normalizedIP, err := normalizeIP(ip)
	if err != nil {
		return UnauthenticatedDecision{}, err
	}

	decision, err := l.acquireWindow(
		ctx,
		"rate_limit:unauthenticated:"+normalizedIP,
		l.config.UnauthenticatedLimit,
	)
	if err != nil {
		return UnauthenticatedDecision{}, fmt.Errorf("run unauthenticated sliding window: %w", err)
	}
	return UnauthenticatedDecision{
		Allowed:    decision.Allowed,
		Limit:      decision.Limit,
		Remaining:  decision.Remaining,
		RetryAfter: decision.RetryAfter,
	}, nil
}

func (l *RedisLimiter) AcquirePublic(ctx context.Context, ip string) (PublicDecision, error) {
	normalizedIP, err := normalizeIP(ip)
	if err != nil {
		return PublicDecision{}, err
	}

	values, err := publicSlidingWindowScript.Run(
		ctx,
		l.client,
		[]string{
			"rate_limit:api:total",
			"rate_limit:unauthenticated:" + normalizedIP,
		},
		l.config.Window.Milliseconds(),
		l.config.APITotalLimit,
		l.config.UnauthenticatedLimit,
		uuid.NewString(),
	).Int64Slice()
	if err != nil {
		return PublicDecision{}, fmt.Errorf("run public sliding window: %w", err)
	}
	if len(values) != 5 {
		return PublicDecision{}, fmt.Errorf("public sliding window returned %d values", len(values))
	}

	return PublicDecision{
		Allowed:                  values[0] == 1,
		Layer:                    PublicLayer(values[1]),
		APITotalLimit:            l.config.APITotalLimit,
		APITotalRemaining:        max(0, l.config.APITotalLimit-int(values[2])),
		UnauthenticatedLimit:     l.config.UnauthenticatedLimit,
		UnauthenticatedRemaining: max(0, l.config.UnauthenticatedLimit-int(values[3])),
		RetryAfter:               time.Duration(values[4]) * time.Millisecond,
	}, nil
}

func normalizeIP(ip string) (string, error) {
	parsedIP := net.ParseIP(strings.TrimSpace(ip))
	if parsedIP == nil {
		return "", fmt.Errorf("invalid unauthenticated client IP %q", ip)
	}
	return parsedIP.String(), nil
}

type windowDecision struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter time.Duration
}

func (l *RedisLimiter) acquireWindow(
	ctx context.Context,
	key string,
	limit int,
) (windowDecision, error) {
	values, err := slidingWindowCounterScript.Run(
		ctx,
		l.client,
		[]string{key},
		l.config.Window.Milliseconds(),
		limit,
		uuid.NewString(),
	).Int64Slice()
	if err != nil {
		return windowDecision{}, err
	}
	if len(values) != 3 {
		return windowDecision{}, fmt.Errorf("sliding window returned %d values", len(values))
	}
	return windowDecision{
		Allowed:    values[0] == 1,
		Limit:      limit,
		Remaining:  max(0, limit-int(values[1])),
		RetryAfter: time.Duration(values[2]) * time.Millisecond,
	}, nil
}

var slidingWindowCounterScript = redis.NewScript(`
local window_ms = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local member = ARGV[3]

local redis_time = redis.call('TIME')
local now_ms = tonumber(redis_time[1]) * 1000 + math.floor(tonumber(redis_time[2]) / 1000)
local cutoff = now_ms - window_ms

redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', cutoff)
local count = redis.call('ZCARD', KEYS[1])
if count >= limit then
    local oldest = redis.call('ZRANGE', KEYS[1], 0, 0, 'WITHSCORES')
    local retry_after = window_ms
    if oldest[2] ~= nil then
        retry_after = math.max(1, tonumber(oldest[2]) + window_ms - now_ms)
    end
    return {0, count, retry_after}
end

redis.call('ZADD', KEYS[1], now_ms, member)
redis.call('PEXPIRE', KEYS[1], window_ms)
return {1, count + 1, 0}
`)

var publicSlidingWindowScript = redis.NewScript(`
local window_ms = tonumber(ARGV[1])
local api_total_limit = tonumber(ARGV[2])
local unauthenticated_limit = tonumber(ARGV[3])
local member = ARGV[4]

local redis_time = redis.call('TIME')
local now_ms = tonumber(redis_time[1]) * 1000 + math.floor(tonumber(redis_time[2]) / 1000)
local cutoff = now_ms - window_ms

redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', cutoff)
redis.call('ZREMRANGEBYSCORE', KEYS[2], '-inf', cutoff)

local api_total_count = redis.call('ZCARD', KEYS[1])
local unauthenticated_count = redis.call('ZCARD', KEYS[2])

if api_total_count >= api_total_limit then
    local oldest = redis.call('ZRANGE', KEYS[1], 0, 0, 'WITHSCORES')
    local retry_after = window_ms
    if oldest[2] ~= nil then
        retry_after = math.max(1, tonumber(oldest[2]) + window_ms - now_ms)
    end
    return {0, 1, api_total_count, unauthenticated_count, retry_after}
end

redis.call('ZADD', KEYS[1], now_ms, member)
redis.call('PEXPIRE', KEYS[1], window_ms)

if unauthenticated_count >= unauthenticated_limit then
    local oldest = redis.call('ZRANGE', KEYS[2], 0, 0, 'WITHSCORES')
    local retry_after = window_ms
    if oldest[2] ~= nil then
        retry_after = math.max(1, tonumber(oldest[2]) + window_ms - now_ms)
    end
    return {0, 2, api_total_count + 1, unauthenticated_count, retry_after}
end

redis.call('ZADD', KEYS[2], now_ms, member)
redis.call('PEXPIRE', KEYS[2], window_ms)

return {1, 0, api_total_count + 1, unauthenticated_count + 1, 0}
`)

var slidingWindowScript = redis.NewScript(`
local window_ms = tonumber(ARGV[1])
local api_limit = tonumber(ARGV[2])
local subject_limit = tonumber(ARGV[3])
local member = ARGV[4]

local redis_time = redis.call('TIME')
local now_ms = tonumber(redis_time[1]) * 1000 + math.floor(tonumber(redis_time[2]) / 1000)
local cutoff = now_ms - window_ms

redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', cutoff)
redis.call('ZREMRANGEBYSCORE', KEYS[2], '-inf', cutoff)

local api_count = redis.call('ZCARD', KEYS[1])
local subject_count = redis.call('ZCARD', KEYS[2])
local layer = 0
local retry_after = 0

if api_count >= api_limit then
    layer = 1
elseif subject_count >= subject_limit then
    layer = 2
end

if layer ~= 0 then
    local oldest = redis.call('ZRANGE', KEYS[layer], 0, 0, 'WITHSCORES')
    retry_after = window_ms
    if oldest[2] ~= nil then
        retry_after = math.max(1, tonumber(oldest[2]) + window_ms - now_ms)
    end
    return {0, layer, api_count, subject_count, retry_after}
end

redis.call('ZADD', KEYS[1], now_ms, member)
redis.call('ZADD', KEYS[2], now_ms, member)
redis.call('PEXPIRE', KEYS[1], window_ms)
redis.call('PEXPIRE', KEYS[2], window_ms)

return {1, layer, api_count + 1, subject_count + 1, retry_after}
`)
