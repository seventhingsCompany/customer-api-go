package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRoomAndLocationResponseFormats(t *testing.T) {
	for _, format := range []string{"flat", "wrapped"} {
		for _, resource := range []string{"room", "location"} {
			for _, operation := range []string{"list", "get", "patch"} {
				t.Run(resource+"/"+operation+"/"+format, func(t *testing.T) {
					payload := `{"uuid":"resource-1","id":42,"name":"Office","cost_center":"CC-1","picture":[{"uuid":"file-1"}]} `
					if format == "wrapped" {
						payload = `{"uuid":"resource-1","fields":{"id":42,"name":"Office","cost_center":"CC-1","picture":[{"uuid":"file-1"}]}}`
					}
					if operation == "list" {
						payload = `{"items":[` + payload + `]}`
					}
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = fmt.Fprint(w, payload)
					}))
					defer server.Close()
					c := newTestClient(t, server)
					var get func(context.Context, string) (map[string]any, error)
					if resource == "room" {
						get = c.RoomGet
					} else {
						get = c.LocationGet
					}
					var got map[string]any
					var err error
					if operation == "list" {
						var items []map[string]any
						if resource == "room" {
							items, err = c.RoomsList(t.Context(), nil)
						} else {
							items, err = c.LocationsList(t.Context(), nil)
						}
						if err == nil && len(items) == 1 {
							got = items[0]
						}
					} else if operation == "patch" {
						if resource == "room" {
							got, err = c.RoomPatch(t.Context(), "resource-1", map[string]any{"name": "Office"})
						} else {
							got, err = c.LocationPatch(t.Context(), "resource-1", map[string]any{"name": "Office"})
						}
					} else {
						got, err = get(t.Context(), "resource-1")
					}
					if err != nil {
						t.Fatal(err)
					}
					want := map[string]any{"uuid": "resource-1", "id": float64(42), "name": "Office", "cost_center": "CC-1", "picture": []any{map[string]any{"uuid": "file-1"}}}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("got %v, want %v", got, want)
					}
				})
			}
		}
	}
}
