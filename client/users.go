package client

import (
	"context"
	"strconv"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// UsersList returns a paginated list of users.
func (c *Client) UsersList(ctx context.Context, opts *models.UserListOptions) (*models.UserListResponse, error) {
	p := "users"
	if qs := opts.Encode(); qs != "" {
		p += "?" + qs
	}
	resp, err := c.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	var result models.UserListResponse
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UserGet returns a single user by UUID.
func (c *Client) UserGet(ctx context.Context, uuid string) (*models.User, error) {
	resp, err := c.Get(ctx, "user/"+uuid)
	if err != nil {
		return nil, err
	}
	var result models.User
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UserGetByID returns a single user by numeric ID.
func (c *Client) UserGetByID(ctx context.Context, id int) (*models.User, error) {
	resp, err := c.Get(ctx, "user/by-id/"+strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	var result models.User
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
