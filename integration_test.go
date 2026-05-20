//go:build integration

package seventhings_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/SeventhingsCompany/customer-api-go/client"
	"github.com/SeventhingsCompany/customer-api-go/models"
)

func integrationClient(t *testing.T) *client.Client {
	t.Helper()

	baseURL := os.Getenv("SEVENTHINGS_BASE_URL")
	username := os.Getenv("SEVENTHINGS_USERNAME")
	password := os.Getenv("SEVENTHINGS_PASSWORD")
	clientID := os.Getenv("SEVENTHINGS_CLIENT_ID")

	if baseURL == "" || username == "" || password == "" || clientID == "" {
		t.Skip("set SEVENTHINGS_BASE_URL, SEVENTHINGS_USERNAME, SEVENTHINGS_PASSWORD, SEVENTHINGS_CLIENT_ID to run integration tests")
	}

	c, err := client.NewWithCredentials(context.Background(), baseURL, username, password, clientID)
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}
	return c
}

// integrationLogin performs a fresh login and returns both the client and the token response.
func integrationLogin(t *testing.T) (*client.Client, *models.TokenResponse) {
	t.Helper()

	baseURL := os.Getenv("SEVENTHINGS_BASE_URL")
	username := os.Getenv("SEVENTHINGS_USERNAME")
	password := os.Getenv("SEVENTHINGS_PASSWORD")
	clientID := os.Getenv("SEVENTHINGS_CLIENT_ID")

	if baseURL == "" || username == "" || password == "" || clientID == "" {
		t.Skip("set SEVENTHINGS_BASE_URL, SEVENTHINGS_USERNAME, SEVENTHINGS_PASSWORD, SEVENTHINGS_CLIENT_ID to run integration tests")
	}

	c := client.New(baseURL)
	tok, err := c.Login(context.Background(), username, password, clientID)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	return c, tok
}

func uniqueSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

// ---------------------------------------------------------------------------
// Ping
// ---------------------------------------------------------------------------

func TestIntegrationPing(t *testing.T) {
	baseURL := os.Getenv("SEVENTHINGS_BASE_URL")
	if baseURL == "" {
		t.Skip("set SEVENTHINGS_BASE_URL to run integration tests")
	}

	c := client.New(baseURL)
	ping, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	if ping.Status != "OK" {
		t.Errorf("expected status OK, got %q", ping.Status)
	}
}

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

func TestIntegrationAuth(t *testing.T) {
	c := integrationClient(t)

	if c.Token() == "" {
		t.Fatal("expected token to be set after login")
	}
}

