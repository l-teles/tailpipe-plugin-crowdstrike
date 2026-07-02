package s3_bucket

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elastic/go-grok"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/v2/filter"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/context_values"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/types"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/crowdstrikeconfig"
)

const (
	CrowdstrikeS3BucketSourceIdentifier = "crowdstrike_s3_bucket"
	defaultBucketRegion                 = "us-east-1"
)

// CrowdstrikeS3BucketSource is an [artifact_source.ArtifactSource] that reads
// CrowdStrike Falcon Data Replicator files from an S3 bucket (or its
// access-point alias).
type CrowdstrikeS3BucketSource struct {
	artifact_source.ArtifactSourceImpl[*CrowdstrikeS3BucketSourceConfig, *crowdstrikeconfig.CrowdstrikeConnection]

	client *s3.Client
}

func (s *CrowdstrikeS3BucketSource) Identifier() string {
	return CrowdstrikeS3BucketSourceIdentifier
}

func (s *CrowdstrikeS3BucketSource) Init(ctx context.Context, params *row_source.RowSourceParams, opts ...row_source.RowSourceOption) error {
	if err := s.ArtifactSourceImpl.Init(ctx, params, opts...); err != nil {
		return err
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return fmt.Errorf("getting S3 client: %w", err)
	}
	s.client = client

	slog.Info("Initialized CrowdstrikeS3BucketSource", "bucket", s.Config.Bucket, "layout", s.Config.FileLayout)
	return nil
}

func (s *CrowdstrikeS3BucketSource) Close() error {
	_ = os.RemoveAll(s.TempDir)
	return nil
}

func (s *CrowdstrikeS3BucketSource) ValidateConfig() error {
	return s.Config.Validate()
}

func (s *CrowdstrikeS3BucketSource) DiscoverArtifacts(ctx context.Context) error {
	layout := typehelpers.SafeString(s.Config.GetFileLayout())
	optionalLayouts := artifact_source.ExpandPatternIntoOptionalAlternatives(layout)

	g := grok.New()
	if err := g.AddPatterns(s.Config.GetPatterns()); err != nil {
		return fmt.Errorf("adding grok patterns: %w", err)
	}

	var prefix string
	if s.Config.Prefix != nil {
		prefix = *s.Config.Prefix
		// when a prefix is provided we expand it into the layout so files at
		// the prefix root match too (mirrors tailpipe-plugin-aws behaviour).
		var withPrefix []string
		for _, l := range optionalLayouts {
			withPrefix = append(withPrefix, fmt.Sprintf("%s%s", prefix, l))
		}
		optionalLayouts = append(optionalLayouts, withPrefix...)
	}

	if err := s.walkS3(ctx, s.Config.Bucket, prefix, optionalLayouts, map[string]*filter.SqlFilter{}, g); err != nil {
		return fmt.Errorf("%s: %s", s.Config.Bucket, err.Error())
	}
	return nil
}

func (s *CrowdstrikeS3BucketSource) DownloadArtifact(ctx context.Context, info *types.ArtifactInfo) error {
	// Validate the S3 key before joining it onto a local path. S3 keys are
	// caller-controlled (anyone with write access to the bucket can choose
	// them) and a key like "../etc/passwd" would otherwise escape TempDir
	// once path.Join cleans the segments.
	if err := validateArtifactKey(info.Name); err != nil {
		return fmt.Errorf("%s: rejecting unsafe S3 key: %w", info.Name, err)
	}

	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.Config.Bucket,
		Key:    &info.Name,
	})
	if err != nil {
		return fmt.Errorf("%s: failed to download artifact from %s: %w", info.Name, s.Config.Bucket, err)
	}
	defer out.Body.Close()

	size := typehelpers.Int64Value(out.ContentLength)

	localFilePath := path.Join(s.TempDir, info.Name)
	if err := os.MkdirAll(filepath.Dir(localFilePath), 0o750); err != nil {
		return fmt.Errorf("%s: creating local directory: %w", info.Name, err)
	}

	f, err := os.Create(localFilePath) // #nosec G304 -- key validated by validateArtifactKey above
	if err != nil {
		return fmt.Errorf("%s: creating local file: %w", info.Name, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, out.Body); err != nil {
		return fmt.Errorf("%s: writing local file: %w", info.Name, err)
	}

	return s.OnArtifactDownloaded(ctx, types.NewDownloadedArtifactInfo(info, localFilePath, size))
}

// validateArtifactKey rejects S3 keys that would let the caller escape the
// per-source temp directory once joined locally. Three classes of unsafe key:
//
//  1. Absolute paths (`/etc/passwd`) — would bypass TempDir entirely.
//  2. Parent-directory segments (`../foo`, `a/../../b`) — would traverse up.
//  3. NUL bytes — interpreted as a path terminator on some kernels.
//
// Empty keys are rejected too; legitimate S3 keys always have a name.
func validateArtifactKey(key string) error {
	if key == "" {
		return fmt.Errorf("empty key")
	}
	if strings.ContainsRune(key, 0) {
		return fmt.Errorf("contains NUL byte")
	}
	if filepath.IsAbs(key) || strings.HasPrefix(key, "/") {
		return fmt.Errorf("absolute path not allowed")
	}
	for seg := range strings.SplitSeq(filepath.ToSlash(key), "/") {
		if seg == ".." {
			return fmt.Errorf("parent-directory segment %q not allowed", seg)
		}
	}
	return nil
}

