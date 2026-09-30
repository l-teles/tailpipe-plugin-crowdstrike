## Activity Examples

### Newly Seen Hosts

List agents first observed in the last 7 days, for example to confirm a rollout or spot unexpected installs.

```sql
select
  aid,
  computer_name,
  event_platform,
  first_seen
from
  crowdstrike_aid_master
where
  first_seen > current_timestamp - interval '7 days'
qualify
  row_number() over (partition by aid order by time desc) = 1
order by
  first_seen desc;
```

```yaml
folder: CrowdStrike
```

### Hosts Not Reporting Recently

Find hosts whose most recent snapshot is older than 7 days.

```sql
select
  aid,
  computer_name,
  event_platform,
  max(time) as last_snapshot
from
  crowdstrike_aid_master
group by
  aid,
  computer_name,
  event_platform
having
  max(time) < current_timestamp - interval '7 days'
order by
  last_snapshot asc;
```

```yaml
folder: CrowdStrike
```

## Operational Examples

### Latest Record per Host

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

```yaml
folder: CrowdStrike
```

### Hosts by Operating System

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

```yaml
folder: CrowdStrike
```

### Falcon Sensor Versions in Use

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

```yaml
folder: CrowdStrike
```

### Hosts by Hardware Model

Break the fleet down by manufacturer and model.

```sql
select
  system_manufacturer,
  system_product_name,
  count(distinct aid) as host_count
from
  crowdstrike_aid_master
group by
  system_manufacturer,
  system_product_name
order by
  host_count desc
limit 20;
```

```yaml
folder: CrowdStrike
```
