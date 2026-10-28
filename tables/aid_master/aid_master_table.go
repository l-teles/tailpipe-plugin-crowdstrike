package aid_master

import (
	"time"

	"github.com/rs/xid"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

const AidMasterTableIdentifier = "crowdstrike_aid_master"

type AidMasterTable struct{}

func (AidMasterTable) Identifier() string { return AidMasterTableIdentifier }

func (AidMasterTable) GetDescription() string {
	return "CrowdStrike FDR AIDMaster snapshots: one row per agent (host) with sensor, OS, and hardware metadata."
}

func (AidMasterTable) GetSourceMetadata() ([]*table.SourceMetadata[*AidMaster], error) {
	return common.SourceMetadata(common.NewJSONLinesMapper("aid_master_mapper", mapAidMaster)), nil
}

func (AidMasterTable) EnrichRow(row *AidMaster, sourceEnrichmentFields schema.SourceEnrichment) (*AidMaster, error) {
	ts, err := common.FirstTime(row.Time, row.FirstSeen)
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
	if row.Aid != nil {
		row.TpAkas = append(row.TpAkas, "crowdstrike:aid:"+*row.Aid)
	}

	return row, nil
}
