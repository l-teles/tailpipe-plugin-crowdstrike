# crowdstrike_s3_bucket

Reads CrowdStrike Falcon Data Replicator (FDR) `.txt.gz` files from an S3 bucket — typically the customer-specific access-point alias provisioned by CrowdStrike (`cs-lion-cannon-*-s3alias`).

Authenticates via the standard AWS credential chain (profile, env vars, IRSA, SSO). Files are discovered with a [grok](https://github.com/elastic/go-grok)-based layout pattern and downloaded to a temp directory before extraction.

## Arguments

| Field | Required | Description |
|---|---|---|
| `connection`   | yes | Reference to a `connection "crowdstrike"` block. |
| `bucket`       | yes | S3 bucket name or access-point alias. |
| `prefix`       | no  | Object-key prefix to scope the listing — strongly recommended for FDR (every customer prefix is tenant-specific). |
| `file_layout`  | no  | Grok pattern used to filter and extract metadata from object keys. Defaults to the standard FDR `batch=<uuid>/year=YYYY/month=MM/day=DD/hour=HH/platform=<plat>/part-*.txt.gz` layout, with the date / platform segments treated as optional. When the layout declares Hive date partitions (`year=`/`month=`/`day=`/`hour=`), the plugin uses them to prune prefixes outside the `--from`/`--to` window — see [Performance](#performance-prefer-a-date-first-layout). |

## Example

```hcl
connection "crowdstrike" "default" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"   # set explicitly — see the region note below
}

partition "crowdstrike_fdr_event" "prod" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.default
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/data/"
  }
}
```

## Performance: prefer a date-first layout

The plugin prunes S3 prefixes against the `--from`/`--to` collection window so a narrow-window collect doesn't walk the whole bucket — but it can only prune on the date partitions your `file_layout` declares, and only from the point in the key where they appear.

CrowdStrike's default layout nests the date partitions **under** a per-batch directory:

```
batch=<uuid>/year=YYYY/month=MM/day=DD/hour=HH/platform=<plat>/part-*.txt.gz
```

Because `batch=<uuid>/` comes first, the walk must still enumerate **every** batch prefix before it can reach — and prune by — the date partitions inside each one. On a large bucket that is thousands of top-level listings regardless of how narrow the window is.

For ideal performance, organise the bucket **date-first**, with no batch directory above the dates:

```
year=YYYY/month=MM/day=DD/hour=HH/platform=<plat>/part-*.txt.gz
```

Now `--from`/`--to` prunes whole years, months, and days at the very top of the walk: a one-day collect lists only that day's prefixes. If you control how FDR data lands in the bucket (e.g. a Lambda that rewrites keys on delivery), this layout is strongly recommended. Point the source at the date root and give it a matching `file_layout`:

```hcl
partition "crowdstrike_fdr_event" "prod" {
  source "crowdstrike_s3_bucket" {
    connection  = connection.crowdstrike.default
    bucket      = "my-fdr-bucket"
    prefix      = "data/"
    file_layout = "year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/platform=%{DATA:platform}/%{DATA}.gz"
  }
}
```

Layouts without Hive date partitions (the flat `<uuid>/part-*.gz` FDR variant, or any non-date scheme) still collect correctly — they simply can't be pruned, so the walk lists the full prefix.

## Notes

- **Region** — set `region` on the connection explicitly. It is **required** for `*-s3alias` access-point aliases (the `manager.GetBucketRegion` probe can't resolve aliases) and when reaching a real bucket through a re-signing access proxy (e.g. Teleport's `tsh proxy aws`, which routes to a single region). For direct access to a real bucket it is optional — the plugin will probe for the region — but setting it is recommended, as it skips the extra HeadBucket round-trip. Resolution order: connection `region` → `AWS_REGION` → probe (real buckets only).
- **S3 access-point aliases** — the AWS Go SDK accepts the `*-s3alias` form as a bucket name. The plugin detects the suffix and skips `manager.GetBucketRegion`, which doesn't support aliases; the connection's `region` (or `AWS_REGION`) wins instead.
- **`_SUCCESS` markers** — every FDR batch directory contains a zero-byte `_SUCCESS` marker. The plugin filters them out during discovery so they don't appear as artifacts.
- **Layout flexibility** — date and platform partitioning may occasionally be missing from a batch (CrowdStrike does this for tiny / late-arriving files). The default grok pattern wraps both segments in optional groups so files at any depth match.
- **Both real bucket names and access-point aliases work** — the plugin doesn't require special configuration for either.

## Local-file alternative

For replay or testing, use the SDK's built-in `file` source instead — every CrowdStrike table registers it. See each table's docs for an example.
