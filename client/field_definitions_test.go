package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestFieldDefinitionsListAsset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/asset-tracking/asset/field-definitions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"uuid":"fd1","field_key":"name","field_type":{"name":"TEXT","constraints":[]},"label":"Name"}]`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	defs, err := c.FieldDefinitionsList(context.Background(), models.AssetTrackingTemplateAsset)
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(defs))
	}
	if defs[0].UUID != "fd1" {
		t.Errorf("expected uuid fd1, got %s", defs[0].UUID)
	}
}

func TestFieldDefinitionsListRoom(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/asset-tracking/room/field-definitions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	defs, err := c.FieldDefinitionsList(context.Background(), models.AssetTrackingTemplateRoom)
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) != 0 {
		t.Fatalf("expected 0 definitions, got %d", len(defs))
	}
}

func TestFieldDefinitionCreateLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/asset-tracking/asset/field-definition" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Location", "/customer-api/v1/asset-tracking/asset/field-definition/new-fd-uuid")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	uuid, err := c.FieldDefinitionCreate(context.Background(), models.AssetTrackingTemplateAsset, models.CreateFieldDefinition{
		FieldType: models.FieldDefinitionFieldType{Name: models.FieldTypeText},
		Label:     "Name",
	})
	if err != nil {
		t.Fatal(err)
	}
	if uuid != "new-fd-uuid" {
		t.Errorf("expected new-fd-uuid, got %q", uuid)
	}
}

func TestFieldDefinitionCreateWithConstraints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var input models.CreateFieldDefinition
		if err := json.Unmarshal(body, &input); err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		if input.FieldType.Name != models.FieldTypeNumber {
			t.Errorf("expected NUMBER, got %s", input.FieldType.Name)
		}
		if len(input.FieldType.Constraints) != 1 {
			t.Fatalf("expected 1 constraint, got %d", len(input.FieldType.Constraints))
		}
		if input.FieldType.Constraints[0].Type != "min" {
			t.Errorf("expected constraint type min, got %s", input.FieldType.Constraints[0].Type)
		}
		w.Header().Set("Location", "/customer-api/v1/asset-tracking/asset/field-definition/uuid")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.FieldDefinitionCreate(context.Background(), models.AssetTrackingTemplateAsset, models.CreateFieldDefinition{
		FieldType: models.FieldDefinitionFieldType{
			Name:        models.FieldTypeNumber,
			Constraints: []models.FieldValueConstraint{{Type: "min", Value: 0}},
		},
		Label: "Count",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFieldDefinitionGetFull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/asset-tracking/asset/field-definition/fd1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"uuid":"fd1",
			"field_key":"status",
			"field_type":{"name":"DROPDOWN","constraints":[{"type":"options","value":["a","b"]}]},
			"label":"Status",
			"attributes":[{"type":"placeholder","value":"Select..."}],
			"relations":[{"type":"depends_on","field_uuid":"fd2","comparison_values":["x"]}],
			"comment":"A dropdown",
			"default_value":"a",
			"possible_values":["a","b","c"]
		}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	fd, err := c.FieldDefinitionGet(context.Background(), models.AssetTrackingTemplateAsset, "fd1")
	if err != nil {
		t.Fatal(err)
	}
	if fd.Label != "Status" {
		t.Errorf("expected label Status, got %s", fd.Label)
	}
	if fd.FieldType.Name != models.FieldTypeDropdown {
		t.Errorf("expected DROPDOWN, got %s", fd.FieldType.Name)
	}
	if len(fd.Relations) != 1 || fd.Relations[0].FieldUUID != "fd2" {
		t.Errorf("unexpected relations: %+v", fd.Relations)
	}
	if fd.Comment == nil || *fd.Comment != "A dropdown" {
		t.Errorf("expected comment, got %v", fd.Comment)
	}
}

func TestFieldDefinitionUpdatePUT204(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/asset-tracking/asset/field-definition/fd1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	err := c.FieldDefinitionUpdate(context.Background(), models.AssetTrackingTemplateAsset, "fd1", models.CreateFieldDefinition{
		FieldType: models.FieldDefinitionFieldType{Name: models.FieldTypeText},
		Label:     "Updated",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFieldDefinitionGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.FieldDefinitionGet(context.Background(), models.AssetTrackingTemplateAsset, "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}

func TestFieldDefinitionNullableFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"uuid":"fd3",
			"field_key":"notes",
			"field_type":{"name":"TEXT","constraints":[]},
			"label":"Notes",
			"attributes":[],
			"comment":null,
			"default_value":null,
			"possible_values":null
		}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	fd, err := c.FieldDefinitionGet(context.Background(), models.AssetTrackingTemplateAsset, "fd3")
	if err != nil {
		t.Fatal(err)
	}
	if fd.Comment != nil {
		t.Errorf("expected nil comment, got %v", fd.Comment)
	}
	if fd.DefaultValue != nil {
		t.Errorf("expected nil default_value, got %v", fd.DefaultValue)
	}
	if fd.PossibleValues != nil {
		t.Errorf("expected nil possible_values, got %v", fd.PossibleValues)
	}
}
