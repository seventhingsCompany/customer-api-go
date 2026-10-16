package models

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAPIErrorRetryAfter(t *testing.T) {
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name, value string
		want        time.Duration
		valid       bool
	}{
		{"seconds", "120", 2 * time.Minute, true},
		{"zero", "0", 0, true},
		{"whitespace", " 42 \t", 42 * time.Second, true},
		{"future date", now.Add(90 * time.Second).Format(http.TimeFormat), 90 * time.Second, true},
		{"past date", now.Add(-time.Hour).Format(http.TimeFormat), 0, true},
		{"current date", now.Format(http.TimeFormat), 0, true},
		{"missing", "", 0, false},
		{"negative", "-1", 0, false},
		{"explicit plus", "+1", 0, false},
		{"fractional", "1.5", 0, false},
		{"invalid date", "tomorrow", 0, false},
		{"multiple values", "1, 2", 0, false},
		{"integer overflow", "18446744073709551616", 0, false},
		{"duration overflow", "9223372037", 0, false},
		{"date duration overflow", "Fri, 31 Dec 9999 23:59:59 GMT", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := &APIError{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": {tc.value}}}
			if delay, ok := e.RetryAfter(now); delay != tc.want || ok != tc.valid {
				t.Fatalf("RetryAfter = %s, %v; want %s, %v", delay, ok, tc.want, tc.valid)
			}
		})
	}
	for _, e := range []*APIError{nil, {StatusCode: http.StatusTooManyRequests}} {
		if delay, ok := e.RetryAfter(now); delay != 0 || ok {
			t.Fatalf("missing error/header: %s, %v", delay, ok)
		}
	}
	e := &APIError{Header: http.Header{"Retry-After": {"1", "2"}}}
	if delay, ok := e.RetryAfter(now); delay != 0 || ok {
		t.Fatalf("repeated header: %s, %v", delay, ok)
	}
}

func TestAPIErrorDoesNotLogHeaders(t *testing.T) {
	e := &APIError{StatusCode: http.StatusTooManyRequests, Status: "429 Too Many Requests", Body: "too many requests", Header: http.Header{"X-Request-Id": {"header-only-value"}}}
	if strings.Contains(e.Error(), "header-only-value") {
		t.Fatal("Error() must not include response headers")
	}
}
