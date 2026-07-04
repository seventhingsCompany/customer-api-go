package models

import (
	"net/url"
	"strconv"
	"strings"
)

// UserSortBy specifies the field to sort users by.
type UserSortBy string

const (
	// UserSortByID sorts users by ID.
	UserSortByID UserSortBy = "id"
	// UserSortByEmail sorts users by email.
	UserSortByEmail UserSortBy = "email"
)

// UserSortOrder specifies the sort direction for user listings.
type UserSortOrder string

const (
	// UserSortOrderAsc sorts in ascending order.
	UserSortOrderAsc UserSortOrder = "asc"
	// UserSortOrderDesc sorts in descending order.
	UserSortOrderDesc UserSortOrder = "desc"
)

// User represents a user in the seventhings API.
type User struct {
	UUID        string  `json:"uuid"`
	ID          int     `json:"id"`
	Email       string  `json:"email"`
	Firstname   *string `json:"firstname"`
	Lastname    *string `json:"lastname"`
	DisplayName *string `json:"display_name"`
}

// UserListResponse is the paginated response for listing users.
type UserListResponse struct {
	Items   []User `json:"items"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	SortBy  string `json:"sort_by"`
	Order   string `json:"order"`
	Total   int    `json:"total"`
}

// UserListOptions configures pagination and sorting for user list requests.
type UserListOptions struct {
	Page    *int
	PerPage *int
	SortBy  *UserSortBy
	Order   *UserSortOrder
}

// NewUserListOptions returns an empty *UserListOptions ready for fluent configuration.
func NewUserListOptions() *UserListOptions {
	return &UserListOptions{}
}

// WithPage sets the page number and returns o for chaining.
func (o *UserListOptions) WithPage(p int) *UserListOptions {
	o.Page = &p
	return o
}

// WithPerPage sets the page size and returns o for chaining.
func (o *UserListOptions) WithPerPage(n int) *UserListOptions {
	o.PerPage = &n
	return o
}

// WithSort sets the sort field and direction and returns o for chaining.
func (o *UserListOptions) WithSort(by UserSortBy, order UserSortOrder) *UserListOptions {
	o.SortBy = &by
	o.Order = &order
	return o
}

// Encode builds a query string from the UserListOptions. A nil receiver returns "".
func (o *UserListOptions) Encode() string {
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
		parts = append(parts, "sort_by="+url.QueryEscape(string(*o.SortBy)))
	}
	if o.Order != nil {
		parts = append(parts, "order="+url.QueryEscape(string(*o.Order)))
	}

	return strings.Join(parts, "&")
}
