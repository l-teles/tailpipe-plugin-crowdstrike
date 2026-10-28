---
title: "Tailpipe Table: crowdstrike_app_info - Query CrowdStrike FDR AppInfo"
description: "CrowdStrike FDR AppInfo snapshots: the application inventory Falcon observes on each host."
---

# Table: crowdstrike_app_info - Query CrowdStrike FDR AppInfo

The `crowdstrike_app_info` table allows you to query AppInfo snapshots from CrowdStrike Falcon Data Replicator (FDR). Each row is one application observed on one host. Use it to hunt unwanted software, build a software inventory, or check how widely a tool is deployed.

## Configure

Create a [partition](https://tailpipe.io/docs/manage/partition) for `crowdstrike_app_info` ([examples](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_app_info#example-configurations)):

```sh
vi ~/.tailpipe/config/crowdstrike.tpc
```

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_app_info" "my_apps" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/appinfo/"
  }
  tp_index = "cid"
}
```

## Collect

[Collect](https://tailpipe.io/docs/manage/collection) snapshots for all `crowdstrike_app_info` partitions:

```sh
tailpipe collect crowdstrike_app_info
```

Or for a single partition:

```sh
tailpipe collect crowdstrike_app_info.my_apps
```

## Query

**[Explore example queries for this table →](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/queries/crowdstrike_app_info)**

### Hosts running a specific binary

Find every host with a given executable and the versions in use.

```sql
select distinct
  aid,
  hostname,
  file_version,
  product_version
from
  crowdstrike_app_info
where
  file_name = 'node';
```

### Top applications by install footprint

Rank applications by how many hosts have them.

```sql
select
  product_name,
  count(distinct aid) as host_count
from
  crowdstrike_app_info
where
  product_name is not null
  and product_name != '#Application'
group by
  product_name
order by
  host_count desc
limit 25;
```

### Applications with detections

List applications that Falcon has recorded detections for.

```sql
select distinct
  aid,
  hostname,
  file_name,
  sha256_hash_data,
  detection_count
from
  crowdstrike_app_info
where
  detection_count > 0
order by
  detection_count desc;
```

## Example Configurations

### Collect AppInfo snapshots from an S3 bucket

Collect AppInfo snapshots for a tenant, indexed by customer ID.

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_app_info" "my_apps" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/appinfo/"
  }
  tp_index = "cid"
}
```

### Collect applications with detections only

Use the partition `filter` to keep only applications that have detections.

```hcl
partition "crowdstrike_app_info" "my_detected_apps" {
  filter = "detection_count > 0"

  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/appinfo/"
  }
  tp_index = "cid"
}
```

### Collect AppInfo snapshots from local files

Replay AppInfo files downloaded outside Tailpipe.

```hcl
partition "crowdstrike_app_info" "local_apps" {
  source "file" {
    paths       = ["/Users/myuser/fdr/appinfo"]
    file_layout = `%{DATA}.gz`
  }
  tp_index = "cid"
}
```

## Notes

- `time` (wire field `_time`) and `installation_timestamp` are parsed into `TIMESTAMP` columns; an `installation_timestamp` of `"0"` means unknown and becomes null. `detection_count` is a `BIGINT`, although FDR sends it as a float string such as `"0.0"`.
- Vendor and product strings are sometimes literal placeholders like `"#Vendor"`, `"#Application"` or `"#Version"` when CrowdStrike couldn't extract a real value.
- `hostname` and `external_ip` identify devices. Restrict access to your Tailpipe data accordingly.

## Source Defaults

### crowdstrike_s3_bucket

This table sets the following defaults for the [crowdstrike_s3_bucket source](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/sources/crowdstrike_s3_bucket#arguments):

| Argument    | Default |
|-------------|---------|
| file_layout | `(batch=)?%{DATA:batch}/(year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/)?(platform=%{DATA:platform}/)?%{DATA}.gz` |
