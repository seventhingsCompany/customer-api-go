package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

type rateLimitTransport func(*http.Request) (*http.Response, error)

func (f rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func rateLimitHTTPClient(record func(time.Time)) *http.Client {
	return &http.Client{Transport: rateLimitTransport(func(req *http.Request) (*http.Response, error) {
		if record != nil {
			record(time.Now())
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"access_token":"token"}`))}, nil
	})}
}

func TestDefaultRateLimitAcrossRequestTypes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var sent []time.Time
		c := New("https://example.test", WithHTTPClient(rateLimitHTTPClient(func(now time.Time) { sent = append(sent, now) })))
		ctx := t.Context()
		calls := []func() error{
			func() error { _, err := c.Get(ctx, "objects"); return err },
			func() error { _, err := c.GetRaw(ctx, "file/f/data"); return err },
			func() error {
				_, err := c.PostMultipart(ctx, "file", strings.NewReader("data"), "multipart/form-data")
				return err
			},
			func() error { _, err := c.Login(ctx, "user", "password", "client"); return err },
		}
		for _, call := range calls {
			if err := call(); err != nil {
				t.Fatal(err)
			}
		}
		if DefaultRequestsPerMinute != 200 {
			t.Fatalf("default = %d", DefaultRequestsPerMinute)
		}
		for i := 1; i < len(sent); i++ {
			if gap := sent[i].Sub(sent[i-1]); gap != 300*time.Millisecond {
				t.Errorf("request spacing = %s", gap)
			}
		}
	})
}

func TestRateLimitOptions(t *testing.T) {
	for _, limit := range []int{60, 600, 0, -1} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				c := NewWithToken("https://example.test", "token", WithHTTPClient(rateLimitHTTPClient(nil)), WithRateLimit(limit))
				start := time.Now()
				for range 3 {
					if _, err := c.Get(t.Context(), "objects"); err != nil {
						t.Fatal(err)
					}
				}
				var want time.Duration
				if limit > 0 {
					want = 2 * (time.Minute / time.Duration(limit))
				}
				if elapsed := time.Since(start); elapsed != want {
					t.Errorf("elapsed = %s, want %s", elapsed, want)
				}
			})
		})
	}
}

func TestRateLimitCancellationDoesNotConsumeSlot(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		count := 0
		c := New("https://example.test", WithHTTPClient(rateLimitHTTPClient(func(time.Time) { count++ })))
		if _, err := c.Get(t.Context(), "objects"); err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()
		if _, err := c.Get(ctx, "objects"); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("error = %v", err)
		}
		if count != 1 {
			t.Fatalf("canceled request reached transport: count = %d", count)
		}
		if _, err := c.Get(t.Context(), "objects"); err != nil {
			t.Fatal(err)
		}
		if elapsed := time.Since(start); elapsed != 300*time.Millisecond {
			t.Fatalf("cancellation consumed a slot: %s", elapsed)
		}
		ctx, cancelNow := context.WithCancel(t.Context())
		cancelNow()
		if _, err := c.Get(ctx, "objects"); !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v", err)
		}
		if count != 2 {
			t.Fatalf("canceled request reached transport: count = %d", count)
		}
	})
}

func TestRateLimitConcurrentRequests(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var mu sync.Mutex
		var sent []time.Time
		c := New("https://example.test", WithHTTPClient(rateLimitHTTPClient(func(now time.Time) {
			mu.Lock()
			defer mu.Unlock()
			sent = append(sent, now)
		})))
		var wg sync.WaitGroup
		for range 10 {
			wg.Go(func() {
				if _, err := c.Get(t.Context(), "objects"); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		for i := 1; i < len(sent); i++ {
			if gap := sent[i].Sub(sent[i-1]); gap < 300*time.Millisecond {
				t.Errorf("concurrent request spacing = %s", gap)
			}
		}
	})
}

func TestRateLimitIdleDoesNotAccumulateBurst(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := New("https://example.test", WithHTTPClient(rateLimitHTTPClient(nil)))
		if _, err := c.Get(t.Context(), "objects"); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Minute)
		start := time.Now()
		for range 2 {
			if _, err := c.Get(t.Context(), "objects"); err != nil {
				t.Fatal(err)
			}
		}
		if elapsed := time.Since(start); elapsed != 300*time.Millisecond {
			t.Fatalf("idle burst: elapsed = %s", elapsed)
		}
	})
}

func TestRateLimitCredentialsAndIndependentClients(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		c, err := NewWithCredentials(t.Context(), "https://example.test", "user", "password", "client", WithHTTPClient(rateLimitHTTPClient(nil)), WithRateLimit(600))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := c.Get(t.Context(), "objects"); err != nil {
			t.Fatal(err)
		}
		if elapsed := time.Since(start); elapsed != 100*time.Millisecond {
			t.Fatalf("login did not share the overridden limit: %s", elapsed)
		}
		other := NewWithToken("https://example.test", "token", WithHTTPClient(rateLimitHTTPClient(nil)))
		start = time.Now()
		if _, err := other.Get(t.Context(), "objects"); err != nil {
			t.Fatal(err)
		}
		if elapsed := time.Since(start); elapsed != 0 {
			t.Fatalf("independent client was throttled: %s", elapsed)
		}
	})
}

func TestRateLimitFIFO(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var mu sync.Mutex
		var order []string
		httpClient := &http.Client{Transport: rateLimitTransport(func(req *http.Request) (*http.Response, error) {
			mu.Lock()
			order = append(order, req.URL.Path)
			mu.Unlock()
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: http.NoBody}, nil
		})}
		c := New("https://example.test", WithHTTPClient(httpClient))
		if _, err := c.Get(t.Context(), "prime"); err != nil {
			t.Fatal(err)
		}
		want := []string{"/customer-api/v1/prime"}
		var wg sync.WaitGroup
		for i := range 30 {
			path := "queued-" + strconv.Itoa(i)
			want = append(want, "/customer-api/v1/"+path)
			wg.Go(func() {
				if _, err := c.Get(t.Context(), path); err != nil {
					t.Error(err)
				}
			})
			// Establish enqueue order before advancing the virtual clock.
			synctest.Wait()
			// Keep new requests arriving faster than the limiter can serve them.
			time.Sleep(50 * time.Millisecond)
		}
		wg.Wait()
		if !reflect.DeepEqual(order, want) {
			t.Fatalf("requests overtook earlier waiters:\n got %v\nwant %v", order, want)
		}
	})
}

func TestRateLimitFIFOCancellation(t *testing.T) {
	for _, canceled := range []int{0, 1, 2} {
		t.Run(strconv.Itoa(canceled), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				limiter := newRateLimiter(200)
				if err := limiter.wait(t.Context()); err != nil {
					t.Fatal(err)
				}
				start := time.Now()
				var mu sync.Mutex
				var order []int
				var wg sync.WaitGroup
				contexts := make([]context.Context, 3)
				cancels := make([]context.CancelFunc, 3)
				for i := range contexts {
					contexts[i], cancels[i] = context.WithCancel(t.Context())
					defer cancels[i]()
					wg.Go(func() {
						err := limiter.wait(contexts[i])
						if i == canceled {
							if !errors.Is(err, context.Canceled) {
								t.Errorf("canceled wait error = %v", err)
							}
							return
						}
						if err != nil {
							t.Error(err)
							return
						}
						mu.Lock()
						order = append(order, i)
						mu.Unlock()
					})
					synctest.Wait()
				}
				time.Sleep(100 * time.Millisecond)
				cancels[canceled]()
				synctest.Wait()
				wg.Wait()
				var want []int
				for i := range contexts {
					if i != canceled {
						want = append(want, i)
					}
				}
				if !reflect.DeepEqual(order, want) {
					t.Fatalf("order = %v, want %v", order, want)
				}
				if elapsed := time.Since(start); elapsed != 600*time.Millisecond {
					t.Fatalf("canceled waiter consumed a slot: %s", elapsed)
				}
				// A drained queue must remain usable.
				if err := limiter.wait(t.Context()); err != nil {
					t.Fatal(err)
				}
				if elapsed := time.Since(start); elapsed != 900*time.Millisecond {
					t.Fatalf("unexpected next slot: %s", elapsed)
				}
			})
		})
	}
}

func TestRateLimitFIFOAllWaitersCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		limiter := newRateLimiter(200)
		if err := limiter.wait(t.Context()); err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		var wg sync.WaitGroup
		for range 10 {
			wg.Go(func() {
				if err := limiter.wait(ctx); !errors.Is(err, context.Canceled) {
					t.Errorf("wait error = %v", err)
				}
			})
			synctest.Wait()
		}
		time.Sleep(100 * time.Millisecond)
		cancel()
		wg.Wait()
		limiter.mu.Lock()
		remaining := limiter.waiters.Len()
		limiter.mu.Unlock()
		if remaining != 0 {
			t.Fatalf("canceled waiters left in queue: %d", remaining)
		}
		if err := limiter.wait(t.Context()); err != nil {
			t.Fatal(err)
		}
		if elapsed := time.Since(start); elapsed != 300*time.Millisecond {
			t.Fatalf("canceled queue consumed slots: %s", elapsed)
		}
	})
}
