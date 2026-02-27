package client

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// CircularityHubSuggestCategory returns category suggestions based on the
// given filter object.
func (c *Client) CircularityHubSuggestCategory(ctx context.Context, filter models.FilterObject) (map[string]string, error) {
	body, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	resp, err := c.Post(ctx, "circularity-hub/suggest-category", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var result map[string]string
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CircularityHubSuggestRestPrice returns rest-price suggestions for the given
// input fields.
func (c *Client) CircularityHubSuggestRestPrice(ctx context.Context, input map[string]string) (map[string]string, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	resp, err := c.Post(ctx, "circularity-hub/suggest-rest-price", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var result map[string]string
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CircularityHubAddObjects adds objects to the CircularityHub in bulk.
func (c *Client) CircularityHubAddObjects(ctx context.Context, entries map[string]models.AddObjectEntry) error {
	body, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	_, err = c.Post(ctx, "circularity-hub/add-objects-to-circularity-hub", bytes.NewReader(body))
	return err
}

// CircularityHubItemsList returns a list of CircularityHub items.
func (c *Client) CircularityHubItemsList(ctx context.Context, opts *models.ListOptions) ([]map[string]any, error) {
	p := "circularity-hub/items"
	if qs := opts.Encode(); qs != "" {
		p += "?" + qs
	}
	resp, err := c.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CircularityHubItemGet returns a single CircularityHub item by integer ID.
func (c *Client) CircularityHubItemGet(ctx context.Context, id int) (map[string]any, error) {
	resp, err := c.Get(ctx, "circularity-hub/item/"+strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CircularityHubItemUpdate updates a CircularityHub item by integer ID.
func (c *Client) CircularityHubItemUpdate(ctx context.Context, id int, fields map[string]any) error {
	body, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	_, err = c.Patch(ctx, "circularity-hub/item/"+strconv.Itoa(id), bytes.NewReader(body))
	return err
}

// CircularityHubItemDelete deletes a CircularityHub item by integer ID.
func (c *Client) CircularityHubItemDelete(ctx context.Context, id int) error {
	_, err := c.Delete(ctx, "circularity-hub/item/"+strconv.Itoa(id))
	return err
}

// CircularityHubOrdersList returns a list of CircularityHub orders.
func (c *Client) CircularityHubOrdersList(ctx context.Context, opts *models.ListOptions) ([]models.CircularityHubOrder, error) {
	p := "circularity-hub/orders"
	if qs := opts.Encode(); qs != "" {
		p += "?" + qs
	}
	resp, err := c.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	var result []models.CircularityHubOrder
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CircularityHubOrderCreate creates a new CircularityHub order from a list of
// item IDs and returns the new order's integer ID.
func (c *Client) CircularityHubOrderCreate(ctx context.Context, itemIDs []int) (int, error) {
	body, err := json.Marshal(itemIDs)
	if err != nil {
		return 0, err
	}
	resp, err := c.Post(ctx, "circularity-hub/orders", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	return IntFromLocationIDHeader(resp)
}

// CircularityHubOrderGet returns a single CircularityHub order by integer ID.
func (c *Client) CircularityHubOrderGet(ctx context.Context, id int) (*models.CircularityHubOrder, error) {
	resp, err := c.Get(ctx, "circularity-hub/order/"+strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	var result models.CircularityHubOrder
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CircularityHubOrderUpdate updates a CircularityHub order by integer ID.
func (c *Client) CircularityHubOrderUpdate(ctx context.Context, id int, fields map[string]any) error {
	body, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	_, err = c.Patch(ctx, "circularity-hub/order/"+strconv.Itoa(id), bytes.NewReader(body))
	return err
}
