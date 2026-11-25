package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisLimiterSlidingWindow(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	limiter := newRedisLimiter(client, Config{
		Window:               time.Second,
		APITotalLimit:        2,
		UnauthenticatedLimit: 2,
		APILimit:             2,
		SubjectLimit:         3,
	})
	ctx := context.Background()
	routeA := Principal{Role: RoleUser, ID: 1, Method: "GET", Route: "/api/v1/orders"}
	routeB := routeA
	routeB.Route = "/api/v1/cart"
	routeC := routeA
	routeC.Route = "/api/v1/feed"

	for request := 0; request < 2; request++ {
		decision, err := limiter.Acquire(ctx, routeA)
		if err != nil {
			t.Fatalf("request %d: %v", request+1, err)
		}
		if !decision.Allowed {
			t.Fatalf("request %d was denied", request+1)
		}
	}

	decision, err := limiter.Acquire(ctx, routeA)
	if err != nil {
		t.Fatalf("deny API layer: %v", err)
	}
	if decision.Allowed || decision.Layer != LayerAPI || decision.RetryAfter <= 0 {
		t.Fatalf("API decision = %+v", decision)
	}

	decision, err = limiter.Acquire(ctx, routeB)
	if err != nil {
		t.Fatalf("allow subject request: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("subject-limited request was denied: %+v", decision)
	}

	decision, err = limiter.Acquire(ctx, routeC)
	if err != nil {
		t.Fatalf("deny subject layer: %v", err)
	}
	if decision.Allowed || decision.Layer != LayerSubject {
		t.Fatalf("subject decision = %+v", decision)
	}

	server.FastForward(1100 * time.Millisecond)
	decision, err = limiter.Acquire(ctx, routeA)
	if err != nil {
		t.Fatalf("allow after window: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("request after window was denied: %+v", decision)
	}

	otherUser := routeA
	otherUser.ID = 2
	decision, err = limiter.Acquire(ctx, otherUser)
	if err != nil {
		t.Fatalf("allow other user: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("other user was denied: %+v", decision)
	}

	otherRole := routeA
	otherRole.Role = RoleShop
	decision, err = limiter.Acquire(ctx, otherRole)
	if err != nil {
		t.Fatalf("allow other role: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("other role was denied: %+v", decision)
	}
}

func TestRedisLimiterAPITotal(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	limiter := newRedisLimiter(client, Config{
		Window:               time.Second,
		APITotalLimit:        2,
		UnauthenticatedLimit: 10,
		APILimit:             10,
		SubjectLimit:         10,
	})
	ctx := context.Background()

	for request := 0; request < 2; request++ {
		decision, err := limiter.AcquireAPI(ctx)
		if err != nil {
			t.Fatalf("request %d: %v", request+1, err)
		}
		if !decision.Allowed {
			t.Fatalf("request %d was denied", request+1)
		}
	}

	decision, err := limiter.AcquireAPI(ctx)
	if err != nil {
		t.Fatalf("deny API total: %v", err)
	}
	if decision.Allowed || decision.RetryAfter <= 0 {
		t.Fatalf("API total decision = %+v", decision)
	}

	server.FastForward(1100 * time.Millisecond)
	decision, err = limiter.AcquireAPI(ctx)
	if err != nil {
		t.Fatalf("allow after window: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("request after window was denied: %+v", decision)
	}
}

func TestRedisLimiterUnauthenticated(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	limiter := newRedisLimiter(client, Config{
		Window:               time.Second,
		APITotalLimit:        10,
		UnauthenticatedLimit: 2,
		APILimit:             10,
		SubjectLimit:         10,
	})
	ctx := context.Background()

	for request := 0; request < 2; request++ {
		decision, err := limiter.AcquireUnauthenticated(ctx, "192.0.2.10")
		if err != nil {
			t.Fatalf("request %d: %v", request+1, err)
		}
		if !decision.Allowed {
			t.Fatalf("request %d was denied", request+1)
		}
	}

	decision, err := limiter.AcquireUnauthenticated(ctx, "192.0.2.10")
	if err != nil {
		t.Fatalf("deny unauthenticated IP: %v", err)
	}
	if decision.Allowed || decision.RetryAfter <= 0 {
		t.Fatalf("unauthenticated decision = %+v", decision)
	}

	decision, err = limiter.AcquireUnauthenticated(ctx, "192.0.2.11")
	if err != nil {
		t.Fatalf("allow other IP: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("other IP was denied: %+v", decision)
	}

	server.FastForward(1100 * time.Millisecond)
	decision, err = limiter.AcquireUnauthenticated(ctx, "192.0.2.10")
	if err != nil {
		t.Fatalf("allow after window: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("request after window was denied: %+v", decision)
	}
}

func TestRedisLimiterPublicSlidingWindow(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	limiter := newRedisLimiter(client, Config{
		Window:               time.Second,
		APITotalLimit:        3,
		UnauthenticatedLimit: 2,
		APILimit:             10,
		SubjectLimit:         10,
	})
	ctx := context.Background()

	for request := 0; request < 2; request++ {
		decision, err := limiter.AcquirePublic(ctx, "192.0.2.10")
		if err != nil {
			t.Fatalf("request %d: %v", request+1, err)
		}
		if !decision.Allowed || decision.Layer != PublicLayerNone {
			t.Fatalf("request %d decision = %+v", request+1, decision)
		}
	}

	decision, err := limiter.AcquirePublic(ctx, "192.0.2.10")
	if err != nil {
		t.Fatalf("deny unauthenticated IP: %v", err)
	}
	if decision.Allowed ||
		decision.Layer != PublicLayerUnauthenticated ||
		decision.APITotalRemaining != 0 ||
		decision.UnauthenticatedRemaining != 0 ||
		decision.RetryAfter <= 0 {
		t.Fatalf("unauthenticated decision = %+v", decision)
	}

	decision, err = limiter.AcquirePublic(ctx, "192.0.2.11")
	if err != nil {
		t.Fatalf("deny API total: %v", err)
	}
	if decision.Allowed || decision.Layer != PublicLayerAPI || decision.RetryAfter <= 0 {
		t.Fatalf("API total decision = %+v", decision)
	}

	server.FastForward(1100 * time.Millisecond)
	decision, err = limiter.AcquirePublic(ctx, "192.0.2.10")
	if err != nil {
		t.Fatalf("allow after window: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("request after window was denied: %+v", decision)
	}
}

func TestRedisLimiterPublicConcurrent(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	limiter := newRedisLimiter(client, Config{
		Window:               time.Minute,
		APITotalLimit:        100,
		UnauthenticatedLimit: 3,
		APILimit:             100,
		SubjectLimit:         100,
	})

	const requests = 20
	type result struct {
		allowed bool
		err     error
	}
	results := make(chan result, requests)
	var waitGroup sync.WaitGroup
	for range requests {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			decision, err := limiter.AcquirePublic(context.Background(), "192.0.2.10")
			results <- result{allowed: decision.Allowed, err: err}
		}()
	}
	waitGroup.Wait()
	close(results)

	allowed := 0
	for result := range results {
		if result.err != nil {
			t.Errorf("acquire unauthenticated rate limit: %v", result.err)
			continue
		}
		if result.allowed {
			allowed++
		}
	}
	if allowed != 3 {
		t.Fatalf("allowed requests = %d, want 3", allowed)
	}
}
