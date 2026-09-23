package models

import (
	"net/url"
	"strconv"
)

// HistoryListOptions configures history pagination. Zero values use the API
// defaults (page 1, per_page 50). The API allows at most 200 entries per page.
type HistoryListOptions struct {
	Page    int
	PerPage int
}

// Encode builds a query string. A nil receiver returns an empty string.
func (o *HistoryListOptions) Encode() string {
	if o == nil {
		return ""
	}
	q := url.Values{}
	if o.Page != 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	if o.PerPage != 0 {
		q.Set("per_page", strconv.Itoa(o.PerPage))
	}
	return q.Encode()
}

// HistoryResponse is a page of recorded changes, newest first.
type HistoryResponse[T any] struct {
	Items   []T `json:"items"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// ObjectHistoryEntry holds a dynamic history event. The type field is asset,
// task, rental_case, or object_merge. Merge events carry user_id and
// absorbedObjectData instead of properties. A map preserves all event fields.
type ObjectHistoryEntry = map[string]any

// RoomHistoryEntry is a recorded change of a room.
type RoomHistoryEntry struct {
	RoomUUID    string `json:"room_uuid"`
	UserUUID    string `json:"user_uuid"`
	OccurredAt  string `json:"occurred_at"`
	EventName   string `json:"event_name"`
	Description string `json:"description"`
	Details     string `json:"details"` // JSON-encoded snapshot, or an empty string.
}

// LocationHistoryEntry is a recorded change of a location.
type LocationHistoryEntry struct {
	LocationUUID string `json:"location_uuid"`
	UserUUID     string `json:"user_uuid"`
	OccurredAt   string `json:"occurred_at"`
	EventName    string `json:"event_name"`
	Description  string `json:"description"`
	Details      string `json:"details"` // JSON-encoded snapshot, or an empty string.
}

// PersonHistoryEntry is a recorded change of a person.
type PersonHistoryEntry struct {
	PersonUUID  string `json:"person_uuid"`
	UserUUID    string `json:"user_uuid"`
	OccurredAt  string `json:"occurred_at"`
	EventName   string `json:"event_name"`
	Description string `json:"description"`
	Details     string `json:"details"` // JSON-encoded snapshot, or an empty string.
}

// TaskHistoryEntry is a recorded change of a task.
type TaskHistoryEntry struct {
	TaskUUID    string `json:"task_uuid"`
	UserUUID    string `json:"user_uuid"`
	OccurredAt  string `json:"occurred_at"`
	EventName   string `json:"event_name"`
	Description string `json:"description"`
	Details     string `json:"details"` // JSON-encoded snapshot, or an empty string.
}

// RentalCaseHistoryEntry is a recorded change of a rental case.
type RentalCaseHistoryEntry struct {
	RentalCaseUUID string `json:"rental_case_uuid"`
	UserUUID       string `json:"user_uuid"`
	OccurredAt     string `json:"occurred_at"`
	EventName      string `json:"event_name"`
	Description    string `json:"description"`
	Details        string `json:"details"` // JSON-encoded snapshot, or an empty string.
}
