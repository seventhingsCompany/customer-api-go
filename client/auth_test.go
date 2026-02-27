package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

const tokenJSON = `{
	"access_token": "eyJhbGciOiJIUzI1NiJ9.test",
	"expires_in": 3600,
	"token_type": "Bearer",
	"scope": null,
	"refresh_token": "refresh-abc",
	"user_id": 42
}`

func tokenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tokenJSON))
	}))
}

func TestLoginSuccess(t *testing.T) {
	server := tokenServer(t)
	defer server.Close()

	c := newTestClient(t, server)
	tok, err := c.Login(context.Background(), "user@example.com", "secret", "my-client")
	if err != nil {
		t.Fatal(err)
	}

	if tok.AccessToken != "eyJhbGciOiJIUzI1NiJ9.test" {
		t.Errorf("unexpected access token: %s", tok.AccessToken)
	}
	if tok.ExpiresIn != 3600 {
		t.Errorf("expected expires_in 3600, got %d", tok.ExpiresIn)
	}
	if tok.RefreshToken != "refresh-abc" {
		t.Errorf("unexpected refresh token: %s", tok.RefreshToken)
	}
	if tok.UserID != 42 {
		t.Errorf("expected user_id 42, got %d", tok.UserID)
	}
	if tok.Scope != nil {
		t.Errorf("expected nil scope, got %v", *tok.Scope)
	}
	if c.Token() != tok.AccessToken {
		t.Errorf("token not auto-set on client")
	}
	if c.ClientID() != "my-client" {
		t.Errorf("expected clientID my-client, got %q", c.ClientID())
	}
}

func TestLoginSendsCorrectBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("expected no Authorization header, got %q", auth)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if req["grant_type"] != "password" {
			t.Errorf("expected grant_type password, got %q", req["grant_type"])
		}
		if req["username"] != "user@example.com" {
			t.Errorf("expected username user@example.com, got %q", req["username"])
		}
		if req["password"] != "secret" {
			t.Errorf("expected password secret, got %q", req["password"])
		}
		if req["client_id"] != "my-client" {
			t.Errorf("expected client_id my-client, got %q", req["client_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tokenJSON))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.Login(context.Background(), "user@example.com", "secret", "my-client")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoginUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.Login(context.Background(), "user@example.com", "wrong", "my-client")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *models.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *models.APIError, got %T: %v", err, err)
	}
	if !apiErr.IsStatusCode(http.StatusUnauthorized) {
		t.Errorf("expected 401, got %d", apiErr.StatusCode)
	}
}

func TestLoginForbiddenWithDetail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"detail":"Banned"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.Login(context.Background(), "user@example.com", "pass", "my-client")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *models.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *models.APIError, got %T: %v", err, err)
	}
	if !apiErr.IsStatusCode(http.StatusForbidden) {
		t.Errorf("expected 403, got %d", apiErr.StatusCode)
	}

	var detail models.LoginDeniedDetail
	if err := json.Unmarshal([]byte(apiErr.Body), &detail); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	if detail.Detail != models.LoginDeniedBanned {
		t.Errorf("expected detail Banned, got %q", detail.Detail)
	}
}

func TestRefreshSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if req["grant_type"] != "refresh_token" {
			t.Errorf("expected grant_type refresh_token, got %q", req["grant_type"])
		}
		if req["refresh_token"] != "refresh-abc" {
			t.Errorf("expected refresh_token refresh-abc, got %q", req["refresh_token"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tokenJSON))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.clientID = "my-client"
	tok, err := c.Refresh(context.Background(), "refresh-abc")
	if err != nil {
		t.Fatal(err)
	}
	if c.Token() != tok.AccessToken {
		t.Errorf("token not auto-set on client")
	}
}

