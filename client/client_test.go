package client

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	return New(server.URL, WithHTTPClient(server.Client()))
}

func TestAuthenticatedRequestSendsBearer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization: Bearer test-token, got %q", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("test-token")

	_, err := c.Get(context.Background(), "/items")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAllRequestsSendAcceptJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if accept != "application/json" {
			t.Errorf("expected Accept: application/json, got %q", accept)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	ctx := context.Background()
	for _, fn := range []func() (*Response, error){
		func() (*Response, error) { return c.Get(ctx, "/a") },
		func() (*Response, error) { return c.Delete(ctx, "/b") },
		func() (*Response, error) {
			return c.DoUnauthenticated(ctx, http.MethodGet, "/c", nil)
		},
	} {
		if _, err := fn(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPostPatchPutSendContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("%s: expected Content-Type: application/json, got %q", r.Method, ct)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")
	ctx := context.Background()
	body := bytes.NewBufferString(`{"key":"value"}`)

	for _, fn := range []func() (*Response, error){
		func() (*Response, error) { return c.Post(ctx, "/a", bytes.NewBufferString(body.String())) },
		func() (*Response, error) { return c.Patch(ctx, "/b", bytes.NewBufferString(body.String())) },
		func() (*Response, error) { return c.Put(ctx, "/c", bytes.NewBufferString(body.String())) },
	} {
		if _, err := fn(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDeleteNilBodyNoContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "" {
			t.Errorf("expected no Content-Type on DELETE, got %q", ct)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.Delete(context.Background(), "/item/1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoUnauthenticatedNoAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("expected no Authorization header, got %q", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("should-not-appear")

	_, err := c.DoUnauthenticated(context.Background(), http.MethodGet, "/public", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestErrorResponseReturnsAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"400 Bad Request", 400, `{"error":"bad request"}`},
		{"401 Unauthorized", 401, "Unauthorized"},
		{"404 Not Found", 404, "not found"},
		{"500 Internal Server Error", 500, "internal error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c := newTestClient(t, server)
			c.SetToken("tok")

			_, err := c.Get(context.Background(), "/fail")
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var apiErr *models.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *models.APIError, got %T: %v", err, err)
			}

			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, apiErr.StatusCode)
			}
			if apiErr.Body != tt.body {
				t.Errorf("expected body %q, got %q", tt.body, apiErr.Body)
			}
			if !apiErr.IsStatusCode(tt.statusCode) {
				t.Errorf("IsStatusCode(%d) returned false", tt.statusCode)
			}
		})
	}
}

func TestResponsePreservesHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Custom", "test-value")
		w.Header().Set("Location", "/items/123")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.DoUnauthenticated(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Header.Get("X-Custom") != "test-value" {
		t.Errorf("expected X-Custom: test-value, got %q", resp.Header.Get("X-Custom"))
	}
	if resp.Header.Get("Location") != "/items/123" {
		t.Errorf("expected Location: /items/123, got %q", resp.Header.Get("Location"))
	}
}

func TestTokenGetterSetter(t *testing.T) {
	c := New("https://example.com")
	if c.Token() != "" {
		t.Errorf("expected empty token, got %q", c.Token())
	}
	c.SetToken("abc")
	if c.Token() != "abc" {
		t.Errorf("expected token abc, got %q", c.Token())
	}
}

func TestDecodeJSON(t *testing.T) {
	resp := &Response{
		StatusCode: 200,
		Body:       []byte(`{"status":"ok","description":"running"}`),
	}
	var result struct {
		Status      string `json:"status"`
		Description string `json:"description"`
	}
	if err := DecodeJSON(resp, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" {
		t.Errorf("expected status ok, got %q", result.Status)
	}
	if result.Description != "running" {
		t.Errorf("expected description running, got %q", result.Description)
	}
}

func TestUUIDFromLocationHeader(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		resp := &Response{Header: http.Header{"Location": {"/customer-api/v1/object/abc-123"}}}
		uuid, err := UUIDFromLocationHeader(resp)
		if err != nil {
			t.Fatal(err)
		}
		if uuid != "abc-123" {
			t.Errorf("expected abc-123, got %q", uuid)
		}
	})

	t.Run("missing header", func(t *testing.T) {
		resp := &Response{Header: http.Header{}}
		_, err := UUIDFromLocationHeader(resp)
		if err == nil {
			t.Fatal("expected error for missing Location header")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		resp := &Response{Header: http.Header{"Location": {"/"}}}
		_, err := UUIDFromLocationHeader(resp)
		if err == nil {
			t.Fatal("expected error for empty path")
		}
	})
}

func TestIntFromLocationIDHeader(t *testing.T) {
	resp := &Response{Header: http.Header{"Location-Id": {"42"}}}
	id, err := IntFromLocationIDHeader(resp)
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestIntFromLocationIDHeaderMissing(t *testing.T) {
	resp := &Response{Header: http.Header{}}
	_, err := IntFromLocationIDHeader(resp)
	if err == nil {
		t.Fatal("expected error for missing Location-Id header")
	}
}

func TestIntFromLocationIDHeaderInvalid(t *testing.T) {
	resp := &Response{Header: http.Header{"Location-Id": {"not-a-number"}}}
	_, err := IntFromLocationIDHeader(resp)
	if err == nil {
		t.Fatal("expected error for non-numeric Location-Id header")
	}
}

func TestGetRawNoAcceptHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if accept := r.Header.Get("Accept"); accept == "application/json" {
			t.Error("GetRaw should not set Accept: application/json")
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization: Bearer test-token, got %q", auth)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("binary-data"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("test-token")

	resp, err := c.GetRaw(context.Background(), "/file/123/data")
	if err != nil {
		t.Fatal(err)
	}
	if string(resp.Body) != "binary-data" {
		t.Errorf("expected binary-data, got %q", string(resp.Body))
	}
}

func TestNewRequestInvalidURL(t *testing.T) {
	c := &Client{
		httpClient: http.DefaultClient,
		baseURL:    "://bad",
	}
	c.SetToken("tok")

	_, err := c.Get(context.Background(), "/items")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestDoTransportError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	c := newTestClient(t, server)
	c.SetToken("tok")
	server.Close()

	_, err := c.Get(context.Background(), "/items")
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestDoRawTransportError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	c := newTestClient(t, server)
	c.SetToken("tok")
	server.Close()

	_, err := c.GetRaw(context.Background(), "/file/123/data")
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestPostMultipartContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "multipart/form-data; boundary=testboundary" {
			t.Errorf("expected multipart Content-Type, got %q", ct)
		}
		if accept := r.Header.Get("Accept"); accept == "application/json" {
			t.Error("PostMultipart should not set Accept: application/json")
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer tok" {
			t.Errorf("expected Bearer tok, got %q", auth)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	resp, err := c.PostMultipart(context.Background(), "/file", bytes.NewBufferString("body"), "multipart/form-data; boundary=testboundary")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}
