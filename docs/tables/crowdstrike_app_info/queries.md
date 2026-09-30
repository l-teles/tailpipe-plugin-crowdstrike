## Activity Examples

### Recently Installed Applications

List applications installed in the last 7 days.

```sql
select distinct
  aid,
  hostname,
  file_name,
  product_name,
  installation_timestamp
from
  crowdstrike_app_info
where
  installation_timestamp > current_timestamp - interval '7 days'
order by
  installation_timestamp desc;
```

```yaml
folder: CrowdStrike
```

## Detection Examples

### Applications with Detections

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

```yaml
folder: CrowdStrike
```

### Hosts with a Specific Hash

Find every host carrying a binary by its SHA-256, for example one flagged by threat intelligence.

```sql
select distinct
  aid,
  hostname,
  file_name,
  product_name
from
  crowdstrike_app_info
where
  sha256_hash_data = '0000000000000000000000000000000000000000000000000000000000000000';
```

```yaml
folder: CrowdStrike
```

## Operational Examples

### Top Applications by Install Footprint

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

```yaml
folder: CrowdStrike
```

### Versions of a Specific Application

See which versions of an application are deployed and on how many hosts.

```sql
select
  product_version,
  count(distinct aid) as host_count
from
  crowdstrike_app_info
where
  file_name = 'node'
group by
  product_version
order by
  host_count desc;
```

```yaml
folder: CrowdStrike
```

### Applications by Vendor

Count distinct applications per vendor.

```sql
select
  company_name,
  count(distinct product_name) as product_count,
  count(distinct aid) as host_count
from
  crowdstrike_app_info
where
  company_name is not null
  and company_name != '#Vendor'
group by
  company_name
order by
  host_count desc
limit 20;
```

```yaml
folder: CrowdStrike
```
