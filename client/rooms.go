package client

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// RoomsList returns a paginated list of rooms.
func (c *Client) RoomsList(ctx context.Context, page, perPage int) ([]map[string]any, error) {
	p := "rooms"
	var params []string
	if page != 0 {
		params = append(params, "page="+strconv.Itoa(page))
	}
	if perPage != 0 {
		params = append(params, "per_page="+strconv.Itoa(perPage))
	}
	if len(params) > 0 {
		p += "?"
		for i, param := range params {
			if i > 0 {
				p += "&"
			}
			p += param
		}
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

// RoomsCount returns the count of rooms matching the given options.
func (c *Client) RoomsCount(ctx context.Context, opts *models.ListOptions) (int, error) {
	p := "rooms/count"
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

// RoomCreate creates a new room and returns its UUID.
func (c *Client) RoomCreate(ctx context.Context, fields map[string]any) (string, error) {
	body, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	resp, err := c.Post(ctx, "room", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return UUIDFromLocationHeader(resp)
}

// RoomGet returns a single room by UUID.
func (c *Client) RoomGet(ctx context.Context, uuid string) (map[string]any, error) {
	resp, err := c.Get(ctx, "room/"+uuid)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RoomPatch updates a room and returns the updated resource.
func (c *Client) RoomPatch(ctx context.Context, uuid string, fields map[string]any) (map[string]any, error) {
	body, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	resp, err := c.Patch(ctx, "room/"+uuid, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RoomDelete deletes a room by UUID.
func (c *Client) RoomDelete(ctx context.Context, uuid string) error {
	_, err := c.Delete(ctx, "room/"+uuid)
	return err
}
