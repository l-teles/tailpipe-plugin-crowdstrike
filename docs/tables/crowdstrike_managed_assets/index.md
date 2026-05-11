# crowdstrike_managed_assets

Network interface and gateway info per Falcon-managed agent — one row per (agent, interface) tuple from the FDR ManagedAssets lookup.

## Quick examples

### Hosts behind a specific gateway

```sql
select aid, local_address_ip4, gateway_ip
from crowdstrike_managed_assets
where gateway_ip = '192.0.2.1';
```

### Most common subnets

```sql
select
  regexp_extract(local_address_ip4, '^(\d+\.\d+\.\d+)\.', 1) as subnet,
  count(distinct aid) as hosts
from crowdstrike_managed_assets
where local_address_ip4 is not null
group by 1
order by 2 desc
limit 20;
```

### MAC vendors (OUI prefixes)

```sql
select mac_prefix, count(*) as c
from crowdstrike_managed_assets
group by 1
order by 2 desc
limit 20;
```

## Notes

- `time` is delivered as `_time` on the wire (epoch seconds string).
- The schema captures the *managed* side only; for hosts seen on the network without a Falcon agent, see the future `crowdstrike_not_managed` table (not in v1).

## Configuration

```hcl
partition "crowdstrike_managed_assets" "prod" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.default
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/managedassets/"
  }
}
```
