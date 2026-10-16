package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestVerifyRentalTimes(t *testing.T) {
	issue := time.Date(2099, 1, 1, 10, 0, 0, 0, time.FixedZone("offset", 7200))
	due := issue.Add(2 * time.Hour)
	issueUTC, dueUTC := issue.UTC().Format(time.RFC3339), due.UTC().Format(time.RFC3339)
	bad := "invalid"
	for _, tc := range []struct {
		name      string
		rental    models.RentalCase
		wantError bool
	}{
		{"normalized UTC", models.RentalCase{IssueAt: &issueUTC, DueAt: &dueUTC}, false},
		{"missing timestamps", models.RentalCase{}, true},
		{"full day", models.RentalCase{IsFullDay: true, IssueAt: &issueUTC, DueAt: &dueUTC}, true},
		{"invalid timestamp", models.RentalCase{IssueAt: &bad, DueAt: &dueUTC}, true},
		{"wrong return time", models.RentalCase{IssueAt: &issueUTC, DueAt: &issueUTC}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := verifyRentalTimes(&tc.rental, issue, due); (err != nil) != tc.wantError {
				t.Fatalf("error = %v, want error = %v", err, tc.wantError)
			}
		})
	}
}

func TestCleanupDemoSurvivesCancellationAndPreservesErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	original, cleanupErr := errors.New("original"), errors.New("cleanup")
	result := original
	cleanupDemo(ctx, &result, func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			t.Errorf("cleanup context canceled: %v", err)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Error("cleanup context has no deadline")
		}
		return cleanupErr
	})
	if !errors.Is(result, original) || !errors.Is(result, cleanupErr) {
		t.Fatalf("errors lost: %v", result)
	}
}
