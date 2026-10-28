package fdr_event

import (
	"time"

	"github.com/rs/xid"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"

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
	return common.SourceMetadata(common.NewJSONLinesMapper("fdr_event_mapper", mapFdrEvent)), nil
}

// EnrichRow populates the schema.CommonFields tp_* columns. A record with no
// timestamp is rejected so it shows up as a row error in the collect summary.
func (FdrEventTable) EnrichRow(row *FdrEvent, sourceEnrichmentFields schema.SourceEnrichment) (*FdrEvent, error) {
	ts, err := common.FirstTime(row.ContextTimeStamp, row.UTCTimestamp, row.Timestamp)
	if err != nil {
		return nil, err
	}

	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpIngestTimestamp = time.Now().UTC()
	row.TpTimestamp = ts
	row.TpDate = ts.Truncate(24 * time.Hour)

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
