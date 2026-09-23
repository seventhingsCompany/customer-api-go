package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestReportTemplatesList(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want []models.ReportTemplate
	}{
		{"templates", `[{"uuid":"template-uuid","name":"Inventory list"}]`, []models.ReportTemplate{{UUID: "template-uuid", Name: "Inventory list"}}},
		{"empty", `[]`, []models.ReportTemplate{}},
		{"malformed", `{`, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/customer-api/v1/report-template" || r.Header.Get("Accept") != "application/json" {
					t.Errorf("unexpected request: %s %s %v", r.Method, r.URL, r.Header)
				}
				_, _ = fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			got, err := newTestClient(t, server).ReportTemplatesList(t.Context())
			if tc.name == "malformed" {
				if err == nil {
					t.Fatal("expected decoding error")
				}
				return
			}
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, %v; want %+v", got, err, tc.want)
			}
		})
	}
}

func TestReportCreate(t *testing.T) {
	pdf := []byte("%PDF-1.7\n\x00\xff\n%%EOF")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/customer-api/v1/report" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Accept") != "application/pdf" || r.Header.Get("Authorization") != "Bearer report-token" {
			t.Errorf("unexpected headers: %v", r.Header)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		want := map[string]any{"report_template_uuid": "template-uuid", "object_uuids": []any{"object-2", "object-1"}}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("body = %#v, want %#v", body, want)
		}
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(pdf)
	}))
	defer server.Close()
	c := newTestClient(t, server)
	c.SetToken("report-token")
	got, err := c.ReportCreate(t.Context(), models.CreateReport{ReportTemplateUUID: "template-uuid", ObjectUUIDs: []string{"object-2", "object-1"}})
	if err != nil || !bytes.Equal(got, pdf) {
		t.Fatalf("PDF = %q, error = %v", got, err)
	}
}

func TestReportsAPIErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprint(w, `{"message":"forbidden"}`)
	}))
	defer server.Close()
	c := newTestClient(t, server)
	_, listErr := c.ReportTemplatesList(t.Context())
	_, createErr := c.ReportCreate(t.Context(), models.CreateReport{})
	for _, err := range []error{listErr, createErr} {
		var apiErr *models.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden {
			t.Errorf("expected APIError 403, got %v", err)
		}
	}
}
