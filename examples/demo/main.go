// Command demo exercises the core seventhings SDK modules (Auth, Objects,
// Files, Tasks) against a real instance. Configure via environment variables:
//
//	SEVENTHINGS_BASE_URL   — e.g. https://example.seventhings.com
//	SEVENTHINGS_USERNAME   — login username
//	SEVENTHINGS_PASSWORD   — login password
//	SEVENTHINGS_CLIENT_ID  — OAuth client ID
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SeventhingsCompany/customer-api-go/client"
	"github.com/SeventhingsCompany/customer-api-go/models"
)

func main() {
	// ── Configuration ────────────────────────────────────────────────────
	baseURL := requireEnv("SEVENTHINGS_BASE_URL")
	username := requireEnv("SEVENTHINGS_USERNAME")
	password := requireEnv("SEVENTHINGS_PASSWORD")
	clientID := requireEnv("SEVENTHINGS_CLIENT_ID")

	ctx := context.Background()

	// ── Auth ─────────────────────────────────────────────────────────────
	section("Auth", "Logging in…")

	c := client.New(baseURL, client.WithClientID(clientID))
	tok, err := c.Login(ctx, username, password, clientID)
	mustDo(err)

	truncated := tok.AccessToken[:20] + "…"
	pf("Auth", "Logged in — user_id=%d, token=%s", tok.UserID, truncated)

	// ── Objects ──────────────────────────────────────────────────────────
	section("Objects", "Listing objects…")

	objs, err := c.ObjectsList(ctx, &models.ListOptions{Page: 1, PerPage: 5})
	mustDo(err)
	pf("Objects", "Listed %d object(s) (first page, max 5)", len(objs))

	// Create
	ts := time.Now().UnixMilli()
	objUUID := must(c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "SDK Demo Object",
		"barcode":        fmt.Sprintf("SDK-DEMO-%d", ts),
	}))
	pf("Objects", "Created object %s", objUUID)

	// Patch
	mustDo(c.ObjectPatch(ctx, objUUID, map[string]any{"inventory_name": "SDK Demo Object (updated)"}))
	updated := must(c.ObjectGet(ctx, objUUID))
	pf("Objects", "Patched object — inventory_name=%s", updated["inventory_name"])

	// Archive / Unarchive
	mustDo(c.ObjectArchive(ctx, objUUID))
	pf("Objects", "Archived object %s", objUUID)

	mustDo(c.ObjectUnarchive(ctx, objUUID))
	pf("Objects", "Unarchived object %s", objUUID)

	// Delete + confirm 404
	mustDo(c.ObjectDelete(ctx, objUUID))
	pf("Objects", "Deleted object %s", objUUID)

	_, err = c.ObjectGet(ctx, objUUID)
	if isNotFound(err) {
		pf("Objects", "Confirmed deletion (404)")
	} else {
		log.Fatalf("[Objects] Expected 404 after deletion, got: %v", err)
	}

	// ── Filtered listing ─────────────────────────────────────────────────
	section("Objects", "Fetching last 5 changed assets (sorted + filtered)…")

	recentObjs, err := c.ObjectsList(ctx, &models.ListOptions{
		Page:    1,
		PerPage: 5,
		Sort:    map[string]models.SortDirection{"updated_at": models.SortDESC},
	})
	mustDo(err)
	pf("Objects", "Got %d recently changed asset(s):", len(recentObjs))
	for i, obj := range recentObjs {
		name, _ := obj["inventory_name"].(string)
		updatedAt, _ := obj["updated_at"].(string)
		pf("Objects", "  %d. %s (updated_at=%s)", i+1, name, updatedAt)
	}

	// Filter by name — find objects whose inventory_name contains "SDK"
	section("Objects", "Filtering assets by name containing \"SDK\"…")

	filtered, err := c.ObjectsList(ctx, &models.ListOptions{
		Page:    1,
		PerPage: 5,
		Filters: []models.FilterEntry{
			{Field: "inventory_name", Operator: models.FilterLike, Values: []string{"SDK"}},
		},
	})
	mustDo(err)
	pf("Objects", "Got %d asset(s) matching filter:", len(filtered))
	for i, obj := range filtered {
		name, _ := obj["inventory_name"].(string)
		pf("Objects", "  %d. %s", i+1, name)
	}

	// ── Files ────────────────────────────────────────────────────────────
	section("Files", "Uploading file…")

	fileContent := "Hello from the seventhings Go SDK demo!\n"
	fileUUID := must(c.FileUpload(ctx, "demo.txt", strings.NewReader(fileContent)))
	pf("Files", "Uploaded file %s (demo.txt, %d bytes)", fileUUID, len(fileContent))

	fileMeta := must(c.FileGet(ctx, fileUUID))
	pf("Files", "File metadata — name=%s, type=%s, size=%d", fileMeta.Name, fileMeta.Type, fileMeta.Size)

	// Create a temporary object to attach the file to
	tmpObjUUID := must(c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "SDK Demo File Host",
		"barcode":        fmt.Sprintf("SDK-FILE-%d", ts),
	}))
	pf("Files", "Created temp object %s for file attachment", tmpObjUUID)

	attachment := []models.FileAttachment{{FieldKey: "documents", FileUUID: fileUUID}}

	_, err = c.ObjectAddFiles(ctx, tmpObjUUID, attachment)
	mustDo(err)
	pf("Files", "Attached file %s to object %s", fileUUID, tmpObjUUID)

	// Clean up: remove file from object, then delete the object
	_, err = c.ObjectRemoveFiles(ctx, tmpObjUUID, attachment)
	mustDo(err)
	pf("Files", "Removed file from object %s", tmpObjUUID)

	mustDo(c.ObjectDelete(ctx, tmpObjUUID))
	pf("Files", "Deleted temp object %s", tmpObjUUID)

	// ── Tasks ────────────────────────────────────────────────────────────
	section("Tasks", "Creating task…")

	// Look up current user UUID for task assignee
	currentUser := must(c.UserGetByID(ctx, tok.UserID))
	pf("Tasks", "Current user UUID: %s", currentUser.UUID)

	// Create a temporary object for the task reference
	taskObjUUID := must(c.ObjectCreate(ctx, map[string]any{
		"inventory_name": "SDK Demo Task Target",
		"barcode":        fmt.Sprintf("SDK-TASK-%d", ts),
	}))
	pf("Tasks", "Created reference object %s", taskObjUUID)

	deadline := "2026-12-31"
	taskUUID := must(c.TaskCreate(ctx, models.CreateTask{
		Title:    "SDK Demo Task",
		Deadline: &deadline,
		Assignees: []string{currentUser.UUID},
		References: []models.TaskReferenceInput{
			{Type: models.TaskReferenceTypeAsset, UUID: taskObjUUID},
		},
		Reminders:         []models.TimeInterval{{Unit: models.TimeIntervalDays, Value: 1}},
		RecurringSchedule: nil,
	}))
	pf("Tasks", "Created task %s referencing object %s", taskUUID, taskObjUUID)

	// Close the task
	mustDo(c.TaskUpdateStatus(ctx, taskUUID, models.TaskStatusClosed))
	pf("Tasks", "Updated task status to closed")

	// Delete the task + confirm 404
	mustDo(c.TaskDelete(ctx, taskUUID))
	pf("Tasks", "Deleted task %s", taskUUID)

	_, err = c.TaskGet(ctx, taskUUID)
	if isNotFound(err) {
		pf("Tasks", "Confirmed deletion (404)")
	} else {
		log.Fatalf("[Tasks] Expected 404 after deletion, got: %v", err)
	}

	// Clean up reference object
	mustDo(c.ObjectDelete(ctx, taskObjUUID))
	pf("Tasks", "Deleted reference object %s", taskObjUUID)

	// ── Auth cleanup ─────────────────────────────────────────────────────
	section("Auth", "Revoking tokens…")

	mustDo(c.RevokeTokens(ctx))
	pf("Auth", "Tokens revoked")

	fmt.Println("\nDone — all steps completed successfully.")
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Missing required environment variable: %s", key)
	}
	return v
}

func must[T any](val T, err error) T {
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	return val
}

func mustDo(err error) {
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func section(tag, msg string) {
	fmt.Printf("\n── %s ──────────────────────────────────────────\n", tag)
	fmt.Printf("[%-7s] %s\n", tag, msg)
}

func pf(tag, format string, args ...any) {
	fmt.Printf("[%-7s] "+format+"\n", append([]any{tag}, args...)...)
}

func isNotFound(err error) bool {
	var apiErr *models.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}
