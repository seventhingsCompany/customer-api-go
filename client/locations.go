package client

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// LocationsList returns a list of locations matching the given options.
func (c *Client) LocationsList(ctx context.Context, opts *models.ListOptions) ([]map[string]any, error) {
	p := "locations"
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

// LocationsCount returns the count of locations matching the given options.
func (c *Client) LocationsCount(ctx context.Context, opts *models.ListOptions) (int, error) {
	p := "locations/count"
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

// LocationCreate creates a new location and returns its UUID.
func (c *Client) LocationCreate(ctx context.Context, fields map[string]any) (string, error) {
	body, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	resp, err := c.Post(ctx, "location", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return UUIDFromLocationHeader(resp)
}

// LocationGet returns a single location by UUID.
func (c *Client) LocationGet(ctx context.Context, uuid string) (map[string]any, error) {
	resp, err := c.Get(ctx, "location/"+uuid)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// LocationPatch updates a location and returns the updated resource.
func (c *Client) LocationPatch(ctx context.Context, uuid string, fields map[string]any) (map[string]any, error) {
	body, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	resp, err := c.Patch(ctx, "location/"+uuid, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// LocationDelete deletes a location by UUID.
func (c *Client) LocationDelete(ctx context.Context, uuid string) error {
	_, err := c.Delete(ctx, "location/"+uuid)
	return err
}
