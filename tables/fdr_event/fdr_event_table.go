package fdr_event

import (
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/sources/s3_bucket"
	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

const FdrEventTableIdentifier = "crowdstrike_fdr_event"

// FdrEventTable exposes CrowdStrike FDR primary events as the
// crowdstrike_fdr_event table.
type FdrEventTable struct{}

func (FdrEventTable) Identifier() string { return FdrEventTableIdentifier }

func (FdrEventTable) GetDescription() string {
	return "CrowdStrike Falcon Data Replicator primary events: sensor telemetry and external-API events delivered via FDR."
}

func (FdrEventTable) GetSourceMetadata() ([]*table.SourceMetadata[*FdrEvent], error) {
	return []*table.SourceMetadata[*FdrEvent]{
		{
			SourceName: s3_bucket.CrowdstrikeS3BucketSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(&artifact_source_config.ArtifactSourceConfigImpl{
					FileLayout: common.DefaultBatchLayout,
				}),
				artifact_source.WithArtifactExtractor(NewFdrEventExtractor()),
			},
		},
		{
			// Generic local-file source provided by the SDK. Lets users replay
			// FDR files downloaded outside Tailpipe (e.g. by aws s3 cp).
			SourceName: constants.ArtifactSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithArtifactExtractor(NewFdrEventExtractor()),
			},
		},
	}, nil
}

// EnrichRow populates the schema.CommonFields tp_* columns. FDR uses
// inconsistent timestamp formats across event families; resolveTimestamp
// tries the most specific field first and falls back gracefully.
func (FdrEventTable) EnrichRow(row *FdrEvent, sourceEnrichmentFields schema.SourceEnrichment) (*FdrEvent, error) {
	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpIngestTimestamp = time.Now().UTC()
	row.TpTimestamp = resolveTimestamp(row).UTC()
	row.TpDate = row.TpTimestamp.Truncate(24 * time.Hour)

	if row.Aip != nil {
		row.TpSourceIP = row.Aip
		row.TpIps = append(row.TpIps, *row.Aip)
	}

	switch {
	case row.Aid != nil:
		row.TpAkas = append(row.TpAkas, "crowdstrike:aid:"+*row.Aid)
	case row.AgentIdString != nil:
		row.TpAkas = append(row.TpAkas, "crowdstrike:aid:"+*row.AgentIdString)
	}

	return row, nil
}

// resolveTimestamp picks the best timestamp present on the row.
//
// Order of preference:
//  1. ContextTimeStamp — sensor events; epoch seconds with optional fractional ms.
//  2. UTCTimestamp     — external-API events; epoch milliseconds as string.
//  3. timestamp        — varies: epoch milliseconds (sensor) OR RFC3339 (external-API).
//
// Falls back to time.Now() if nothing parses, so EnrichRow never produces a
// zero-value tp_timestamp (which would break Hive partitioning).
func resolveTimestamp(row *FdrEvent) time.Time {
	if row.ContextTimeStamp != nil {
		if t, ok := parseEpochSeconds(*row.ContextTimeStamp); ok {
			return t
		}
	}
	if row.UTCTimestamp != nil {
		if t, ok := parseEpochMillis(*row.UTCTimestamp); ok {
			return t
		}
	}
	if row.Timestamp != nil {
		if t, ok := parseRFC3339(*row.Timestamp); ok {
			return t
		}
		if t, ok := parseEpochMillis(*row.Timestamp); ok {
			return t
		}
	}
	return time.Now()
}

// parseEpochSeconds handles "1778159119.283" (seconds with optional fractional
// milliseconds). FDR's fractional component is millisecond-precision; we
// round to the nearest millisecond to avoid float64 precision artifacts.
func parseEpochSeconds(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || f <= 0 {
		return time.Time{}, false
	}
	ms := int64(f*1000 + 0.5) // round to nearest millisecond
	return time.UnixMilli(ms), true
}

// parseEpochMillis handles "1778159121826".
func parseEpochMillis(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil || n <= 0 {
		return time.Time{}, false
	}
	return time.UnixMilli(n), true
}

// parseRFC3339 handles "2026-05-07T12:59:18Z" and similar.
func parseRFC3339(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	return time.Time{}, false
}
