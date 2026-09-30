package common

import (
	"errors"
	"testing"
	"time"
)

func TestTypedParsers(t *testing.T) {
	t.Parallel()

	doc := map[string]any{
		"secs":      "1778159119.283",
		"zero":      "0",
		"empty":     "",
		"junk":      "N/A",
		"ms":        "1778159121826",
		"rfc":       "2026-05-07T12:59:18Z",
		"floatInt":  "3.0",
		"fraction":  "1.5",
		"one":       "1",
		"falseWord": "false",
		"number":    42.0, // not a string on the wire
	}

	if got := EpochSecondsFromMap(doc, "secs"); got == nil || !got.Equal(time.UnixMilli(1778159119283)) || got.Location() != time.UTC {
		t.Errorf("EpochSeconds: got %v", got)
	}
	for _, k := range []string{"zero", "empty", "junk", "missing", "number"} {
		if got := EpochSecondsFromMap(doc, k); got != nil {
			t.Errorf("EpochSeconds(%s): got %v, want nil", k, got)
		}
	}
	if got := EpochMillisFromMap(doc, "ms"); got == nil || !got.Equal(time.UnixMilli(1778159121826)) {
		t.Errorf("EpochMillis: got %v", got)
	}
	if got := RFC3339OrEpochMillisFromMap(doc, "rfc"); got == nil || !got.Equal(time.Date(2026, 5, 7, 12, 59, 18, 0, time.UTC)) {
		t.Errorf("RFC3339OrEpochMillis(rfc): got %v", got)
	}
	if got := RFC3339OrEpochMillisFromMap(doc, "ms"); got == nil || !got.Equal(time.UnixMilli(1778159121826)) {
		t.Errorf("RFC3339OrEpochMillis(ms): got %v", got)
	}
	if got := IntFromMap(doc, "floatInt"); got == nil || *got != 3 {
		t.Errorf("Int(floatInt): got %v", got)
	}
	for _, k := range []string{"fraction", "junk", "empty"} {
		if got := IntFromMap(doc, k); got != nil {
			t.Errorf("Int(%s): got %v, want nil", k, *got)
		}
	}
	if got := BoolFromMap(doc, "one"); got == nil || !*got {
		t.Errorf("Bool(one): got %v", got)
	}
	if got := BoolFromMap(doc, "falseWord"); got == nil || *got {
		t.Errorf("Bool(falseWord): got %v", got)
	}
	if got := BoolFromMap(doc, "junk"); got != nil {
		t.Errorf("Bool(junk): got %v, want nil", *got)
	}
}

func TestFirstTime(t *testing.T) {
	t.Parallel()

	a := time.Unix(1, 0)
	if got, err := FirstTime(nil, &a); err != nil || !got.Equal(a) {
		t.Errorf("got %v, %v", got, err)
	}
	if _, err := FirstTime(nil, nil); !errors.Is(err, ErrNoTimestamp) {
		t.Errorf("got %v, want ErrNoTimestamp", err)
	}
}
