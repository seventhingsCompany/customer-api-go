package models

import (
	"net/url"
	"strconv"
	"strings"
)

// SortDirection specifies ascending or descending sort order.
type SortDirection string

const (
	// SortASC sorts in ascending order.
	SortASC SortDirection = "ASC"
	// SortDESC sorts in descending order.
	SortDESC SortDirection = "DESC"
)

// FilterOperator specifies the comparison operator for a filter entry.
type FilterOperator string

const (
	// FilterEq matches values equal to the given value.
	FilterEq FilterOperator = "eq"
	// FilterNeq matches values not equal to the given value.
	FilterNeq FilterOperator = "neq"
	// FilterGt matches values greater than the given value.
	FilterGt FilterOperator = "gt"
	// FilterGtOrNull matches values greater than the given value or null.
	FilterGtOrNull FilterOperator = "gt_or_null"
	// FilterGte matches values greater than or equal to the given value.
	FilterGte FilterOperator = "gte"
	// FilterGteOrNull matches values greater than or equal to the given value or null.
	FilterGteOrNull FilterOperator = "gte_or_null"
	// FilterLt matches values less than the given value.
	FilterLt FilterOperator = "lt"
	// FilterLtOrNull matches values less than the given value or null.
	FilterLtOrNull FilterOperator = "lt_or_null"
	// FilterLte matches values less than or equal to the given value.
	FilterLte FilterOperator = "lte"
	// FilterLteOrNull matches values less than or equal to the given value or null.
	FilterLteOrNull FilterOperator = "lte_or_null"
	// FilterLike matches values containing the given substring.
	FilterLike FilterOperator = "like"
	// FilterNotLike matches values not containing the given substring.
	FilterNotLike FilterOperator = "not_like"
	// FilterIn matches values present in the given set.
	FilterIn FilterOperator = "in"
	// FilterNin matches values not present in the given set.
	FilterNin FilterOperator = "nin"
)

// FilterEntry represents a single filter condition.
type FilterEntry struct {
	Field    string
	Operator FilterOperator
	Values   []string
}

// Filter constructors build a FilterEntry for use with ListOptions.Where.

// Eq matches values equal to value.
func Eq(field, value string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterEq, Values: []string{value}}
}

// Neq matches values not equal to value.
func Neq(field, value string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterNeq, Values: []string{value}}
}

// Gt matches values greater than value.
func Gt(field, value string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterGt, Values: []string{value}}
}

// Gte matches values greater than or equal to value.
func Gte(field, value string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterGte, Values: []string{value}}
}

// Lt matches values less than value.
func Lt(field, value string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterLt, Values: []string{value}}
}

// Lte matches values less than or equal to value.
func Lte(field, value string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterLte, Values: []string{value}}
}

// Like matches values containing value.
func Like(field, value string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterLike, Values: []string{value}}
}

// NotLike matches values not containing value.
func NotLike(field, value string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterNotLike, Values: []string{value}}
}

// In matches values present in the given set.
func In(field string, values ...string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterIn, Values: values}
}

// Nin matches values not present in the given set.
func Nin(field string, values ...string) FilterEntry {
	return FilterEntry{Field: field, Operator: FilterNin, Values: values}
}

// ListOptions configures pagination, sorting, and filtering for list endpoints.
type ListOptions struct {
	Page    int
	PerPage int
	Sort    map[string]SortDirection
	Filters []FilterEntry
}

// NewListOptions returns an empty *ListOptions ready for fluent configuration.
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

// WithPage sets the page number and returns o for chaining.
func (o *ListOptions) WithPage(p int) *ListOptions {
	o.Page = p
	return o
}

// WithPerPage sets the page size and returns o for chaining.
func (o *ListOptions) WithPerPage(n int) *ListOptions {
	o.PerPage = n
	return o
}

// SortBy adds a sort on field in the given direction and returns o for chaining.
func (o *ListOptions) SortBy(field string, dir SortDirection) *ListOptions {
	if o.Sort == nil {
		o.Sort = map[string]SortDirection{}
	}
	o.Sort[field] = dir
	return o
}

// Where appends a filter condition and returns o for chaining. Build entries
// with the filter constructors (Eq, In, Like, ...).
func (o *ListOptions) Where(f FilterEntry) *ListOptions {
	o.Filters = append(o.Filters, f)
	return o
}

// isMultiValueOp returns true for operators that use array-style encoding.
func isMultiValueOp(op FilterOperator) bool {
	return op == FilterLike || op == FilterNotLike || op == FilterIn || op == FilterNin
}

// Encode builds a query string from the ListOptions. Brackets in parameter
// names are kept literal (not percent-encoded) to match the PHP deep-object
// format expected by the seventhings API. A nil receiver returns "".
func (o *ListOptions) Encode() string {
	if o == nil {
		return ""
	}

	var parts []string

	if o.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(o.Page))
	}
	if o.PerPage != 0 {
		parts = append(parts, "per_page="+strconv.Itoa(o.PerPage))
	}

	for field, dir := range o.Sort {
		parts = append(parts, "sort["+field+"]="+string(dir))
	}

	for _, f := range o.Filters {
		if isMultiValueOp(f.Operator) {
			for _, v := range f.Values {
				parts = append(parts, "filter["+f.Field+"]["+string(f.Operator)+"][]="+url.QueryEscape(v))
			}
		} else {
			val := ""
			if len(f.Values) > 0 {
				val = f.Values[0]
			}
			parts = append(parts, "filter["+f.Field+"]["+string(f.Operator)+"]="+url.QueryEscape(val))
		}
	}

	return strings.Join(parts, "&")
}
