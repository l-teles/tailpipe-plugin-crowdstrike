---
title: "Tailpipe Table: crowdstrike_user_info - Query CrowdStrike FDR UserInfo"
description: "CrowdStrike FDR UserInfo snapshots: the accounts Falcon observes signing in on each host."
---

# Table: crowdstrike_user_info - Query CrowdStrike FDR UserInfo

The `crowdstrike_user_info` table allows you to query UserInfo snapshots from CrowdStrike Falcon Data Replicator (FDR). Each row is one account observed on one host, with its type, last logon and password age.

> **PII**: `user`, `user_name` and `user_sid_readable` identify people. Restrict access to your Tailpipe data accordingly.

## Configure

Create a [partition](https://tailpipe.io/docs/manage/partition) for `crowdstrike_user_info` ([examples](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_user_info#example-configurations)):

```sh
vi ~/.tailpipe/config/crowdstrike.tpc
```

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_user_info" "my_users" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/userinfo/"
  }
  tp_index = "cid"
}
```

## Collect

[Collect](https://tailpipe.io/docs/manage/collection) snapshots for all `crowdstrike_user_info` partitions:

```sh
tailpipe collect crowdstrike_user_info
```

Or for a single partition:

```sh
tailpipe collect crowdstrike_user_info.my_users
```

## Query

**[Explore example queries for this table →](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/queries/crowdstrike_user_info)**

### Local administrators

List accounts that are local administrators on a host.

```sql
select distinct
  user,
  last_logged_on_host,
  account_type
from
  crowdstrike_user_info
where
  user_is_admin;
```

### Logon activity by account

Summarise how each account signs in.

```sql
select
  user_name,
  account_type,
  logon_type,
  count(*) as record_count
from
  crowdstrike_user_info
group by
  user_name,
  account_type,
  logon_type
order by
  record_count desc
limit 25;
```

### Stale passwords

Find accounts whose password was last changed more than 180 days ago.

```sql
select distinct
  user,
  last_logged_on_host,
  password_last_set
from
  crowdstrike_user_info
where
  password_last_set < current_timestamp - interval '180 days'
order by
  password_last_set;
```

## Example Configurations

### Collect UserInfo snapshots from an S3 bucket

Collect UserInfo snapshots for a tenant, indexed by customer ID.

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_user_info" "my_users" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/userinfo/"
  }
  tp_index = "cid"
}
```

### Collect administrator accounts only

Use the partition `filter` to keep only local administrators.

```hcl
partition "crowdstrike_user_info" "my_admins" {
  filter = "user_is_admin"

  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/userinfo/"
  }
  tp_index = "cid"
}
```

### Collect UserInfo snapshots from local files

Replay UserInfo files downloaded outside Tailpipe.

```hcl
partition "crowdstrike_user_info" "local_users" {
  source "file" {
    paths       = ["/Users/myuser/fdr/userinfo"]
    file_layout = `%{DATA}.gz`
  }
  tp_index = "cid"
}
```

## Notes

- `time` (wire field `_time`), `logon_time` and `password_last_set` are parsed into `TIMESTAMP` columns; `"0"` means unknown and becomes null. `user_is_admin` is a `BOOLEAN`, and `months_since_reset` a `BIGINT` that is null when FDR sends `"N/A"`.
- This table has **no `aid` column**, because the wire format does not consistently include one. Use `last_logged_on_host` plus `cid` to associate rows with an agent, or join `crowdstrike_aid_master` on `cid` and `last_logged_on_host = computer_name`.

## Source Defaults

### crowdstrike_s3_bucket

This table sets the following defaults for the [crowdstrike_s3_bucket source](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/sources/crowdstrike_s3_bucket#arguments):

| Argument    | Default |
|-------------|---------|
| file_layout | `(batch=)?%{DATA:batch}/(year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/)?(platform=%{DATA:platform}/)?%{DATA}.gz` |
