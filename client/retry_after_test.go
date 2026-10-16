package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestErrorHeadersAndNoAutomaticRetry(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Retry-After", "120")
		w.Header().Set("X-Request-ID", "request-id")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, "Too Many Requests")
	}))
	defer server.Close()
	c := newTestClient(t, server)
	for name, call := range map[string]func(context.Context) (*Response, error){
		"authenticated": func(ctx context.Context) (*Response, error) { return c.Get(ctx, "objects") },
		"unauthenticated": func(ctx context.Context) (*Response, error) {
			return c.DoUnauthenticated(ctx, http.MethodPost, "auth_token", strings.NewReader(`{}`))
		},
		"download": func(ctx context.Context) (*Response, error) { return c.GetRaw(ctx, "file/f/data") },
		"upload": func(ctx context.Context) (*Response, error) {
			return c.PostMultipart(ctx, "file", strings.NewReader("data"), "multipart/form-data")
		},
	} {
		t.Run(name, func(t *testing.T) {
			before := requests.Load()
			resp, err := call(t.Context())
			var apiErr *models.APIError
			if resp != nil || !errors.As(fmt.Errorf("wrapped: %w", err), &apiErr) {
				t.Fatalf("response = %v, error = %v", resp, err)
			}
			if !apiErr.IsRateLimited() || apiErr.Body != "Too Many Requests" || apiErr.Header.Get("X-Request-ID") != "request-id" {
				t.Fatalf("unexpected API error: %+v", apiErr)
			}
			if delay, ok := apiErr.RetryAfter(time.Now()); !ok || delay != 2*time.Minute {
				t.Fatalf("RetryAfter = %s, %v", delay, ok)
			}
			if got := requests.Load() - before; got != 1 {
				t.Fatalf("sent %d requests; want exactly one", got)
			}
		})
	}
}

func TestAPIErrorHeadersAreCloned(t *testing.T) {
	header := http.Header{"Retry-After": {"42"}, "X-Request-Id": {"original"}}
	c := New("https://example.test", WithHTTPClient(&http.Client{Transport: rateLimitTransport(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: header, Body: http.NoBody}, nil
	})}))
	_, err := c.Get(t.Context(), "objects")
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v", err)
	}
	header["Retry-After"][0] = "99"
	header.Set("X-Request-ID", "changed")
	if apiErr.Header.Get("Retry-After") != "42" || apiErr.Header.Get("X-Request-ID") != "original" {
		t.Fatalf("error header aliased the transport header: %v", apiErr.Header)
	}
}
