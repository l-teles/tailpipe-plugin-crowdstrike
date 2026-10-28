package aid_master

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

func TestMapAndEnrich_AidMaster(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/sample.jsonl")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	line, _, _ := strings.Cut(string(raw), "\n")
	first, err := common.NewJSONLinesMapper("aid_master_mapper", mapAidMaster).Map(context.Background(), line)
	if err != nil {
		t.Fatalf("map: %v", err)
	}

	if first.Aid == nil || *first.Aid == "" {
		t.Errorf("Aid: empty")
	}
	if first.AgentVersion == nil || *first.AgentVersion == "" {
		t.Errorf("AgentVersion: empty")
	}
	if first.Time == nil || !first.Time.Equal(time.Unix(1700000060, 0)) {
		t.Errorf("Time: got %v", first.Time)
	}
	if first.FirstSeen == nil || !first.FirstSeen.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("FirstSeen: got %v", first.FirstSeen)
	}

	out, err := (AidMasterTable{}).EnrichRow(first, schema.SourceEnrichment{})
	if err != nil {
		t.Fatalf("enrich: %v", err)
	}
	if !out.TpTimestamp.Equal(*first.Time) {
		t.Errorf("TpTimestamp: got %v, want %v", out.TpTimestamp, first.Time)
	}
	if out.TpDate.Location() != time.UTC {
		t.Errorf("TpDate not UTC: %v", out.TpDate.Location())
	}
	if len(out.TpAkas) == 0 || out.TpAkas[0] == "" {
		t.Errorf("TpAkas: expected crowdstrike:aid:* entry, got %v", out.TpAkas)
	}

	if _, err := (AidMasterTable{}).EnrichRow(&AidMaster{}, schema.SourceEnrichment{}); !errors.Is(err, common.ErrNoTimestamp) {
		t.Errorf("no timestamp: got %v, want ErrNoTimestamp", err)
	}
}
