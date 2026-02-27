package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestPingSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","description":"seventhings API is running"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.Ping(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %q", resp.Status)
	}
	if resp.Description != "seventhings API is running" {
		t.Errorf("expected description 'seventhings API is running', got %q", resp.Description)
	}
}

func TestPingNoAuthHeaderEvenWithToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("expected no Authorization header on Ping, got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","description":"running"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("should-not-appear")

	_, err := c.Ping(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestPingServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *models.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *models.APIError, got %T: %v", err, err)
	}
	if !apiErr.IsStatusCode(http.StatusInternalServerError) {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

func TestPingCancelledContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := c.Ping(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
