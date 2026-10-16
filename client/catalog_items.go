package client

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// CatalogItemsList returns catalog items with pagination, filtering and sorting.
// The instance must have the product catalog feature enabled.
func (c *Client) CatalogItemsList(ctx context.Context, opts *models.ListOptions) (*models.CatalogItemListResponse, error) {
	p := "catalog-items"
	if qs := opts.Encode(); qs != "" {
		p += "?" + qs
	}
	resp, err := c.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	var result models.CatalogItemListResponse
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CatalogItemGet returns a catalog item by UUID.
func (c *Client) CatalogItemGet(ctx context.Context, uuid string) (*models.CatalogItem, error) {
	resp, err := c.Get(ctx, "catalog-item/"+uuid)
	if err != nil {
		return nil, err
	}
	var result models.CatalogItem
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CatalogItemCreate creates a catalog item and returns its UUID.
func (c *Client) CatalogItemCreate(ctx context.Context, fields map[string]any) (string, error) {
	body, err := json.Marshal(map[string]any{"fields": fields})
	if err != nil {
		return "", err
	}
	resp, err := c.Post(ctx, "catalog-item", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	if uuid := resp.Header.Get("Location-UUID"); uuid != "" {
		return uuid, nil
	}
	return UUIDFromLocationHeader(resp)
}

// CatalogItemUpdate replaces all catalog fields. Missing fields are emptied;
// objects previously created from this item are not changed.
func (c *Client) CatalogItemUpdate(ctx context.Context, uuid string, fields map[string]any) error {
	body, err := json.Marshal(map[string]any{"fields": fields})
	if err != nil {
		return err
	}
	_, err = c.Put(ctx, "catalog-item/"+uuid, bytes.NewReader(body))
	return err
}

// CatalogItemDelete deletes a catalog item, leaving its created objects intact.
func (c *Client) CatalogItemDelete(ctx context.Context, uuid string) error {
	_, err := c.Delete(ctx, "catalog-item/"+uuid)
	return err
}
