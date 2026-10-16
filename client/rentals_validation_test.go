package client

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestRentalCaseRejectsMissingRenterBeforeRequest(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Location", "/customer-api/v1/rental-management/rental-case/unexpected")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	c := newTestClient(t, server)
	if uuid, err := c.RentalCaseCreate(t.Context(), models.CreateRentalCase{}); !errors.Is(err, ErrRentalCaseRenterRequired) || uuid != "" {
		t.Errorf("create = %q, %v; want missing-renter error", uuid, err)
	}
	if err := c.RentalCaseUpdate(t.Context(), "existing", models.UpdateRentalCase{}); !errors.Is(err, ErrRentalCaseRenterRequired) {
		t.Errorf("update error = %v; want missing-renter error", err)
	}
	if got := requests.Load(); got != 0 {
		t.Errorf("sent %d HTTP requests for invalid inputs", got)
	}
}
