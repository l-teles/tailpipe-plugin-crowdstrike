package app_info

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

const AppInfoTableIdentifier = "crowdstrike_app_info"

type AppInfoTable struct{}

func (AppInfoTable) Identifier() string { return AppInfoTableIdentifier }

func (AppInfoTable) GetDescription() string {
	return "CrowdStrike FDR AppInfo snapshots: application inventory observed by Falcon, one row per (agent, application) tuple."
}

func (AppInfoTable) GetSourceMetadata() ([]*table.SourceMetadata[*AppInfo], error) {
	return []*table.SourceMetadata[*AppInfo]{
		{
			SourceName: s3_bucket.CrowdstrikeS3BucketSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(&artifact_source_config.ArtifactSourceConfigImpl{
					FileLayout: common.DefaultBatchLayout,
				}),
				artifact_source.WithArtifactExtractor(NewAppInfoExtractor()),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithArtifactExtractor(NewAppInfoExtractor()),
			},
		},
	}, nil
}

func (AppInfoTable) EnrichRow(row *AppInfo, sourceEnrichmentFields schema.SourceEnrichment) (*AppInfo, error) {
	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpIngestTimestamp = time.Now().UTC()
	row.TpTimestamp = common.PickEpochSeconds(row.Time)
	row.TpDate = row.TpTimestamp.Truncate(24 * time.Hour)

	if row.ExternalIP != nil {
		row.TpIps = append(row.TpIps, *row.ExternalIP)
	}
	if row.Aid != nil {
		row.TpAkas = append(row.TpAkas, "crowdstrike:aid:"+*row.Aid)
	}
	if row.SHA256HashData != nil {
		row.TpAkas = append(row.TpAkas, "crowdstrike:sha256:"+*row.SHA256HashData)
	}

	return row, nil
}
