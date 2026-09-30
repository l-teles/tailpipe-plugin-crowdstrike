---
title: "Tailpipe Table: crowdstrike_aid_master - Query CrowdStrike FDR AIDMaster"
description: "CrowdStrike FDR AIDMaster snapshots: sensor, operating system and hardware metadata for every Falcon agent."
---

# Table: crowdstrike_aid_master - Query CrowdStrike FDR AIDMaster

The `crowdstrike_aid_master` table allows you to query AIDMaster snapshots from CrowdStrike Falcon Data Replicator (FDR). Each row describes one agent (host) seen by Falcon, with its sensor version, operating system and hardware details.

CrowdStrike emits these snapshots several times a day, so the same agent appears many times. Use the most recent `time` per `aid` for the current state of a host.

## Configure

Create a [partition](https://tailpipe.io/docs/manage/partition) for `crowdstrike_aid_master` ([examples](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_aid_master#example-configurations)):

```sh
vi ~/.tailpipe/config/crowdstrike.tpc
```

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_aid_master" "my_hosts" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/aidmaster/"
  }
  tp_index = "cid"
}
```

## Collect

[Collect](https://tailpipe.io/docs/manage/collection) snapshots for all `crowdstrike_aid_master` partitions:

```sh
tailpipe collect crowdstrike_aid_master
```

Or for a single partition:

```sh
tailpipe collect crowdstrike_aid_master.my_hosts
```

## Query

**[Explore example queries for this table →](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/queries/crowdstrike_aid_master)**

### Latest record per host

Show the current state of each host from its most recent snapshot.

```sql
select
  aid,
  computer_name,
  event_platform,
  version,
  agent_version,
  time
from
  crowdstrike_aid_master
qualify
  row_number() over (partition by aid order by time desc) = 1
order by
  computer_name;
```

### Hosts by operating system

Count hosts per platform and OS version.

```sql
select
  event_platform,
  version,
  count(distinct aid) as host_count
from
  crowdstrike_aid_master
group by
  event_platform,
  version
order by
  host_count desc;
```

### Falcon sensor versions in use

See how far sensor upgrades have rolled out.

```sql
select
  agent_version,
  count(distinct aid) as host_count
from
  crowdstrike_aid_master
group by
  agent_version
order by
  host_count desc;
```

## Example Configurations

### Collect AIDMaster snapshots from an S3 bucket

Collect AIDMaster snapshots for a tenant, indexed by customer ID.

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_aid_master" "my_hosts" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/aidmaster/"
  }
  tp_index = "cid"
}
```

### Collect Windows hosts only

Use the partition `filter` to keep only Windows hosts.

```hcl
partition "crowdstrike_aid_master" "my_windows_hosts" {
  filter = "event_platform = 'Win'"

  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/aidmaster/"
  }
  tp_index = "cid"
}
```

### Collect AIDMaster snapshots from local files

Replay AIDMaster files downloaded outside Tailpipe.

```hcl
partition "crowdstrike_aid_master" "local_hosts" {
  source "file" {
    paths       = ["/Users/myuser/fdr/aidmaster"]
    file_layout = `%{DATA}.gz`
  }
  tp_index = "cid"
}
```

## Notes

- `time` (wire field `Time`) is when the snapshot was emitted and `first_seen` when the agent was first observed; both are parsed into `TIMESTAMP` columns. `tp_timestamp` uses `time`, falling back to `first_seen`.
- Other values stay as text, as delivered. The `payload` JSON column carries any field not promoted to a typed column.
- `aip`, `computer_name`, `machine_domain` and the location columns identify devices and their owners. Restrict access to your Tailpipe data accordingly.

## Source Defaults

### crowdstrike_s3_bucket

This table sets the following defaults for the [crowdstrike_s3_bucket source](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/sources/crowdstrike_s3_bucket#arguments):

| Argument    | Default |
|-------------|---------|
| file_layout | `(batch=)?%{DATA:batch}/(year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/)?(platform=%{DATA:platform}/)?%{DATA}.gz` |
