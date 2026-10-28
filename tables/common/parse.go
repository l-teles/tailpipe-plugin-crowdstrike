package common

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

// FDR delivers every value as a JSON string. These helpers turn the ones we
// promote to typed columns into Go types; a missing, empty or unparseable
// value becomes nil (the raw value is still in `payload`).

// ErrNoTimestamp is returned by EnrichRow when a record has no usable
// timestamp, so it surfaces as a row error rather than landing in today's
// partition.
var ErrNoTimestamp = errors.New("record has no parseable timestamp")

// StringFromMap returns the value at key when it is a non-empty string.
func StringFromMap(m map[string]any, key string) *string {
	s, ok := m[key].(string)
	if !ok || s == "" {
		return nil
	}
	return &s
}

// EpochSecondsFromMap parses epoch seconds with optional fractional
// milliseconds (e.g. "1778159119.283"). "0" means unknown and becomes nil.
func EpochSecondsFromMap(m map[string]any, key string) *time.Time {
	f, ok := floatFromMap(m, key)
	if !ok || f <= 0 {
		return nil
	}
	// Round to the nearest ms to avoid float64 artifacts (.283 → 282999992ns).
	t := time.UnixMilli(int64(math.Round(f * 1000))).UTC()
	return &t
}

// EpochMillisFromMap parses epoch milliseconds (e.g. "1778159121826").
func EpochMillisFromMap(m map[string]any, key string) *time.Time {
	s := StringFromMap(m, key)
	if s == nil {
		return nil
	}
	n, err := strconv.ParseInt(strings.TrimSpace(*s), 10, 64)
	if err != nil || n <= 0 {
		return nil
	}
	t := time.UnixMilli(n).UTC()
	return &t
}

// RFC3339OrEpochMillisFromMap parses RFC3339 (external-API events) or epoch
// milliseconds (sensor events), the two forms FDR uses for `timestamp`.
func RFC3339OrEpochMillisFromMap(m map[string]any, key string) *time.Time {
	s := StringFromMap(m, key)
	if s == nil {
		return nil
	}
	if t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(*s)); err == nil {
		t = t.UTC()
		return &t
	}
	return EpochMillisFromMap(m, key)
}

// IntFromMap parses an integral count. FDR sometimes renders counts as
// floats ("0.0"); non-integral values and placeholders like "N/A" become nil.
func IntFromMap(m map[string]any, key string) *int64 {
	f, ok := floatFromMap(m, key)
	if !ok || f != math.Trunc(f) || math.Abs(f) > 1<<53 {
		return nil
	}
	n := int64(f)
	return &n
}

// BoolFromMap parses "1"/"0" (and "true"/"false").
func BoolFromMap(m map[string]any, key string) *bool {
	s := StringFromMap(m, key)
	if s == nil {
		return nil
	}
	b, err := strconv.ParseBool(strings.TrimSpace(*s))
	if err != nil {
		return nil
	}
	return &b
}

// FirstTime returns the first non-nil candidate.
func FirstTime(cands ...*time.Time) (time.Time, error) {
	for _, c := range cands {
		if c != nil {
			return *c, nil
		}
	}
	return time.Time{}, ErrNoTimestamp
}

func floatFromMap(m map[string]any, key string) (float64, bool) {
	s := StringFromMap(m, key)
	if s == nil {
		return 0, false
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(*s), 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
}