func TestIntegrationAuthRefreshAndRevoke(t *testing.T) {
	c, tok := integrationLogin(t)
	ctx := context.Background()

	if tok.RefreshToken == "" {
		t.Fatal("expected refresh_token from login")
	}

	// Refresh
	newTok, err := c.Refresh(ctx, tok.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if newTok.AccessToken == "" {
		t.Fatal("expected new access_token from refresh")
	}
	if c.Token() != newTok.AccessToken {
		t.Error("expected client token to be updated after refresh")
	}

	// Revoke
	if err := c.RevokeTokens(ctx); err != nil {
		t.Fatalf("RevokeTokens: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Objects — basic CRUD (existing)
// ---------------------------------------------------------------------------

func TestIntegrationObjects(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Create
	uuid, err := c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "integration-test-object",
		"barcode":        "INT-TEST-" + uniqueSuffix(),
	})
	if err != nil {
		t.Fatalf("ObjectCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.ObjectDelete(ctx, uuid)
	})

	// Get
	obj, err := c.ObjectGet(ctx, uuid)
	if err != nil {
		t.Fatalf("ObjectGet: %v", err)
	}
	if obj["inventory_name"] != "integration-test-object" {
		t.Errorf("expected inventory_name integration-test-object, got %v", obj["inventory_name"])
	}

	// List
	objects, err := c.ObjectsList(ctx, &models.ListOptions{PerPage: 1})
	if err != nil {
		t.Fatalf("ObjectsList: %v", err)
	}
	if len(objects) == 0 {
		t.Error("expected at least 1 object")
	}

	// Update
	if err := c.ObjectPatch(ctx, uuid, map[string]any{"inventory_name": "updated-object"}); err != nil {
		t.Fatalf("ObjectPatch: %v", err)
	}
	updated, err := c.ObjectGet(ctx, uuid)
	if err != nil {
		t.Fatalf("ObjectGet after patch: %v", err)
	}
	if updated["inventory_name"] != "updated-object" {
		t.Errorf("expected updated inventory_name, got %v", updated["inventory_name"])
	}

	// Delete
	if err := c.ObjectDelete(ctx, uuid); err != nil {
		t.Fatalf("ObjectDelete: %v", err)
	}

	// Verify 404
	_, err = c.ObjectGet(ctx, uuid)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ---------------------------------------------------------------------------
// Objects — filters, sort, count
// ---------------------------------------------------------------------------

func TestIntegrationObjectsListWithFilters(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	name := "filter-test-" + uniqueSuffix()
	uuid, err := c.ObjectCreate(ctx, map[string]any{
		"inventory_name": name,
		"barcode":        "INT-FILTER-" + uniqueSuffix(),
	})
	if err != nil {
		t.Fatalf("ObjectCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.ObjectDelete(ctx, uuid)
	})

	// List with filter on inventory_name
	objects, err := c.ObjectsList(ctx, &models.ListOptions{
		PerPage: 10,
		Filters: []models.FilterEntry{
			{Field: "inventory_name", Operator: models.FilterEq, Values: []string{name}},
		},
		Sort: map[string]models.SortDirection{
			"inventory_name": models.SortASC,
		},
	})
	if err != nil {
		t.Fatalf("ObjectsList with filter: %v", err)
	}
	if len(objects) == 0 {
		t.Error("expected at least 1 object matching filter")
	}
	for _, obj := range objects {
		if obj["inventory_name"] != name {
			t.Errorf("expected inventory_name %q, got %v", name, obj["inventory_name"])
		}
	}
}

func TestIntegrationObjectsCount(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Count all
	count, err := c.ObjectsCount(ctx, nil)
	if err != nil {
		t.Fatalf("ObjectsCount (all): %v", err)
	}
	if count < 0 {
		t.Errorf("expected non-negative count, got %d", count)
	}

	// Count with filter (may be 0 on fresh instance)
	_, err = c.ObjectsCount(ctx, &models.ListOptions{
		Filters: []models.FilterEntry{
			{Field: "inventory_name", Operator: models.FilterEq, Values: []string{"nonexistent-object-xyz"}},
		},
	})
	if err != nil {
		t.Fatalf("ObjectsCount (filtered): %v", err)
	}
}

// ---------------------------------------------------------------------------
// Objects — archive / unarchive
// ---------------------------------------------------------------------------

func TestIntegrationObjectArchiveUnarchive(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	uuid, err := c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "archive-test-" + uniqueSuffix(),
		"barcode":        "INT-ARCH-" + uniqueSuffix(),
	})
	if err != nil {
		t.Fatalf("ObjectCreate: %v", err)
	}
	t.Cleanup(func() {
		// Unarchive first in case it's still archived, then delete.
		_ = c.ObjectUnarchive(ctx, uuid)
		_ = c.ObjectDelete(ctx, uuid)
	})

	// Archive
	if err := c.ObjectArchive(ctx, uuid); err != nil {
		t.Fatalf("ObjectArchive: %v", err)
	}

	// Unarchive
	if err := c.ObjectUnarchive(ctx, uuid); err != nil {
		t.Fatalf("ObjectUnarchive: %v", err)
	}

	// Should still be accessible after unarchive
	obj, err := c.ObjectGet(ctx, uuid)
	if err != nil {
		t.Fatalf("ObjectGet after unarchive: %v", err)
	}
	if obj == nil {
		t.Error("expected object to be accessible after unarchive")
	}
}

// ---------------------------------------------------------------------------
// Objects — add/remove files
// ---------------------------------------------------------------------------

func TestIntegrationObjectFiles(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Create an object
	objUUID, err := c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "file-test-" + uniqueSuffix(),
		"barcode":        "INT-FILE-" + uniqueSuffix(),
	})
	if err != nil {
		t.Fatalf("ObjectCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.ObjectDelete(ctx, objUUID)
	})

	// Upload a small test file
	fileContent := []byte("integration test file content")
	fileUUID, err := c.FileUpload(ctx, "test.txt", bytes.NewReader(fileContent))
	if err != nil {
		t.Fatalf("FileUpload: %v", err)
	}

	// We need to know a valid attachment field key. Use a generic one.
	// Look up the asset field definitions to find an ATTACHMENT field.
	defs, err := c.FieldDefinitionsList(ctx, models.AssetTrackingTemplateAsset)
	if err != nil {
		t.Fatalf("FieldDefinitionsList: %v", err)
	}

	var attachmentFieldKey string
	for _, d := range defs {
		if d.FieldType.Name == models.FieldTypeAttachment {
			attachmentFieldKey = d.FieldKey
			break
		}
	}
	if attachmentFieldKey == "" {
		t.Skip("no ATTACHMENT field definition found on asset template; skipping file attach test")
	}

	attachment := models.FileAttachment{
		FieldKey: attachmentFieldKey,
		FileUUID: fileUUID,
	}

	// Add file to object
	_, err = c.ObjectAddFiles(ctx, objUUID, []models.FileAttachment{attachment})
	if err != nil {
		t.Fatalf("ObjectAddFiles: %v", err)
	}

	// Remove file from object
	_, err = c.ObjectRemoveFiles(ctx, objUUID, []models.FileAttachment{attachment})
	if err != nil {
		t.Fatalf("ObjectRemoveFiles: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Locations — full CRUD
// ---------------------------------------------------------------------------

func TestIntegrationLocationsCRUD(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	name := "int-loc-" + uniqueSuffix()

	// Create
	uuid, err := c.LocationCreate(ctx, map[string]any{
		"name": name,
	})
	if err != nil {
		t.Fatalf("LocationCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.LocationDelete(ctx, uuid)
	})

	// Get
	loc, err := c.LocationGet(ctx, uuid)
	if err != nil {
		t.Fatalf("LocationGet: %v", err)
	}
	if loc["name"] != name {
		t.Errorf("expected name %q, got %v", name, loc["name"])
	}

	// Patch
	updatedName := "int-loc-updated-" + uniqueSuffix()
	_, err = c.LocationPatch(ctx, uuid, map[string]any{"name": updatedName})
	if err != nil {
		t.Fatalf("LocationPatch: %v", err)
	}
	// Verify via GET since PATCH may return an empty body.
	locAfterPatch, err := c.LocationGet(ctx, uuid)
	if err != nil {
		t.Fatalf("LocationGet after patch: %v", err)
	}
	if locAfterPatch["name"] != updatedName {
		t.Errorf("expected patched name %q, got %v", updatedName, locAfterPatch["name"])
	}

	// List
	locations, err := c.LocationsList(ctx, &models.ListOptions{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("LocationsList: %v", err)
	}
	if len(locations) == 0 {
		t.Error("expected at least 1 location")
	}

	// Count
	count, err := c.LocationsCount(ctx, nil)
	if err != nil {
		t.Fatalf("LocationsCount: %v", err)
	}
	if count < 1 {
		t.Errorf("expected at least 1 location in count, got %d", count)
	}

	// Delete
	if err := c.LocationDelete(ctx, uuid); err != nil {
		t.Fatalf("LocationDelete: %v", err)
	}

	// Verify 404
	_, err = c.LocationGet(ctx, uuid)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ---------------------------------------------------------------------------
// Rooms — full CRUD
// ---------------------------------------------------------------------------

func TestIntegrationRoomsCRUD(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// A room requires a linked location (building_id). Find one first.
	locations, err := c.LocationsList(ctx, &models.ListOptions{Page: 1, PerPage: 1})
	if err != nil {
		t.Fatalf("LocationsList: %v", err)
	}
	if len(locations) == 0 {
		t.Skip("no locations available; skipping room CRUD (rooms require a building_id)")
	}

	// building_id is the numeric location ID.
	buildingID := locations[0]["id"]

	// Discover required fields from room field definitions. Some instances
	// have custom mandatory fields (e.g. a room-type dropdown).
	roomFields := map[string]any{
		"name":       "int-room-" + uniqueSuffix(),
		"number":     "IR-" + uniqueSuffix(),
		"building_id": buildingID,
	}
	roomDefs, err := c.FieldDefinitionsList(ctx, models.AssetTrackingTemplateRoom)
	if err == nil {
		for _, d := range roomDefs {
			isMandatory := false
			for _, attr := range d.Attributes {
				if attr.Type == "mandatory" && attr.Value == "yes" {
					isMandatory = true
					break
				}
			}
			if !isMandatory {
				continue
			}
			// Skip fields we already set or system-managed fields.
			if _, exists := roomFields[d.FieldKey]; exists {
				continue
			}
			// For DROPDOWN fields, use the first allowed value.
			if d.FieldType.Name == models.FieldTypeDropdown {
				for _, c := range d.FieldType.Constraints {
					if c.Type == "allowed_values" {
						if vals, ok := c.Value.([]any); ok && len(vals) > 0 {
							roomFields[d.FieldKey] = vals[0]
						}
					}
				}
			}
		}
	}

	name := roomFields["name"].(string)

	// Create
	uuid, err := c.RoomCreate(ctx, roomFields)
	if err != nil {
		t.Fatalf("RoomCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.RoomDelete(ctx, uuid)
	})

	// Get
	room, err := c.RoomGet(ctx, uuid)
	if err != nil {
		t.Fatalf("RoomGet: %v", err)
	}
	if room["name"] != name {
		t.Errorf("expected name %q, got %v", name, room["name"])
	}

	// Patch
	updatedName := "int-room-updated-" + uniqueSuffix()
	_, err = c.RoomPatch(ctx, uuid, map[string]any{"name": updatedName})
	if err != nil {
		t.Fatalf("RoomPatch: %v", err)
	}
	// Verify via GET since PATCH may return an empty body.
	roomAfterPatch, err := c.RoomGet(ctx, uuid)
	if err != nil {
		t.Fatalf("RoomGet after patch: %v", err)
	}
	if roomAfterPatch["name"] != updatedName {
		t.Errorf("expected patched name %q, got %v", updatedName, roomAfterPatch["name"])
	}

	// List
	rooms, err := c.RoomsList(ctx, &models.ListOptions{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("RoomsList: %v", err)
	}
	if len(rooms) == 0 {
		t.Error("expected at least 1 room")
	}

	// Count
	count, err := c.RoomsCount(ctx, nil)
	if err != nil {
		t.Fatalf("RoomsCount: %v", err)
	}
	if count < 1 {
		t.Errorf("expected at least 1 room in count, got %d", count)
	}

	// Delete
	if err := c.RoomDelete(ctx, uuid); err != nil {
		t.Fatalf("RoomDelete: %v", err)
	}

	// Verify 404
	_, err = c.RoomGet(ctx, uuid)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ---------------------------------------------------------------------------
// Tasks — basic CRUD (existing)
// ---------------------------------------------------------------------------

func TestIntegrationTasks(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Look up current user for assignee
	userResp, err := c.UsersList(ctx, nil)
	if err != nil {
		t.Fatalf("UsersList: %v", err)
	}
	if len(userResp.Items) == 0 {
		t.Fatal("expected at least 1 user")
	}
	userUUID := userResp.Items[0].UUID

	// Create a temporary object as task reference
	refUUID, err := c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "integration-task-ref",
		"barcode":        "INT-TASKREF-" + uniqueSuffix(),
	})
	if err != nil {
		t.Fatalf("ObjectCreate for task ref: %v", err)
	}
	t.Cleanup(func() {
		_ = c.ObjectDelete(ctx, refUUID)
	})

	// Create task
	deadline := "2099-12-31"
	uuid, err := c.TaskCreate(ctx, models.CreateTask{
		Title:    "integration-test-task",
		Deadline: &deadline,
		Assignees: []string{userUUID},
		References: []models.TaskReferenceInput{
			{Type: models.TaskReferenceTypeAsset, UUID: refUUID},
		},
		Reminders: []models.TimeInterval{{Unit: models.TimeIntervalDays, Value: 1}},
	})
	if err != nil {
		t.Fatalf("TaskCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.TaskDelete(ctx, uuid)
	})

	// Get
	task, err := c.TaskGet(ctx, uuid)
	if err != nil {
		t.Fatalf("TaskGet: %v", err)
	}
	if task.Title != "integration-test-task" {
		t.Errorf("expected title integration-test-task, got %q", task.Title)
	}

	// List
	tasks, err := c.TasksList(ctx, nil)
	if err != nil {
		t.Fatalf("TasksList: %v", err)
	}
	if len(tasks) == 0 {
		t.Error("expected at least 1 task")
	}

	// Delete
	if err := c.TaskDelete(ctx, uuid); err != nil {
		t.Fatalf("TaskDelete: %v", err)
	}

	// Verify 404
	_, err = c.TaskGet(ctx, uuid)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ---------------------------------------------------------------------------
// Tasks — update
// ---------------------------------------------------------------------------

func TestIntegrationTaskUpdate(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	userResp, err := c.UsersList(ctx, nil)
	if err != nil {
		t.Fatalf("UsersList: %v", err)
	}
	if len(userResp.Items) == 0 {
		t.Fatal("expected at least 1 user")
	}
	userUUID := userResp.Items[0].UUID

	refUUID, err := c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "task-update-ref",
		"barcode":        "INT-TUPD-" + uniqueSuffix(),
	})
	if err != nil {
		t.Fatalf("ObjectCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.ObjectDelete(ctx, refUUID)
	})

	deadline := "2099-12-31"
	uuid, err := c.TaskCreate(ctx, models.CreateTask{
		Title:     "task-update-test",
		Deadline:  &deadline,
		Assignees: []string{userUUID},
		References: []models.TaskReferenceInput{
			{Type: models.TaskReferenceTypeAsset, UUID: refUUID},
		},
		Reminders: []models.TimeInterval{{Unit: models.TimeIntervalDays, Value: 1}},
	})
	if err != nil {
		t.Fatalf("TaskCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.TaskDelete(ctx, uuid)
	})

	// Update title and comment
	comment := "updated comment"
	err = c.TaskUpdate(ctx, uuid, models.UpdateTask{
		Title:     "task-update-test-renamed",
		Deadline:  &deadline,
		Assignees: []string{userUUID},
		References: []models.TaskReferenceInput{
			{Type: models.TaskReferenceTypeAsset, UUID: refUUID},
		},
		Reminders: []models.TimeInterval{{Unit: models.TimeIntervalDays, Value: 1}},
		Comment:   &comment,
	})
	if err != nil {
		t.Fatalf("TaskUpdate: %v", err)
	}

	updated, err := c.TaskGet(ctx, uuid)
	if err != nil {
		t.Fatalf("TaskGet after update: %v", err)
	}
	if updated.Title != "task-update-test-renamed" {
		t.Errorf("expected updated title, got %q", updated.Title)
	}
	if updated.Comment == nil || *updated.Comment != comment {
		t.Errorf("expected comment %q, got %v", comment, updated.Comment)
	}
}

// ---------------------------------------------------------------------------
// Tasks — status transition
// ---------------------------------------------------------------------------

func TestIntegrationTaskStatusTransition(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	userResp, err := c.UsersList(ctx, nil)
	if err != nil {
		t.Fatalf("UsersList: %v", err)
	}
	if len(userResp.Items) == 0 {
		t.Fatal("expected at least 1 user")
	}
	userUUID := userResp.Items[0].UUID

	refUUID, err := c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "task-status-ref",
		"barcode":        "INT-TSTAT-" + uniqueSuffix(),
	})
	if err != nil {
		t.Fatalf("ObjectCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.ObjectDelete(ctx, refUUID)
	})

	deadline := "2099-12-31"
	uuid, err := c.TaskCreate(ctx, models.CreateTask{
		Title:     "task-status-test",
		Deadline:  &deadline,
		Assignees: []string{userUUID},
		References: []models.TaskReferenceInput{
			{Type: models.TaskReferenceTypeAsset, UUID: refUUID},
		},
		Reminders: []models.TimeInterval{{Unit: models.TimeIntervalDays, Value: 1}},
	})
	if err != nil {
		t.Fatalf("TaskCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.TaskDelete(ctx, uuid)
	})

	// Close the task
	if err := c.TaskUpdateStatus(ctx, uuid, models.TaskStatusClosed); err != nil {
		t.Fatalf("TaskUpdateStatus (closed): %v", err)
	}

	task, err := c.TaskGet(ctx, uuid)
	if err != nil {
		t.Fatalf("TaskGet after close: %v", err)
	}
	if task.Status != models.TaskStatusClosed {
		t.Errorf("expected status %q, got %q", models.TaskStatusClosed, task.Status)
	}

	// Re-open the task
	if err := c.TaskUpdateStatus(ctx, uuid, models.TaskStatusOpen); err != nil {
		t.Fatalf("TaskUpdateStatus (open): %v", err)
	}

	task, err = c.TaskGet(ctx, uuid)
	if err != nil {
		t.Fatalf("TaskGet after re-open: %v", err)
	}
	if task.Status != models.TaskStatusOpen {
		t.Errorf("expected status %q, got %q", models.TaskStatusOpen, task.Status)
	}
}

// ---------------------------------------------------------------------------
// Tasks — list with filters
// ---------------------------------------------------------------------------

func TestIntegrationTasksListWithFilters(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	statusOpen := models.TaskStatusOpen
	refType := models.TaskReferenceTypeAsset
	deadlineFrom := "2000-01-01"
	deadlineTo := "2199-12-31"

	tasks, err := c.TasksList(ctx, &models.TaskListOptions{
		Status:        &statusOpen,
		DeadlineFrom:  &deadlineFrom,
		DeadlineTo:    &deadlineTo,
		ReferenceType: &refType,
	})
	if err != nil {
		t.Fatalf("TasksList with filters: %v", err)
	}
	// Result may be empty but the call should succeed.
	_ = tasks
}

// ---------------------------------------------------------------------------
// Rentals — full CRUD
// ---------------------------------------------------------------------------

func TestIntegrationRentalsCRUD(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Create a temporary object for reference
	refUUID, err := c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "rental-ref-" + uniqueSuffix(),
		"barcode":        "INT-RENT-" + uniqueSuffix(),
	})
	if err != nil {
		t.Fatalf("ObjectCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.ObjectDelete(ctx, refUUID)
	})

	// Get a user UUID for responsible_user_uuid
	perPage := 1
	usersResp, err := c.UsersList(ctx, &models.UserListOptions{PerPage: &perPage})
	if err != nil {
		t.Fatalf("UsersList: %v", err)
	}
	if len(usersResp.Items) == 0 {
		t.Fatal("no users found for responsible_user_uuid")
	}
	userUUID := usersResp.Items[0].UUID

	// Create
	uuid, err := c.RentalCaseCreate(ctx, models.CreateRentalCase{
		Title: "Integration Test Rental",
		Renter: &models.RentalCaseRenter{
			Type:  models.RenterTypePlain,
			Value: "Integration Tester",
		},
		References: []models.RentalCaseReferenceInput{
			{Type: models.RentalCaseReferenceTypeAsset, UUID: refUUID},
		},
		IssueDate:           "2099-01-01",
		DueDate:             "2099-06-01",
		Comment:             "integration test rental",
		ResponsibleUserUUID: userUUID,
		Attachments:         []string{},
	})
	if err != nil {
		t.Fatalf("RentalCaseCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.RentalCaseDelete(ctx, uuid)
	})

	// Get
	rental, err := c.RentalCaseGet(ctx, uuid)
	if err != nil {
		t.Fatalf("RentalCaseGet: %v", err)
	}
	if rental.UUID != uuid {
		t.Errorf("expected UUID %q, got %q", uuid, rental.UUID)
	}

	// Update
	err = c.RentalCaseUpdate(ctx, uuid, models.UpdateRentalCase{
		Title: "Updated Integration Test Rental",
		Renter: &models.RentalCaseRenter{
			Type:  models.RenterTypePlain,
			Value: "Updated Tester",
		},
		References: []models.RentalCaseReferenceInput{
			{Type: models.RentalCaseReferenceTypeAsset, UUID: refUUID},
		},
		IssueDate:           "2099-01-01",
		DueDate:             "2099-06-01",
		Comment:             "updated rental comment",
		ResponsibleUserUUID: userUUID,
		Attachments:         []string{},
	})
	if err != nil {
		t.Fatalf("RentalCaseUpdate: %v", err)
	}

	// List with pagination
	cases, err := c.RentalCasesList(ctx, &models.ListOptions{PerPage: 10})
	if err != nil {
		t.Fatalf("RentalCasesList: %v", err)
	}
	if len(cases) == 0 {
		t.Error("expected at least 1 rental case")
	}

	// Delete
	if err := c.RentalCaseDelete(ctx, uuid); err != nil {
		t.Fatalf("RentalCaseDelete: %v", err)
	}

	// Verify 404
	_, err = c.RentalCaseGet(ctx, uuid)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ---------------------------------------------------------------------------
