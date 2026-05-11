# crowdstrike_s3_bucket

Reads CrowdStrike Falcon Data Replicator (FDR) `.txt.gz` files from an S3 bucket — typically the customer-specific access-point alias provisioned by CrowdStrike (`cs-lion-cannon-*-s3alias`).

Authenticates via the standard AWS credential chain (profile, env vars, IRSA, SSO). Files are discovered with a [grok](https://github.com/elastic/go-grok)-based layout pattern and downloaded to a temp directory before extraction.

## Arguments

| Field | Required | Description |
|---|---|---|
| `connection`   | yes | Reference to a `connection "crowdstrike"` block. |
| `bucket`       | yes | S3 bucket name or access-point alias. |
| `prefix`       | no  | Object-key prefix to scope the listing — strongly recommended for FDR (every customer prefix is tenant-specific). |
| `file_layout`  | no  | Grok pattern used to filter and extract metadata from object keys. Defaults to the standard FDR `batch=<uuid>/year=YYYY/month=MM/day=DD/hour=HH/platform=<plat>/part-*.txt.gz` layout, with the date / platform segments treated as optional. |

## Example

```hcl
connection "crowdstrike" "default" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"   # required for access-point aliases — bucket-region probe doesn't work for them
}

partition "crowdstrike_fdr_event" "prod" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.default
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/data/"
  }
}
```

## Notes

- **S3 access-point aliases** — the AWS Go SDK accepts the `*-s3alias` form as a bucket name. The plugin detects the suffix and skips `manager.GetBucketRegion`, which doesn't support aliases; the connection's `region` (or `AWS_REGION`) wins instead.
- **`_SUCCESS` markers** — every FDR batch directory contains a zero-byte `_SUCCESS` marker. The plugin filters them out during discovery so they don't appear as artifacts.
- **Layout flexibility** — date and platform partitioning may occasionally be missing from a batch (CrowdStrike does this for tiny / late-arriving files). The default grok pattern wraps both segments in optional groups so files at any depth match.
- **Both real bucket names and access-point aliases work** — the plugin doesn't require special configuration for either.

## Local-file alternative

For replay or testing, use the SDK's built-in `file` source instead — every CrowdStrike table registers it. See each table's docs for an example.
