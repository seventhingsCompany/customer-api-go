//go:build integration

package seventhings_test

import (
	"net/http"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/SeventhingsCompany/customer-api-go/client"
)

type liveRateLimitTransport struct {
	mu     sync.Mutex
	starts []time.Time
}

func (r *liveRateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.starts = append(r.starts, time.Now())
	r.mu.Unlock()
	return http.DefaultTransport.RoundTrip(req)
}

func TestIntegrationRateLimit(t *testing.T) {
	baseURL := os.Getenv("SEVENTHINGS_BASE_URL")
	if baseURL == "" {
		t.Skip("set SEVENTHINGS_BASE_URL to run integration tests")
	}
	for _, tc := range []struct {
		name     string
		options  []client.Option
		interval time.Duration
	}{
		{name: "default200", interval: 300 * time.Millisecond},
		{name: "override400", options: []client.Option{client.WithRateLimit(400)}, interval: 150 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			transport := &liveRateLimitTransport{}
			options := append([]client.Option{client.WithHTTPClient(&http.Client{Transport: transport, Timeout: 30 * time.Second})}, tc.options...)
			c := client.New(baseURL, options...)
			var wg sync.WaitGroup
			for range 3 {
				wg.Go(func() {
					if _, err := c.Ping(t.Context()); err != nil {
						t.Error(err)
					}
				})
			}
			wg.Wait()
			if len(transport.starts) != 3 {
				t.Fatalf("sent %d requests, want 3", len(transport.starts))
			}
			sort.Slice(transport.starts, func(i, j int) bool { return transport.starts[i].Before(transport.starts[j]) })
			for i := 1; i < len(transport.starts); i++ {
				gap := transport.starts[i].Sub(transport.starts[i-1])
				t.Logf("live request spacing: %s (minimum %s)", gap, tc.interval)
				if gap < tc.interval-time.Millisecond {
					t.Errorf("requests sent too close together: %s", gap)
				}
			}
		})
	}
}
