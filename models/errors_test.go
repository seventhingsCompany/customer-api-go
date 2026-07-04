package models

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIErrorError(t *testing.T) {
	e := &APIError{
		StatusCode: 404,
		Status:     "404 Not Found",
		Body:       `{"message":"not found"}`,
	}
	got := e.Error()
	expected := `seventhings API error 404 (404 Not Found): {"message":"not found"}`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestAPIErrorIsStatusCodeMatch(t *testing.T) {
	e := &APIError{StatusCode: 403}
	if !e.IsStatusCode(403) {
		t.Error("expected IsStatusCode(403) to return true")
	}
}

func TestAPIErrorIsStatusCodeNoMatch(t *testing.T) {
	e := &APIError{StatusCode: 403}
	if e.IsStatusCode(404) {
		t.Error("expected IsStatusCode(404) to return false")
	}
}

func TestAPIErrorPredicates(t *testing.T) {
	cases := []struct {
		code int
		pred func(*APIError) bool
		name string
	}{
		{404, (*APIError).IsNotFound, "IsNotFound"},
		{401, (*APIError).IsUnauthorized, "IsUnauthorized"},
		{403, (*APIError).IsForbidden, "IsForbidden"},
		{409, (*APIError).IsConflict, "IsConflict"},
		{429, (*APIError).IsRateLimited, "IsRateLimited"},
		{500, (*APIError).IsServerError, "IsServerError"},
		{503, (*APIError).IsServerError, "IsServerError-503"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !c.pred(&APIError{StatusCode: c.code}) {
				t.Errorf("%s should be true for %d", c.name, c.code)
			}
			if c.pred(&APIError{StatusCode: 200}) {
				t.Errorf("%s should be false for 200", c.name)
			}
		})
	}
}

func TestPackageLevelErrorHelpers(t *testing.T) {
	wrapped := fmt.Errorf("request failed: %w", &APIError{StatusCode: 404})
	if !IsNotFound(wrapped) {
		t.Error("IsNotFound should unwrap a wrapped *APIError")
	}
	if IsUnauthorized(wrapped) {
		t.Error("IsUnauthorized should be false for a 404")
	}
	if IsNotFound(errors.New("plain error")) {
		t.Error("IsNotFound should be false for a non-APIError")
	}
}
