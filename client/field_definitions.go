package client

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// FieldDefinitionsList returns all field definitions for the given template.
func (c *Client) FieldDefinitionsList(ctx context.Context, template models.AssetTrackingTemplate) ([]models.FieldDefinition, error) {
	resp, err := c.Get(ctx, "asset-tracking/"+string(template)+"/field-definitions")
	if err != nil {
		return nil, err
	}
	var result []models.FieldDefinition
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// FieldDefinitionCreate creates a new field definition and returns its UUID.
func (c *Client) FieldDefinitionCreate(ctx context.Context, template models.AssetTrackingTemplate, input models.CreateFieldDefinition) (string, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	resp, err := c.Post(ctx, "asset-tracking/"+string(template)+"/field-definition", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return UUIDFromLocationHeader(resp)
}

// FieldDefinitionGet returns a single field definition by UUID.
func (c *Client) FieldDefinitionGet(ctx context.Context, template models.AssetTrackingTemplate, uuid string) (*models.FieldDefinition, error) {
	resp, err := c.Get(ctx, "asset-tracking/"+string(template)+"/field-definition/"+uuid)
	if err != nil {
		return nil, err
	}
	var result models.FieldDefinition
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FieldDefinitionUpdate updates a field definition by UUID (PUT, returns 204).
func (c *Client) FieldDefinitionUpdate(ctx context.Context, template models.AssetTrackingTemplate, uuid string, input models.UpdateFieldDefinition) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}
	_, err = c.Put(ctx, "asset-tracking/"+string(template)+"/field-definition/"+uuid, bytes.NewReader(body))
	return err
}
