package common

import (
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
	"github.com/turbot/tailpipe-plugin-sdk/mappers"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/sources/s3_bucket"
)

// SourceMetadata is the source list shared by every CrowdStrike table: the
// plugin's own S3 source (with the FDR default layout) plus any artifact
// source — the SDK's `file` source or another plugin's, e.g. `aws_s3_bucket`.
func SourceMetadata[R any](mapper mappers.Mapper[R]) []*table.SourceMetadata[R] {
	return []*table.SourceMetadata[R]{
		{
			SourceName: s3_bucket.CrowdstrikeS3BucketSourceIdentifier,
			Mapper:     mapper,
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(&artifact_source_config.ArtifactSourceConfigImpl{
					FileLayout: DefaultBatchLayout,
				}),
				artifact_source.WithArtifactLoader(NewJSONLinesLoader()),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Mapper:     mapper,
			Options: []row_source.RowSourceOption{
				artifact_source.WithArtifactLoader(NewJSONLinesLoader()),
			},
		},
	}
}
