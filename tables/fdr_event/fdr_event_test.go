package fdr_event

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
)

func TestExtract_HandlesSensorAndExternalApiEvents(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/sample.jsonl")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	rows, err := (FdrEventExtractor{}).Extract(context.Background(), raw)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if got, want := len(rows), 2; got != want {
		t.Fatalf("row count: got %d, want %d", got, want)
	}

	// Row 0 is a Win sensor event (EndOfProcess); row 1 is an external-API event
	// (Event_ModuleSummaryInfoEvent on platform=Other).
	sensor, ok := rows[0].(*FdrEvent)
	if !ok {
		t.Fatalf("row[0] type: got %T", rows[0])
	}
	external, ok := rows[1].(*FdrEvent)
	if !ok {
		t.Fatalf("row[1] type: got %T", rows[1])
	}

	// ---- sensor row ----
	if got := strDeref(sensor.EventSimpleName); got != "EndOfProcess" {
		t.Errorf("sensor.EventSimpleName: got %q, want EndOfProcess", got)
	}
	if got := strDeref(sensor.EventPlatform); got != "Win" {
		t.Errorf("sensor.EventPlatform: got %q, want Win", got)
	}
	if got := strDeref(sensor.Aid); got == "" {
		t.Errorf("sensor.Aid: empty")
	}
	if got := strDeref(sensor.Aip); got == "" {
		t.Errorf("sensor.Aip: empty")
	}
	// Synthetic fixture uses a fixed epoch (1700000000.000 = 2023-11-14 UTC) to
	// keep the test reproducible and the fixture free of real timestamps.
	if got := strDeref(sensor.ContextTimeStamp); got != "1700000000.000" {
		t.Errorf("sensor.ContextTimeStamp: got %q, want 1700000000.000", got)
	}
	// Synthetic fixture has 25 keys; assertion is informational, not strict,
	// so adding fields to the fixture doesn't break the test.
	t.Logf("sensor.Payload key count: %d", len(sensor.Payload))
	if _, ok := sensor.Payload["LocalAddressIP4"]; !ok {
		t.Errorf("sensor.Payload missing LocalAddressIP4 — uncommon fields should still be queryable via payload")
	}

	// ---- external-API row ----
	if got := strDeref(external.EventType); got != "Event_ExternalApiEvent" {
		t.Errorf("external.EventType: got %q", got)
	}
	if got := strDeref(external.ExternalApiType); got != "Event_ModuleSummaryInfoEvent" {
		t.Errorf("external.ExternalApiType: got %q", got)
	}
	if got := strDeref(external.AgentIdString); got == "" {
		t.Errorf("external.AgentIdString: empty")
	}
	// Cross-fill: external-API events have no top-level `aid` field but should
	// still populate Cid (the build helper copies CustomerIdString → Cid when
	// the lowercase `cid` is missing).
	if got := strDeref(external.Cid); got == "" {
		t.Errorf("external.Cid: should be cross-filled from CustomerIdString")
	}
	if external.UTCTimestamp == nil {
		t.Errorf("external.UTCTimestamp: nil")
	}
	if external.Aip != nil {
		t.Errorf("external.Aip: external-API events do not carry aip; got %q", *external.Aip)
	}
}

func TestEnrichRow_PopulatesTpFieldsForBothFlavours(t *testing.T) {
	t.Parallel()

	raw, _ := os.ReadFile("testdata/sample.jsonl")
	rows, err := (FdrEventExtractor{}).Extract(context.Background(), raw)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	tbl := FdrEventTable{}
	for i, r := range rows {
		row := r.(*FdrEvent)
		out, err := tbl.EnrichRow(row, schema.SourceEnrichment{})
		if err != nil {
			t.Fatalf("row %d enrich: %v", i, err)
		}
		if out.TpID == "" {
			t.Errorf("row %d: TpID empty", i)
		}
		if out.TpTimestamp.IsZero() {
			t.Errorf("row %d: TpTimestamp zero", i)
		}
		if out.TpDate.IsZero() {
			t.Errorf("row %d: TpDate zero", i)
		}
		if h, m, s := out.TpDate.Clock(); h != 0 || m != 0 || s != 0 {
			t.Errorf("row %d: TpDate not midnight UTC: %v", i, out.TpDate)
		}
		if out.TpDate.Location() != time.UTC {
			t.Errorf("row %d: TpDate not UTC: %v", i, out.TpDate.Location())
		}
		if len(out.TpAkas) == 0 {
			t.Errorf("row %d: TpAkas empty (expected crowdstrike:aid:* entry)", i)
		}
	}

	// Row 0: sensor → tp_source_ip should be cross-filled from aip.
	sensor := rows[0].(*FdrEvent)
	if sensor.TpSourceIP == nil || sensor.Aip == nil || *sensor.TpSourceIP != *sensor.Aip {
		t.Errorf("sensor: TpSourceIP=%v should equal Aip=%v", sensor.TpSourceIP, sensor.Aip)
	}

	// Row 1: external-API → no aip, so TpSourceIP must be nil.
	external := rows[1].(*FdrEvent)
	if external.TpSourceIP != nil {
		t.Errorf("external: TpSourceIP got %q, want nil", *external.TpSourceIP)
	}
}

func TestResolveTimestamp_PreferenceOrder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		row  *FdrEvent
		want time.Time
	}{
		{
			name: "context_time_stamp wins (sensor epoch seconds with millis)",
			row: &FdrEvent{
				ContextTimeStamp: ptr("1778159119.283"),
				UTCTimestamp:     ptr("9999999999999"),
				Timestamp:        ptr("9999999999999"),
			},
			want: time.UnixMilli(1778159119283),
		},
		{
			name: "utc_timestamp wins when context_time_stamp absent (epoch ms)",
			row:  &FdrEvent{UTCTimestamp: ptr("1778158758261")},
			want: time.UnixMilli(1778158758261),
		},
		{
			name: "rfc3339 timestamp parsed when only timestamp present",
			row:  &FdrEvent{Timestamp: ptr("2026-05-07T12:59:18Z")},
			want: time.Date(2026, 5, 7, 12, 59, 18, 0, time.UTC),
		},
		{
			name: "epoch ms timestamp parsed when only timestamp present and not RFC3339",
			row:  &FdrEvent{Timestamp: ptr("1778159121826")},
			want: time.UnixMilli(1778159121826),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := resolveTimestamp(tc.row)
			if !got.Equal(tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func ptr(s string) *string { return &s }

func strDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
