## Operational Examples

### Latest Interfaces per Host

Show each host's interfaces from its most recent snapshot.

```sql
select
  aid,
  interface_alias,
  local_address_ip4,
  gateway_ip,
  mac,
  time
from
  crowdstrike_managed_asset
qualify
  rank() over (partition by aid order by time desc) = 1
order by
  aid,
  interface_alias;
```

```yaml
folder: CrowdStrike
```

### Most Common Subnets

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

```yaml
folder: CrowdStrike
```

### Hosts per Gateway

Count hosts behind each default gateway.

```sql
select
  gateway_ip,
  count(distinct aid) as host_count
from
  crowdstrike_managed_asset
where
  gateway_ip is not null
group by
  gateway_ip
order by
  host_count desc
limit 20;
```

```yaml
folder: CrowdStrike
```

### Network Hardware Vendors

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

```yaml
folder: CrowdStrike
```

## Detection Examples

### Hosts Sharing a MAC Address

Find MAC addresses reported by more than one agent, which can indicate cloned images or spoofing.

```sql
select
  mac,
  count(distinct aid) as host_count,
  list(distinct aid) as aids
from
  crowdstrike_managed_asset
where
  mac is not null
group by
  mac
having
  count(distinct aid) > 1
order by
  host_count desc;
```

```yaml
folder: CrowdStrike
```
