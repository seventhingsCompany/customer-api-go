//go:build integration

package seventhings_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"testing"
	"time"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func upgradeCleanup(t *testing.T, remove func(context.Context) error) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), 30*time.Second)
		defer cancel()
		if err := remove(ctx); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})
}

func TestIntegrationCatalogItemsCRUD(t *testing.T) {
	c := integrationClient(t)
	ctx := t.Context()
	name := "sdk-catalog-" + uniqueSuffix()
	fields := map[string]any{"catalog_object_inventory_name": name}
	uuid, err := c.CatalogItemCreate(ctx, fields)
	if err != nil {
		t.Fatalf("CatalogItemCreate: %v", err)
	}
	deleted := false
	upgradeCleanup(t, func(ctx context.Context) error {
		if deleted {
			return nil
		}
		return c.CatalogItemDelete(ctx, uuid)
	})
	item, err := c.CatalogItemGet(ctx, uuid)
	if err != nil {
		t.Fatal(err)
	}
	if item.UUID != uuid || item.Fields["catalog_object_inventory_name"] != name {
		t.Fatalf("created item = %+v", item)
	}
	item.Fields["catalog_object_inventory_name"] = name + "-updated"
	if err := c.CatalogItemUpdate(ctx, uuid, item.Fields); err != nil {
		t.Fatalf("CatalogItemUpdate: %v", err)
	}
	item, err = c.CatalogItemGet(ctx, uuid)
	if err != nil {
		t.Fatal(err)
	}
	if item.Fields["catalog_object_inventory_name"] != name+"-updated" {
		t.Fatalf("updated item = %+v", item)
	}
	page, err := c.CatalogItemsList(ctx, &models.ListOptions{Page: 1, PerPage: 10, Filters: []models.FilterEntry{{Field: "catalog_object_inventory_name", Operator: models.FilterEq, Values: []string{name + "-updated"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].UUID != uuid {
		t.Fatalf("filtered list = %+v", page)
	}
	if err := c.CatalogItemDelete(ctx, uuid); err != nil {
		t.Fatal(err)
	}
	deleted = true
	_, err = c.CatalogItemGet(ctx, uuid)
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Fatalf("get after delete = %v", err)
	}
}

func TestIntegrationTimedRentalCRUD(t *testing.T) {
	c := integrationClient(t)
	ctx := t.Context()
	objectUUID, err := c.ObjectCreate(ctx, map[string]any{"inventory_name": "sdk-timed-rental-" + uniqueSuffix(), "barcode": "SDK-TIMED-" + uniqueSuffix()})
	if err != nil {
		t.Fatal(err)
	}
	upgradeCleanup(t, func(ctx context.Context) error { return c.ObjectDelete(ctx, objectUUID) })
	users, err := c.UsersList(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(users.Items) == 0 {
		t.Fatal("no responsible user available")
	}
	fullDay := false
	issue, due := "2099-01-01T10:00:00+02:00", "2099-01-01T12:00:00+02:00"
	refs := []models.RentalCaseReferenceInput{{Type: models.RentalCaseReferenceTypeAsset, UUID: objectUUID}}
	uuid, err := c.RentalCaseCreate(ctx, models.CreateRentalCase{
		Title: "sdk-timed-rental-" + uniqueSuffix(), References: refs,
		Renter:    &models.RentalCaseRenter{Type: models.RenterTypePlain, Value: "SDK integration test"},
		IsFullDay: &fullDay, IssueAt: &issue, DueAt: &due,
		IssueDate: "2099-01-01", DueDate: "2099-01-01",
		ResponsibleUserUUID: users.Items[0].UUID, Attachments: []string{},
	})
	if err != nil {
		t.Fatalf("create timed rental: %v", err)
	}
	deleted := false
	upgradeCleanup(t, func(ctx context.Context) error {
		if deleted {
			return nil
		}
		return c.RentalCaseDelete(ctx, uuid)
	})
	verify := func(wantIssue, wantDue string, wantFullDay bool) {
		t.Helper()
		rental, err := c.RentalCaseGet(ctx, uuid)
		if err != nil {
			t.Fatal(err)
		}
		if rental.IsFullDay != wantFullDay {
			t.Fatalf("IsFullDay = %v", rental.IsFullDay)
		}
		if wantFullDay {
			if rental.IssueAt != nil || rental.DueAt != nil {
				t.Fatalf("full-day timestamps: %+v", rental)
			}
			return
		}
		for _, pair := range []struct {
			got  *string
			want string
		}{{rental.IssueAt, wantIssue}, {rental.DueAt, wantDue}} {
			if pair.got == nil {
				t.Fatal("missing rental timestamp")
			}
			gotTime, err := time.Parse(time.RFC3339, *pair.got)
			if err != nil {
				t.Fatal(err)
			}
			wantTime, _ := time.Parse(time.RFC3339, pair.want)
			if !gotTime.Equal(wantTime) {
				t.Fatalf("timestamp = %s, want %s", *pair.got, pair.want)
			}
		}
	}
	verify(issue, due, false)
	newDue := "2099-01-01T13:00:00+02:00"
	update := models.UpdateRentalCase{Title: "updated timed rental", References: refs, IsFullDay: &fullDay, IssueAt: &issue, DueAt: &newDue, IssueDate: "2099-01-01", DueDate: "2099-01-01", ResponsibleUserUUID: users.Items[0].UUID, Attachments: []string{}}
	update.Renter = &models.RentalCaseRenter{Type: models.RenterTypePlain, Value: "SDK integration test"}
	if err := c.RentalCaseUpdate(ctx, uuid, update); err != nil {
		t.Fatal(err)
	}
	verify(issue, newDue, false)
	fullDay = true
	update.IssueAt, update.DueAt = nil, nil
	update.DueDate = "2099-01-02"
	if err := c.RentalCaseUpdate(ctx, uuid, update); err != nil {
		t.Fatal(err)
	}
	verify("", "", true)
	if err := c.RentalCaseDelete(ctx, uuid); err != nil {
		t.Fatal(err)
	}
	deleted = true
}

func TestIntegrationThumbnailCustomSize(t *testing.T) {
	c := integrationClient(t)
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 80, G: 140, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	uuid, err := c.FileUpload(t.Context(), "sdk-thumbnail-"+uniqueSuffix()+".png", &buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("uploaded thumbnail fixture %s (API has no file-delete endpoint)", uuid)
	data, err := c.FileGetThumbnailWithSize(t.Context(), uuid, 320, 240)
	if err != nil {
		t.Fatal(err)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if config.Width != 320 || config.Height != 240 {
		t.Fatalf("thumbnail size = %dx%d", config.Width, config.Height)
	}
	cached, err := c.FileGetThumbnailWithSize(t.Context(), uuid, 320, 240)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, cached) {
		t.Error("cached thumbnail differs")
	}
}

func TestIntegrationTaskOptionalReminders(t *testing.T) {
	c := integrationClient(t)
	ctx := t.Context()
	users, err := c.UsersList(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(users.Items) == 0 {
		t.Fatal("no assignee available")
	}
	objectUUID, err := c.ObjectCreate(ctx, map[string]any{"inventory_name": "sdk-reminders-" + uniqueSuffix(), "barcode": "SDK-REMINDERS-" + uniqueSuffix()})
	if err != nil {
		t.Fatal(err)
	}
	upgradeCleanup(t, func(ctx context.Context) error { return c.ObjectDelete(ctx, objectUUID) })
	deadline := "2099-12-31"
	notify := false
	assignees := []string{users.Items[0].UUID}
	refs := []models.TaskReferenceInput{{Type: models.TaskReferenceTypeAsset, UUID: objectUUID}}
	uuid, err := c.TaskCreate(ctx, models.CreateTask{Title: "sdk-optional-reminders", Deadline: &deadline, Assignees: assignees, References: refs, Notify: &notify})
	if err != nil {
		t.Fatalf("create with omitted reminders: %v", err)
	}
	upgradeCleanup(t, func(ctx context.Context) error { return c.TaskDelete(ctx, uuid) })
	for _, reminders := range [][]models.TimeInterval{{{Unit: models.TimeIntervalDays, Value: 0}}, {}, nil} {
		if err := c.TaskUpdate(ctx, uuid, models.UpdateTask{Title: "sdk-optional-reminders", Deadline: &deadline, Assignees: assignees, References: refs, Reminders: reminders, Notify: &notify}); err != nil {
			t.Fatalf("update reminders %v: %v", reminders, err)
		}
		task, err := c.TaskGet(ctx, uuid)
		if err != nil {
			t.Fatal(err)
		}
		if reminders != nil && len(task.Reminders) != len(reminders) {
			t.Fatalf("reminders = %+v, want %+v", task.Reminders, reminders)
		}
		if len(reminders) == 1 && task.Reminders[0] != reminders[0] {
			t.Fatalf("zero-interval reminder changed: %+v", task.Reminders)
		}
	}
}

// Read-only: never edits existing catalog records on the configured instance.
func TestIntegrationCatalogItemsRead(t *testing.T) {
	c := integrationClient(t)
	page, err := c.CatalogItemsList(t.Context(), &models.ListOptions{Page: 1, PerPage: 1})
	if err != nil {
		var apiErr *models.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 403 {
			t.Skip("product catalog feature or read permission unavailable")
		}
		t.Fatal(err)
	}
	if page.Page != 1 || page.PerPage != 1 {
		t.Fatalf("unexpected pagination: %+v", page)
	}
	if len(page.Items) == 0 {
		t.Log("catalog is empty; detail read not exercised")
		return
	}
	item := page.Items[0]
	if item.UUID == "" || item.Fields == nil {
		t.Fatalf("invalid catalog item: %+v", item)
	}
	detail, err := c.CatalogItemGet(t.Context(), item.UUID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.UUID != item.UUID || detail.Fields == nil {
		t.Fatalf("invalid catalog detail: %+v", detail)
	}
}
