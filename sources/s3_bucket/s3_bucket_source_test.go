package s3_bucket

import (
	"strings"
	"testing"
	"time"
)

// TestValidateArtifactKey covers the path-traversal defence in DownloadArtifact:
// S3 keys are caller-controlled, so anything that would escape the per-source
// temp directory once joined locally must be rejected before os.Create runs.
func TestValidateArtifactKey(t *testing.T) {
	t.Parallel()

	safe := []string{
		"part-00000.gz",
		"abc123/part-00000.gz",
		"batch=abc123/year=2026/month=05/day=07/hour=13/platform=Win/part-00000.c000.txt.gz",
		"2a0e29555370beab-0a908e98/fdrv2/aidmaster/uuid/part-00000.gz",
		"file.with.many.dots.gz",
		"file with spaces.gz",
		"unicode-éñ-名前.gz",
	}
	for _, key := range safe {
		t.Run("safe/"+key, func(t *testing.T) {
			t.Parallel()
			if err := validateArtifactKey(key); err != nil {
				t.Errorf("expected %q to validate, got error: %v", key, err)
			}
		})
	}

	unsafe := []struct {
		key  string
		want string // expected substring of error message
	}{
		{"", "empty"},
		{"../etc/passwd", "parent-directory"},
		{"a/../../b.gz", "parent-directory"},
		{"a/b/../../../etc/passwd", "parent-directory"},
		{"/etc/passwd", "absolute"},
		{"/tmp/escape.gz", "absolute"},
		{"valid\x00.gz", "NUL"},
		{"\x00.gz", "NUL"},
	}
	for _, tc := range unsafe {
		t.Run("unsafe/"+tc.key, func(t *testing.T) {
			t.Parallel()
			err := validateArtifactKey(tc.key)
			if err == nil {
				t.Fatalf("expected %q to be rejected", tc.key)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q should mention %q", err.Error(), tc.want)
			}
		})
	}
}

// TestDirStartsAfter covers the upper-boundary prefix pruning that keeps a
// narrow-window collect from walking every partition newer than --to. Pruning
// is layout-driven: it must only act on date partitions the file_layout
// declares, and must be a no-op for any other bucket structure.
func TestDirStartsAfter(t *testing.T) {
	t.Parallel()

	// Collecting a single day: --from 2026-06-02 --to 2026-06-03 gives an upper
	// boundary at the start of 2026-06-03.
	to := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)

	// A date-first layout and the Hive FDR variant both partition by
	// year=/month=/day=/hour=.
	hive := "year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/platform=%{DATA:platform}/%{DATA}.gz"
	// The flat FDR variant has no date partitions at all.
	flat := "%{DATA:batch}/%{DATA}.gz"

	cases := []struct {
		name   string
		prefix string
		layout string
		to     time.Time
		want   bool
	}{
		// Inside or before the window — must be walked.
		{"year enclosing window", "data/year=2026/", hive, to, false},
		{"month enclosing window", "data/year=2026/month=06/", hive, to, false},
		{"target day", "data/year=2026/month=06/day=02/", hive, to, false},
		{"hour within target day", "data/year=2026/month=06/day=02/hour=13/", hive, to, false},
		// The boundary day itself starts exactly at `to` (not strictly after),
		// so it is kept — it can hold the 00:00 instant.
		{"boundary day kept", "data/year=2026/month=06/day=03/", hive, to, false},
		{"earlier month", "data/year=2026/month=05/", hive, to, false},

		// Strictly after the window — must be pruned.
		{"next day", "data/year=2026/month=06/day=04/", hive, to, true},
		{"next month", "data/year=2026/month=07/", hive, to, true},
		{"next year", "data/year=2027/", hive, to, true},

		// Layout has no date partitions — never prune, even if a key happens to
		// contain a "year=" substring.
		{"flat layout, plain prefix", "abc123def/", flat, to, false},
		{"flat layout, coincidental year= token", "weird/year=2099/", flat, to, false},

		// No date tokens in the prefix (Hive layout, but shallow / alias-style) —
		// never prune.
		{"hive layout, no tokens in prefix", "tenant-id/fdrv2/aidmaster/", hive, to, false},
		// Zero upper boundary (no --to) — never prune.
		{"zero to", "data/year=2027/", hive, time.Time{}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := dirStartsAfter(tc.prefix, tc.layout, tc.to); got != tc.want {
				t.Errorf("dirStartsAfter(%q, layout, %s) = %v, want %v", tc.prefix, tc.to.Format("2006-01-02"), got, tc.want)
			}
		})
	}
}
