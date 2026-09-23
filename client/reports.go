package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// ReportTemplatesList returns all PDF templates available on the instance.
func (c *Client) ReportTemplatesList(ctx context.Context) ([]models.ReportTemplate, error) {
	resp, err := c.Get(ctx, "report-template")
	if err != nil {
		return nil, err
	}
	var result []models.ReportTemplate
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ReportCreate renders objects into a template and returns the PDF bytes.
// The API generates a new document on each call and does not store it.
func (c *Client) ReportCreate(ctx context.Context, input models.CreateReport) ([]byte, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "report", bytes.NewReader(body), true)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/pdf")
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
