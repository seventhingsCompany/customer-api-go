package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestUsersListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/users" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		page := 1
		perPage := 10
		sortBy := models.UserSortByEmail
		order := models.UserSortOrderAsc
		_ = page
		_ = perPage
		_ = sortBy
		_ = order
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("expected page=1, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("per_page") != "10" {
			t.Errorf("expected per_page=10, got %s", r.URL.Query().Get("per_page"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"uuid":"u1","id":1,"email":"a@b.com"}],"page":1,"per_page":10,"sort_by":"email","order":"asc","total":1}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	page := 1
	perPage := 10
	sortBy := models.UserSortByEmail
	order := models.UserSortOrderAsc
	result, err := c.UsersList(context.Background(), &models.UserListOptions{
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
	if result.Items[0].UUID != "u1" {
		t.Errorf("expected uuid u1, got %s", result.Items[0].UUID)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestUsersListNilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query string, got %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[],"page":1,"per_page":25,"sort_by":"id","order":"asc","total":0}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	result, err := c.UsersList(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result.Items))
	}
}

func TestUserGetByUUID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/user/u1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"u1","id":1,"email":"a@b.com","firstname":"Alice","lastname":"Smith","display_name":"Alice S"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	user, err := c.UserGet(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "a@b.com" {
		t.Errorf("expected a@b.com, got %s", user.Email)
	}
	if user.Firstname == nil || *user.Firstname != "Alice" {
		t.Errorf("expected firstname Alice, got %v", user.Firstname)
	}
}

func TestUserGetByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/user/by-id/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"u42","id":42,"email":"b@c.com"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	user, err := c.UserGetByID(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != 42 {
		t.Errorf("expected id 42, got %d", user.ID)
	}
}

func TestUserGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.UserGet(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}
