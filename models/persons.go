package models

import (
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
