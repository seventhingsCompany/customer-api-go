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

func TestCircularityHubSuggestCategory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/suggest-category" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"category":"Electronics"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	result, err := c.CircularityHubSuggestCategory(context.Background(), models.FilterObject{})
	if err != nil {
		t.Fatal(err)
	}
	if result["category"] != "Electronics" {
		t.Errorf("expected Electronics, got %q", result["category"])
	}
}

func TestCircularityHubSuggestCategoryBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fo models.FilterObject
		if err := json.Unmarshal(body, &fo); err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		if fo.Sort == nil || fo.Sort["name"] != models.SortASC {
			t.Errorf("unexpected sort: %v", fo.Sort)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubSuggestCategory(context.Background(), models.FilterObject{
		Sort: map[string]models.SortDirection{"name": models.SortASC},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCircularityHubSuggestRestPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/suggest-rest-price" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"price":"12.50"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	result, err := c.CircularityHubSuggestRestPrice(context.Background(), map[string]string{"category": "Laptop"})
	if err != nil {
		t.Fatal(err)
	}
	if result["price"] != "12.50" {
		t.Errorf("expected 12.50, got %q", result["price"])
	}
}

func TestCircularityHubAddObjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/add-objects-to-circularity-hub" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.CircularityHubAddObjects(context.Background(), map[string]models.AddObjectEntry{
		"obj-1": {Category: "Chair", Price: "50.00"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCircularityHubAddObjectsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var entries map[string]models.AddObjectEntry
		if err := json.Unmarshal(body, &entries); err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		if e, ok := entries["obj-1"]; !ok || e.Category != "Chair" || e.Price != "50.00" {
			t.Errorf("unexpected body: %v", entries)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.CircularityHubAddObjects(context.Background(), map[string]models.AddObjectEntry{
		"obj-1": {Category: "Chair", Price: "50.00"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCircularityHubItemsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/items" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"id":1,"name":"Laptop"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	items, err := c.CircularityHubItemsList(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestCircularityHubItemsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "page=2&per_page=10" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubItemsList(context.Background(), &models.ListOptions{Page: 2, PerPage: 10})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCircularityHubItemGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/item/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":42,"name":"Monitor"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	item, err := c.CircularityHubItemGet(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if item["name"] != "Monitor" {
		t.Errorf("expected Monitor, got %v", item["name"])
	}
}

func TestCircularityHubSuggestCategoryError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubSuggestCategory(context.Background(), models.FilterObject{})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestCircularityHubSuggestCategoryEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	result, err := c.CircularityHubSuggestCategory(context.Background(), models.FilterObject{})
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Errorf("expected nil for empty array, got %v", result)
	}
}

func TestCircularityHubSuggestRestPriceError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubSuggestRestPrice(context.Background(), map[string]string{"category": "Laptop"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestCircularityHubAddObjectsError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.CircularityHubAddObjects(context.Background(), map[string]models.AddObjectEntry{
		"obj-1": {Category: "Chair", Price: "50.00"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestCircularityHubItemsListError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubItemsList(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestCircularityHubItemUpdateError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.CircularityHubItemUpdate(context.Background(), 42, map[string]any{"price": "99.00"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestCircularityHubItemGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubItemGet(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}

func TestCircularityHubItemUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/item/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.CircularityHubItemUpdate(context.Background(), 42, map[string]any{"price": "99.00"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCircularityHubItemDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/item/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.CircularityHubItemDelete(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCircularityHubOrdersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/orders" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"id":1,"order_number":"ORD-001","created_at":"2024-01-01","completed":false,"cancelled":false,"articles":[]}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	orders, err := c.CircularityHubOrdersList(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	if orders[0].OrderNumber != "ORD-001" {
		t.Errorf("expected ORD-001, got %q", orders[0].OrderNumber)
	}
}

func TestCircularityHubOrdersListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "page=1&per_page=5" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubOrdersList(context.Background(), &models.ListOptions{Page: 1, PerPage: 5})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCircularityHubOrderCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/orders" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Location-Id", "77")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	id, err := c.CircularityHubOrderCreate(context.Background(), []int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if id != 77 {
		t.Errorf("expected 77, got %d", id)
	}
}

func TestCircularityHubOrderCreateBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var ids []int
		if err := json.Unmarshal(body, &ids); err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		if len(ids) != 2 || ids[0] != 10 || ids[1] != 20 {
			t.Errorf("unexpected body: %v", ids)
		}
		w.Header().Set("Location-Id", "1")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubOrderCreate(context.Background(), []int{10, 20})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCircularityHubOrderGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/order/77" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":77,"order_number":"ORD-077","created_at":"2024-01-01","completed":true,"cancelled":false,"articles":[]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	order, err := c.CircularityHubOrderGet(context.Background(), 77)
	if err != nil {
		t.Fatal(err)
	}
	if order.ID != 77 {
		t.Errorf("expected ID 77, got %d", order.ID)
	}
	if !order.Completed {
		t.Error("expected completed to be true")
	}
}

func TestCircularityHubOrdersListError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubOrdersList(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestCircularityHubOrderCreateError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubOrderCreate(context.Background(), []int{1, 2})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestCircularityHubOrderUpdateError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.CircularityHubOrderUpdate(context.Background(), 77, map[string]any{"completed": true})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestCircularityHubOrderGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.CircularityHubOrderGet(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}

func TestCircularityHubOrderUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/circularity-hub/order/77" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.CircularityHubOrderUpdate(context.Background(), 77, map[string]any{"completed": true})
	if err != nil {
		t.Fatal(err)
	}
}
