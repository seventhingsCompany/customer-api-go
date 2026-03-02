package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestLocationsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/locations" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.RawQuery != "page=2&per_page=20" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"uuid":"l1","name":"HQ"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	result, err := c.LocationsList(context.Background(), &models.ListOptions{Page: 2, PerPage: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestLocationsCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/locations/count" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":3}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	count, err := c.LocationsCount(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestLocationCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/location" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Location", "/customer-api/v1/location/loc-uuid")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	uuid, err := c.LocationCreate(context.Background(), map[string]any{"name": "Branch"})
	if err != nil {
		t.Fatal(err)
	}
	if uuid != "loc-uuid" {
		t.Errorf("expected loc-uuid, got %q", uuid)
	}
}

func TestLocationGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/location/l1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"l1","name":"HQ"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	loc, err := c.LocationGet(context.Background(), "l1")
	if err != nil {
		t.Fatal(err)
	}
	if loc["name"] != "HQ" {
		t.Errorf("expected HQ, got %v", loc["name"])
	}
}

func TestLocationPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"l1","name":"Updated"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	loc, err := c.LocationPatch(context.Background(), "l1", map[string]any{"name": "Updated"})
	if err != nil {
		t.Fatal(err)
	}
	if loc["name"] != "Updated" {
		t.Errorf("expected Updated, got %v", loc["name"])
	}
}

func TestLocationDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.LocationDelete(context.Background(), "l1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLocationsListError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.LocationsList(context.Background(), &models.ListOptions{Page: 1, PerPage: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestLocationsCountError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.LocationsCount(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestLocationCreateError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.LocationCreate(context.Background(), map[string]any{"name": "Branch"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestLocationPatchError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.LocationPatch(context.Background(), "l1", map[string]any{"name": "Updated"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestLocationGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.LocationGet(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}
