package middleware

import (
	"cmp"
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultAccessLogQueueSize = 8192

type accessLogEntry struct {
	method    string
	path      string
	route     string
	clientIP  string
	userAgent string
	status    int
	bytes     int
	latency   time.Duration
}

// AccessLogger writes structured request logs asynchronously.
type AccessLogger struct {
	logger  *slog.Logger
	entries chan accessLogEntry
	closing chan struct{}
	done    chan struct{}

	closeOnce sync.Once
	lock      sync.RWMutex
	closed    bool
	dropped   atomic.Uint64
}

func NewAccessLogger(logger *slog.Logger) *AccessLogger {
	return newAccessLogger(logger, defaultAccessLogQueueSize)
}

func newAccessLogger(logger *slog.Logger, queueSize int) *AccessLogger {
	if logger == nil {
		logger = slog.Default()
	}
	accessLogger := &AccessLogger{
		logger:  logger,
		entries: make(chan accessLogEntry, queueSize),
		closing: make(chan struct{}),
		done:    make(chan struct{}),
	}
	go accessLogger.run()
	return accessLogger
}

// Middleware records the final HTTP status and latency after the request.
func (l *AccessLogger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		route := c.FullPath()

		entry := accessLogEntry{
			method:    c.Request.Method,
			path:      cmp.Or(route, "<unmatched>"),
			route:     route,
			clientIP:  c.ClientIP(),
			userAgent: c.Request.UserAgent(),
			status:    c.Writer.Status(),
			bytes:     c.Writer.Size(),
			latency:   time.Since(started),
		}
		l.enqueue(entry)
	}
}

func (l *AccessLogger) Close(ctx context.Context) error {
	l.closeOnce.Do(func() {
		l.lock.Lock()
		l.closed = true
		close(l.closing)
		l.lock.Unlock()
	})

	select {
	case <-l.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *AccessLogger) enqueue(entry accessLogEntry) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.closed {
		return
	}

	select {
	case l.entries <- entry:
	default:
		l.dropped.Add(1)
	}
}

func (l *AccessLogger) run() {
	defer close(l.done)
	for {
		select {
		case entry := <-l.entries:
			l.write(entry)
		case <-l.closing:
			for {
				select {
				case entry := <-l.entries:
					l.write(entry)
				default:
					if dropped := l.dropped.Swap(0); dropped > 0 {
						l.logger.Error("dropped access logs", "count", dropped)
					}
					return
				}
			}
		}
	}
}

func (l *AccessLogger) write(entry accessLogEntry) {
	l.logger.Info(
		"http request",
		"method",
		entry.method,
		"path",
		entry.path,
		"route",
		entry.route,
		"status",
		entry.status,
		"latency_ms",
		float64(entry.latency.Microseconds())/1000,
		"bytes",
		entry.bytes,
		"client_ip",
		entry.clientIP,
		"user_agent",
		entry.userAgent,
	)
}
