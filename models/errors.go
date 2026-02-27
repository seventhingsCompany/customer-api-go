// Package models provides data structures for the seventhings API client.
package models

import "fmt"

// APIError represents an error response from the seventhings API.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("seventhings API error %d (%s): %s", e.StatusCode, e.Status, e.Body)
}

// IsStatusCode returns true if the error has the given HTTP status code.
func (e *APIError) IsStatusCode(code int) bool {
	return e.StatusCode == code
}
