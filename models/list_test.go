package models

import (
	"testing"
)

func TestEncodeNilReceiver(t *testing.T) {
	var o *ListOptions
	if got := o.Encode(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestEncodeEmpty(t *testing.T) {
	o := &ListOptions{}
	if got := o.Encode(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestEncodePageOnly(t *testing.T) {
	o := &ListOptions{Page: 2}
	if got := o.Encode(); got != "page=2" {
		t.Errorf("expected page=2, got %q", got)
	}
}

func TestEncodePerPageOnly(t *testing.T) {
	o := &ListOptions{PerPage: 50}
	if got := o.Encode(); got != "per_page=50" {
		t.Errorf("expected per_page=50, got %q", got)
	}
}

func TestEncodePagination(t *testing.T) {
	o := &ListOptions{Page: 3, PerPage: 25}
	got := o.Encode()
	if got != "page=3&per_page=25" {
		t.Errorf("expected page=3&per_page=25, got %q", got)
	}
}

func TestEncodeSort(t *testing.T) {
	o := &ListOptions{
		Sort: map[string]SortDirection{"name": SortASC},
	}
	got := o.Encode()
	if got != "sort[name]=ASC" {
		t.Errorf("expected sort[name]=ASC, got %q", got)
	}
}

func TestEncodeSingleValueFilter(t *testing.T) {
	o := &ListOptions{
		Filters: []FilterEntry{
			{Field: "status", Operator: FilterEq, Values: []string{"active"}},
		},
	}
	got := o.Encode()
	if got != "filter[status][eq]=active" {
		t.Errorf("expected filter[status][eq]=active, got %q", got)
	}
}

func TestEncodeMultiValueFilter(t *testing.T) {
	o := &ListOptions{
		Filters: []FilterEntry{
			{Field: "tag", Operator: FilterIn, Values: []string{"a", "b"}},
		},
	}
	got := o.Encode()
	expected := "filter[tag][in][]=a&filter[tag][in][]=b"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEncodeLikeFilter(t *testing.T) {
	o := &ListOptions{
		Filters: []FilterEntry{
			{Field: "name", Operator: FilterLike, Values: []string{"foo"}},
		},
	}
	got := o.Encode()
	expected := "filter[name][like][]=foo"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEncodeSpecialChars(t *testing.T) {
	o := &ListOptions{
		Filters: []FilterEntry{
			{Field: "name", Operator: FilterEq, Values: []string{"hello world&more=yes"}},
		},
	}
	got := o.Encode()
	expected := "filter[name][eq]=hello+world%26more%3Dyes"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEncodeRangeFilter(t *testing.T) {
	o := &ListOptions{
		Filters: []FilterEntry{
			{Field: "price", Operator: FilterGt, Values: []string{"10"}},
			{Field: "price", Operator: FilterLt, Values: []string{"100"}},
		},
	}
	got := o.Encode()
	expected := "filter[price][gt]=10&filter[price][lt]=100"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEncodeCombined(t *testing.T) {
	o := &ListOptions{
		Page:    1,
		PerPage: 10,
		Sort:    map[string]SortDirection{"created_at": SortDESC},
		Filters: []FilterEntry{
			{Field: "status", Operator: FilterEq, Values: []string{"active"}},
			{Field: "tag", Operator: FilterIn, Values: []string{"x", "y"}},
		},
	}
	got := o.Encode()
	expected := "page=1&per_page=10&sort[created_at]=DESC&filter[status][eq]=active&filter[tag][in][]=x&filter[tag][in][]=y"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEncodeNinFilter(t *testing.T) {
	o := &ListOptions{
		Filters: []FilterEntry{
			{Field: "color", Operator: FilterNin, Values: []string{"red", "blue"}},
		},
	}
	got := o.Encode()
	expected := "filter[color][nin][]=red&filter[color][nin][]=blue"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEncodeOrNullOperators(t *testing.T) {
	o := &ListOptions{
		Filters: []FilterEntry{
			{Field: "date", Operator: FilterGtOrNull, Values: []string{"2024-01-01"}},
			{Field: "date", Operator: FilterLteOrNull, Values: []string{"2024-12-31"}},
		},
	}
	got := o.Encode()
	expected := "filter[date][gt_or_null]=2024-01-01&filter[date][lte_or_null]=2024-12-31"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestListOptionsBuilder(t *testing.T) {
	o := NewListOptions().
		WithPage(1).
		WithPerPage(10).
		SortBy("created_at", SortDESC).
		Where(Eq("status", "active")).
		Where(In("tag", "x", "y"))

	got := o.Encode()
	expected := "page=1&per_page=10&sort[created_at]=DESC&filter[status][eq]=active&filter[tag][in][]=x&filter[tag][in][]=y"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFilterConstructors(t *testing.T) {
	tests := []struct {
		name  string
		entry FilterEntry
		op    FilterOperator
		vals  []string
	}{
		{"Eq", Eq("f", "v"), FilterEq, []string{"v"}},
		{"Neq", Neq("f", "v"), FilterNeq, []string{"v"}},
		{"Gt", Gt("f", "v"), FilterGt, []string{"v"}},
		{"Gte", Gte("f", "v"), FilterGte, []string{"v"}},
		{"Lt", Lt("f", "v"), FilterLt, []string{"v"}},
		{"Lte", Lte("f", "v"), FilterLte, []string{"v"}},
		{"Like", Like("f", "v"), FilterLike, []string{"v"}},
		{"NotLike", NotLike("f", "v"), FilterNotLike, []string{"v"}},
		{"In", In("f", "a", "b"), FilterIn, []string{"a", "b"}},
		{"Nin", Nin("f", "a", "b"), FilterNin, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.entry.Field != "f" {
				t.Errorf("field = %q", tt.entry.Field)
			}
			if tt.entry.Operator != tt.op {
				t.Errorf("operator = %q, want %q", tt.entry.Operator, tt.op)
			}
			if len(tt.entry.Values) != len(tt.vals) {
				t.Fatalf("values = %v, want %v", tt.entry.Values, tt.vals)
			}
			for i := range tt.vals {
				if tt.entry.Values[i] != tt.vals[i] {
					t.Errorf("values[%d] = %q, want %q", i, tt.entry.Values[i], tt.vals[i])
				}
			}
		})
	}
}

func TestUserListOptionsBuilder(t *testing.T) {
	o := NewUserListOptions().WithPage(2).WithPerPage(5).WithSort(UserSortByEmail, UserSortOrderAsc)
	got := o.Encode()
	expected := "page=2&per_page=5&sort_by=email&order=asc"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPersonListOptionsBuilder(t *testing.T) {
	o := NewPersonListOptions().WithPage(1).WithPerPage(20).WithSort("last_name", UserSortOrderDesc)
	got := o.Encode()
	expected := "page=1&per_page=20&sort_by=last_name&order=desc"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
