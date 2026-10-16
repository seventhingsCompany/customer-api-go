package models

import (
	"bytes"
	"encoding/json"
)

// CatalogItem contains reusable object master data. Object field keys have a
// catalog_object_ prefix; catalog metadata such as id has no prefix.
type CatalogItem struct {
	UUID   string `json:"uuid"`
	Fields Fields `json:"fields"`
}

// CatalogItemListResponse is a paginated product catalog response.
type CatalogItemListResponse struct {
	Items   []CatalogItem            `json:"items"`
	Page    int                      `json:"page"`
	PerPage int                      `json:"per_page"`
	Total   int                      `json:"total"`
	Sort    map[string]SortDirection `json:"sort"`
}

// UnmarshalJSON accepts the empty array emitted by the API for an empty sort
// map, while preserving the documented object shape for non-empty sorting.
func (r *CatalogItemListResponse) UnmarshalJSON(data []byte) error {
	type alias CatalogItemListResponse
	var result alias
	wire := struct {
		*alias
		Sort json.RawMessage `json:"sort"`
	}{alias: &result}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if len(wire.Sort) > 0 && !bytes.Equal(bytes.TrimSpace(wire.Sort), []byte("[]")) {
		if err := json.Unmarshal(wire.Sort, &result.Sort); err != nil {
			return err
		}
	}
	*r = CatalogItemListResponse(result)
	return nil
}
