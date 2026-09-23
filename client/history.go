package client

import (
	"context"
	"net/url"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func getHistory[T any](ctx context.Context, c *Client, path string, opts *models.HistoryListOptions) (*models.HistoryResponse[T], error) {
	if qs := opts.Encode(); qs != "" {
		path += "?" + qs
	}
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result models.HistoryResponse[T]
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ObjectHistory returns a page of object changes, newest first.
func (c *Client) ObjectHistory(ctx context.Context, uuid string, opts *models.HistoryListOptions) (*models.HistoryResponse[models.ObjectHistoryEntry], error) {
	return getHistory[models.ObjectHistoryEntry](ctx, c, "object/"+url.PathEscape(uuid)+"/history", opts)
}

// RoomHistory returns a page of room changes, newest first.
func (c *Client) RoomHistory(ctx context.Context, uuid string, opts *models.HistoryListOptions) (*models.HistoryResponse[models.RoomHistoryEntry], error) {
	return getHistory[models.RoomHistoryEntry](ctx, c, "room/"+url.PathEscape(uuid)+"/history", opts)
}

// LocationHistory returns a page of location changes, newest first.
func (c *Client) LocationHistory(ctx context.Context, uuid string, opts *models.HistoryListOptions) (*models.HistoryResponse[models.LocationHistoryEntry], error) {
	return getHistory[models.LocationHistoryEntry](ctx, c, "location/"+url.PathEscape(uuid)+"/history", opts)
}

// PersonHistory returns a page of person changes, newest first.
func (c *Client) PersonHistory(ctx context.Context, uuid string, opts *models.HistoryListOptions) (*models.HistoryResponse[models.PersonHistoryEntry], error) {
	return getHistory[models.PersonHistoryEntry](ctx, c, "person/"+url.PathEscape(uuid)+"/history", opts)
}

// TaskHistory returns a page of task changes, newest first.
func (c *Client) TaskHistory(ctx context.Context, uuid string, opts *models.HistoryListOptions) (*models.HistoryResponse[models.TaskHistoryEntry], error) {
	return getHistory[models.TaskHistoryEntry](ctx, c, "task-management/task/"+url.PathEscape(uuid)+"/history", opts)
}

// RentalCaseHistory returns a page of rental case changes, newest first.
func (c *Client) RentalCaseHistory(ctx context.Context, uuid string, opts *models.HistoryListOptions) (*models.HistoryResponse[models.RentalCaseHistoryEntry], error) {
	return getHistory[models.RentalCaseHistoryEntry](ctx, c, "rental-management/rental-case/"+url.PathEscape(uuid)+"/history", opts)
}
