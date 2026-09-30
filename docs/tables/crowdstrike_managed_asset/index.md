---
title: "Tailpipe Table: crowdstrike_managed_asset - Query CrowdStrike FDR Managed Assets"
description: "CrowdStrike FDR ManagedAssets snapshots: network interface and gateway info for each Falcon-managed host."
---

# Table: crowdstrike_managed_asset - Query CrowdStrike FDR Managed Assets

The `crowdstrike_managed_asset` table allows you to query ManagedAssets snapshots from CrowdStrike Falcon Data Replicator (FDR). Each row is one network interface on one Falcon-managed host, with its addresses and default gateway.

## Configure

Create a [partition](https://tailpipe.io/docs/manage/partition) for `crowdstrike_managed_asset` ([examples](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_managed_asset#example-configurations)):

```sh
vi ~/.tailpipe/config/crowdstrike.tpc
```

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_managed_asset" "my_assets" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/managedassets/"
  }
  tp_index = "cid"
}
```

## Collect

[Collect](https://tailpipe.io/docs/manage/collection) snapshots for all `crowdstrike_managed_asset` partitions:

```sh
tailpipe collect crowdstrike_managed_asset
```

Or for a single partition:

```sh
tailpipe collect crowdstrike_managed_asset.my_assets
```

## Query

**[Explore example queries for this table →](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/queries/crowdstrike_managed_asset)**

### Hosts behind a specific gateway

Find the hosts and addresses that route through a given gateway.

```sql
select distinct
  aid,
  local_address_ip4,
  gateway_ip
from
  crowdstrike_managed_asset
where
  gateway_ip = '203.0.113.1';
```

### Most common subnets

Group hosts by their /24 subnet.

```sql
select
  regexp_extract(local_address_ip4, '^(\d+\.\d+\.\d+)\.', 1) as subnet,
  count(distinct aid) as host_count
from
  crowdstrike_managed_asset
where
  local_address_ip4 is not null
group by
  subnet
order by
  host_count desc
limit 20;
```

### Network hardware vendors

Count interfaces per MAC vendor prefix (OUI).

```sql
select
  mac_prefix,
  count(*) as interface_count
from
  crowdstrike_managed_asset
group by
  mac_prefix
order by
  interface_count desc
limit 20;
```

## Example Configurations

### Collect ManagedAssets snapshots from an S3 bucket

Collect ManagedAssets snapshots for a tenant, indexed by customer ID.

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_managed_asset" "my_assets" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/managedassets/"
  }
  tp_index = "cid"
}
```

### Collect interfaces with a gateway only

Use the partition `filter` to skip interfaces without a default gateway.

```hcl
partition "crowdstrike_managed_asset" "my_routed_assets" {
  filter = "gateway_ip is not null"

  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/managedassets/"
  }
  tp_index = "cid"
}
```

### Collect ManagedAssets snapshots from local files

Replay ManagedAssets files downloaded outside Tailpipe.

```hcl
partition "crowdstrike_managed_asset" "local_assets" {
  source "file" {
    paths       = ["/Users/myuser/fdr/managedassets"]
    file_layout = `%{DATA}.gz`
  }
  tp_index = "cid"
}
```

## Notes

- `time` (wire field `_time`) is parsed into a `TIMESTAMP` column. Other values stay as text, as delivered.
- The table covers managed hosts only. Hosts seen on the network without a Falcon agent (the FDR NotManaged lookup) are not collected yet.
- IP and MAC addresses identify devices. Restrict access to your Tailpipe data accordingly.

## Source Defaults

### crowdstrike_s3_bucket

This table sets the following defaults for the [crowdstrike_s3_bucket source](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/sources/crowdstrike_s3_bucket#arguments):

| Argument    | Default |
|-------------|---------|
| file_layout | `(batch=)?%{DATA:batch}/(year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/)?(platform=%{DATA:platform}/)?%{DATA}.gz` |
