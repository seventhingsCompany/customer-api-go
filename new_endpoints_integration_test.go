//go:build integration

package seventhings_test

import (
	"bytes"
	"context"
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

func verifyLiveHistory[T any](t *testing.T, uuid string, fetch func(context.Context, string, *models.HistoryListOptions) (*models.HistoryResponse[T], error)) {
	t.Helper()
	if uuid == "" {
		t.Fatal("empty entity UUID")
	}
	for _, page := range []int{1, 2} {
		got, err := fetch(t.Context(), uuid, &models.HistoryListOptions{Page: page, PerPage: 1})
		if err != nil {
			t.Fatalf("history page %d: %v", page, err)
		}
		if got.Page != page || got.PerPage != 1 || got.Total < 0 || len(got.Items) > 1 || got.Items == nil {
			t.Fatalf("invalid history pagination: page=%d per_page=%d total=%d items=%d", got.Page, got.PerPage, got.Total, len(got.Items))
		}
		t.Logf("page=%d per_page=%d total=%d decoded_entries=%d", got.Page, got.PerPage, got.Total, len(got.Items))
	}
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
		verifyLiveHistory(t, uuid, c.ObjectHistory)
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
		verifyLiveHistory(t, liveEntityUUID(t, items[0], "room_uuid"), c.RoomHistory)
	})
	t.Run("LocationHistory", func(t *testing.T) {
		items, err := c.LocationsList(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) == 0 {
			t.Skip("instance has no locations")
		}
		verifyLiveHistory(t, liveEntityUUID(t, items[0], "location_uuid"), c.LocationHistory)
	})
	t.Run("PersonHistory", func(t *testing.T) {
		perPage := 1
		items, err := c.PersonsList(ctx, &models.PersonListOptions{PerPage: &perPage})
		if err != nil {
			t.Fatal(err)
		}
		if len(items.Items) == 0 {
			t.Skip("instance has no persons")
		}
		uuid := items.Items[0].UUID
		if uuid == "" {
			uuid = liveEntityUUID(t, items.Items[0].Fields, "person_uuid")
			t.Log("person UUID recovered from raw fields; typed Person.UUID was empty")
		}
		verifyLiveHistory(t, uuid, c.PersonHistory)
	})
	t.Run("TaskHistory", func(t *testing.T) {
		items, err := c.TasksList(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) == 0 {
			t.Skip("instance has no tasks")
		}
		verifyLiveHistory(t, items[0].UUID, c.TaskHistory)
	})
	t.Run("RentalCaseHistory", func(t *testing.T) {
		items, err := c.RentalCasesList(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) == 0 {
			t.Skip("instance has no rental cases")
		}
		verifyLiveHistory(t, items[0].UUID, c.RentalCaseHistory)
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