func (s *CrowdstrikeS3BucketSource) getClient(ctx context.Context) (*s3.Client, error) {
	// Has the caller pinned a region explicitly, via the connection `region`
	// attribute or the AWS_REGION env var? If so we trust it and skip the
	// HeadBucket region probe below.
	regionConfigured := (s.Connection.Region != nil && *s.Connection.Region != "") ||
		os.Getenv("AWS_REGION") != ""

	// GetBucketRegion does not work against S3 access-point aliases (the
	// `*-s3alias` form) — it requires a real bucket name. For aliases we never
	// probe and let the connection's Region (or AWS_REGION env var) win.
	isAlias := strings.HasSuffix(s.Config.Bucket, "-s3alias")

	// Let the connection resolve the region (connection.Region > AWS_REGION >
	// us-east-1 default). We only override it below, and only when we actually
	// run the probe.
	cfg, err := s.Connection.GetClientConfiguration(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("client configuration: %w", err)
	}

	// Apply the optional HCL endpoint override via s3.Options.BaseEndpoint
	// (the SDK v2 idiom). When unset this is a no-op and the SDK picks the
	// default S3 endpoint for cfg.Region. The AWS_ENDPOINT_URL[_S3] env vars
	// are handled by the SDK itself and need no special handling here.
	s3OptFns := func(o *s3.Options) {
		if s.Connection.EndpointUrl != nil && *s.Connection.EndpointUrl != "" {
			o.BaseEndpoint = s.Connection.EndpointUrl
		}
	}

	// Only probe for the bucket's region when we have to: a real bucket name
	// AND no region configured. The probe is a *cross-region* HeadBucket
	// (seeded at us-east-1, relying on S3's x-amz-bucket-region hint), which
	// fails behind a re-signing access proxy — Teleport's `tsh proxy aws`
	// routes to a single region and returns a 504 GatewayTimeout for the
	// cross-region hint dance. When the region is known we must not run it, and
	// respecting an explicit region is the more correct behaviour regardless.
	if !isAlias && !regionConfigured {
		// Seed us-east-1 purely as the probe's starting endpoint.
		probeCfg := *cfg
		probeCfg.Region = defaultBucketRegion

		// Resolve the region with a *signed* probe.
		//
		// manager.GetBucketRegion forces anonymous, unsigned requests by
		// default (it sets options.Credentials = nil internally). Against AWS
		// directly that works — S3 returns the x-amz-bucket-region header even
		// on an unauthenticated HeadBucket. But behind a re-signing access
		// proxy an unsigned request has nothing to validate, so the proxy
		// rejects it with 403. Restoring the configured credentials makes the
		// probe a normal signed request; the optFn runs after
		// GetBucketRegion's internal nil-out, so it wins.
		region, err := manager.GetBucketRegion(ctx, s3.NewFromConfig(probeCfg, s3OptFns), s.Config.Bucket,
			func(o *s3.Options) { o.Credentials = cfg.Credentials })
		if err != nil {
			return nil, fmt.Errorf("resolving bucket region: %w", err)
		}
		cfg.Region = region
	}

	return s3.NewFromConfig(*cfg, s3OptFns), nil
}

func (s *CrowdstrikeS3BucketSource) walkS3(ctx context.Context, bucket, prefix string, layouts []string, filterMap map[string]*filter.SqlFilter, g *grok.Grok) error {
	executionId, err := context_values.ExecutionIdFromContext(ctx)
	if err != nil {
		return err
	}

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing objects: %w", err)
		}

		for _, dir := range page.CommonPrefixes {
			dirPrefix := typehelpers.SafeString(*dir.Prefix)
			err = s.WalkNode(ctx, dirPrefix, "", layouts, true, g, filterMap)
			if err != nil {
				if errors.Is(err, fs.SkipDir) {
					continue
				}
				slog.Error("walk directory failed", "key", dirPrefix, "error", err)
				s.NotifyError(ctx, executionId, fmt.Errorf("%s: failed to obtain directory info", dirPrefix))
				continue
			}
			if err := s.walkS3(ctx, bucket, dirPrefix, layouts, filterMap, g); err != nil {
				slog.Error("walk subtree failed", "bucket", bucket, "prefix", dirPrefix, "error", err)
				s.NotifyError(ctx, executionId, fmt.Errorf("%s: %s", bucket, err.Error()))
			}
		}

		for _, obj := range page.Contents {
			objKey := typehelpers.SafeString(*obj.Key)
			if objKey == "" {
				continue
			}
			// FDR drops zero-byte _SUCCESS markers in every batch directory; skip them.
			if strings.HasSuffix(objKey, "/_SUCCESS") || strings.HasSuffix(objKey, "_SUCCESS") {
				if obj.Size != nil && *obj.Size == 0 {
					continue
				}
			}
			if err := s.WalkNode(ctx, objKey, "", layouts, false, g, filterMap); err != nil {
				slog.Error("walk file failed", "key", objKey, "error", err)
				s.NotifyError(ctx, executionId, fmt.Errorf("%s: failed to obtain artifact info", objKey))
			}
		}
	}
	return nil
}
