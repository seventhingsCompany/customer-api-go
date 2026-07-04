package models

import (
	"encoding/json"
	"time"
)

// Fields wraps a dynamic resource body (object/asset, room, location,
// circularity-hub item) with type-safe accessors. These resources are returned
// as map[string]any because their schema is instance-defined; convert one with
// models.Fields(m) to read values without panic-prone type assertions.
//
// Every getter follows the comma-ok idiom: it returns the zero value and false
// when the key is absent, nil, or holds a different type.
type Fields map[string]any

// timeLayouts are the formats Time attempts, in order. The API returns bare
// dates (YYYY-MM-DD) for most fields but may return datetimes for others.
var timeLayouts = []string{
	"2006-01-02",
	"2006-01-02 15:04:05",
	time.RFC3339,
}

// Raw returns the underlying value for key and whether it was present.
func (f Fields) Raw(key string) (any, bool) {
	v, ok := f[key]
	return v, ok
}

// Has reports whether key is present with a non-nil value.
func (f Fields) Has(key string) bool {
	v, ok := f[key]
	return ok && v != nil
}

// String returns the value for key if it is a string.
func (f Fields) String(key string) (string, bool) {
	s, ok := f[key].(string)
	return s, ok
}

// Float returns the value for key as a float64. It accepts JSON's default
// float64 as well as json.Number.
func (f Fields) Float(key string) (float64, bool) {
	switch v := f[key].(type) {
	case float64:
		return v, true
	case json.Number:
		n, err := v.Float64()
		return n, err == nil
	default:
		return 0, false
	}
}

// Int returns the value for key as an int. It accepts JSON's default float64
// (when integral), json.Number, and int. A non-integral float returns false.
func (f Fields) Int(key string) (int, bool) {
	switch v := f[key].(type) {
	case int:
		return v, true
	case float64:
		if v != float64(int(v)) {
			return 0, false
		}
		return int(v), true
	case json.Number:
		n, err := v.Int64()
		return int(n), err == nil
	default:
		return 0, false
	}
}

// Bool returns the value for key if it is a bool.
func (f Fields) Bool(key string) (bool, bool) {
	b, ok := f[key].(bool)
	return b, ok
}

// Time returns the value for key parsed as a time. It accepts date-only,
// datetime, and RFC3339 layouts. Non-string or unparseable values return false.
func (f Fields) Time(key string) (time.Time, bool) {
	s, ok := f[key].(string)
	if !ok || s == "" {
		return time.Time{}, false
	}
	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// UUID returns the "uuid" field as a string, or "" if absent or not a string.
func (f Fields) UUID() string {
	s, _ := f.String("uuid")
	return s
}

// Name returns the "name" field as a string, or "" if absent or not a string.
func (f Fields) Name() string {
	s, _ := f.String("name")
	return s
}
