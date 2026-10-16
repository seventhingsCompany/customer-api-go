package client

import (
	"context"
	"sync"
	"time"
)

// DefaultRequestsPerMinute is the default per-client request limit.
const DefaultRequestsPerMinute = 200

// WithRateLimit overrides the per-client request limit in requests per minute.
// Requests are evenly spaced, without bursts; the first request is immediate.
// A value of zero or less disables throttling. All SDK requests share the limit,
// including authentication, uploads and downloads. Separate clients do not share
// a limit. Waiting respects the request's context cancellation and deadline.
func WithRateLimit(requestsPerMinute int) Option {
	return func(c *Client) {
		c.rateLimiter = newRateLimiter(requestsPerMinute)
	}
}

type rateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time
}

func newRateLimiter(requestsPerMinute int) *rateLimiter {
	if requestsPerMinute <= 0 {
		return nil
	}
	interval := time.Minute / time.Duration(requestsPerMinute)
	if interval < time.Nanosecond {
		interval = time.Nanosecond
	}
	return &rateLimiter{interval: interval}
}

func (l *rateLimiter) wait(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if l == nil {
			return nil
		}
		l.mu.Lock()
		now := time.Now()
		delay := l.next.Sub(now)
		if delay <= 0 {
			l.next = now.Add(l.interval)
			l.mu.Unlock()
			return nil
		}
		l.mu.Unlock()

		// Do not reserve a future slot: canceled waiters must not delay other
		// requests. Recheck availability after waking to serialize concurrent callers.
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
