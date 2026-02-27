package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestRoomsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/rooms" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.RawQuery != "page=1&per_page=5" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"uuid":"r1","name":"Room A"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	result, err := c.RoomsList(context.Background(), 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
}

func TestRoomsListNoPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query string, got %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.RoomsList(context.Background(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRoomsCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/rooms/count" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":7}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	count, err := c.RoomsCount(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 7 {
		t.Errorf("expected 7, got %d", count)
	}
}

func TestRoomCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/room" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Location", "/customer-api/v1/room/room-uuid")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	uuid, err := c.RoomCreate(context.Background(), map[string]any{"name": "Room B"})
	if err != nil {
		t.Fatal(err)
	}
	if uuid != "room-uuid" {
		t.Errorf("expected room-uuid, got %q", uuid)
	}
}

func TestRoomGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/room/r1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"r1","name":"Room A"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	room, err := c.RoomGet(context.Background(), "r1")
	if err != nil {
		t.Fatal(err)
	}
	if room["name"] != "Room A" {
		t.Errorf("expected Room A, got %v", room["name"])
	}
}

func TestRoomPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"r1","name":"Updated"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	room, err := c.RoomPatch(context.Background(), "r1", map[string]any{"name": "Updated"})
	if err != nil {
		t.Fatal(err)
	}
	if room["name"] != "Updated" {
		t.Errorf("expected Updated, got %v", room["name"])
	}
}

func TestRoomDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.RoomDelete(context.Background(), "r1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRoomGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.RoomGet(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}