// Users — get by UUID, get by ID, list with options
// ---------------------------------------------------------------------------

func TestIntegrationUserGetAndOptions(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// List with options
	perPage := 5
	page := 1
	sortBy := models.UserSortByEmail
	order := models.UserSortOrderAsc

	resp, err := c.UsersList(ctx, &models.UserListOptions{
		Page:    &page,
		PerPage: &perPage,
		SortBy:  &sortBy,
		Order:   &order,
	})
	if err != nil {
		t.Fatalf("UsersList with options: %v", err)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected at least 1 user")
	}

	user := resp.Items[0]

	// Get by UUID
	byUUID, err := c.UserGet(ctx, user.UUID)
	if err != nil {
		t.Fatalf("UserGet: %v", err)
	}
	if byUUID.UUID != user.UUID {
		t.Errorf("expected UUID %q, got %q", user.UUID, byUUID.UUID)
	}

	// Get by ID
	byID, err := c.UserGetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("UserGetByID: %v", err)
	}
	if byID.ID != user.ID {
		t.Errorf("expected ID %d, got %d", user.ID, byID.ID)
	}
	if byID.UUID != user.UUID {
		t.Errorf("expected UUID %q from GetByID, got %q", user.UUID, byID.UUID)
	}
}

// ---------------------------------------------------------------------------
// Files — upload, get metadata, download data/thumbnail
// ---------------------------------------------------------------------------

