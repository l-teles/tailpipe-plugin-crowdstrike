package managed_assets

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

const ManagedAssetsTableIdentifier = "crowdstrike_managed_assets"

type ManagedAssetsTable struct{}

func (ManagedAssetsTable) Identifier() string { return ManagedAssetsTableIdentifier }

func (ManagedAssetsTable) GetDescription() string {
	return "CrowdStrike FDR ManagedAssets snapshots: network interface and gateway info per Falcon-managed agent."
}

func (ManagedAssetsTable) GetSourceMetadata() ([]*table.SourceMetadata[*ManagedAsset], error) {
	return []*table.SourceMetadata[*ManagedAsset]{
		{
			SourceName: s3_bucket.CrowdstrikeS3BucketSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(&artifact_source_config.ArtifactSourceConfigImpl{
					FileLayout: common.DefaultBatchLayout,
				}),
				artifact_source.WithArtifactExtractor(NewManagedAssetsExtractor()),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithArtifactExtractor(NewManagedAssetsExtractor()),
			},
		},
	}, nil
}

func (ManagedAssetsTable) EnrichRow(row *ManagedAsset, sourceEnrichmentFields schema.SourceEnrichment) (*ManagedAsset, error) {
	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpIngestTimestamp = time.Now().UTC()
	row.TpTimestamp = common.PickEpochSeconds(row.Time)
	row.TpDate = row.TpTimestamp.Truncate(24 * time.Hour)

	if row.LocalAddressIP4 != nil {
		row.TpSourceIP = row.LocalAddressIP4
		row.TpIps = append(row.TpIps, *row.LocalAddressIP4)
	}
	if row.GatewayIP != nil {
		row.TpIps = append(row.TpIps, *row.GatewayIP)
	}
	if row.Aid != nil {
		row.TpAkas = append(row.TpAkas, "crowdstrike:aid:"+*row.Aid)
	}

	return row, nil
}
