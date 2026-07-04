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

// MandatoryFieldDefinitions returns the field definitions for the given
// template that are configured as required for this instance. Use it before
// creating a resource (object/asset, room, person) to discover which fields
// must be supplied — this can vary per instance.
//
// System-managed keys (see models.SystemManagedFieldKeys) are excluded, since
// they may be reported as mandatory but must not be sent on create.
func (c *Client) MandatoryFieldDefinitions(ctx context.Context, template models.AssetTrackingTemplate) ([]models.FieldDefinition, error) {
	defs, err := c.FieldDefinitionsList(ctx, template)
	if err != nil {
		return nil, err
	}
	var mandatory []models.FieldDefinition
	for _, d := range defs {
		if d.IsMandatory() && !models.SystemManagedFieldKeys[d.FieldKey] {
			mandatory = append(mandatory, d)
		}
	}
	return mandatory, nil
}

// MissingMandatoryFields returns the field keys that are required for the given
// template (per this instance) but are absent — or present with a nil value —
// in fields. System-managed keys are excluded (see MandatoryFieldDefinitions),
// so they never appear in the result. An empty slice means the payload
// satisfies every instance-required field.
//
// Use it to fail fast before ObjectCreate/RoomCreate/PersonCreate. It performs
// one request (to fetch field definitions) and checks presence only — it does
// not validate values.
func (c *Client) MissingMandatoryFields(ctx context.Context, template models.AssetTrackingTemplate, fields map[string]any) ([]string, error) {
	defs, err := c.MandatoryFieldDefinitions(ctx, template)
	if err != nil {
		return nil, err
	}
	var missing []string
	for _, d := range defs {
		if v, ok := fields[d.FieldKey]; !ok || v == nil {
			missing = append(missing, d.FieldKey)
		}
	}
	return missing, nil
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
