package rateLimiter

import (
	"context"
	"time"
)

type RateLimiter interface {
	Wait(ctx context.Context) error
}

type tokenBucketRateLimiter struct {
	ticker   *time.Ticker
	tokens   chan struct{}
	capacity int
}

func NewRateLimiter(requestsPerSecond int, interval time.Duration) RateLimiter {
	capacity := requestsPerSecond
	rl := &tokenBucketRateLimiter{
		ticker:   time.NewTicker(interval / time.Duration(requestsPerSecond)),
		tokens:   make(chan struct{}, capacity),
		capacity: capacity,
	}

	// Fill initial tokens
	for i := 0; i < capacity; i++ {
		rl.tokens <- struct{}{}
	}

	// Start token refill goroutine
	go rl.refillTokens()

	return rl
}

func (rl *tokenBucketRateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *tokenBucketRateLimiter) refillTokens() {
	for range rl.ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full, skip
		}
	}
}
