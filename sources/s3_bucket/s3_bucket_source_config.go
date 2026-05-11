package s3_bucket

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
)

// CrowdstrikeS3BucketSourceConfig is the HCL configuration for a
// crowdstrike_s3_bucket source.
type CrowdstrikeS3BucketSourceConfig struct {
	// Remain captures any unknown HCL attributes so partial decoding does not
	// fail when users add fields the SDK uses (e.g. file_layout).
	Remain hcl.Body `hcl:",remain" json:"-"`

	artifact_source_config.ArtifactSourceConfigImpl

	Bucket string  `hcl:"bucket"`
	Prefix *string `hcl:"prefix,optional"`
}

func (c *CrowdstrikeS3BucketSourceConfig) Validate() error {
	if c.Bucket == "" {
		return fmt.Errorf("bucket is required")
	}
	return nil
}

func (c *CrowdstrikeS3BucketSourceConfig) Identifier() string {
	return CrowdstrikeS3BucketSourceIdentifier
}
