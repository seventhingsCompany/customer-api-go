//go:build integration

package seventhings_test

import (
	"context"
	"os"
	"testing"

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
	if ping.Status != "ok" {
		t.Errorf("expected status ok, got %q", ping.Status)
	}
}

func TestIntegrationAuth(t *testing.T) {
	c := integrationClient(t)

	if c.Token() == "" {
		t.Fatal("expected token to be set after login")
	}
}

func TestIntegrationObjects(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Create
	uuid, err := c.ObjectCreate(ctx, map[string]any{"name": "integration-test-object"})
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
	if obj["name"] != "integration-test-object" {
		t.Errorf("expected name integration-test-object, got %v", obj["name"])
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
	updated, err := c.ObjectPatch(ctx, uuid, map[string]any{"name": "updated-object"})
	if err != nil {
		t.Fatalf("ObjectPatch: %v", err)
	}
	if updated["name"] != "updated-object" {
		t.Errorf("expected updated name, got %v", updated["name"])
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

func TestIntegrationFiles(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	files, err := c.FilesList(ctx)
	if err != nil {
		t.Fatalf("FilesList: %v", err)
	}
	_ = files // may be empty
}

func TestIntegrationTasks(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Create
	uuid, err := c.TaskCreate(ctx, models.CreateTask{Title: "integration-test-task"})
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

	// Update
	if err := c.TaskUpdate(ctx, uuid, models.UpdateTask{Title: "updated-task"}); err != nil {
		t.Fatalf("TaskUpdate: %v", err)
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

func TestIntegrationRentals(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Create
	uuid, err := c.RentalCaseCreate(ctx, models.CreateRentalCase{})
	if err != nil {
		t.Fatalf("RentalCaseCreate: %v", err)
	}
	t.Cleanup(func() {
		_ = c.RentalCaseDelete(ctx, uuid)
	})

	// Get
	rc, err := c.RentalCaseGet(ctx, uuid)
	if err != nil {
		t.Fatalf("RentalCaseGet: %v", err)
	}
	if rc.UUID != uuid {
		t.Errorf("expected uuid %s, got %s", uuid, rc.UUID)
	}

	// List
	cases, err := c.RentalCasesList(ctx, nil)
	if err != nil {
		t.Fatalf("RentalCasesList: %v", err)
	}
	if len(cases) == 0 {
		t.Error("expected at least 1 rental case")
	}

	// Update
	comment := "integration test"
	if err := c.RentalCaseUpdate(ctx, uuid, models.UpdateRentalCase{Comment: &comment}); err != nil {
		t.Fatalf("RentalCaseUpdate: %v", err)
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

func TestIntegrationLocations(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// List
	locations, err := c.LocationsList(ctx, 0, 0)
	if err != nil {
		t.Fatalf("LocationsList: %v", err)
	}
	_ = locations
}

func TestIntegrationRooms(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// List
	rooms, err := c.RoomsList(ctx, 0, 0)
	if err != nil {
		t.Fatalf("RoomsList: %v", err)
	}
	_ = rooms
}

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

func TestIntegrationFieldDefinitions(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	defs, err := c.FieldDefinitionsList(ctx, models.AssetTrackingTemplateAsset)
	if err != nil {
		t.Fatalf("FieldDefinitionsList: %v", err)
	}
	_ = defs
}

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
