// Package crowdstrikeconfig holds the plugin's connection (HCL) configuration
// and AWS-credential helpers. It lives at this path (rather than the more
// conventional `config/`) to sidestep a local-tooling quirk that occasionally
// removes a top-level `config/` directory between tool invocations.
package crowdstrikeconfig

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// PluginName is the OCI-style name reported by `<plugin> metadata` and used
// by Tailpipe to locate the installed binary. The namespace must match the
// directory under hub.tailpipe.io/plugins/, otherwise `tailpipe collect` will
// look up the wrong path.
const PluginName = "l-teles/crowdstrike"

// ConnectionType is the short identifier used in HCL connection blocks
// (`connection "crowdstrike" "default" { ... }`). It is intentionally distinct
// from PluginName: PluginName is namespaced (l-teles/crowdstrike) so the OCI
// ref points at the right install directory, while ConnectionType matches the
// short HCL type users actually write.
const ConnectionType = "crowdstrike"

// CrowdstrikeConnection holds the AWS credentials used to read the customer's
// Falcon Data Replicator (FDR) S3 bucket. Auth follows the standard AWS
// credential chain: explicit static keys > named profile > environment vars >
// instance / IRSA / SSO. Values not set fall through to the chain.
type CrowdstrikeConnection struct {
	Profile      *string `hcl:"profile,optional"`
	AccessKey    *string `hcl:"access_key,optional"`
	SecretKey    *string `hcl:"secret_key,optional"`
	SessionToken *string `hcl:"session_token,optional"`
	Region       *string `hcl:"region,optional"`
	EndpointUrl  *string `hcl:"endpoint_url,optional"`
}

func (c *CrowdstrikeConnection) Validate() error {
	if c.AccessKey != nil && c.SecretKey == nil {
		return fmt.Errorf("access_key set without secret_key")
	}
	if c.AccessKey == nil && c.SecretKey != nil {
		return fmt.Errorf("secret_key set without access_key")
	}
	return nil
}

func (c *CrowdstrikeConnection) Identifier() string {
	return ConnectionType
}

// GetClientConfiguration returns an *aws.Config populated from the connection.
// If overrideRegion is non-nil it wins, then connection.Region, then the
// AWS_REGION env var, then us-east-1.
func (c *CrowdstrikeConnection) GetClientConfiguration(ctx context.Context, overrideRegion *string) (*aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	if c.Profile != nil {
		opts = append(opts, config.WithSharedConfigProfile(aws.ToString(c.Profile)))
	}

	if c.AccessKey != nil && c.SecretKey != nil {
		token := ""
		if c.SessionToken != nil {
			token = aws.ToString(c.SessionToken)
		}
		provider := credentials.NewStaticCredentialsProvider(
			aws.ToString(c.AccessKey),
			aws.ToString(c.SecretKey),
			token,
		)
		opts = append(opts, config.WithCredentialsProvider(provider))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	switch {
	case overrideRegion != nil && *overrideRegion != "":
		cfg.Region = *overrideRegion
	case c.Region != nil && *c.Region != "":
		cfg.Region = *c.Region
	case cfg.Region == "":
		if env := os.Getenv("AWS_REGION"); env != "" {
			cfg.Region = env
		} else {
			cfg.Region = "us-east-1"
		}
	}

	// Custom-endpoint override for HCL `endpoint_url`: applied at S3-client
	// build time via s3.Options.BaseEndpoint (see s3_bucket_source.go).
	// AWS_ENDPOINT_URL / AWS_ENDPOINT_URL_S3 env vars are handled natively by
	// the SDK v2 loader (loaded by config.LoadDefaultConfig above) so we don't
	// re-read them ourselves.

	return &cfg, nil
}
