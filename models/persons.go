package models

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

// Person represents a person in the seventhings asset-tracking system.
//
// Field tags follow the live API response, which uses snake_case and
// differs from the OpenAPI spec (the spec documents `uuid`/`firstname`/
// `lastname`, but the wire format is `person_uuid`/`first_name`/`last_name`).
type Person struct {
	UUID      string  `json:"person_uuid"`
	ID        int     `json:"id"`
	UserUUID  string  `json:"user_uuid"`
	Email     string  `json:"email"`
	Firstname *string `json:"first_name"`
	Lastname  *string `json:"last_name"`

	Department *string `json:"department"`
	// Picture and Documents are ATTACHMENT field values. They may be
	// returned as null, [], or arrays of file objects depending on
	// whether values are set; left untyped for now.
	Picture   any `json:"picture,omitempty"`
	Documents any `json:"documents,omitempty"`

	UpdatedByUserID *int    `json:"updated_by_user_id"`
	UpdatedAt       *string `json:"updated_at"`
	CreatedAt       *string `json:"created_at"`

	ImportedByUserID              *int    `json:"imported_by_user_id"`
	ImportedWithTemplateID        *int    `json:"imported_with_template_id"`
	ImportedAt                    *string `json:"imported_at"`
	CreatedOnImportWithTemplateID *int    `json:"created_on_import_with_template_id"`

	// Fields holds the complete, untouched field map as returned by the API,
	// including instance-defined custom fields that have no typed property
	// above. Person schemas are template-defined and vary per instance, so the
	// typed fields cover only the common columns; read custom values (and the
	// common ones) from Fields using its typed accessors. Populated on decode
	// by every SDK read (PersonGet/PersonGetByID/PersonsList).
	Fields Fields `json:"-"`
}

// UnmarshalJSON decodes a person, populating both the typed convenience fields
// and the full raw Fields map so instance-defined custom fields are never lost.
func (p *Person) UnmarshalJSON(data []byte) error {
	// personAlias avoids recursing into this method while decoding the typed
	// fields. The Fields tag is json:"-", so it is not touched here.
	type personAlias Person
	var typed personAlias
	if err := json.Unmarshal(data, &typed); err != nil {
		return err
	}
	*p = Person(typed)

	// Capture the complete payload verbatim, including unmapped custom fields.
	var raw Fields
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.Fields = raw
	return nil
}

// PersonListResponse is the paginated response for listing persons.
type PersonListResponse struct {
	Items   []Person `json:"items"`
	Page    int      `json:"page"`
	PerPage int      `json:"per_page"`
	SortBy  string   `json:"sort_by"`
	Order   string   `json:"order"`
	Total   int      `json:"total"`
}

// PersonListOptions configures pagination and sorting for person list requests.
// SortBy is a free-form field key (see field definitions); Order reuses UserSortOrder.
type PersonListOptions struct {
	Page    *int
	PerPage *int
	SortBy  *string
	Order   *UserSortOrder
}

// NewPersonListOptions returns an empty *PersonListOptions ready for fluent configuration.
func NewPersonListOptions() *PersonListOptions {
	return &PersonListOptions{}
}

// WithPage sets the page number and returns o for chaining.
func (o *PersonListOptions) WithPage(p int) *PersonListOptions {
	o.Page = &p
	return o
}

// WithPerPage sets the page size and returns o for chaining.
func (o *PersonListOptions) WithPerPage(n int) *PersonListOptions {
	o.PerPage = &n
	return o
}

// WithSort sets the sort field (a person field key) and direction and returns o
// for chaining.
func (o *PersonListOptions) WithSort(fieldKey string, order UserSortOrder) *PersonListOptions {
	o.SortBy = &fieldKey
	o.Order = &order
	return o
}

// Encode builds a query string from the PersonListOptions. A nil receiver returns "".
func (o *PersonListOptions) Encode() string {
	if o == nil {
		return ""
	}

	var parts []string

	if o.Page != nil {
		parts = append(parts, "page="+strconv.Itoa(*o.Page))
	}
	if o.PerPage != nil {
		parts = append(parts, "per_page="+strconv.Itoa(*o.PerPage))
	}
	if o.SortBy != nil {
		parts = append(parts, "sort_by="+url.QueryEscape(*o.SortBy))
	}
	if o.Order != nil {
		parts = append(parts, "order="+url.QueryEscape(string(*o.Order)))
	}

	return strings.Join(parts, "&")
}
