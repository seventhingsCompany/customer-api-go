package models

import "testing"

func TestTaskListOptionsEncodeNilReceiver(t *testing.T) {
	var o *TaskListOptions
	if got := o.Encode(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestTaskListOptionsEncodeEmpty(t *testing.T) {
	o := &TaskListOptions{}
	if got := o.Encode(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestTaskListOptionsEncodeStatus(t *testing.T) {
	s := TaskStatusOpen
	o := &TaskListOptions{Status: &s}
	if got := o.Encode(); got != "status=open" {
		t.Errorf("expected status=open, got %q", got)
	}
}

func TestTaskListOptionsEncodeDeadlineFrom(t *testing.T) {
	d := "2024-01-01"
	o := &TaskListOptions{DeadlineFrom: &d}
	if got := o.Encode(); got != "deadline_from=2024-01-01" {
		t.Errorf("expected deadline_from=2024-01-01, got %q", got)
	}
}

func TestTaskListOptionsEncodeDeadlineTo(t *testing.T) {
	d := "2024-12-31"
	o := &TaskListOptions{DeadlineTo: &d}
	if got := o.Encode(); got != "deadline_to=2024-12-31" {
		t.Errorf("expected deadline_to=2024-12-31, got %q", got)
	}
}

func TestTaskListOptionsEncodeAssignee(t *testing.T) {
	a := "user@example.com"
	o := &TaskListOptions{Assignee: &a}
	if got := o.Encode(); got != "assignee=user%40example.com" {
		t.Errorf("expected assignee=user%%40example.com, got %q", got)
	}
}

func TestTaskListOptionsEncodeAuthor(t *testing.T) {
	a := "author@example.com"
	o := &TaskListOptions{Author: &a}
	if got := o.Encode(); got != "author=author%40example.com" {
		t.Errorf("expected author=author%%40example.com, got %q", got)
	}
}

func TestTaskListOptionsEncodeReferenceType(t *testing.T) {
	r := TaskReferenceTypeAsset
	o := &TaskListOptions{ReferenceType: &r}
	if got := o.Encode(); got != "reference_type=asset" {
		t.Errorf("expected reference_type=asset, got %q", got)
	}
}

func TestTaskListOptionsEncodeAllFields(t *testing.T) {
	s := TaskStatusClosed
	df := "2024-01-01"
	dt := "2024-12-31"
	a := "user@example.com"
	au := "admin@example.com"
	r := TaskReferenceTypeAsset
	o := &TaskListOptions{
		Status:        &s,
		DeadlineFrom:  &df,
		DeadlineTo:    &dt,
		Assignee:      &a,
		Author:        &au,
		ReferenceType: &r,
	}
	got := o.Encode()
	expected := "status=closed&deadline_from=2024-01-01&deadline_to=2024-12-31&assignee=user%40example.com&author=admin%40example.com&reference_type=asset"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestTaskListOptionsEncodeSpecialChars(t *testing.T) {
	a := "user name&more=yes"
	o := &TaskListOptions{Assignee: &a}
	got := o.Encode()
	expected := "assignee=user+name%26more%3Dyes"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
