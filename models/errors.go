// Package models provides data structures for the seventhings API client.
package models

import (
	"errors"
	"fmt"
	"net/http"
)

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

// IsNotFound reports whether the error is a 404 Not Found.
func (e *APIError) IsNotFound() bool { return e.StatusCode == http.StatusNotFound }

// IsUnauthorized reports whether the error is a 401 Unauthorized.
func (e *APIError) IsUnauthorized() bool { return e.StatusCode == http.StatusUnauthorized }

// IsForbidden reports whether the error is a 403 Forbidden.
func (e *APIError) IsForbidden() bool { return e.StatusCode == http.StatusForbidden }

// IsConflict reports whether the error is a 409 Conflict.
func (e *APIError) IsConflict() bool { return e.StatusCode == http.StatusConflict }

// IsRateLimited reports whether the error is a 429 Too Many Requests.
func (e *APIError) IsRateLimited() bool { return e.StatusCode == http.StatusTooManyRequests }

// IsServerError reports whether the error is a 5xx server error.
func (e *APIError) IsServerError() bool { return e.StatusCode >= 500 }

// asAPIError unwraps err to an *APIError, returning nil if err is not one.
func asAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}

// IsNotFound reports whether err wraps an *APIError with a 404 status.
func IsNotFound(err error) bool {
	e := asAPIError(err)
	return e != nil && e.IsNotFound()
}

// IsUnauthorized reports whether err wraps an *APIError with a 401 status.
func IsUnauthorized(err error) bool {
	e := asAPIError(err)
	return e != nil && e.IsUnauthorized()
}

// IsForbidden reports whether err wraps an *APIError with a 403 status.
func IsForbidden(err error) bool {
	e := asAPIError(err)
	return e != nil && e.IsForbidden()
}

// IsConflict reports whether err wraps an *APIError with a 409 status.
func IsConflict(err error) bool {
	e := asAPIError(err)
	return e != nil && e.IsConflict()
}

// IsRateLimited reports whether err wraps an *APIError with a 429 status.
func IsRateLimited(err error) bool {
	e := asAPIError(err)
	return e != nil && e.IsRateLimited()
}

// IsServerError reports whether err wraps an *APIError with a 5xx status.
func IsServerError(err error) bool {
	e := asAPIError(err)
	return e != nil && e.IsServerError()
}
