package client

import (
	"container/list"
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
// Concurrent waiters are admitted in FIFO enqueue order; canceled waiters are
// removed without consuming a request slot.
func WithRateLimit(requestsPerMinute int) Option {
	return func(c *Client) {
		c.rateLimiter = newRateLimiter(requestsPerMinute)
	}
}

type rateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time
	waiters  list.List // FIFO of chan struct{}; only the head may wait for a time slot.
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
	if err := ctx.Err(); err != nil {
		return err
	}
	if l == nil {
		return nil
	}

	ready := make(chan struct{})
	l.mu.Lock()
	element := l.waiters.PushBack(ready)
	if l.waiters.Front() == element {
		close(ready)
	}
	l.mu.Unlock()
	defer l.removeWaiter(element)

	// Only the queue head waits for the next time slot. New arrivals cannot
	// overtake existing waiters, and non-head cancellation never reserves a slot.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ready:
	}

	for {
		l.mu.Lock()
		if err := ctx.Err(); err != nil {
			l.mu.Unlock()
			return err
		}
		now := time.Now()
		delay := l.next.Sub(now)
		if delay <= 0 {
			l.next = now.Add(l.interval)
			l.mu.Unlock()
			return nil
		}
		l.mu.Unlock()

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (l *rateLimiter) removeWaiter(element *list.Element) {
	l.mu.Lock()
	defer l.mu.Unlock()
	wasHead := l.waiters.Front() == element
	l.waiters.Remove(element)
	if next := l.waiters.Front(); wasHead && next != nil {
		close(next.Value.(chan struct{}))
	}
}