func TestIntegrationFileUploadAndDownload(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	content := []byte("hello from integration test")
	fileUUID, err := c.FileUpload(ctx, "integration-test.txt", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("FileUpload: %v", err)
	}

	// Get metadata
	meta, err := c.FileGet(ctx, fileUUID)
	if err != nil {
		t.Fatalf("FileGet: %v", err)
	}
	if meta.UUID != fileUUID {
		t.Errorf("expected UUID %q, got %q", fileUUID, meta.UUID)
	}
	if meta.Name != "integration-test.txt" {
		t.Errorf("expected file name integration-test.txt, got %q", meta.Name)
	}

	// Download data
	data, err := c.FileGetData(ctx, fileUUID)
	if err != nil {
		t.Fatalf("FileGetData: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty file data")
	}

	// Download thumbnail (may fail for non-image files, so we just check the call doesn't panic)
	_, _ = c.FileGetThumbnail(ctx, fileUUID)
}

// ---------------------------------------------------------------------------
// Files — basic list (existing)
// ---------------------------------------------------------------------------

func TestIntegrationFiles(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	files, err := c.FilesList(ctx)
	if err != nil {
		t.Fatalf("FilesList: %v", err)
	}
	_ = files // may be empty
}

// ---------------------------------------------------------------------------
// Field Definitions — CRUD for asset and room templates
// ---------------------------------------------------------------------------

