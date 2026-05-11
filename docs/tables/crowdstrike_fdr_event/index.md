# crowdstrike_fdr_event

Primary CrowdStrike Falcon Data Replicator (FDR) events. Two flavours share this table:

1. **Sensor telemetry** — flat records keyed by `event_simpleName` (e.g. `ProcessRollup2`, `EndOfProcess`, `DnsRequest`, `NetworkConnectIP4`). These carry `aid`, `aip`, `event_platform`, `ContextTimeStamp`, plus dozens of event-specific fields.
2. **External-API events** — wrapped records with `EventType = "Event_ExternalApiEvent"`, a more specific `ExternalApiType` (e.g. `Event_ModuleSummaryInfoEvent`, `Event_AuthActivityAuditEvent`), and PascalCase identifiers (`AgentIdString`, `CustomerIdString`, `UTCTimestamp`).

Hot identifiers are typed columns; everything else (which varies dramatically across events) lives in the JSON `payload` column. Cast or extract via `payload->>'$.SomeField'` when you need an event-specific field.

## Quick examples

### Top sensor event types

```sql
select event_simple_name, count(*) as c
from crowdstrike_fdr_event
where event_simple_name is not null
group by 1
order by 2 desc
limit 10;
```

### External-API event breakdown

```sql
select external_api_type, count(*) as c
from crowdstrike_fdr_event
where external_api_type is not null
group by 1
order by 2 desc;
```

### DNS lookups by host

```sql
select
  computer_name,
  payload->>'$.DomainName' as domain,
  count(*) as c
from crowdstrike_fdr_event
where event_simple_name = 'DnsRequest'
group by 1, 2
order by 3 desc
limit 20;
```

### Process executions referencing PowerShell

```sql
select
  tp_timestamp,
  computer_name,
  payload->>'$.UserName' as user_name,
  payload->>'$.CommandLine' as cmd
from crowdstrike_fdr_event
where event_simple_name = 'ProcessRollup2'
  and lower(payload->>'$.CommandLine') like '%powershell%'
order by tp_timestamp desc
limit 20;
```

## Wire-format notes

- **Every value is a string on the wire** — including timestamps, integers, and floats. All columns are `*string`. Cast in SQL when you need numeric semantics, e.g. `cast(payload->>'$.RawProcessId' as bigint)`.
- **Timestamps**: sensor `ContextTimeStamp` is epoch-seconds with optional fractional milliseconds (`"1778159119.283"`); sensor `timestamp` is epoch-milliseconds (`"1778159121826"`); external-API `UTCTimestamp` is epoch-milliseconds; external-API `timestamp` is RFC3339 (`"2026-05-07T12:59:18Z"`). The plugin resolves `tp_timestamp` from these in priority order.
- **Cross-fill**: external-API events have no top-level `aid` — `aid` will be `NULL` and you should use `agent_id_string` instead. `cid` is cross-filled from `CustomerIdString` so it is always populated.
- **Large lines**: some FDR events exceed 1 MiB (e.g. `PeFileWritten` with embedded info). The extractor uses a 16 MiB scanner buffer so these aren't truncated.

## Configuration

### From a CrowdStrike S3 bucket

```hcl
connection "crowdstrike" "default" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_fdr_event" "prod" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.default
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/data/"
  }
}
```

### From local files (replay)

```hcl
partition "crowdstrike_fdr_event" "local" {
  source "file" {
    paths       = ["/path/to/fdr-samples"]
    file_layout = "%{DATA}.txt.gz"
  }
}
```
