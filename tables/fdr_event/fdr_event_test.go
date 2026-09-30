package fdr_event

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

func loadFixture(t *testing.T) []*FdrEvent {
	t.Helper()
	raw, err := os.ReadFile("testdata/sample.jsonl")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	m := common.NewJSONLinesMapper("fdr_event_mapper", mapFdrEvent)
	var rows []*FdrEvent
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		row, err := m.Map(context.Background(), line)
		if err != nil {
			t.Fatalf("map: %v", err)
		}
		rows = append(rows, row)
	}
	return rows
}

func TestMap_HandlesSensorAndExternalApiEvents(t *testing.T) {
	t.Parallel()

	rows := loadFixture(t)
	if got, want := len(rows), 2; got != want {
		t.Fatalf("row count: got %d, want %d", got, want)
	}

	// Row 0 is a Win sensor event (EndOfProcess); row 1 is an external-API event
	// (Event_ModuleSummaryInfoEvent on platform=Other).
	sensor, external := rows[0], rows[1]

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
	// Synthetic fixture uses a fixed epoch (1700000000 = 2023-11-14 UTC).
	if sensor.ContextTimeStamp == nil || !sensor.ContextTimeStamp.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("sensor.ContextTimeStamp: got %v", sensor.ContextTimeStamp)
	}
	// Sensor `timestamp` is epoch ms.
	if sensor.Timestamp == nil || !sensor.Timestamp.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("sensor.Timestamp: got %v", sensor.Timestamp)
	}
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
	if got := strDeref(external.Cid); got == "" {
		t.Errorf("external.Cid: should be cross-filled from CustomerIdString")
	}
	if external.UTCTimestamp == nil || !external.UTCTimestamp.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("external.UTCTimestamp: got %v", external.UTCTimestamp)
	}
	// External-API `timestamp` is RFC3339.
	if external.Timestamp == nil || !external.Timestamp.Equal(time.Date(2023, 11, 14, 22, 13, 20, 0, time.UTC)) {
		t.Errorf("external.Timestamp: got %v", external.Timestamp)
	}
	if external.Aip != nil {
		t.Errorf("external.Aip: external-API events do not carry aip; got %q", *external.Aip)
	}
}

func TestEnrichRow_PopulatesTpFieldsForBothFlavours(t *testing.T) {
	t.Parallel()

	rows := loadFixture(t)
	tbl := FdrEventTable{}
	for i, row := range rows {
		out, err := tbl.EnrichRow(row, schema.SourceEnrichment{})
		if err != nil {
			t.Fatalf("row %d enrich: %v", i, err)
		}
		if out.TpID == "" {
			t.Errorf("row %d: TpID empty", i)
		}
		if !out.TpTimestamp.Equal(time.Unix(1700000000, 0)) {
			t.Errorf("row %d: TpTimestamp got %v", i, out.TpTimestamp)
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
	if s := rows[0]; s.TpSourceIP == nil || s.Aip == nil || *s.TpSourceIP != *s.Aip {
		t.Errorf("sensor: TpSourceIP=%v should equal Aip=%v", s.TpSourceIP, s.Aip)
	}
	// Row 1: external-API → no aip, so TpSourceIP must be nil.
	if e := rows[1]; e.TpSourceIP != nil {
		t.Errorf("external: TpSourceIP got %q, want nil", *e.TpSourceIP)
	}
}

func TestEnrichRow_TimestampPreferenceAndMissing(t *testing.T) {
	t.Parallel()

	ctx, utc, ts := time.UnixMilli(1), time.UnixMilli(2), time.UnixMilli(3)
	cases := []struct {
		name string
		row  *FdrEvent
		want time.Time
	}{
		{"context_time_stamp wins", &FdrEvent{ContextTimeStamp: &ctx, UTCTimestamp: &utc, Timestamp: &ts}, ctx},
		{"utc_timestamp next", &FdrEvent{UTCTimestamp: &utc, Timestamp: &ts}, utc},
		{"timestamp last", &FdrEvent{Timestamp: &ts}, ts},
	}
	for _, tc := range cases {
		out, err := (FdrEventTable{}).EnrichRow(tc.row, schema.SourceEnrichment{})
		if err != nil || !out.TpTimestamp.Equal(tc.want) {
			t.Errorf("%s: got %v, %v", tc.name, out, err)
		}
	}

	if _, err := (FdrEventTable{}).EnrichRow(&FdrEvent{}, schema.SourceEnrichment{}); !errors.Is(err, common.ErrNoTimestamp) {
		t.Errorf("no timestamp: got %v, want ErrNoTimestamp", err)
	}
}

func strDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
