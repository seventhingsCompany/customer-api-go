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
