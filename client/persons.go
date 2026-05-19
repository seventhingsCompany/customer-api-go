package client

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// PersonsList returns a paginated list of persons.
func (c *Client) PersonsList(ctx context.Context, opts *models.PersonListOptions) (*models.PersonListResponse, error) {
	p := "persons"
	if qs := opts.Encode(); qs != "" {
		p += "?" + qs
	}
	resp, err := c.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	var result models.PersonListResponse
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PersonGet returns a single person by UUID.
func (c *Client) PersonGet(ctx context.Context, uuid string) (*models.Person, error) {
	resp, err := c.Get(ctx, "person/"+uuid)
	if err != nil {
		return nil, err
	}
	var result models.Person
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PersonGetByID returns a single person by numeric ID.
func (c *Client) PersonGetByID(ctx context.Context, id int) (*models.Person, error) {
	resp, err := c.Get(ctx, "person/by-id/"+strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	var result models.Person
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PersonCreate creates a new person and returns the UUID parsed from the
// Location header of the 201 response.
func (c *Client) PersonCreate(ctx context.Context, fields map[string]any) (string, error) {
	body, err := json.Marshal(map[string]any{"fields": fields})
	if err != nil {
		return "", err
	}
	resp, err := c.Post(ctx, "persons", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return UUIDFromLocationHeader(resp)
}

// PersonCreateUser triggers user creation for the person(s) matched by the
// given filter. Only the Filter field is sent in the request body.
func (c *Client) PersonCreateUser(ctx context.Context, filter models.FilterObject) error {
	body, err := json.Marshal(map[string]any{"filter": filter.Filter})
	if err != nil {
		return err
	}
	_, err = c.Post(ctx, "persons/create-user", bytes.NewReader(body))
	return err
}
