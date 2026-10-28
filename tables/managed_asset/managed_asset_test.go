package managed_asset

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

func TestMapAndEnrich_ManagedAsset(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/sample.jsonl")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	m := common.NewJSONLinesMapper("managed_asset_mapper", mapManagedAsset)
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")

	first, err := m.Map(context.Background(), lines[0])
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	out, err := (ManagedAssetTable{}).EnrichRow(first, schema.SourceEnrichment{})
	if err != nil {
		t.Fatalf("enrich: %v", err)
	}
	if !out.TpTimestamp.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("TpTimestamp: got %v", out.TpTimestamp)
	}
	if len(out.TpIps) != 2 {
		t.Errorf("TpIps: got %v, want local and gateway IPs", out.TpIps)
	}

	// Second fixture row has an empty GatewayIP, which must not reach tp_ips.
	second, err := m.Map(context.Background(), lines[1])
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if second.GatewayIP != nil {
		t.Errorf("GatewayIP: got %q, want nil", *second.GatewayIP)
	}
}
