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

func TestRentalCasesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/rental-management/rental-cases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"uuid":"rc1","status":"borrowed","renter":{"type":"plain","value":"John"}}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	cases, err := c.RentalCasesList(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(cases))
	}
	if cases[0].UUID != "rc1" {
		t.Errorf("expected uuid rc1, got %s", cases[0].UUID)
	}
}

func TestRentalCasesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "page=2&per_page=5" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.RentalCasesList(context.Background(), &models.ListOptions{Page: 2, PerPage: 5})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRentalCaseCreateLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/rental-management/rental-case" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Location", "/customer-api/v1/rental-management/rental-case/new-rc-uuid")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	uuid, err := c.RentalCaseCreate(context.Background(), models.CreateRentalCase{})
	if err != nil {
		t.Fatal(err)
	}
	if uuid != "new-rc-uuid" {
		t.Errorf("expected new-rc-uuid, got %q", uuid)
	}
}

func TestRentalCaseCreateBodyWithRenter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var input models.CreateRentalCase
		if err := json.Unmarshal(body, &input); err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		if input.Renter == nil {
			t.Fatal("expected renter to be set")
		}
		if input.Renter.Type != models.RenterTypePlain || input.Renter.Value != "John" {
			t.Errorf("unexpected renter: %+v", input.Renter)
		}
		w.Header().Set("Location", "/customer-api/v1/rental-management/rental-case/uuid")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.RentalCaseCreate(context.Background(), models.CreateRentalCase{
		Renter: &models.RentalCaseRenter{Type: models.RenterTypePlain, Value: "John"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRentalCaseGetWithRenterString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/rental-management/rental-case/rc1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"rc1","status":"borrowed","renter":{"type":"plain","value":"John"},"created_at":"2024-01-01","updated_at":"2024-01-02"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	rc, err := c.RentalCaseGet(context.Background(), "rc1")
	if err != nil {
		t.Fatal(err)
	}
	if rc.Renter == nil || rc.Renter.Value != "John" {
		t.Errorf("expected renter John, got %v", rc.Renter)
	}
}

func TestRentalCaseGetWithNullRenter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"rc2","status":"requested","renter":null,"created_at":"2024-01-01","updated_at":"2024-01-02"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	rc, err := c.RentalCaseGet(context.Background(), "rc2")
	if err != nil {
		t.Fatal(err)
	}
	if rc.Renter != nil {
		t.Errorf("expected nil renter, got %v", rc.Renter)
	}
}

func TestRentalCaseUpdatePUT204(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/rental-management/rental-case/rc1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.RentalCaseUpdate(context.Background(), "rc1", models.UpdateRentalCase{Comment: "updated"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRentalCaseDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/rental-management/rental-case/rc1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.RentalCaseDelete(context.Background(), "rc1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRentalCasesListError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.RentalCasesList(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestRentalCaseCreateError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.RentalCaseCreate(context.Background(), models.CreateRentalCase{})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestRentalCaseUpdateError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.RentalCaseUpdate(context.Background(), "rc1", models.UpdateRentalCase{Comment: "updated"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestRentalCaseGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.RentalCaseGet(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}
