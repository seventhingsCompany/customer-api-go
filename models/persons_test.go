package models

import (
	"encoding/json"
	"testing"
)

func TestPersonUnmarshalPreservesCustomFields(t *testing.T) {
	// A person payload with the common typed columns plus two instance-defined
	// custom fields (cost_center, employee_number) that have no typed property.
	raw := `{
		"person_uuid": "p-1",
		"id": 7,
		"email": "a@b.com",
		"first_name": "Ada",
		"cost_center": "CC-100",
		"employee_number": 4242
	}`

	var p Person
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Typed convenience fields still decode.
	if p.UUID != "p-1" || p.ID != 7 || p.Email != "a@b.com" {
		t.Errorf("typed fields wrong: uuid=%q id=%d email=%q", p.UUID, p.ID, p.Email)
	}
	if p.Firstname == nil || *p.Firstname != "Ada" {
		t.Errorf("first_name = %v", p.Firstname)
	}

	// The full raw map is captured, including keys with a typed property.
	// Note the person wire format uses "person_uuid", not "uuid", so the raw
	// bag keys mirror the API exactly.
	if p.Fields == nil {
		t.Fatal("Fields was not populated")
	}
	if got, ok := p.Fields.String("person_uuid"); !ok || got != "p-1" {
		t.Errorf("Fields[person_uuid] = %q, %v", got, ok)
	}
	if got, ok := p.Fields.String("email"); !ok || got != "a@b.com" {
		t.Errorf("Fields[email] = %q, %v", got, ok)
	}

	// Custom fields survive and are reachable via the typed accessors.
	if cc, ok := p.Fields.String("cost_center"); !ok || cc != "CC-100" {
		t.Errorf("Fields[cost_center] = %q, %v", cc, ok)
	}
	if num, ok := p.Fields.Int("employee_number"); !ok || num != 4242 {
		t.Errorf("Fields[employee_number] = %d, %v", num, ok)
	}
}

func TestPersonUnmarshalWithinListResponse(t *testing.T) {
	raw := `{
		"items": [
			{"person_uuid": "p-1", "id": 1, "email": "a@b.com", "cost_center": "CC-1"},
			{"person_uuid": "p-2", "id": 2, "email": "c@d.com"}
		],
		"page": 1, "per_page": 50, "sort_by": "id", "order": "asc", "total": 2
	}`

	var resp PersonListResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	// Custom field on the first item is preserved through the list decode.
	if cc, ok := resp.Items[0].Fields.String("cost_center"); !ok || cc != "CC-1" {
		t.Errorf("item[0] cost_center = %q, %v", cc, ok)
	}
	// The second item has no custom fields but still gets a populated bag.
	if resp.Items[1].Fields == nil {
		t.Error("item[1] Fields should be populated even without custom fields")
	}
}

func TestPersonMarshalRoundTripUnaffected(t *testing.T) {
	// Marshaling a Person should not emit a "Fields" key (json:"-"), so
	// serialization behavior is unchanged for existing callers.
	p := Person{UUID: "p-1", ID: 1, Email: "a@b.com"}
	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var check map[string]any
	if err := json.Unmarshal(out, &check); err != nil {
		t.Fatal(err)
	}
	if _, exists := check["Fields"]; exists {
		t.Error("marshaled Person unexpectedly contains a Fields key")
	}
	if _, exists := check["fields"]; exists {
		t.Error("marshaled Person unexpectedly contains a fields key")
	}
}
