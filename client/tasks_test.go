package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestTasksList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/task-management/tasks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"uuid":"t1","title":"Fix bug","status":"open"}]`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	tasks, err := c.TasksList(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].UUID != "t1" {
		t.Errorf("expected uuid t1, got %s", tasks[0].UUID)
	}
}

func TestTasksListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "open" {
			t.Errorf("expected status=open, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("assignee") != "user-1" {
			t.Errorf("expected assignee=user-1, got %s", r.URL.Query().Get("assignee"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	status := models.TaskStatusOpen
	assignee := "user-1"
	_, err := c.TasksList(context.Background(), &models.TaskListOptions{
		Status:   &status,
		Assignee: &assignee,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskCreateLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/task-management/task" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Location", "/customer-api/v1/task-management/task/new-task-uuid")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	uuid, err := c.TaskCreate(context.Background(), models.CreateTask{Title: "New task"})
	if err != nil {
		t.Fatal(err)
	}
	if uuid != "new-task-uuid" {
		t.Errorf("expected new-task-uuid, got %q", uuid)
	}
}

func TestTaskCreateBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var input models.CreateTask
		if err := json.Unmarshal(body, &input); err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		if input.Title != "Test task" {
			t.Errorf("expected title 'Test task', got %q", input.Title)
		}
		if len(input.References) != 1 || input.References[0].Type != models.TaskReferenceTypeAsset {
			t.Errorf("unexpected references: %+v", input.References)
		}
		w.Header().Set("Location", "/customer-api/v1/task-management/task/uuid")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.TaskCreate(context.Background(), models.CreateTask{
		Title: "Test task",
		References: []models.TaskReferenceInput{
			{Type: models.TaskReferenceTypeAsset, UUID: "asset-1"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/task-management/task/t1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"t1","title":"Fix bug","status":"open","author":"user-1","created_at":"2024-01-01","updated_at":"2024-01-02","comment":"notes","recurring_schedule":{"unit":"weeks","value":2}}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	task, err := c.TaskGet(context.Background(), "t1")
	if err != nil {
		t.Fatal(err)
	}
	if task.Title != "Fix bug" {
		t.Errorf("expected title 'Fix bug', got %q", task.Title)
	}
	if task.RecurringSchedule == nil || task.RecurringSchedule.Value != 2 {
		t.Errorf("expected recurring schedule with value 2, got %+v", task.RecurringSchedule)
	}
}

func TestTaskUpdatePUT204(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/task-management/task/t1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.TaskUpdate(context.Background(), "t1", models.UpdateTask{Title: "Updated"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/task-management/task/t1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.TaskDelete(context.Background(), "t1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskUpdateStatusBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/task-management/task/t1/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var su models.TaskStatusUpdate
		if err := json.Unmarshal(body, &su); err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		if su.Status != models.TaskStatusClosed {
			t.Errorf("expected status closed, got %s", su.Status)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.TaskUpdateStatus(context.Background(), "t1", models.TaskStatusClosed)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTasksListError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.TasksList(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestTaskCreateError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.TaskCreate(context.Background(), models.CreateTask{Title: "New task"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestTaskUpdateError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.TaskUpdate(context.Background(), "t1", models.UpdateTask{Title: "Updated"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestTaskUpdateStatusError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.TaskUpdateStatus(context.Background(), "t1", models.TaskStatusClosed)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Errorf("expected 500 APIError, got %v", err)
	}
}

func TestTaskGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.TaskGet(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}
