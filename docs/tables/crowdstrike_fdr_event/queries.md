## Activity Examples

### Daily Event Trends

Count events per day to establish a baseline and spot unusual spikes or gaps in sensor telemetry.

```sql
select
  strftime(tp_timestamp, '%Y-%m-%d') as event_date,
  count(*) as event_count
from
  crowdstrike_fdr_event
group by
  event_date
order by
  event_date asc;
```

```yaml
folder: CrowdStrike
```

### Top 10 Sensor Event Types

Rank sensor events by volume to see which telemetry dominates your data.

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

```yaml
folder: CrowdStrike
```

### External-API Event Breakdown

Break external-API events down by subtype.

```sql
select
  external_api_type,
  count(*) as event_count
from
  crowdstrike_fdr_event
where
  external_api_type is not null
group by
  external_api_type
order by
  event_count desc;
```

```yaml
folder: CrowdStrike
```

### Events by Platform

Compare event volume across operating system families.

```sql
select
  event_platform,
  count(*) as event_count,
  count(distinct aid) as host_count
from
  crowdstrike_fdr_event
where
  event_platform is not null
group by
  event_platform
order by
  event_count desc;
```

```yaml
folder: CrowdStrike
```

## Detection Examples

### PowerShell Executions

Find processes whose command line mentions PowerShell, a frequent vehicle for attacker tooling.

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

```yaml
folder: CrowdStrike
```

### Encoded PowerShell Commands

Look for PowerShell launched with an encoded command, which is often used to hide what a script does.

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
  and regexp_matches(lower(payload->>'$.CommandLine'), '\s-(e|en|enc|enco|encod|encode|encoded|encodedcommand)\s')
order by
  tp_timestamp desc;
```

```yaml
folder: CrowdStrike
```

### Processes Run by a Specific Hash

Find every execution of a binary by its SHA-256, for example one flagged by threat intelligence.

```sql
select
  tp_timestamp,
  computer_name,
  aid,
  payload->>'$.CommandLine' as command_line
from
  crowdstrike_fdr_event
where
  event_simple_name = 'ProcessRollup2'
  and (payload->>'$.SHA256HashData') = '0000000000000000000000000000000000000000000000000000000000000000'
order by
  tp_timestamp desc;
```

```yaml
folder: CrowdStrike
```

## Operational Examples

### Hosts by Event Volume

Identify the noisiest hosts, which can point to misbehaving software or ongoing activity.

```sql
select
  computer_name,
  aid,
  count(*) as event_count
from
  crowdstrike_fdr_event
where
  aid is not null
group by
  computer_name,
  aid
order by
  event_count desc
limit 20;
```

```yaml
folder: CrowdStrike
```

### Most Resolved Domains

List the domains most often looked up across the fleet.

```sql
select
  payload->>'$.DomainName' as domain_name,
  count(*) as lookup_count,
  count(distinct aid) as host_count
from
  crowdstrike_fdr_event
where
  event_simple_name = 'DnsRequest'
group by
  domain_name
order by
  lookup_count desc
limit 20;
```

```yaml
folder: CrowdStrike
```
