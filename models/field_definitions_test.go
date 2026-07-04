package models

import "testing"

func TestFieldDefinitionAttribute(t *testing.T) {
	d := FieldDefinition{
		Attributes: []FieldAttribute{
			{Type: "placeholder", Value: "Select..."},
			{Type: FieldAttributeMandatory, Value: "yes"},
		},
	}

	if v, ok := d.Attribute("placeholder"); !ok || v != "Select..." {
		t.Errorf("Attribute(placeholder) = %v, %v", v, ok)
	}
	if v, ok := d.Attribute(FieldAttributeMandatory); !ok || v != "yes" {
		t.Errorf("Attribute(mandatory) = %v, %v", v, ok)
	}
	if v, ok := d.Attribute("missing"); ok || v != nil {
		t.Errorf("Attribute(missing) = %v, %v; want nil, false", v, ok)
	}
}

func TestFieldDefinitionIsMandatory(t *testing.T) {
	tests := []struct {
		name  string
		attrs []FieldAttribute
		want  bool
	}{
		{"mandatory yes", []FieldAttribute{{Type: "mandatory", Value: "yes"}}, true},
		{"mandatory no", []FieldAttribute{{Type: "mandatory", Value: "no"}}, false},
		{"no attribute", nil, false},
		{"other attribute", []FieldAttribute{{Type: "placeholder", Value: "x"}}, false},
		{"non-string value", []FieldAttribute{{Type: "mandatory", Value: true}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := FieldDefinition{Attributes: tt.attrs}
			if got := d.IsMandatory(); got != tt.want {
				t.Errorf("IsMandatory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldDefinitionFieldTypeAllowedValues(t *testing.T) {
	ft := FieldDefinitionFieldType{
		Name: FieldTypeDropdown,
		Constraints: []FieldValueConstraint{
			{Type: "allowed_values", Value: []any{"A", "B"}},
		},
	}
	vals, ok := ft.AllowedValues()
	if !ok {
		t.Fatal("expected allowed_values constraint to be found")
	}
	if len(vals) != 2 || vals[0] != "A" || vals[1] != "B" {
		t.Errorf("AllowedValues() = %v", vals)
	}

	// No constraint present.
	empty := FieldDefinitionFieldType{Name: FieldTypeText}
	if _, ok := empty.AllowedValues(); ok {
		t.Error("expected no allowed_values for TEXT field")
	}
}

func TestSystemManagedFieldKeys(t *testing.T) {
	for _, k := range []string{"id", "uuid", "created_at", "updated_at"} {
		if !SystemManagedFieldKeys[k] {
			t.Errorf("expected %q to be system-managed", k)
		}
	}
	if SystemManagedFieldKeys["name"] {
		t.Error("name should not be system-managed")
	}
}
