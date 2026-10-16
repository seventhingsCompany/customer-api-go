package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// ErrRentalCaseRenterRequired is returned before sending a rental create or
// update without a renter. Although the API description marks renter optional,
// the current server rejects null and fails with HTTP 500 when it is omitted.
var ErrRentalCaseRenterRequired = errors.New("rental case renter is required: provide a renter with type and value; the current API fails when renter is omitted")

// RentalCasesList returns a list of rental cases.
func (c *Client) RentalCasesList(ctx context.Context, opts *models.ListOptions) ([]models.RentalCase, error) {
	p := "rental-management/rental-cases"
	if qs := opts.Encode(); qs != "" {
		p += "?" + qs
	}
	resp, err := c.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Items []models.RentalCase `json:"items"`
	}
	if err := DecodeJSON(resp, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Items, nil
}

// RentalCaseCreate creates a new rental case and returns its UUID.
// A nil Renter returns ErrRentalCaseRenterRequired without sending a request.
func (c *Client) RentalCaseCreate(ctx context.Context, input models.CreateRentalCase) (string, error) {
	if input.Renter == nil {
		return "", ErrRentalCaseRenterRequired
	}
	body, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	resp, err := c.Post(ctx, "rental-management/rental-case", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return UUIDFromLocationHeader(resp)
}

// RentalCaseGet returns a single rental case by UUID.
func (c *Client) RentalCaseGet(ctx context.Context, uuid string) (*models.RentalCase, error) {
	resp, err := c.Get(ctx, "rental-management/rental-case/"+uuid)
	if err != nil {
		return nil, err
	}
	var result models.RentalCase
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RentalCaseUpdate replaces a rental case by UUID (PUT, returns 204).
// Send the complete state, not a partial patch: omitted values may be cleared.
// A nil Renter returns ErrRentalCaseRenterRequired without sending a request.
func (c *Client) RentalCaseUpdate(ctx context.Context, uuid string, input models.UpdateRentalCase) error {
	if input.Renter == nil {
		return ErrRentalCaseRenterRequired
	}
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}
	_, err = c.Put(ctx, "rental-management/rental-case/"+uuid, bytes.NewReader(body))
	return err
}

// RentalCaseDelete deletes a rental case by UUID.
func (c *Client) RentalCaseDelete(ctx context.Context, uuid string) error {
	_, err := c.Delete(ctx, "rental-management/rental-case/"+uuid)
	return err
}
