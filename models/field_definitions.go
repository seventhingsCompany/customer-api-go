package models

// AssetTrackingTemplate identifies the template type for field definitions.
type AssetTrackingTemplate string

const (
	// AssetTrackingTemplateAsset is the asset template.
	AssetTrackingTemplateAsset AssetTrackingTemplate = "asset"
	// AssetTrackingTemplateRoom is the room template.
	AssetTrackingTemplateRoom AssetTrackingTemplate = "room"
	// AssetTrackingTemplatePerson is the person template.
	AssetTrackingTemplatePerson AssetTrackingTemplate = "person"
)

// FieldTypeName identifies the type of a field definition.
type FieldTypeName string

const (
	// FieldTypeAttachment is an attachment field.
	FieldTypeAttachment FieldTypeName = "ATTACHMENT"
	// FieldTypeBarcode is a barcode field.
	FieldTypeBarcode FieldTypeName = "BARCODE"
	// FieldTypeBoolean is a boolean field.
	FieldTypeBoolean FieldTypeName = "BOOLEAN"
	// FieldTypeDate is a date field.
	FieldTypeDate FieldTypeName = "DATE"
	// FieldTypeDatetime is a datetime field.
	FieldTypeDatetime FieldTypeName = "DATETIME"
	// FieldTypeDecimal is a decimal field.
	FieldTypeDecimal FieldTypeName = "DECIMAL"
	// FieldTypeDropdown is a dropdown field.
	FieldTypeDropdown FieldTypeName = "DROPDOWN"
	// FieldTypeFieldValueComparison is a field value comparison field.
	FieldTypeFieldValueComparison FieldTypeName = "FIELD_VALUE_COMPARISON"
	// FieldTypeLink is a link field.
	FieldTypeLink FieldTypeName = "LINK"
	// FieldTypeLinkedAssets is a linked assets field.
	FieldTypeLinkedAssets FieldTypeName = "LINKED_ASSETS"
	// FieldTypeLinkedLocation is a linked location field.
	FieldTypeLinkedLocation FieldTypeName = "LINKED_LOCATION"
	// FieldTypeLinkedRoom is a linked room field.
	FieldTypeLinkedRoom FieldTypeName = "LINKED_ROOM"
	// FieldTypeLinkedUser is a linked user field.
	FieldTypeLinkedUser FieldTypeName = "LINKED_USER"
	// FieldTypeLongText is a long text field.
	FieldTypeLongText FieldTypeName = "LONG_TEXT"
	// FieldTypeMoney is a money field.
	FieldTypeMoney FieldTypeName = "MONEY"
	// FieldTypeNumber is a number field.
	FieldTypeNumber FieldTypeName = "NUMBER"
	// FieldTypeReminder is a reminder field.
	FieldTypeReminder FieldTypeName = "REMINDER"
	// FieldTypeText is a text field.
	FieldTypeText FieldTypeName = "TEXT"
)

// FieldValueConstraint defines a constraint on a field's value.
// Value can be a string, int, or []string depending on the constraint type.
type FieldValueConstraint struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// FieldAttribute defines an attribute on a field definition.
// Value is a string or number depending on the attribute type.
type FieldAttribute struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// FieldRelation defines a relation between field definitions.
type FieldRelation struct {
	Type             string `json:"type"`
	FieldUUID        string `json:"field_uuid"`
	ComparisonValues []any  `json:"comparison_values,omitempty"`
}

// FieldDefinitionFieldType describes the type and constraints of a field definition.
type FieldDefinitionFieldType struct {
	Name        FieldTypeName          `json:"name"`
	Constraints []FieldValueConstraint `json:"constraints"`
}

// FieldDefinition represents a field definition in the seventhings API.
type FieldDefinition struct {
	UUID           string                   `json:"uuid"`
	FieldKey       string                   `json:"field_key"`
	FieldType      FieldDefinitionFieldType `json:"field_type"`
	Label          string                   `json:"label"`
	Attributes     []FieldAttribute         `json:"attributes"`
	Relations      []FieldRelation          `json:"relations,omitempty"`
	Comment        *string                  `json:"comment"`
	DefaultValue   any                      `json:"default_value"`
	PossibleValues []any                    `json:"possible_values"`
}

// CreateFieldDefinition is the request body for creating a field definition.
type CreateFieldDefinition struct {
	FieldType      FieldDefinitionFieldType `json:"field_type"`
	Label          string                   `json:"label"`
	Attributes     []FieldAttribute         `json:"attributes"`
	Relations      []FieldRelation          `json:"relations"`
	Comment        *string                  `json:"comment"`
	DefaultValue   any                      `json:"default_value"`
	PossibleValues []any                    `json:"possible_values"`
}

// UpdateFieldDefinition is the request body for updating a field definition.
// The PUT endpoint requires UUID and FieldKey in addition to the creation fields.
type UpdateFieldDefinition struct {
	UUID           string                   `json:"uuid"`
	FieldKey       string                   `json:"field_key"`
	FieldType      FieldDefinitionFieldType `json:"field_type"`
	Label          string                   `json:"label"`
	Attributes     []FieldAttribute         `json:"attributes"`
	Relations      []FieldRelation          `json:"relations"`
	Comment        *string                  `json:"comment"`
	DefaultValue   any                      `json:"default_value"`
	PossibleValues []any                    `json:"possible_values"`
}
