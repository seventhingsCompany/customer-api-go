package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"time"

	"github.com/SeventhingsCompany/customer-api-go/client"
	"github.com/SeventhingsCompany/customer-api-go/models"
)

// Cleanup uses a fresh deadline so canceled demo requests do not leak fixtures.
// Join cleanup failures into the returned error instead of claiming success.
func cleanupDemo(ctx context.Context, result *error, remove func(context.Context) error) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	if err := remove(cleanupCtx); err != nil {
		*result = errors.Join(*result, fmt.Errorf("demo cleanup: %w", err))
	}
}

func demoCatalog(ctx context.Context, c *client.Client) (result error) {
	section("Catalog", "Creating a temporary catalog item…")
	name := fmt.Sprintf("SDK Demo Catalog %d", time.Now().UnixNano())
	uuid, err := c.CatalogItemCreate(ctx, map[string]any{"catalog_object_inventory_name": name})
	if err != nil {
		if models.IsForbidden(err) {
			pf("Catalog", "Skipped: product catalog feature or create permission unavailable")
			return nil
		}
		return err
	}
	deleted := false
	defer func() {
		if !deleted {
			cleanupDemo(ctx, &result, func(ctx context.Context) error { return c.CatalogItemDelete(ctx, uuid) })
		}
	}()
	item, err := c.CatalogItemGet(ctx, uuid)
	if err != nil {
		return err
	}
	if item.UUID != uuid || item.Fields["catalog_object_inventory_name"] != name {
		return fmt.Errorf("catalog create/read mismatch")
	}
	// PUT replaces the entire state. Send the read field map back, not a patch.
	item.Fields["catalog_object_inventory_name"] = name + " updated"
	if err := c.CatalogItemUpdate(ctx, uuid, item.Fields); err != nil {
		return err
	}
	item, err = c.CatalogItemGet(ctx, uuid)
	if err != nil {
		return err
	}
	if item.Fields["catalog_object_inventory_name"] != name+" updated" {
		return fmt.Errorf("catalog update/read mismatch")
	}
	page, err := c.CatalogItemsList(ctx, models.NewListOptions().WithPage(1).WithPerPage(5).Where(models.Eq("catalog_object_inventory_name", name+" updated")))
	if err != nil {
		return err
	}
	if len(page.Items) != 1 || page.Items[0].UUID != uuid {
		return fmt.Errorf("catalog filtered list did not return the temporary item")
	}
	pf("Catalog", "Verified create, get, PUT and filtered list — total=%d", page.Total)
	if err := c.CatalogItemDelete(ctx, uuid); err != nil {
		return err
	}
	deleted = true
	if _, err := c.CatalogItemGet(ctx, uuid); !models.IsNotFound(err) {
		return fmt.Errorf("catalog get after delete: expected 404, got %v", err)
	}
	pf("Catalog", "Deleted temporary item; confirmed 404")
	return nil
}

