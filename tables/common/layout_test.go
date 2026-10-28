package common

import (
	"strings"
	"testing"

	"github.com/elastic/go-grok"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
)

// TestDefaultBatchLayout_MatchesBothFdrLayouts feeds real-shaped paths from
// both observed FDR layouts (classic Hive-style + newer flat) through the
// same grok engine the SDK uses, and asserts that at least one of the
// expanded alternatives matches each — with the expected `batch` capture.
//
// This catches regressions where someone tightens the grok pattern, switches
// `(...)?` back to `(?:...)?` (rejected by the SDK's grok engine), or drops
// the `.txt` tolerance on the file extension.
func TestDefaultBatchLayout_MatchesBothFdrLayouts(t *testing.T) {
	t.Parallel()

	layout := *DefaultBatchLayout

	// 3 optional groups in the layout → 2^3 = 8 expanded alternatives.
	alts := artifact_source.ExpandPatternIntoOptionalAlternatives(layout)
	if got, want := len(alts), 8; got != want {
		t.Fatalf("expected %d expanded alternatives, got %d:\n%s",
			want, got, strings.Join(alts, "\n"))
	}

	// Every alternative must be a syntactically valid grok pattern; otherwise
	// the SDK's walk fails with "error parsing regexp" (the bug we hit on the
	// `(?:...)?` form).
	for _, alt := range alts {
		g := grok.New()
		if err := g.Compile(alt, true); err != nil {
			t.Errorf("alternative %q failed to compile: %v", alt, err)
		}
	}

	cases := []struct {
		name         string
		path         string
		wantBatch    string
		wantYear     string // empty when not captured
		wantPlatform string // empty when not captured
	}{
		{
			name:      "new (flat) FDR layout",
			path:      "abc12345-def6-7890-1234-fedcba987654/part-00000.gz",
			wantBatch: "abc12345-def6-7890-1234-fedcba987654",
		},
		{
			name:         "old (Hive) FDR layout, fully partitioned, .txt.gz extension",
			path:         "batch=0034ad7a-0a9f-4ef5-a62c-3cb650f1c86a/year=2026/month=05/day=07/hour=13/platform=Win/part-00000-505a736b-d5af-4ca5-ad77-648a10f40f26.c000.txt.gz",
			wantBatch:    "0034ad7a-0a9f-4ef5-a62c-3cb650f1c86a",
			wantYear:     "2026",
			wantPlatform: "Win",
		},
		{
			name:      "old (Hive) FDR layout, no platform segment",
			path:      "batch=0034ad7a-0a9f-4ef5-a62c-3cb650f1c86a/year=2026/month=05/day=07/hour=13/part-00000.c000.txt.gz",
			wantBatch: "0034ad7a-0a9f-4ef5-a62c-3cb650f1c86a",
			wantYear:  "2026",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var matchedAlt string
			var captures map[string]string

			for _, alt := range alts {
				g := grok.New()
				if err := g.Compile(alt, true); err != nil {
					continue
				}
				m, err := g.ParseString(tc.path)
				if err != nil || len(m) == 0 {
					continue
				}
				// Require the alternative to extract the expected batch UUID;
				// some alternatives match the path but with a wrong capture
				// (e.g. consuming `<uuid>/part-00000` as the batch).
				if got := m["batch"]; got != tc.wantBatch {
					continue
				}
				matchedAlt = alt
				captures = m
				break
			}

			if matchedAlt == "" {
				t.Fatalf("no alternative produced batch=%q for path %q.\nTried:\n  %s",
					tc.wantBatch, tc.path, strings.Join(alts, "\n  "))
			}

			if tc.wantYear != "" && captures["year"] != tc.wantYear {
				t.Errorf("year capture: got %q, want %q (alt=%q)",
					captures["year"], tc.wantYear, matchedAlt)
			}
			if tc.wantPlatform != "" && captures["platform"] != tc.wantPlatform {
				t.Errorf("platform capture: got %q, want %q (alt=%q)",
					captures["platform"], tc.wantPlatform, matchedAlt)
			}

			t.Logf("matched via alternative: %s\ncaptures: %+v", matchedAlt, captures)
		})
	}
}
