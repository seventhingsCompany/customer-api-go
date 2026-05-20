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

func TestPersonsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/persons" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("per_page") != "10" {
			t.Errorf("expected per_page=10, got %s", r.URL.Query().Get("per_page"))
		}
		if r.URL.Query().Get("sort_by") != "email" {
			t.Errorf("expected sort_by=email, got %s", r.URL.Query().Get("sort_by"))
		}
		if r.URL.Query().Get("order") != "desc" {
			t.Errorf("expected order=desc, got %s", r.URL.Query().Get("order"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"person_uuid":"p1","id":1,"email":"a@b.com"}],"page":2,"per_page":10,"sort_by":"email","order":"desc","total":1}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	page := 2
	perPage := 10
	sortBy := "email"
	order := models.UserSortOrderDesc
	result, err := c.PersonsList(context.Background(), &models.PersonListOptions{
		Page:    &page,
		PerPage: &perPage,
		SortBy:  &sortBy,
		Order:   &order,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].UUID != "p1" {
		t.Errorf("expected uuid p1, got %s", result.Items[0].UUID)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestPersonsListNilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query string, got %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[],"page":1,"per_page":50,"sort_by":"id","order":"asc","total":0}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	result, err := c.PersonsList(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result.Items))
	}
}

func TestPersonGetByUUID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/person/p1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"person_uuid":"p1","id":1,"email":"a@b.com","first_name":"Alice","last_name":"Smith"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	person, err := c.PersonGet(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if person.Email != "a@b.com" {
		t.Errorf("expected a@b.com, got %s", person.Email)
	}
	if person.Firstname == nil || *person.Firstname != "Alice" {
		t.Errorf("expected firstname Alice, got %v", person.Firstname)
	}
}

func TestPersonGetByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/person/by-id/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"person_uuid":"p42","id":42,"email":"b@c.com"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	person, err := c.PersonGetByID(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if person.ID != 42 {
		t.Errorf("expected id 42, got %d", person.ID)
	}
}

func TestPersonCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/persons" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("body not valid JSON: %v", err)
		}
		fields, ok := got["fields"].(map[string]any)
		if !ok {
			t.Fatalf("expected fields map, got %v", got)
		}
		if fields["email"] != "max@example.com" {
			t.Errorf("expected fields.email=max@example.com, got %v", fields["email"])
		}
		w.Header().Set("Location", "/customer-api/v1/person/new-uuid-123")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	uuid, err := c.PersonCreate(context.Background(), map[string]any{
		"email":     "max@example.com",
		"firstname": "Max",
	})
	if err != nil {
		t.Fatal(err)
	}
	if uuid != "new-uuid-123" {
		t.Errorf("expected new-uuid-123, got %s", uuid)
	}
}

func TestPersonCreateUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/persons/create-user" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("body not valid JSON: %v", err)
		}
		if _, hasSort := got["sort"]; hasSort {
			t.Errorf("sort must not be sent, body=%s", string(body))
		}
		filter, ok := got["filter"].(map[string]any)
		if !ok {
			t.Fatalf("expected filter map, got %v", got)
		}
		email, ok := filter["email"].(map[string]any)
		if !ok {
			t.Fatalf("expected filter.email map, got %v", filter)
		}
		if email["eq"] != "tester@domain.de" {
			t.Errorf("expected filter.email.eq=tester@domain.de, got %v", email["eq"])
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.PersonCreateUser(context.Background(), models.FilterObject{
		Filter: map[string]map[models.FilterOperator]any{
			"email": {models.FilterEq: "tester@domain.de"},
		},
		Sort: map[string]models.SortDirection{"name": models.SortASC}, // must be dropped
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPersonsListError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.PersonsList(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestPersonGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.PersonGet(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}
