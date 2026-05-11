package s3_bucket

import (
	"strings"
	"testing"
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
