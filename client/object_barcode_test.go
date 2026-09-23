package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestObjectGetByBarcode(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"found", 200, `{"uuid":"object-uuid","archived":true,"name":"Desk"}`},
		{"missing", 404, `{"message":"not found"}`},
		{"malformed", 200, `{`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.EscapedPath() != "/customer-api/v1/object/by-barcode/INV%2F100%20%3F%23%25" || r.URL.RawQuery != "" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				w.WriteHeader(tc.status)
				_, _ = fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			got, err := newTestClient(t, server).ObjectGetByBarcode(t.Context(), "INV/100 ?#%")
			switch tc.name {
			case "missing":
				var apiErr *models.APIError
				if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
					t.Fatalf("expected APIError 404, got %v", err)
				}
			case "malformed":
				if err == nil {
					t.Fatal("expected decoding error")
				}
			default:
				if err != nil || got["uuid"] != "object-uuid" || got["archived"] != true || got["name"] != "Desk" {
					t.Fatalf("got %v, %v", got, err)
				}
			}
		})
	}
}
