// Package common holds helpers shared by the CrowdStrike table packages.
package common

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/v2/utils"
)

// DefaultBatchLayout is the shared FDR file layout (relative to the
// per-tenant prefix). It matches both FDR layout variants observed in the
// wild:
//
//  1. Hive-style ("classic" FDR):
//     batch=<uuid>/year=YYYY/month=MM/day=DD/hour=HH/platform=<plat>/part-*.txt.gz
//  2. Flat ("newer" FDR / recently provisioned tenants):
//     <uuid>/part-*.gz
//
// The `batch=` prefix, date partitioning, platform partitioning, and the
// `.txt.` extension prefix are all wrapped in optional groups so the SDK's
// ExpandPatternIntoOptionalAlternatives expansion covers every combination.
//
// Users supply a `prefix` like "<tenant-id>/data/" or
// "<tenant-id>/fdrv2/aidmaster/"; the SDK prepends the prefix to this layout
// during artifact discovery.
var DefaultBatchLayout = utils.ToStringPointer(
	"(batch=)?%{DATA:batch}/" +
		"(year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/)?" +
		"(platform=%{DATA:platform}/)?" +
		"%{DATA}.gz")

// ExtractJSONLines splits a gunzipped FDR file (passed as the raw decompressed
// bytes by the SDK's GzipLoader) into one row per JSON line, returning a slice
// of row pointers ready for the SDK to call EnrichRow on.
//
// The caller provides build(doc) which constructs the typed row struct from
// the raw map. Lines that fail to parse are logged and skipped — FDR
// occasionally writes truncated/malformed records and we want collection to
// continue rather than abort the whole file.
//
// FDR records can exceed 1 MiB, so we override bufio.Scanner's default 64 KiB
// buffer (which would silently truncate large lines).
func ExtractJSONLines[T any](raw []byte, source string, build func(map[string]any) *T) ([]any, error) {
	const initBuf = 1 << 20  // 1 MiB initial
	const maxLine = 16 << 20 // 16 MiB cap

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, initBuf), maxLine)

	var rows []any
	var lineNo int
	for scanner.Scan() {
		lineNo++
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var doc map[string]any
		if err := json.Unmarshal(line, &doc); err != nil {
			slog.Warn("crowdstrike: skipping unparseable line",
				"source", source, "line_no", lineNo, "error", err, "preview", preview(line))
			continue
		}

		rows = append(rows, build(doc))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: scanner error: %w", source, err)
	}
	return rows, nil
}

// PickEpochSeconds returns the first parseable epoch-seconds timestamp from
// the provided candidates as a UTC time.Time. Falls back to time.Now() so
// EnrichRow always produces a non-zero tp_timestamp (zero values break Hive
// partitioning). FDR's secondary records all use epoch-seconds strings (with
// optional fractional millis); this helper is shared across the secondary
// tables.
func PickEpochSeconds(cands ...*string) time.Time {
	for _, c := range cands {
		if c == nil {
			continue
		}
		s := strings.TrimSpace(*c)
		if s == "" {
			continue
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil || f <= 0 {
			continue
		}
		// FDR fractional epoch seconds are millisecond-precision; round to
		// the nearest ms to avoid float64 artifacts (e.g. 1778159119.283
		// otherwise lands at 282999992ns rather than 283000000ns).
		ms := int64(f*1000 + 0.5)
		return time.UnixMilli(ms).UTC()
	}
	return time.Now().UTC()
}

// StringFromMap returns *string when the key exists and the value is a
// non-empty JSON string. FDR delivers all values as strings on the wire so
// type-coercion is unnecessary.
func StringFromMap(m map[string]any, key string) *string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	return &s
}

func preview(b []byte) string {
	const max = 120
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "…"
}
