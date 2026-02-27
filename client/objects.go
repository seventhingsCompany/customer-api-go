package client

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// ObjectsList returns a list of objects matching the given options.
func (c *Client) ObjectsList(ctx context.Context, opts *models.ListOptions) ([]map[string]any, error) {
	p := "objects"
	if qs := opts.Encode(); qs != "" {
		p += "?" + qs
	}
	resp, err := c.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Items []map[string]any `json:"items"`
	}
	if err := DecodeJSON(resp, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Items, nil
}

// ObjectsCount returns the count of objects matching the given options.
func (c *Client) ObjectsCount(ctx context.Context, opts *models.ListOptions) (int, error) {
	p := "objects/count"
	if qs := opts.Encode(); qs != "" {
		p += "?" + qs
	}
	resp, err := c.Get(ctx, p)
	if err != nil {
		return 0, err
	}
	var cr models.CountResponse
	if err := DecodeJSON(resp, &cr); err != nil {
		return 0, err
	}
	return cr.Count, nil
}

// ObjectCreate creates a new object and returns its UUID.
func (c *Client) ObjectCreate(ctx context.Context, fields map[string]any) (string, error) {
	body, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	resp, err := c.Post(ctx, "object", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return UUIDFromLocationHeader(resp)
}

// ObjectGet returns a single object by UUID.
func (c *Client) ObjectGet(ctx context.Context, uuid string) (map[string]any, error) {
	resp, err := c.Get(ctx, "object/"+uuid)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ObjectPatch updates an object by UUID (PATCH, returns 204).
func (c *Client) ObjectPatch(ctx context.Context, uuid string, fields map[string]any) error {
	body, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	_, err = c.Patch(ctx, "object/"+uuid, bytes.NewReader(body))
	return err
}

// ObjectDelete deletes an object by UUID.
func (c *Client) ObjectDelete(ctx context.Context, uuid string) error {
	_, err := c.Delete(ctx, "object/"+uuid)
	return err
}

// ObjectArchive archives an object by UUID.
func (c *Client) ObjectArchive(ctx context.Context, uuid string) error {
	_, err := c.Post(ctx, "object/"+uuid+"/archive", nil)
	return err
}

// ObjectUnarchive unarchives an object by UUID.
func (c *Client) ObjectUnarchive(ctx context.Context, uuid string) error {
	_, err := c.Post(ctx, "object/"+uuid+"/unarchive", nil)
	return err
}

// ObjectAddFiles attaches files to an object. The response is returned directly
// because the API may return 200 or 207 (multi-status).
func (c *Client) ObjectAddFiles(ctx context.Context, uuid string, attachments []models.FileAttachment) (*Response, error) {
	body, err := json.Marshal(attachments)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "object/"+uuid+"/add-file", bytes.NewReader(body))
}

// ObjectRemoveFiles removes file attachments from an object. The response is
// returned directly because the API may return 200 or 207 (multi-status).
func (c *Client) ObjectRemoveFiles(ctx context.Context, uuid string, attachments []models.FileAttachment) (*Response, error) {
	body, err := json.Marshal(attachments)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "object/"+uuid+"/remove-file", bytes.NewReader(body))
}
