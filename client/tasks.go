package client

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// TasksList returns a list of tasks matching the given options.
func (c *Client) TasksList(ctx context.Context, opts *models.TaskListOptions) ([]models.Task, error) {
	p := "task-management/tasks"
	if qs := opts.Encode(); qs != "" {
		p += "?" + qs
	}
	resp, err := c.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	var result []models.Task
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TaskCreate creates a new task and returns its UUID.
func (c *Client) TaskCreate(ctx context.Context, input models.CreateTask) (string, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	resp, err := c.Post(ctx, "task-management/task", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return UUIDFromLocationHeader(resp)
}

// TaskGet returns a single task by UUID.
func (c *Client) TaskGet(ctx context.Context, uuid string) (*models.Task, error) {
	resp, err := c.Get(ctx, "task-management/task/"+uuid)
	if err != nil {
		return nil, err
	}
	var result models.Task
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TaskUpdate updates a task by UUID (PUT, returns 204).
func (c *Client) TaskUpdate(ctx context.Context, uuid string, input models.UpdateTask) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}
	_, err = c.Put(ctx, "task-management/task/"+uuid, bytes.NewReader(body))
	return err
}

// TaskDelete deletes a task by UUID.
func (c *Client) TaskDelete(ctx context.Context, uuid string) error {
	_, err := c.Delete(ctx, "task-management/task/"+uuid)
	return err
}

// TaskUpdateStatus updates a task's status by UUID (PUT, returns 204).
func (c *Client) TaskUpdateStatus(ctx context.Context, uuid string, status models.TaskStatus) error {
	body, err := json.Marshal(models.TaskStatusUpdate{Status: status})
	if err != nil {
		return err
	}
	_, err = c.Put(ctx, "task-management/task/"+uuid+"/status", bytes.NewReader(body))
	return err
}
