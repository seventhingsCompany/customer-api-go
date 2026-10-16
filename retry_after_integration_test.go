//go:build integration

package seventhings_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// Verify real error headers without deliberately exceeding the live rate limit.
func TestIntegrationAPIErrorHeaders(t *testing.T) {
	c := integrationClient(t)
	uuid := fmt.Sprintf("00000000-0000-4000-8000-%012x", uint64(time.Now().UnixNano())&0xffffffffffff)
	_, err := c.ObjectGet(t.Context(), uuid)
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected 404 APIError for nonexistent object, got %v", err)
	}
	if len(apiErr.Header) == 0 {
		t.Fatal("live error response headers were not retained")
	}
	t.Logf("retained %d live response headers on HTTP %d", len(apiErr.Header), apiErr.StatusCode)
}
