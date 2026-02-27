package models

import (
	"net/url"
	"strings"
)

// TimeIntervalUnit specifies the unit for a time interval.
type TimeIntervalUnit string

const (
	// TimeIntervalDays is a day-based interval.
	TimeIntervalDays TimeIntervalUnit = "days"
	// TimeIntervalWeeks is a week-based interval.
	TimeIntervalWeeks TimeIntervalUnit = "weeks"
	// TimeIntervalMonths is a month-based interval.
	TimeIntervalMonths TimeIntervalUnit = "months"
	// TimeIntervalYears is a year-based interval.
	TimeIntervalYears TimeIntervalUnit = "years"
)

// TimeInterval represents a recurring schedule interval.
type TimeInterval struct {
	Unit  TimeIntervalUnit `json:"unit"`
	Value int              `json:"value"`
}

// AttachmentFile represents a file attachment on a task or rental case.
type AttachmentFile struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         int    `json:"size"`
	DataURI      string `json:"data_uri"`
	ThumbnailURI string `json:"thumbnail_uri"`
}

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	// TaskStatusOpen indicates an open task.
	TaskStatusOpen TaskStatus = "open"
	// TaskStatusClosed indicates a closed task.
	TaskStatusClosed TaskStatus = "closed"
)

// TaskReferenceType is the type of entity a task references.
type TaskReferenceType string

const (
	// TaskReferenceTypeAsset references an asset.
	TaskReferenceTypeAsset TaskReferenceType = "asset"
)

// TaskReferenceStatus is the status of a task reference.
type TaskReferenceStatus string

const (
	// TaskReferenceStatusOpen indicates an open reference.
	TaskReferenceStatusOpen TaskReferenceStatus = "open"
	// TaskReferenceStatusDone indicates a completed reference.
	TaskReferenceStatusDone TaskReferenceStatus = "done"
)

// TaskReference is a reference to an entity in a task response.
type TaskReference struct {
	Type   TaskReferenceType   `json:"type"`
	UUID   string              `json:"uuid"`
	Name   string              `json:"name"`
	ID     int                 `json:"id"`
	Status TaskReferenceStatus `json:"status"`
}

// TaskReferenceInput is a reference to an entity in a task request.
type TaskReferenceInput struct {
	Type TaskReferenceType `json:"type"`
	UUID string            `json:"uuid"`
}

// Task represents a task in the seventhings API.
type Task struct {
	UUID              string           `json:"uuid"`
	Title             string           `json:"title"`
	Status            TaskStatus       `json:"status"`
	Deadline          *string          `json:"deadline"`
	Assignees         []string         `json:"assignees"`
	Author            string           `json:"author"`
	References        []TaskReference  `json:"references"`
	Reminders         []TimeInterval   `json:"reminders"`
	RecurringSchedule *TimeInterval    `json:"recurring_schedule"`
	Comment           *string          `json:"comment"`
	Attachments       []AttachmentFile `json:"attachments"`
	CreatedAt         string           `json:"created_at"`
	UpdatedAt         string           `json:"updated_at"`
}

// CreateTask is the request body for creating a task.
type CreateTask struct {
	Title             string               `json:"title"`
	Deadline          *string              `json:"deadline"`
	Assignees         []string             `json:"assignees"`
	References        []TaskReferenceInput `json:"references"`
	Reminders         []TimeInterval       `json:"reminders"`
	RecurringSchedule *TimeInterval        `json:"recurring_schedule"`
	Comment           *string              `json:"comment,omitempty"`
	Attachments       []string             `json:"attachments,omitempty"`
	Notify            *bool                `json:"notify,omitempty"`
}

// UpdateTask is the request body for updating a task.
type UpdateTask struct {
	Title             string               `json:"title"`
	Deadline          *string              `json:"deadline"`
	Assignees         []string             `json:"assignees"`
	References        []TaskReferenceInput `json:"references"`
	Reminders         []TimeInterval       `json:"reminders"`
	RecurringSchedule *TimeInterval        `json:"recurring_schedule"`
	Comment           *string              `json:"comment,omitempty"`
	Attachments       []string             `json:"attachments,omitempty"`
	Notify            *bool                `json:"notify,omitempty"`
}

// TaskStatusUpdate is the request body for updating a task's status.
type TaskStatusUpdate struct {
	Status TaskStatus `json:"status"`
}

// TaskListOptions configures filtering for task list requests.
type TaskListOptions struct {
	Status        *TaskStatus
	DeadlineFrom  *string
	DeadlineTo    *string
	Assignee      *string
	Author        *string
	ReferenceType *TaskReferenceType
}

// Encode builds a query string from the TaskListOptions. A nil receiver returns "".
func (o *TaskListOptions) Encode() string {
	if o == nil {
		return ""
	}

	var parts []string

	if o.Status != nil {
		parts = append(parts, "status="+url.QueryEscape(string(*o.Status)))
	}
	if o.DeadlineFrom != nil {
		parts = append(parts, "deadline_from="+url.QueryEscape(*o.DeadlineFrom))
	}
	if o.DeadlineTo != nil {
		parts = append(parts, "deadline_to="+url.QueryEscape(*o.DeadlineTo))
	}
	if o.Assignee != nil {
		parts = append(parts, "assignee="+url.QueryEscape(*o.Assignee))
	}
	if o.Author != nil {
		parts = append(parts, "author="+url.QueryEscape(*o.Author))
	}
	if o.ReferenceType != nil {
		parts = append(parts, "reference_type="+url.QueryEscape(string(*o.ReferenceType)))
	}

	return strings.Join(parts, "&")
}
