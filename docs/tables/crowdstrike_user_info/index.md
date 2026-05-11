# crowdstrike_user_info

Local-account inventory observed on each Falcon-managed host — one row per (host, account) tuple from the FDR UserInfo lookup.

> **PII** — `user`, `user_name`, and `user_sid_readable` are personally-identifying. Treat collected data accordingly.

## Quick examples

### Local administrators

```sql
select user, last_logged_on_host, account_type
from crowdstrike_user_info
where user_is_admin = '1';
```

### Domain account login activity

```sql
select user_name, account_type, logon_type, count(*) as c
from crowdstrike_user_info
group by 1, 2, 3
order by 4 desc
limit 25;
```

### Stale passwords (>180 days, where known)

```sql
select user, last_logged_on_host, password_last_set
from crowdstrike_user_info
where cast(password_last_set as bigint) > 0
  and to_timestamp(cast(password_last_set as bigint)) < (current_timestamp - interval '180 days');
```

## Notes

- `time` is delivered as `_time` on the wire; `logon_time` and `password_last_set` are also epoch-seconds strings (`"0"` means unknown).
- This table has **no `aid` column** — the wire format does not consistently include one. Use `last_logged_on_host` plus `cid` to associate rows with an agent (or join through `crowdstrike_aid_master` on `computer_name`).

## Configuration

```hcl
partition "crowdstrike_user_info" "prod" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.default
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/fdrv2/userinfo/"
  }
}
```
