## Detection Examples

### Local Administrators

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

```yaml
folder: CrowdStrike
```

### Accounts Administering Many Hosts

Find accounts with local administrator rights on more than five hosts, a common lateral-movement risk.

```sql
select
  user,
  count(distinct last_logged_on_host) as host_count
from
  crowdstrike_user_info
where
  user_is_admin
group by
  user
having
  count(distinct last_logged_on_host) > 5
order by
  host_count desc;
```

```yaml
folder: CrowdStrike
```

### Stale Passwords

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

```yaml
folder: CrowdStrike
```

## Activity Examples

### Recent Logons

List the most recent logon per account and host.

```sql
select
  user_name,
  last_logged_on_host,
  logon_type,
  max(logon_time) as last_logon
from
  crowdstrike_user_info
group by
  user_name,
  last_logged_on_host,
  logon_type
order by
  last_logon desc
limit 50;
```

```yaml
folder: CrowdStrike
```

### Logon Types by Account Type

Summarise how different kinds of account sign in.

```sql
select
  account_type,
  logon_type,
  count(distinct user_sid_readable) as account_count
from
  crowdstrike_user_info
group by
  account_type,
  logon_type
order by
  account_count desc;
```

```yaml
folder: CrowdStrike
```
