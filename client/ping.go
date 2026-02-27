package client

import (
	"context"
	"net/http"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// Ping calls the unauthenticated root endpoint and returns the API status.
func (c *Client) Ping(ctx context.Context) (*models.PingResponse, error) {
	resp, err := c.DoUnauthenticated(ctx, http.MethodGet, "", nil)
	if err != nil {
		return nil, err
	}

	var ping models.PingResponse
	if err := DecodeJSON(resp, &ping); err != nil {
		return nil, err
	}
	return &ping, nil
}
