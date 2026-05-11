package aid_master

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
)

func TestExtractAndEnrich_AidMaster(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/sample.jsonl")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	rows, err := (AidMasterExtractor{}).Extract(context.Background(), raw)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("no rows")
	}

	first, ok := rows[0].(*AidMaster)
	if !ok {
		t.Fatalf("row[0] type: %T", rows[0])
	}
	if first.Aid == nil || *first.Aid == "" {
		t.Errorf("Aid: empty")
	}
	if first.Cid == nil || *first.Cid == "" {
		t.Errorf("Cid: empty")
	}
	if first.AgentVersion == nil || *first.AgentVersion == "" {
		t.Errorf("AgentVersion: empty")
	}
	if first.Time == nil {
		t.Errorf("Time: nil — required for tp_timestamp")
	}

	out, err := (AidMasterTable{}).EnrichRow(first, schema.SourceEnrichment{})
	if err != nil {
		t.Fatalf("enrich: %v", err)
	}
	if out.TpID == "" {
		t.Errorf("TpID empty")
	}
	if out.TpTimestamp.IsZero() {
		t.Errorf("TpTimestamp zero")
	}
	if out.TpDate.Location() != time.UTC {
		t.Errorf("TpDate not UTC: %v", out.TpDate.Location())
	}
	if len(out.TpAkas) == 0 || out.TpAkas[0] == "" {
		t.Errorf("TpAkas: expected crowdstrike:aid:* entry, got %v", out.TpAkas)
	}
}
