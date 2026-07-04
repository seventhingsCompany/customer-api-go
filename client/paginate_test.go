package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// pagedObjectServer serves `pages` full pages of perPage items followed by a
// short final page, so the iterator terminates. It records how many requests
// it received so tests can assert early-stop behavior.
func pagedObjectServer(t *testing.T, perPage, fullPages, lastCount int, hits *int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*hits++
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		count := perPage
		if page > fullPages {
			count = lastCount
		}
		items := make([]string, 0, count)
		for i := 0; i < count; i++ {
			id := (page-1)*perPage + i
			items = append(items, fmt.Sprintf(`{"uuid":"obj-%d"}`, id))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"items":[%s]}`, join(items))
	}))
}

func join(items []string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += ","
		}
		out += s
	}
	return out
}

func TestObjectsAllWalksAllPages(t *testing.T) {
	hits := 0
	// 2 full pages of 3, then a final page of 2 → 8 items total, 3 requests.
	server := pagedObjectServer(t, 3, 2, 2, &hits)
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	var got []string
	for obj, err := range c.ObjectsAll(context.Background(), &models.ListOptions{PerPage: 3}) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got = append(got, obj.UUID())
	}
	if len(got) != 8 {
		t.Fatalf("expected 8 items, got %d (%v)", len(got), got)
	}
	if hits != 3 {
		t.Errorf("expected 3 requests, got %d", hits)
	}
	if got[0] != "obj-0" || got[7] != "obj-7" {
		t.Errorf("unexpected first/last: %q / %q", got[0], got[7])
	}
}

func TestObjectsAllStopsOnBreak(t *testing.T) {
	hits := 0
	server := pagedObjectServer(t, 3, 10, 0, &hits) // effectively unbounded full pages
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	count := 0
	for _, err := range c.ObjectsAll(context.Background(), &models.ListOptions{PerPage: 3}) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		if count == 4 { // stop mid-second-page
			break
		}
	}
	if count != 4 {
		t.Fatalf("expected to stop at 4, got %d", count)
	}
	// Only 2 pages should have been fetched (items 0-2, then 3-5).
	if hits != 2 {
		t.Errorf("expected 2 requests before break, got %d", hits)
	}
}

func TestObjectsAllPropagatesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	var gotErr error
	n := 0
	for _, err := range c.ObjectsAll(context.Background(), &models.ListOptions{PerPage: 3}) {
		n++
		if err != nil {
			gotErr = err
			break
		}
	}
	if gotErr == nil {
		t.Fatal("expected an error to be yielded")
	}
	var apiErr *models.APIError
	if !errors.As(gotErr, &apiErr) || !apiErr.IsServerError() {
		t.Errorf("expected a 5xx APIError, got %v", gotErr)
	}
}

func TestObjectsAllContextCancelled(t *testing.T) {
	server := pagedObjectServer(t, 3, 10, 0, new(int))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before iterating

	var gotErr error
	for _, err := range c.ObjectsAll(ctx, &models.ListOptions{PerPage: 3}) {
		if err != nil {
			gotErr = err
			break
		}
	}
	if !errors.Is(gotErr, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", gotErr)
	}
}

func TestPersonsAllWalksAllPages(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		count := 2
		if page > 2 {
			count = 1 // short final page
		}
		items := make([]string, 0, count)
		for i := 0; i < count; i++ {
			id := (page-1)*2 + i
			items = append(items, fmt.Sprintf(`{"person_uuid":"p-%d"}`, id))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"items":[%s]}`, join(items))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	perPage := 2
	var got []string
	for p, err := range c.PersonsAll(context.Background(), &models.PersonListOptions{PerPage: &perPage}) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got = append(got, p.UUID)
	}
	if len(got) != 5 { // 2 + 2 + 1
		t.Fatalf("expected 5 persons, got %d (%v)", len(got), got)
	}
	if hits != 3 {
		t.Errorf("expected 3 requests, got %d", hits)
	}
}
