# crowdstrike_app_info

Application inventory observed by Falcon — one row per (agent, application) tuple from the FDR AppInfo lookup. Useful for hunting unwanted software, building a software bill-of-materials, or checking deployment of a specific tool across the fleet.

## Quick examples

### Hosts running a specific binary

```sql
select aid, hostname, file_version, product_version
from crowdstrike_app_info
where file_name = 'node';
```

### Top apps by install footprint

```sql
select product_name, count(distinct aid) as hosts
from crowdstrike_app_info
where product_name is not null and product_name != '#Application'
group by 1
order by 2 desc
limit 25;
```

### Apps with detection events recorded

```sql
select aid, file_name, sha256_hash_data, detection_count
from crowdstrike_app_info
where cast(detection_count as double) > 0;
```

## Notes

- `time` is delivered as `_time` on the wire (epoch seconds string).
- Vendor / product strings are sometimes literal placeholders like `"#Vendor"`, `"#Application"`, `"#Version"` when CrowdStrike couldn't extract a real value.

## Configuration

```hcl
partition "crowdstrike_app_info" "prod" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.default
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/appinfo/"
  }
}
```
