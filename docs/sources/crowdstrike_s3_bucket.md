---
title: "Source: crowdstrike_s3_bucket - Collect CrowdStrike FDR data from S3"
description: "Allows users to collect CrowdStrike Falcon Data Replicator (FDR) files from an S3 bucket or access-point alias."
---

# Source: crowdstrike_s3_bucket - Collect CrowdStrike FDR data from S3

CrowdStrike Falcon Data Replicator (FDR) delivers gzipped JSON-lines files to an S3 bucket, typically a customer-specific access-point alias provisioned by CrowdStrike (`cs-lion-cannon-*-s3alias`).

This source discovers those files with a [grok](https://github.com/elastic/go-grok)-based layout pattern and downloads them for collection. It authenticates with a [`crowdstrike` connection](https://hub.tailpipe.io/plugins/l-teles/crowdstrike#connection-credentials), which follows the standard AWS credential chain.

Every CrowdStrike table sets a default `file_layout` that matches both FDR key layouts, so you normally only need `bucket` and a tenant `prefix`. The trailing `/` is not added to `prefix` automatically; include it when your path needs it.

## Example Configurations

### Collect primary events

Collect sensor and external-API events for a tenant.

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_fdr_event" "my_events" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/data/"
  }
  tp_index = "cid"
}
```

### Collect a secondary lookup

Secondary lookups (AIDMaster, AppInfo, ManagedAssets, UserInfo) live under `fdrv2/` in the same bucket.

```hcl
partition "crowdstrike_aid_master" "my_hosts" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/aidmaster/"
  }
  tp_index = "cid"
}
```

### Collect from a date-first bucket layout

If FDR data is copied into your own bucket with the date partitions first, set a matching `file_layout` so `--from`/`--to` can prune whole days (see [Performance](#performance-prefer-a-date-first-layout)).

```hcl
partition "crowdstrike_fdr_event" "my_events_by_date" {
  source "crowdstrike_s3_bucket" {
    connection  = connection.crowdstrike.fdr
    bucket      = "my-fdr-bucket"
    prefix      = "data/"
    file_layout = `year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/platform=%{DATA:platform}/%{DATA}.gz`
  }
  tp_index = "cid"
}
```

### Collect Windows events only

Restrict collection to one platform by fixing the `platform` segment of the default layout.

```hcl
partition "crowdstrike_fdr_event" "my_windows_events" {
  source "crowdstrike_s3_bucket" {
    connection  = connection.crowdstrike.fdr
    bucket      = "cs-lion-cannon-XXXXXX-s3alias"
    prefix      = "<tenant-id>/data/"
    file_layout = `(batch=)?%{DATA:batch}/year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/platform=Win/%{DATA}.gz`
  }
  tp_index = "cid"
}
```

## Arguments

| Argument    | Type                     | Required | Default | Description |
|-------------|--------------------------|----------|---------|-------------|
| bucket      | String                   | Yes      |         | S3 bucket name or access-point alias. |
| connection  | `connection.crowdstrike` | Yes      |         | The [CrowdStrike connection](https://hub.tailpipe.io/plugins/l-teles/crowdstrike#connection-credentials) holding the AWS credentials for the bucket. |
| file_layout | String                   | No       | Set by each table | Grok pattern used to match object keys and capture metadata. When it declares Hive date partitions (`year=`/`month=`/`day=`/`hour=`), the source uses them to skip prefixes outside the `--from`/`--to` window. |
| prefix      | String                   | No       |         | Object-key prefix to scope the listing. Strongly recommended for FDR, since every tenant has its own prefix. |

### Table Defaults

The following tables define their own default values for certain source arguments:

- **[crowdstrike_fdr_event](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_fdr_event#crowdstrike_s3_bucket)**
- **[crowdstrike_aid_master](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_aid_master#crowdstrike_s3_bucket)**
- **[crowdstrike_app_info](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_app_info#crowdstrike_s3_bucket)**
- **[crowdstrike_managed_asset](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_managed_asset#crowdstrike_s3_bucket)**
- **[crowdstrike_user_info](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_user_info#crowdstrike_s3_bucket)**

## Performance: prefer a date-first layout

The source prunes S3 prefixes against the `--from`/`--to` collection window, but it can only prune on the date partitions your `file_layout` declares, and only from the point in the key where they appear.

CrowdStrike's default layout nests the date partitions **under** a per-batch directory:

```
batch=<uuid>/year=YYYY/month=MM/day=DD/hour=HH/platform=<plat>/part-*.txt.gz
```

Because `batch=<uuid>/` comes first, the walk must still enumerate **every** batch prefix before it can prune by date inside each one. On a large bucket that is thousands of listings regardless of how narrow the window is.

For ideal performance, organise the bucket **date-first**, with no batch directory above the dates:

```
year=YYYY/month=MM/day=DD/hour=HH/platform=<plat>/part-*.txt.gz
```

Now `--from`/`--to` prunes whole years, months and days at the top of the walk, so a one-day collect lists only that day's prefixes. If you control how FDR data lands in the bucket (e.g. a Lambda that rewrites keys on delivery), this layout is strongly recommended; see the [date-first example](#collect-from-a-date-first-bucket-layout).

Layouts without Hive date partitions (the flat `<uuid>/part-*.gz` FDR variant, or any non-date scheme) still collect correctly. They just can't be pruned, so the walk lists the full prefix.

## Notes

- **Region**: set `region` on the connection explicitly. It is **required** for `*-s3alias` access-point aliases (the region probe can't resolve aliases) and when reaching a bucket through a re-signing access proxy (e.g. Teleport's `tsh proxy aws`, which routes to a single region). For direct access to a real bucket it is optional, because the source probes for the region, but setting it skips that extra round-trip. Resolution order: connection `region`, then `AWS_REGION`, then the probe (real buckets only).
- **`_SUCCESS` markers**: every FDR batch directory contains a zero-byte `_SUCCESS` marker. The source skips them during discovery.
- **Missing partitions**: CrowdStrike occasionally omits the date or platform segments for tiny or late-arriving files. The default layout treats both as optional, so files at any depth match.
- **Other sources**: the tables also accept any artifact source, such as the SDK's `file` source for local replay or the [AWS plugin's](https://hub.tailpipe.io/plugins/turbot/aws) `aws_s3_bucket` source. Those sources don't get the table default `file_layout`, alias handling or date pruning, so set `file_layout` yourself.
