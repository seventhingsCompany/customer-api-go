package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestCatalogItemsCRUD(t *testing.T) {
	fields := models.Fields{"catalog_object_inventory_name": "Laptop"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /customer-api/v1/catalog-items":
			if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("sort[id]") != "DESC" {
				t.Errorf("query = %s", r.URL.RawQuery)
			}
			fmt.Fprint(w, `{"items":[{"uuid":"c-1","fields":{"catalog_object_inventory_name":"Laptop"}}],"page":2,"per_page":10,"total":11,"sort":{"id":"DESC"}}`)
		case "GET /customer-api/v1/catalog-item/c-1":
			fmt.Fprint(w, `{"uuid":"c-1","fields":{"catalog_object_inventory_name":"Laptop"}}`)
		case "POST /customer-api/v1/catalog-item", "PUT /customer-api/v1/catalog-item/c-1":
			var body struct {
				Fields models.Fields `json:"fields"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !reflect.DeepEqual(body.Fields, fields) {
				t.Errorf("body = %+v, error = %v", body, err)
			}
			if r.Method == http.MethodPost {
				w.Header().Set("Location-UUID", "c-1")
				w.WriteHeader(http.StatusCreated)
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
		case "DELETE /customer-api/v1/catalog-item/c-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	c := newTestClient(t, server)
	list, err := c.CatalogItemsList(t.Context(), &models.ListOptions{Page: 2, PerPage: 10, Sort: map[string]models.SortDirection{"id": models.SortDESC}})
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 11 || list.Page != 2 || len(list.Items) != 1 || !reflect.DeepEqual(list.Items[0].Fields, fields) {
		t.Fatalf("list = %+v", list)
	}
	item, err := c.CatalogItemGet(t.Context(), "c-1")
	if err != nil {
		t.Fatal(err)
	}
	if item.UUID != "c-1" || !reflect.DeepEqual(item.Fields, fields) {
		t.Fatalf("item = %+v", item)
	}
	uuid, err := c.CatalogItemCreate(t.Context(), fields)
	if err != nil || uuid != "c-1" {
		t.Fatalf("create = %q, %v", uuid, err)
	}
	if err := c.CatalogItemUpdate(t.Context(), uuid, fields); err != nil {
		t.Fatal(err)
	}
	if err := c.CatalogItemDelete(t.Context(), uuid); err != nil {
		t.Fatal(err)
	}
}

func TestLoginAuthorizationCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		body := make(map[string]string)
		for key := range r.PostForm {
			body[key] = r.PostForm.Get(key)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" || r.Header.Get("Authorization") != "" {
			t.Errorf("headers = %v", r.Header)
		}
		want := map[string]string{"grant_type": "authorization_code", "code": "code", "code_verifier": "verifier", "redirect_uri": "https://example.test/callback", "client_id": "client"}
		if r.Method != http.MethodPost || r.URL.Path != "/customer-api/v1/auth_token" || !reflect.DeepEqual(body, want) {
			t.Errorf("request = %s %s %+v", r.Method, r.URL, body)
		}
		fmt.Fprint(w, `{"access_token":"access","refresh_token":"refresh"}`)
	}))
	defer server.Close()
	c := newTestClient(t, server)
	tok, err := c.LoginAuthorizationCode(t.Context(), "code", "verifier", "https://example.test/callback", "client")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "access" || c.token != "access" || c.clientID != "client" {
		t.Fatalf("token = %+v", tok)
	}
}

func TestFileGetThumbnailWithSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/file/f-1/thumbnail" || r.URL.Query().Get("width") != "320" || r.URL.Query().Get("height") != "240" {
			t.Errorf("URL = %s", r.URL)
		}
		fmt.Fprint(w, "image")
	}))
	defer server.Close()
	c := newTestClient(t, server)
	data, err := c.FileGetThumbnailWithSize(t.Context(), "f-1", 320, 240)
	if err != nil || string(data) != "image" {
		t.Fatalf("thumbnail = %q, %v", data, err)
	}
	for _, size := range [][2]int{{0, 100}, {100, 0}, {-1, 100}, {2001, 100}, {100, 2001}} {
		if _, err := c.FileGetThumbnailWithSize(t.Context(), "f-1", size[0], size[1]); err == nil {
			t.Errorf("expected error for %v", size)
		}
	}
}

func TestUpgradeAPIErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"not allowed"}`)
	}))
	defer server.Close()
	c := newTestClient(t, server)
	for name, call := range map[string]func() error{
		"catalog list":   func() error { _, err := c.CatalogItemsList(t.Context(), nil); return err },
		"catalog get":    func() error { _, err := c.CatalogItemGet(t.Context(), "c-1"); return err },
		"catalog create": func() error { _, err := c.CatalogItemCreate(t.Context(), nil); return err },
		"catalog update": func() error { return c.CatalogItemUpdate(t.Context(), "c-1", nil) },
		"catalog delete": func() error { return c.CatalogItemDelete(t.Context(), "c-1") },
		"PKCE": func() error {
			_, err := c.LoginAuthorizationCode(t.Context(), "code", "verifier", "uri", "client")
			return err
		},
		"thumbnail": func() error { _, err := c.FileGetThumbnailWithSize(t.Context(), "f-1", 100, 100); return err },
	} {
		t.Run(name, func(t *testing.T) {
			var apiErr *models.APIError
			if err := call(); !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestCatalogItemCreateLocationFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/customer-api/v1/catalog-item/c-2")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	c := newTestClient(t, server)
	if uuid, err := c.CatalogItemCreate(t.Context(), map[string]any{}); err != nil || uuid != "c-2" {
		t.Fatalf("create = %q, %v", uuid, err)
	}
	if _, err := c.CatalogItemCreate(t.Context(), map[string]any{"invalid": make(chan int)}); err == nil {
		t.Error("expected JSON encoding error")
	}
	if err := c.CatalogItemUpdate(t.Context(), "c-2", map[string]any{"invalid": make(chan int)}); err == nil {
		t.Error("expected JSON encoding error")
	}
}
