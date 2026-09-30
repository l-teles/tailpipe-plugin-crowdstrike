---
title: "Tailpipe Table: crowdstrike_fdr_event - Query CrowdStrike FDR Events"
description: "CrowdStrike FDR primary events: sensor telemetry and external-API events delivered by Falcon Data Replicator."
---

# Table: crowdstrike_fdr_event - Query CrowdStrike FDR Events

The `crowdstrike_fdr_event` table allows you to query primary events from CrowdStrike Falcon Data Replicator (FDR). Two kinds of event share this table:

1. **Sensor telemetry**: flat records keyed by `event_simpleName` (e.g. `ProcessRollup2`, `EndOfProcess`, `DnsRequest`, `NetworkConnectIP4`). These carry `aid`, `aip`, `event_platform` and `ContextTimeStamp`, plus dozens of event-specific fields.
2. **External-API events**: wrapped records with `EventType = "Event_ExternalApiEvent"`, a more specific `ExternalApiType` (e.g. `Event_ModuleSummaryInfoEvent`, `Event_AuthActivityAuditEvent`), and PascalCase identifiers (`AgentIdString`, `CustomerIdString`, `UTCTimestamp`).

Common identifiers are typed columns. Everything else, which varies a lot between events, stays in the JSON `payload` column; read an event-specific field with `payload->>'$.SomeField'`.

## Configure

Create a [partition](https://tailpipe.io/docs/manage/partition) for `crowdstrike_fdr_event` ([examples](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_fdr_event#example-configurations)):

```sh
vi ~/.tailpipe/config/crowdstrike.tpc
```

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

## Collect

[Collect](https://tailpipe.io/docs/manage/collection) events for all `crowdstrike_fdr_event` partitions:

```sh
tailpipe collect crowdstrike_fdr_event
```

Or for a single partition:

```sh
tailpipe collect crowdstrike_fdr_event.my_events
```

## Query

**[Explore example queries for this table →](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/queries/crowdstrike_fdr_event)**

### Top sensor event types

Find the most frequent sensor events.

```sql
select
  event_simple_name,
  count(*) as event_count
from
  crowdstrike_fdr_event
where
  event_simple_name is not null
group by
  event_simple_name
order by
  event_count desc
limit 10;
```

### DNS lookups by host

See which domains each host resolved.

```sql
select
  computer_name,
  payload->>'$.DomainName' as domain_name,
  count(*) as lookup_count
from
  crowdstrike_fdr_event
where
  event_simple_name = 'DnsRequest'
group by
  computer_name,
  domain_name
order by
  lookup_count desc
limit 20;
```

### Recent PowerShell executions

List recent processes whose command line mentions PowerShell.

```sql
select
  tp_timestamp,
  computer_name,
  payload->>'$.UserName' as user_name,
  payload->>'$.CommandLine' as command_line
from
  crowdstrike_fdr_event
where
  event_simple_name = 'ProcessRollup2'
  and lower(payload->>'$.CommandLine') like '%powershell%'
order by
  tp_timestamp desc
limit 20;
```

## Example Configurations

### Collect events from an S3 bucket

Collect primary events for a tenant, indexed by customer ID.

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

### Collect only selected event types

Use the partition `filter` to keep only the events you need, reducing local storage.

```hcl
partition "crowdstrike_fdr_event" "my_process_and_dns_events" {
  filter = "event_simple_name in ('ProcessRollup2', 'DnsRequest')"

  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/data/"
  }
  tp_index = "cid"
}
```

### Collect events from a date-first bucket layout

If FDR data is copied into your own bucket with the date partitions first, a matching `file_layout` lets `--from`/`--to` skip whole days.

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

### Collect events from local files

Replay FDR files downloaded outside Tailpipe (e.g. with `aws s3 cp`). Both gzipped and plain JSON-lines files work.

```hcl
partition "crowdstrike_fdr_event" "local_events" {
  source "file" {
    paths       = ["/Users/myuser/fdr/data"]
    file_layout = `%{DATA}.gz`
  }
  tp_index = "cid"
}
```

### Collect events with the AWS plugin's S3 source

With the [AWS plugin](https://hub.tailpipe.io/plugins/turbot/aws) installed, the `aws_s3_bucket` source also works. It has no FDR defaults, so set `file_layout` yourself.

```hcl
connection "aws" "fdr" {
  profile = "crowdstrike-fdr"
}

partition "crowdstrike_fdr_event" "my_events_via_aws" {
  source "aws_s3_bucket" {
    connection  = connection.aws.fdr
    bucket      = "my-fdr-bucket"
    prefix      = "<tenant-id>/data/"
    file_layout = `(batch=)?%{DATA:batch}/(year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/)?(platform=%{DATA:platform}/)?%{DATA}.gz`
  }
  tp_index = "cid"
}
```

## Wire-format notes

- **Timestamps**: FDR delivers every value as a string. The plugin parses the three event-time fields into `TIMESTAMP` columns. `context_time_stamp` (sensor events) arrives as epoch seconds with optional fractional milliseconds, `utc_timestamp` (external-API events) as epoch milliseconds, and `timestamp` as epoch milliseconds for sensor events or RFC3339 for external-API events. `tp_timestamp` is the first of `context_time_stamp`, `utc_timestamp` and `timestamp` that is present. A record with none of them is reported as a row error instead of being collected.
- **Other values** stay as text, as delivered. Cast them in SQL when you need numbers, e.g. `cast(payload->>'$.RawProcessId' as bigint)`.
- **Cross-fill**: external-API events have no top-level `aid`, so `aid` is null and you should use `agent_id_string`. `cid` is filled from `CustomerIdString` so it is always populated.
- **Large lines**: some FDR events exceed 1 MiB (e.g. `PeFileWritten`). Lines up to 16 MiB are collected; longer or malformed lines are reported as row errors in the collection summary.
- **PII**: `aip`, `computer_name`, and payload fields such as `UserName`, `UserSid` and `LocalAddressIP4` identify people and devices. Restrict access to your Tailpipe data accordingly.

## Source Defaults

### crowdstrike_s3_bucket

This table sets the following defaults for the [crowdstrike_s3_bucket source](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/sources/crowdstrike_s3_bucket#arguments):

| Argument    | Default |
|-------------|---------|
| file_layout | `(batch=)?%{DATA:batch}/(year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/)?(platform=%{DATA:platform}/)?%{DATA}.gz` |
