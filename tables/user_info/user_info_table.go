package user_info

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

const UserInfoTableIdentifier = "crowdstrike_user_info"

type UserInfoTable struct{}

func (UserInfoTable) Identifier() string { return UserInfoTableIdentifier }

func (UserInfoTable) GetDescription() string {
	return "CrowdStrike FDR UserInfo snapshots: local-account inventory observed on each Falcon-managed host."
}

func (UserInfoTable) GetSourceMetadata() ([]*table.SourceMetadata[*UserInfo], error) {
	return []*table.SourceMetadata[*UserInfo]{
		{
			SourceName: s3_bucket.CrowdstrikeS3BucketSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(&artifact_source_config.ArtifactSourceConfigImpl{
					FileLayout: common.DefaultBatchLayout,
				}),
				artifact_source.WithArtifactExtractor(NewUserInfoExtractor()),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithArtifactExtractor(NewUserInfoExtractor()),
			},
		},
	}, nil
}

func (UserInfoTable) EnrichRow(row *UserInfo, sourceEnrichmentFields schema.SourceEnrichment) (*UserInfo, error) {
	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpIngestTimestamp = time.Now().UTC()
	row.TpTimestamp = common.PickEpochSeconds(row.Time, row.LogonTime)
	row.TpDate = row.TpTimestamp.Truncate(24 * time.Hour)

	if row.UserName != nil {
		row.TpUsernames = append(row.TpUsernames, *row.UserName)
	}
	if row.User != nil && (row.UserName == nil || *row.User != *row.UserName) {
		row.TpUsernames = append(row.TpUsernames, *row.User)
	}
	if row.UserSidReadable != nil {
		row.TpAkas = append(row.TpAkas, "crowdstrike:sid:"+*row.UserSidReadable)
	}

	return row, nil
}
