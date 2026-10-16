package models

import (
	"encoding/json"
	"testing"
)

func TestTimedRentalCase(t *testing.T) {
	fullDay := false
	issue, due := "2026-10-16T08:00:00+00:00", "2026-10-16T10:00:00+00:00"
	for _, request := range []any{
		CreateRentalCase{IsFullDay: &fullDay, IssueAt: &issue, DueAt: &due},
		UpdateRentalCase{IsFullDay: &fullDay, IssueAt: &issue, DueAt: &due},
	} {
		data, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		var got RentalCase
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if got.IsFullDay || got.IssueAt == nil || *got.IssueAt != issue || got.DueAt == nil || *got.DueAt != due {
			t.Fatalf("rental = %s", data)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatal(err)
		}
		if v, ok := raw["is_full_day"]; !ok || v != false {
			t.Fatalf("explicit false lost: %s", data)
		}
	}
	for _, request := range []any{CreateRentalCase{}, UpdateRentalCase{}} {
		data, _ := json.Marshal(request)
		var raw map[string]any
		json.Unmarshal(data, &raw)
		for _, key := range []string{"is_full_day", "issue_at", "due_at"} {
			if _, ok := raw[key]; ok {
				t.Errorf("optional %s must be omitted: %s", key, data)
			}
		}
	}
}

func TestTaskOptionalReminders(t *testing.T) {
	for _, reminders := range [][]TimeInterval{nil, {}, {{Unit: TimeIntervalDays, Value: 0}}} {
		for _, request := range []any{CreateTask{Reminders: reminders}, UpdateTask{Reminders: reminders}} {
			data, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatal(err)
			}
			value, exists := raw["reminders"]
			if reminders == nil {
				if exists {
					t.Errorf("unspecified reminders emitted: %s", data)
				}
			} else {
				want, _ := json.Marshal(reminders)
				if string(value) != string(want) {
					t.Errorf("reminders = %s, want %s", value, want)
				}
			}
		}
	}
}

func TestCatalogEmptySorting(t *testing.T) {
	for _, raw := range []string{`{"items":[],"sort":[]}`, `{"items":[],"sort":{}}`, `{"items":[],"sort":null}`, `{"items":[]}`} {
		var page CatalogItemListResponse
		if err := json.Unmarshal([]byte(raw), &page); err != nil {
			t.Errorf("decode %s: %v", raw, err)
		}
		if len(page.Sort) != 0 {
			t.Errorf("unexpected sorting: %v", page.Sort)
		}
	}
	for _, raw := range []string{`{"sort":["id"]}`, `{"sort":42}`} {
		var page CatalogItemListResponse
		if err := json.Unmarshal([]byte(raw), &page); err == nil {
			t.Errorf("expected error for %s", raw)
		}
	}
}

func TestRentalRequestsEncodeRenter(t *testing.T) {
	renter := &RentalCaseRenter{Type: RenterTypePlain, Value: "SDK test"}
	for _, request := range []any{CreateRentalCase{Renter: renter}, UpdateRentalCase{Renter: renter}} {
		data, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatal(err)
		}
		var got RentalCaseRenter
		if err := json.Unmarshal(raw["renter"], &got); err != nil {
			t.Fatal(err)
		}
		if got != *renter {
			t.Errorf("renter = %+v, want %+v", got, *renter)
		}
	}
}
