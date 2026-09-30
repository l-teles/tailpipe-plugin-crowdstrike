package app_info

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

func TestMapAndEnrich_AppInfo(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/sample.jsonl")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	line, _, _ := strings.Cut(string(raw), "\n")
	row, err := common.NewJSONLinesMapper("app_info_mapper", mapAppInfo).Map(context.Background(), line)
	if err != nil {
		t.Fatalf("map: %v", err)
	}

	// detectionCount arrives as "0.0".
	if row.DetectionCount == nil || *row.DetectionCount != 0 {
		t.Errorf("DetectionCount: got %v, want 0", row.DetectionCount)
	}
	// installationTimestamp "0" means unknown.
	if row.InstallationTimestamp != nil {
		t.Errorf("InstallationTimestamp: got %v, want nil", row.InstallationTimestamp)
	}
	if row.Time == nil || !row.Time.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("Time: got %v", row.Time)
	}

	out, err := (AppInfoTable{}).EnrichRow(row, schema.SourceEnrichment{})
	if err != nil {
		t.Fatalf("enrich: %v", err)
	}
	if !out.TpTimestamp.Equal(*row.Time) {
		t.Errorf("TpTimestamp: got %v", out.TpTimestamp)
	}
	if len(out.TpAkas) != 2 {
		t.Errorf("TpAkas: got %v, want aid and sha256 entries", out.TpAkas)
	}
}
