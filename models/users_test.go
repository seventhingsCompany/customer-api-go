package models

import "testing"

func TestUserListOptionsEncodeNilReceiver(t *testing.T) {
	var o *UserListOptions
	if got := o.Encode(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestUserListOptionsEncodeEmpty(t *testing.T) {
	o := &UserListOptions{}
	if got := o.Encode(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestUserListOptionsEncodePage(t *testing.T) {
	p := 3
	o := &UserListOptions{Page: &p}
	if got := o.Encode(); got != "page=3" {
		t.Errorf("expected page=3, got %q", got)
	}
}

func TestUserListOptionsEncodePerPage(t *testing.T) {
	pp := 50
	o := &UserListOptions{PerPage: &pp}
	if got := o.Encode(); got != "per_page=50" {
		t.Errorf("expected per_page=50, got %q", got)
	}
}

func TestUserListOptionsEncodeSortBy(t *testing.T) {
	s := UserSortByEmail
	o := &UserListOptions{SortBy: &s}
	if got := o.Encode(); got != "sort_by=email" {
		t.Errorf("expected sort_by=email, got %q", got)
	}
}

func TestUserListOptionsEncodeOrder(t *testing.T) {
	ord := UserSortOrderDesc
	o := &UserListOptions{Order: &ord}
	if got := o.Encode(); got != "order=desc" {
		t.Errorf("expected order=desc, got %q", got)
	}
}

func TestUserListOptionsEncodeAllFields(t *testing.T) {
	p := 2
	pp := 25
	s := UserSortByID
	ord := UserSortOrderAsc
	o := &UserListOptions{
		Page:    &p,
		PerPage: &pp,
		SortBy:  &s,
		Order:   &ord,
	}
	got := o.Encode()
	expected := "page=2&per_page=25&sort_by=id&order=asc"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
