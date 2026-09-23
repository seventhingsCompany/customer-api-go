package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func testHistory[T any](t *testing.T, path, entry string, want T, fetch func(*Client, context.Context, string, *models.HistoryListOptions) (*models.HistoryResponse[T], error)) {
	t.Helper()
	for _, tc := range []struct {
		name   string
		status int
		body   string
		opts   *models.HistoryListOptions
		query  string
	}{
		{"page", 200, `{"items":[` + entry + `],"page":2,"per_page":10,"total":11}`, &models.HistoryListOptions{Page: 2, PerPage: 10}, "page=2&per_page=10"},
		{"empty", 200, `{"items":[],"page":1,"per_page":50,"total":0}`, nil, ""},
		{"malformed", 200, `{`, nil, ""},
		{"forbidden", 403, `{"message":"forbidden"}`, nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/customer-api/v1/"+path {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				if r.URL.RawQuery != tc.query {
					t.Errorf("query = %q, want %q", r.URL.RawQuery, tc.query)
				}
				if r.Header.Get("Authorization") != "Bearer history-token" || r.Header.Get("Accept") != "application/json" {
					t.Errorf("unexpected headers: %v", r.Header)
				}
				w.WriteHeader(tc.status)
				_, _ = fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			c := newTestClient(t, server)
			c.SetToken("history-token")
			got, err := fetch(c, t.Context(), "entity-uuid", tc.opts)
			switch tc.name {
			case "forbidden":
				var apiErr *models.APIError
				if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.status {
					t.Fatalf("expected APIError %d, got %v", tc.status, err)
				}
			case "malformed":
				if err == nil {
					t.Fatal("expected JSON decoding error")
				}
			case "empty":
				if err != nil {
					t.Fatal(err)
				}
				if got.Items == nil || len(got.Items) != 0 || got.Page != 1 || got.PerPage != 50 || got.Total != 0 {
					t.Fatalf("unexpected empty page: %+v", got)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
				if got.Page != 2 || got.PerPage != 10 || got.Total != 11 || !reflect.DeepEqual(got.Items, []T{want}) {
					t.Fatalf("unexpected history page: %+v", got)
				}
			}
		})
	}
}

func TestObjectHistory(t *testing.T) {
	t.Run("asset", func(t *testing.T) {
		testHistory(t, "object/entity-uuid/history", `{"type":"asset","date":"2026-09-15T12:00:00+00:00","properties":{"name":"Desk"}}`, models.ObjectHistoryEntry{
			"type": "asset", "date": "2026-09-15T12:00:00+00:00", "properties": map[string]any{"name": "Desk"},
		}, (*Client).ObjectHistory)
	})
	t.Run("merge", func(t *testing.T) {
		testHistory(t, "object/entity-uuid/history", `{"type":"object_merge","user_id":42,"absorbedObjectData":{"name":"Desk"}}`, models.ObjectHistoryEntry{
			"type": "object_merge", "user_id": float64(42), "absorbedObjectData": map[string]any{"name": "Desk"},
		}, (*Client).ObjectHistory)
	})
}

const historyFields = `"user_uuid":"user-uuid","occurred_at":"2026-09-15T12:00:00+00:00","event_name":"Modified","description":"Changed name","details":"{\"name\":\"Desk\"}"`

func TestRoomHistory(t *testing.T) {
	testHistory(t, "room/entity-uuid/history", `{"room_uuid":"entity-uuid",`+historyFields+`}`, models.RoomHistoryEntry{
		RoomUUID: "entity-uuid", UserUUID: "user-uuid", OccurredAt: "2026-09-15T12:00:00+00:00", EventName: "Modified", Description: "Changed name", Details: `{"name":"Desk"}`,
	}, (*Client).RoomHistory)
}

func TestLocationHistory(t *testing.T) {
	testHistory(t, "location/entity-uuid/history", `{"location_uuid":"entity-uuid",`+historyFields+`}`, models.LocationHistoryEntry{
		LocationUUID: "entity-uuid", UserUUID: "user-uuid", OccurredAt: "2026-09-15T12:00:00+00:00", EventName: "Modified", Description: "Changed name", Details: `{"name":"Desk"}`,
	}, (*Client).LocationHistory)
}

func TestPersonHistory(t *testing.T) {
	testHistory(t, "person/entity-uuid/history", `{"person_uuid":"entity-uuid",`+historyFields+`}`, models.PersonHistoryEntry{
		PersonUUID: "entity-uuid", UserUUID: "user-uuid", OccurredAt: "2026-09-15T12:00:00+00:00", EventName: "Modified", Description: "Changed name", Details: `{"name":"Desk"}`,
	}, (*Client).PersonHistory)
}

func TestTaskHistory(t *testing.T) {
	testHistory(t, "task-management/task/entity-uuid/history", `{"task_uuid":"entity-uuid",`+historyFields+`}`, models.TaskHistoryEntry{
		TaskUUID: "entity-uuid", UserUUID: "user-uuid", OccurredAt: "2026-09-15T12:00:00+00:00", EventName: "Modified", Description: "Changed name", Details: `{"name":"Desk"}`,
	}, (*Client).TaskHistory)
}

func TestRentalCaseHistory(t *testing.T) {
	testHistory(t, "rental-management/rental-case/entity-uuid/history", `{"rental_case_uuid":"entity-uuid",`+historyFields+`}`, models.RentalCaseHistoryEntry{
		RentalCaseUUID: "entity-uuid", UserUUID: "user-uuid", OccurredAt: "2026-09-15T12:00:00+00:00", EventName: "Modified", Description: "Changed name", Details: `{"name":"Desk"}`,
	}, (*Client).RentalCaseHistory)
}
