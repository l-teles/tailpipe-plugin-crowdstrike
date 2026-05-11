# crowdstrike_aid_master

Periodic AIDMaster snapshots from FDR. One row per agent (host) seen by Falcon, with sensor / OS / hardware metadata. CrowdStrike emits these snapshots multiple times per day, so a single agent appears many times — use the most recent `time` per `aid` for "current" host state.

## Quick examples

### Latest record per host

```sql
select *
from crowdstrike_aid_master
qualify row_number() over (partition by aid order by tp_timestamp desc) = 1;
```

### Host count by OS / platform

```sql
select event_platform, version, count(*) as c
from crowdstrike_aid_master
group by 1, 2
order by 3 desc;
```

### Falcon sensor versions in use

```sql
select agent_version, count(distinct aid) as hosts
from crowdstrike_aid_master
group by 1
order by 2 desc;
```

## Notes

- **`time`** is delivered as the wire field `Time` (epoch seconds string). `first_seen` is when the agent was first observed.
- All values are strings on the wire — cast when needed.
- The `payload` JSON column carries any field not promoted to a typed column.

## Configuration

```hcl
partition "crowdstrike_aid_master" "prod" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.default
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/aidmaster/"
  }
}

# Local replay
partition "crowdstrike_aid_master" "local" {
  source "file" {
    paths       = ["/path/to/aidmaster-samples"]
    file_layout = "%{DATA}.txt.gz"
  }
}
```
