package models

import "testing"

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
