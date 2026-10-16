package client

import (
	"context"
	"errors"
	"io"
	"net/http"
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
			func() error { _, err := c.LoginAuthorizationCode(ctx, "code", "verifier", "uri", "client"); return err },
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