func TestIntegrationFieldDefinitionsCRUD(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	templates := []models.AssetTrackingTemplate{
		models.AssetTrackingTemplateAsset,
		models.AssetTrackingTemplateRoom,
	}

	for _, tmpl := range templates {
		t.Run(string(tmpl), func(t *testing.T) {
			label := "int-test-field-" + uniqueSuffix()
			comment := ""

			// Create — include all fields to satisfy strict schema validation.
			uuid, err := c.FieldDefinitionCreate(ctx, tmpl, models.CreateFieldDefinition{
				FieldType: models.FieldDefinitionFieldType{
					Name: models.FieldTypeText,
					Constraints: []models.FieldValueConstraint{
						{Type: "max_length", Value: 255},
					},
				},
				Label:          label,
				Attributes:     []models.FieldAttribute{},
				Relations:      []models.FieldRelation{},
				Comment:        &comment,
				DefaultValue:   nil,
				PossibleValues: []any{},
			})
			if err != nil {
				t.Fatalf("FieldDefinitionCreate: %v", err)
			}

			// Get
			def, err := c.FieldDefinitionGet(ctx, tmpl, uuid)
			if err != nil {
				t.Fatalf("FieldDefinitionGet: %v", err)
			}
			if def.Label != label {
				t.Errorf("expected label %q, got %q", label, def.Label)
			}
			if def.FieldType.Name != models.FieldTypeText {
				t.Errorf("expected field type TEXT, got %q", def.FieldType.Name)
			}

			// Update — PUT requires uuid and field_key in the body.
			updatedLabel := "int-test-field-updated-" + uniqueSuffix()
			err = c.FieldDefinitionUpdate(ctx, tmpl, uuid, models.UpdateFieldDefinition{
				UUID:     def.UUID,
				FieldKey: def.FieldKey,
				FieldType: models.FieldDefinitionFieldType{
					Name: models.FieldTypeText,
					Constraints: []models.FieldValueConstraint{
						{Type: "max_length", Value: 255},
					},
				},
				Label:          updatedLabel,
				Attributes:     def.Attributes,
				Relations:      []models.FieldRelation{},
				Comment:        &comment,
				DefaultValue:   nil,
				PossibleValues: []any{},
			})
			if err != nil {
				t.Fatalf("FieldDefinitionUpdate: %v", err)
			}

			updatedDef, err := c.FieldDefinitionGet(ctx, tmpl, uuid)
			if err != nil {
				t.Fatalf("FieldDefinitionGet after update: %v", err)
			}
			if updatedDef.Label != updatedLabel {
				t.Errorf("expected updated label %q, got %q", updatedLabel, updatedDef.Label)
			}

			// List
			defs, err := c.FieldDefinitionsList(ctx, tmpl)
			if err != nil {
				t.Fatalf("FieldDefinitionsList: %v", err)
			}
			if len(defs) == 0 {
				t.Error("expected at least 1 field definition")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Field Definitions — basic list (existing)
// ---------------------------------------------------------------------------

func TestIntegrationFieldDefinitions(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	defs, err := c.FieldDefinitionsList(ctx, models.AssetTrackingTemplateAsset)
	if err != nil {
		t.Fatalf("FieldDefinitionsList: %v", err)
	}
	_ = defs
}

// ---------------------------------------------------------------------------
// Rentals — basic list (existing)
// ---------------------------------------------------------------------------

func TestIntegrationRentals(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	cases, err := c.RentalCasesList(ctx, nil)
	if err != nil {
		t.Fatalf("RentalCasesList: %v", err)
	}
	_ = cases
}

// ---------------------------------------------------------------------------
// Locations — basic list (existing)
// ---------------------------------------------------------------------------

func TestIntegrationLocations(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	locations, err := c.LocationsList(ctx, nil)
	if err != nil {
		t.Fatalf("LocationsList: %v", err)
	}
	_ = locations
}

// ---------------------------------------------------------------------------
// Rooms — basic list (existing)
// ---------------------------------------------------------------------------

func TestIntegrationRooms(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	rooms, err := c.RoomsList(ctx, nil)
	if err != nil {
		t.Fatalf("RoomsList: %v", err)
	}
	_ = rooms
}

// ---------------------------------------------------------------------------
// Users — basic list (existing)
// ---------------------------------------------------------------------------

func TestIntegrationUsers(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	resp, err := c.UsersList(ctx, nil)
	if err != nil {
		t.Fatalf("UsersList: %v", err)
	}
	if len(resp.Items) == 0 {
		t.Error("expected at least 1 user")
	}
}

// ---------------------------------------------------------------------------
// CircularityHub — existing basic test
// ---------------------------------------------------------------------------

func TestIntegrationCircularityHub(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// List items
	items, err := c.CircularityHubItemsList(ctx, &models.ListOptions{PerPage: 5})
	if err != nil {
		t.Fatalf("CircularityHubItemsList: %v", err)
	}
	_ = items

	// List orders
	orders, err := c.CircularityHubOrdersList(ctx, &models.ListOptions{PerPage: 5})
	if err != nil {
		t.Fatalf("CircularityHubOrdersList: %v", err)
	}
	_ = orders

	// Suggest category (may return empty result depending on data)
	_, err = c.CircularityHubSuggestCategory(ctx, models.FilterObject{})
	if err != nil {
		t.Fatalf("CircularityHubSuggestCategory: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CircularityHub — items CRUD
// ---------------------------------------------------------------------------

func TestIntegrationCircularityHubItemsCRUD(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// List items to find one we can work with
	items, err := c.CircularityHubItemsList(ctx, &models.ListOptions{PerPage: 1})
	if err != nil {
		t.Fatalf("CircularityHubItemsList: %v", err)
	}
	if len(items) == 0 {
		t.Skip("no CircularityHub items available; skipping CRUD test")
	}

	// Extract the item ID (JSON numbers decode as float64)
	idFloat, ok := items[0]["id"].(float64)
	if !ok {
		t.Fatalf("expected item id to be float64, got %T", items[0]["id"])
	}
	id := int(idFloat)

	// Get
	item, err := c.CircularityHubItemGet(ctx, id)
	if err != nil {
		t.Fatalf("CircularityHubItemGet: %v", err)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}

	// Update (set a field that's safe to toggle)
	if err := c.CircularityHubItemUpdate(ctx, id, map[string]any{}); err != nil {
		t.Fatalf("CircularityHubItemUpdate: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CircularityHub — orders CRUD
// ---------------------------------------------------------------------------

func TestIntegrationCircularityHubOrdersCRUD(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// We need at least one item to create an order
	items, err := c.CircularityHubItemsList(ctx, &models.ListOptions{PerPage: 1})
	if err != nil {
		t.Fatalf("CircularityHubItemsList: %v", err)
	}
	if len(items) == 0 {
		t.Skip("no CircularityHub items available; skipping orders CRUD test")
	}

	idFloat, ok := items[0]["id"].(float64)
	if !ok {
		t.Fatalf("expected item id to be float64, got %T", items[0]["id"])
	}
	itemID := int(idFloat)

	// Create order
	orderID, err := c.CircularityHubOrderCreate(ctx, []int{itemID})
	if err != nil {
		t.Fatalf("CircularityHubOrderCreate: %v", err)
	}

	// Get order
	order, err := c.CircularityHubOrderGet(ctx, orderID)
	if err != nil {
		t.Fatalf("CircularityHubOrderGet: %v", err)
	}
	if order.ID != orderID {
		t.Errorf("expected order ID %d, got %d", orderID, order.ID)
	}

	// Update order
	if err := c.CircularityHubOrderUpdate(ctx, orderID, map[string]any{}); err != nil {
		t.Fatalf("CircularityHubOrderUpdate: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CircularityHub — add objects
// ---------------------------------------------------------------------------

func TestIntegrationCircularityHubAddObjects(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Create a temporary object — purchasing_price is required for CHUB to
	// process the object without an "Array to string conversion" error.
	objUUID, err := c.ObjectCreate(ctx, map[string]any{
		"inventory_name":  "ch-add-obj-" + uniqueSuffix(),
		"barcode":         "INT-CHADD-" + uniqueSuffix(),
		"purchasing_price": 100.00,
	})
	if err != nil {
		t.Fatalf("ObjectCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.ObjectDelete(ctx, objUUID)
	})

	err = c.CircularityHubAddObjects(ctx, map[string]models.AddObjectEntry{
		objUUID: {
			Category: "category_furniture",
			Price:    "10.00",
		},
	})
	if err != nil {
		t.Fatalf("CircularityHubAddObjects: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CircularityHub — suggest rest price
// ---------------------------------------------------------------------------

func TestIntegrationCircularityHubSuggestRestPrice(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// The endpoint may return an empty result depending on data, but should not error.
	_, err := c.CircularityHubSuggestRestPrice(ctx, map[string]string{})
	if err != nil {
		t.Fatalf("CircularityHubSuggestRestPrice: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Persons — list, get-by-uuid, get-by-id
// ---------------------------------------------------------------------------

func TestIntegrationPersonsListAndGet(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	page := 1
	perPage := 5
	sortBy := "id"
	order := models.UserSortOrderAsc

	resp, err := c.PersonsList(ctx, &models.PersonListOptions{
		Page:    &page,
		PerPage: &perPage,
		SortBy:  &sortBy,
		Order:   &order,
	})
	if err != nil {
		t.Fatalf("PersonsList with options: %v", err)
	}
	if len(resp.Items) == 0 {
		t.Skip("no persons on instance; skipping get-by-uuid and get-by-id checks")
	}

	first := resp.Items[0]

	byUUID, err := c.PersonGet(ctx, first.UUID)
	if err != nil {
		t.Fatalf("PersonGet: %v", err)
	}
	if byUUID.UUID != first.UUID {
		t.Errorf("expected UUID %q, got %q", first.UUID, byUUID.UUID)
	}

	byID, err := c.PersonGetByID(ctx, first.ID)
	if err != nil {
		t.Fatalf("PersonGetByID: %v", err)
	}
	if byID.ID != first.ID {
		t.Errorf("expected ID %d, got %d", first.ID, byID.ID)
	}
	if byID.UUID != first.UUID {
		t.Errorf("expected UUID %q from GetByID, got %q", first.UUID, byID.UUID)
	}
}

// ---------------------------------------------------------------------------
// Persons — create + create-user
//
// Note: the customer API does not expose a DELETE endpoint for persons.
// Records created here remain on the instance — use a non-production
// instance for these tests.
// ---------------------------------------------------------------------------

func TestIntegrationPersonCreateAndCreateUser(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	suffix := uniqueSuffix()
	email := "sdk-int-" + suffix + "@example.test"

	personFields := map[string]any{
		"email":      email,
		"first_name": "SDK",
		"last_name":  "Integration " + suffix,
	}

	// System-managed fields that are reported as mandatory by field-definitions
	// but must not be sent on POST. The server fills these in.
	systemFields := map[string]bool{
		"id":                                true,
		"uuid":                              true,
		"person_uuid":                       true,
		"user_uuid":                         true,
		"created_at":                        true,
		"updated_at":                        true,
		"updated_by_user_id":                true,
		"imported_by_user_id":               true,
		"imported_with_template_id":         true,
		"imported_at":                       true,
		"created_on_import_with_template_id": true,
	}

	// Best-effort: discover any *real* mandatory person fields and synthesise
	// values. System fields are skipped.
	defs, err := c.FieldDefinitionsList(ctx, models.AssetTrackingTemplatePerson)
	if err == nil {
		for _, d := range defs {
			mandatory := false
			for _, attr := range d.Attributes {
				if attr.Type == "mandatory" && attr.Value == "yes" {
					mandatory = true
					break
				}
			}
			if !mandatory {
				continue
			}
			if systemFields[d.FieldKey] {
				continue
			}
			if _, exists := personFields[d.FieldKey]; exists {
				continue
			}
			switch d.FieldType.Name {
			case models.FieldTypeDropdown:
				for _, con := range d.FieldType.Constraints {
					if con.Type == "allowed_values" {
						if vals, ok := con.Value.([]any); ok && len(vals) > 0 {
							personFields[d.FieldKey] = vals[0]
						}
					}
				}
			case models.FieldTypeText, models.FieldTypeLongText:
				personFields[d.FieldKey] = "int-" + suffix
			default:
				t.Skipf("mandatory person field %q of type %q not auto-fillable", d.FieldKey, d.FieldType.Name)
			}
		}
	}

	uuid, err := c.PersonCreate(ctx, personFields)
	if err != nil {
		t.Fatalf("PersonCreate: %v", err)
	}
	if uuid == "" {
		t.Fatal("expected non-empty UUID from Location header")
	}

	// Verify the new person is retrievable.
	person, err := c.PersonGet(ctx, uuid)
	if err != nil {
		t.Fatalf("PersonGet after create: %v", err)
	}
	if person.Email != email {
		t.Errorf("expected email %q, got %q", email, person.Email)
	}

	// Trigger user creation by filtering on the unique email.
	err = c.PersonCreateUser(ctx, models.FilterObject{
		Filter: map[string]map[models.FilterOperator]any{
			"email": {models.FilterEq: email},
		},
	})
	if err != nil {
		t.Fatalf("PersonCreateUser: %v", err)
	}
}