func TestRefreshUsesStoredClientID(t *testing.T) {
	var receivedClientID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		_ = json.Unmarshal(body, &req)
		receivedClientID = req["client_id"]

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tokenJSON))
	}))
	defer server.Close()

	c := newTestClient(t, server)

	// Login stores the clientID
	_, err := c.Login(context.Background(), "user@example.com", "pass", "stored-client")
	if err != nil {
		t.Fatal(err)
	}

	// Refresh should reuse it
	_, err = c.Refresh(context.Background(), "refresh-abc")
	if err != nil {
		t.Fatal(err)
	}
	if receivedClientID != "stored-client" {
		t.Errorf("expected client_id stored-client, got %q", receivedClientID)
	}
}

func TestLoginSSOSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if req["grant_type"] != "sso_auth_code" {
			t.Errorf("expected grant_type sso_auth_code, got %q", req["grant_type"])
		}
		if req["provider_name"] != "azure-open-id-connect" {
			t.Errorf("expected provider_name azure-open-id-connect, got %q", req["provider_name"])
		}
		if req["auth_code"] != "code-123" {
			t.Errorf("expected auth_code code-123, got %q", req["auth_code"])
		}
		if req["client_id"] != "sso-client" {
			t.Errorf("expected client_id sso-client, got %q", req["client_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tokenJSON))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	tok, err := c.LoginSSO(context.Background(), models.SSOProviderAzure, "code-123", "sso-client", nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Token() != tok.AccessToken {
		t.Errorf("token not auto-set on client")
	}
	if c.ClientID() != "sso-client" {
		t.Errorf("expected clientID sso-client, got %q", c.ClientID())
	}
}

func TestLoginSSOWithAppTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		at, ok := req["app_target"]
		if !ok {
			t.Error("expected app_target to be present")
		}
		if at != "mobile" {
			t.Errorf("expected app_target mobile, got %v", at)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tokenJSON))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	target := models.SSOAppTargetMobile
	_, err := c.LoginSSO(context.Background(), models.SSOProviderGoogle, "code-456", "sso-client", &target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoginSSOWithoutAppTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if _, ok := req["app_target"]; ok {
			t.Error("expected app_target to be absent when nil")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tokenJSON))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.LoginSSO(context.Background(), models.SSOProviderOneLogin, "code-789", "sso-client", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRevokeTokensSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %q", auth)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("test-token")

	err := c.RevokeTokens(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestRevokeTokensUnauthenticated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.RevokeTokens(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *models.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *models.APIError, got %T: %v", err, err)
	}
	if !apiErr.IsStatusCode(http.StatusUnauthorized) {
		t.Errorf("expected 401, got %d", apiErr.StatusCode)
	}
}

func TestNewWithCredentialsSuccess(t *testing.T) {
	server := tokenServer(t)
	defer server.Close()

	c, err := NewWithCredentials(context.Background(), server.URL, "user@example.com", "secret", "my-client", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatal(err)
	}
	if c.Token() != "eyJhbGciOiJIUzI1NiJ9.test" {
		t.Errorf("expected token to be set, got %q", c.Token())
	}
	if c.ClientID() != "my-client" {
		t.Errorf("expected clientID my-client, got %q", c.ClientID())
	}
}

func TestNewWithCredentialsLoginFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	_, err := NewWithCredentials(context.Background(), server.URL, "user@example.com", "wrong", "my-client", WithHTTPClient(server.Client()))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *models.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *models.APIError, got %T: %v", err, err)
	}
}

func TestNewWithToken(t *testing.T) {
	c := NewWithToken("https://example.seventhings.com", "pre-existing-token")
	if c.Token() != "pre-existing-token" {
		t.Errorf("expected token pre-existing-token, got %q", c.Token())
	}
}

func TestNewWithTokenAndClientID(t *testing.T) {
	c := NewWithToken("https://example.seventhings.com", "pre-existing-token", WithClientID("cid"))
	if c.Token() != "pre-existing-token" {
		t.Errorf("expected token pre-existing-token, got %q", c.Token())
	}
	if c.ClientID() != "cid" {
		t.Errorf("expected clientID cid, got %q", c.ClientID())
	}
}
