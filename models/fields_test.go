package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFieldsString(t *testing.T) {
	f := Fields{"name": "widget", "count": 3.0}
	if v, ok := f.String("name"); !ok || v != "widget" {
		t.Errorf("String(name) = %q, %v", v, ok)
	}
	if _, ok := f.String("count"); ok {
		t.Error("String(count) should be false for a number")
	}
	if _, ok := f.String("missing"); ok {
		t.Error("String(missing) should be false")
	}
}

func TestFieldsInt(t *testing.T) {
	f := Fields{
		"float_int":  42.0,
		"float_frac": 42.5,
		"json_num":   json.Number("7"),
		"native_int": 9,
		"str":        "x",
	}
	if v, ok := f.Int("float_int"); !ok || v != 42 {
		t.Errorf("Int(float_int) = %d, %v", v, ok)
	}
	if _, ok := f.Int("float_frac"); ok {
		t.Error("Int(float_frac) should be false for non-integral float")
	}
	if v, ok := f.Int("json_num"); !ok || v != 7 {
		t.Errorf("Int(json_num) = %d, %v", v, ok)
	}
	if v, ok := f.Int("native_int"); !ok || v != 9 {
		t.Errorf("Int(native_int) = %d, %v", v, ok)
	}
	if _, ok := f.Int("str"); ok {
		t.Error("Int(str) should be false")
	}
}

func TestFieldsFloat(t *testing.T) {
	f := Fields{"f": 1.5, "n": json.Number("2.25"), "s": "x"}
	if v, ok := f.Float("f"); !ok || v != 1.5 {
		t.Errorf("Float(f) = %v, %v", v, ok)
	}
	if v, ok := f.Float("n"); !ok || v != 2.25 {
		t.Errorf("Float(n) = %v, %v", v, ok)
	}
	if _, ok := f.Float("s"); ok {
		t.Error("Float(s) should be false")
	}
}

func TestFieldsBool(t *testing.T) {
	f := Fields{"active": true, "other": "true"}
	if v, ok := f.Bool("active"); !ok || !v {
		t.Errorf("Bool(active) = %v, %v", v, ok)
	}
	if _, ok := f.Bool("other"); ok {
		t.Error("Bool(other) should be false for a string")
	}
}

func TestFieldsTime(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string // RFC3339 of expected, or "" for not-ok
	}{
		{"date only", "2024-01-02", "2024-01-02T00:00:00Z"},
		{"datetime space", "2024-01-02 15:04:05", "2024-01-02T15:04:05Z"},
		{"rfc3339", "2024-01-02T15:04:05Z", "2024-01-02T15:04:05Z"},
		{"empty", "", ""},
		{"garbage", "not-a-date", ""},
		{"non-string", 12345, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Fields{"k": tt.value}
			got, ok := f.Time("k")
			if tt.want == "" {
				if ok {
					t.Errorf("Time(%v) unexpectedly ok: %v", tt.value, got)
				}
				return
			}
			if !ok {
				t.Fatalf("Time(%v) not ok", tt.value)
			}
			if got.Format(time.RFC3339) != tt.want {
				t.Errorf("Time(%v) = %s, want %s", tt.value, got.Format(time.RFC3339), tt.want)
			}
		})
	}
}

func TestFieldsHasAndRaw(t *testing.T) {
	f := Fields{"present": "x", "nilval": nil}
	if !f.Has("present") {
		t.Error("Has(present) should be true")
	}
	if f.Has("nilval") {
		t.Error("Has(nilval) should be false for nil value")
	}
	if f.Has("missing") {
		t.Error("Has(missing) should be false")
	}
	if v, ok := f.Raw("present"); !ok || v != "x" {
		t.Errorf("Raw(present) = %v, %v", v, ok)
	}
}

func TestFieldsUUIDAndName(t *testing.T) {
	f := Fields{"uuid": "abc", "name": "widget"}
	if f.UUID() != "abc" {
		t.Errorf("UUID() = %q", f.UUID())
	}
	if f.Name() != "widget" {
		t.Errorf("Name() = %q", f.Name())
	}
	empty := Fields{}
	if empty.UUID() != "" || empty.Name() != "" {
		t.Error("UUID/Name should be empty on missing keys")
	}
}