func demoTimedRental(ctx context.Context, c *client.Client, userUUID string) (result error) {
	section("Rentals", "Creating and updating a timed rental…")
	objectUUID, err := c.ObjectCreate(ctx, map[string]any{"inventory_name": "SDK Demo Rental Target", "barcode": fmt.Sprintf("SDK-RENTAL-%d", time.Now().UnixNano())})
	if err != nil {
		return err
	}
	defer cleanupDemo(ctx, &result, func(ctx context.Context) error { return c.ObjectDelete(ctx, objectUUID) })
	// Use a future date and a non-UTC offset to demonstrate UTC normalization.
	start := time.Now().UTC().AddDate(1, 0, 0)
	start = time.Date(start.Year(), start.Month(), start.Day(), 10, 0, 0, 0, time.FixedZone("demo", 2*60*60))
	issue, due := start.Format(time.RFC3339), start.Add(2*time.Hour).Format(time.RFC3339)
	date := start.Format("2006-01-02")
	fullDay := false
	refs := []models.RentalCaseReferenceInput{{Type: models.RentalCaseReferenceTypeAsset, UUID: objectUUID}}
	renter := &models.RentalCaseRenter{Type: models.RenterTypePlain, Value: "SDK Demo"}
	// Supply a renter: the current server returns 500 when it is omitted.
	uuid, err := c.RentalCaseCreate(ctx, models.CreateRentalCase{
		Title: "SDK Demo Timed Rental", References: refs, Renter: renter,
		IsFullDay: &fullDay, IssueAt: &issue, DueAt: &due,
		IssueDate: date, DueDate: date, ResponsibleUserUUID: userUUID, Attachments: []string{},
	})
	if err != nil {
		return err
	}
	defer cleanupDemo(ctx, &result, func(ctx context.Context) error { return c.RentalCaseDelete(ctx, uuid) })
	rental, err := c.RentalCaseGet(ctx, uuid)
	if err != nil {
		return err
	}
	if err := verifyRentalTimes(rental, start, start.Add(2*time.Hour)); err != nil {
		return err
	}
	pf("Rentals", "Verified timed rental — issue_at=%s, due_at=%s (exclusive)", *rental.IssueAt, *rental.DueAt)
	due = start.Add(3 * time.Hour).Format(time.RFC3339)
	update := models.UpdateRentalCase{
		Title: "SDK Demo Timed Rental updated", References: refs, Renter: renter,
		IsFullDay: &fullDay, IssueAt: &issue, DueAt: &due,
		IssueDate: date, DueDate: date, ResponsibleUserUUID: userUUID, Attachments: []string{},
	}
	if err := c.RentalCaseUpdate(ctx, uuid, update); err != nil {
		return err
	}
	rental, err = c.RentalCaseGet(ctx, uuid)
	if err != nil {
		return err
	}
	if err := verifyRentalTimes(rental, start, start.Add(3*time.Hour)); err != nil {
		return err
	}
	fullDay = true
	update.IssueAt, update.DueAt = nil, nil
	update.DueDate = start.AddDate(0, 0, 1).Format("2006-01-02")
	if err := c.RentalCaseUpdate(ctx, uuid, update); err != nil {
		return err
	}
	rental, err = c.RentalCaseGet(ctx, uuid)
	if err != nil {
		return err
	}
	if !rental.IsFullDay || rental.IssueAt != nil || rental.DueAt != nil {
		return fmt.Errorf("full-day conversion did not clear rental times")
	}
	pf("Rentals", "Verified extended return time and conversion to full-day; cleaning up rental and object")
	return nil
}

func verifyRentalTimes(rental *models.RentalCase, issue, due time.Time) error {
	if rental.IsFullDay || rental.IssueAt == nil || rental.DueAt == nil {
		return fmt.Errorf("expected a timed rental with both timestamps")
	}
	for _, pair := range []struct {
		raw  string
		want time.Time
	}{{*rental.IssueAt, issue}, {*rental.DueAt, due}} {
		got, err := time.Parse(time.RFC3339, pair.raw)
		if err != nil {
			return err
		}
		if !got.Equal(pair.want) {
			return fmt.Errorf("rental timestamp %s does not match %s", pair.raw, pair.want)
		}
	}
	return nil
}

func demoThumbnail(ctx context.Context, c *client.Client) error {
	section("Files", "Uploading an image and verifying a custom thumbnail…")
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 80, G: 140, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	uuid, err := c.FileUpload(ctx, fmt.Sprintf("sdk-demo-thumbnail-%d.png", time.Now().UnixNano()), &buf)
	if err != nil {
		return err
	}
	pf("Files", "Uploaded image %s; file remains because the API has no delete endpoint", uuid)
	data, err := c.FileGetThumbnailWithSize(ctx, uuid, 320, 240)
	if err != nil {
		return err
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return err
	}
	if config.Width != 320 || config.Height != 240 {
		return fmt.Errorf("thumbnail dimensions = %dx%d, want 320x240", config.Width, config.Height)
	}
	cached, err := c.FileGetThumbnailWithSize(ctx, uuid, 320, 240)
	if err != nil {
		return err
	}
	if !bytes.Equal(data, cached) {
		return fmt.Errorf("cached thumbnail differs from first response")
	}
	pf("Files", "Verified 320x240 thumbnail and cached response (%d bytes)", len(data))
	return nil
}
