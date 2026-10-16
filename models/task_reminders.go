package models

import "encoding/json"

// MarshalJSON omits unspecified reminders but sends an explicit empty list to
// disable reminders. Zero-valued intervals still mean "on the deadline".
func (r CreateTask) MarshalJSON() ([]byte, error) {
	type alias CreateTask
	var reminders *[]TimeInterval
	if r.Reminders != nil {
		reminders = &r.Reminders
	}
	return json.Marshal(struct {
		alias
		Reminders *[]TimeInterval `json:"reminders,omitempty"`
	}{alias(r), reminders})
}

// MarshalJSON distinguishes unspecified reminders from an explicit empty list.
func (r UpdateTask) MarshalJSON() ([]byte, error) {
	type alias UpdateTask
	var reminders *[]TimeInterval
	if r.Reminders != nil {
		reminders = &r.Reminders
	}
	return json.Marshal(struct {
		alias
		Reminders *[]TimeInterval `json:"reminders,omitempty"`
	}{alias(r), reminders})
}
