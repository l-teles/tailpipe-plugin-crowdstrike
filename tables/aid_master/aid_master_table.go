package aid_master

import (
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

const AidMasterTableIdentifier = "crowdstrike_aid_master"

type AidMasterTable struct{}

func (AidMasterTable) Identifier() string { return AidMasterTableIdentifier }

func (AidMasterTable) GetDescription() string {
	return "CrowdStrike FDR AIDMaster snapshots: one row per agent (host) with sensor, OS, and hardware metadata."
}

func (AidMasterTable) GetSourceMetadata() ([]*table.SourceMetadata[*AidMaster], error) {
	return []*table.SourceMetadata[*AidMaster]{
		{
			SourceName: s3_bucket.CrowdstrikeS3BucketSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(&artifact_source_config.ArtifactSourceConfigImpl{
					FileLayout: common.DefaultBatchLayout,
				}),
				artifact_source.WithArtifactExtractor(NewAidMasterExtractor()),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithArtifactExtractor(NewAidMasterExtractor()),
			},
		},
	}, nil
}

func (AidMasterTable) EnrichRow(row *AidMaster, sourceEnrichmentFields schema.SourceEnrichment) (*AidMaster, error) {
	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpIngestTimestamp = time.Now().UTC()
	row.TpTimestamp = common.PickEpochSeconds(row.Time, row.FirstSeen)
	row.TpDate = row.TpTimestamp.Truncate(24 * time.Hour)

	if row.Aip != nil {
		row.TpSourceIP = row.Aip
		row.TpIps = append(row.TpIps, *row.Aip)
	}
	if row.Aid != nil {
		row.TpAkas = append(row.TpAkas, "crowdstrike:aid:"+*row.Aid)
	}

	return row, nil
}
