//go:build integration

package seventhings_test

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func liveEntityUUID(t *testing.T, fields map[string]any, key string) string {
	t.Helper()
	for _, k := range []string{key, "uuid"} {
		if uuid, ok := fields[k].(string); ok && uuid != "" {
			return uuid
		}
	}
	t.Fatalf("list response has no %s or uuid", key)
	return ""
}

func liveEntityUUIDs(t *testing.T, items []map[string]any, key string) []string {
	t.Helper()
	uuids := make([]string, 0, len(items))
	for _, item := range items {
		uuids = append(uuids, liveEntityUUID(t, item, key))
	}
	return uuids
}

func verifyLiveHistory[T any](t *testing.T, fetch func(context.Context, string, *models.HistoryListOptions) (*models.HistoryResponse[T], error), uuids ...string) {
	t.Helper()
	for _, uuid := range uuids {
		if uuid == "" {
			t.Fatal("empty entity UUID")
		}
		first, err := fetch(t.Context(), uuid, &models.HistoryListOptions{Page: 1, PerPage: 1})
		if err != nil {
			t.Fatalf("history page 1: %v", err)
		}
		if first.Page != 1 || first.PerPage != 1 || first.Total < 0 || len(first.Items) != min(first.Total, 1) || first.Items == nil {
			t.Fatalf("invalid first page: page=%d per_page=%d total=%d items=%d", first.Page, first.PerPage, first.Total, len(first.Items))
		}
		if first.Total < 2 {
			// The spec does not prescribe out-of-range behavior. The live API
			// clamps to the last page, so look for a valid second page instead.
			continue
		}
		second, err := fetch(t.Context(), uuid, &models.HistoryListOptions{Page: 2, PerPage: 1})
		if err != nil {
			t.Fatalf("history page 2: %v", err)
		}
		if second.Page != 2 || second.PerPage != 1 || second.Total < 2 || len(second.Items) != 1 {
			t.Fatalf("invalid second page: page=%d per_page=%d total=%d items=%d", second.Page, second.PerPage, second.Total, len(second.Items))
		}
		// Compare with a two-entry page to verify the offset and ordering.
		combined, err := fetch(t.Context(), uuid, &models.HistoryListOptions{Page: 1, PerPage: 2})
		if err != nil {
			t.Fatal(err)
		}
		want := append(append([]T{}, first.Items...), second.Items...)
		if combined.Page != 1 || combined.PerPage != 2 || !reflect.DeepEqual(combined.Items, want) {
			t.Fatal("separate history pages do not match a two-entry page (or history changed during the test)")
		}
		t.Logf("verified two populated history pages and ordering; total=%d", first.Total)
		return
	}
	t.Logf("validated first pages for %d resources; none had multiple history entries", len(uuids))
}

// TestIntegrationNewEndpoints exercises the new spec endpoints using existing
// instance data. Report generation returns a PDF without storing a document.
func TestIntegrationNewEndpoints(t *testing.T) {
	c := integrationClient(t)
	ctx := t.Context()
	opts := &models.ListOptions{Page: 1, PerPage: 10}

	t.Run("ObjectHistory", func(t *testing.T) {
		items, err := c.ObjectsList(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) == 0 {
			t.Skip("instance has no objects")
		}
		uuid := liveEntityUUID(t, items[0], "asset_uuid")
		if _, err := c.ObjectGet(ctx, uuid); err != nil {
			t.Fatalf("existing ObjectGet endpoint: %v", err)
		}
		t.Log("existing ObjectGet endpoint succeeded for the same UUID")
		verifyLiveHistory(t, c.ObjectHistory, liveEntityUUIDs(t, items, "asset_uuid")...)
	})
	t.Run("ObjectGetByBarcode", func(t *testing.T) {
		items, err := c.ObjectsList(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range items {
			barcode, ok := item["barcode"].(string)
			if !ok || barcode == "" {
				continue
			}
			got, err := c.ObjectGetByBarcode(ctx, barcode)
			if err != nil {
				t.Fatal(err)
			}
			if liveEntityUUID(t, got, "asset_uuid") != liveEntityUUID(t, item, "asset_uuid") {
				t.Fatal("barcode lookup returned a different object")
			}
			t.Log("barcode lookup returned the expected object")
			return
		}
		t.Skip("sample contains no objects with a barcode")
	})
	t.Run("RoomHistory", func(t *testing.T) {
		items, err := c.RoomsList(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) == 0 {
			t.Skip("instance has no rooms")
		}
		verifyLiveHistory(t, c.RoomHistory, liveEntityUUIDs(t, items, "room_uuid")...)
	})
	t.Run("LocationHistory", func(t *testing.T) {
		items, err := c.LocationsList(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) == 0 {
			t.Skip("instance has no locations")
		}
		verifyLiveHistory(t, c.LocationHistory, liveEntityUUIDs(t, items, "location_uuid")...)
	})
	t.Run("PersonHistory", func(t *testing.T) {
		perPage := 10
		items, err := c.PersonsList(ctx, &models.PersonListOptions{PerPage: &perPage})
		if err != nil {
			t.Fatal(err)
		}
		if len(items.Items) == 0 {
			t.Skip("instance has no persons")
		}
		uuids := make([]string, 0, len(items.Items))
		for _, person := range items.Items {
			uuids = append(uuids, person.UUID)
		}
		verifyLiveHistory(t, c.PersonHistory, uuids...)
	})
	t.Run("TaskHistory", func(t *testing.T) {
		items, err := c.TasksList(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) == 0 {
			t.Skip("instance has no tasks")
		}
		uuids := make([]string, 0, min(len(items), 10))
		for _, task := range items[:min(len(items), 10)] {
			uuids = append(uuids, task.UUID)
		}
		verifyLiveHistory(t, c.TaskHistory, uuids...)
	})
	t.Run("RentalCaseHistory", func(t *testing.T) {
		items, err := c.RentalCasesList(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) == 0 {
			t.Skip("instance has no rental cases")
		}
		uuids := make([]string, 0, len(items))
		for _, rental := range items {
			uuids = append(uuids, rental.UUID)
		}
		verifyLiveHistory(t, c.RentalCaseHistory, uuids...)
	})
	t.Run("Reports", func(t *testing.T) {
		templates, err := c.ReportTemplatesList(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("decoded %d report templates", len(templates))
		t.Run("ReportCreate", func(t *testing.T) {
			if len(templates) == 0 {
				t.Skip("instance has no report templates")
			}
			items, err := c.ObjectsList(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}
			if len(items) == 0 {
				t.Skip("instance has no objects")
			}
			pdf, err := c.ReportCreate(ctx, models.CreateReport{
				ReportTemplateUUID: templates[0].UUID,
				ObjectUUIDs:        []string{liveEntityUUID(t, items[0], "asset_uuid")},
			})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
				t.Fatal("response is not a PDF")
			}
			t.Logf("generated PDF: %d bytes", len(pdf))
		})
	})
}
