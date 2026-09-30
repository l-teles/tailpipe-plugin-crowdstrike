---
organization: l-teles
category: ["security"]
icon_url: "/images/plugins/l-teles/crowdstrike.svg"
brand_color: "#FC0000"
display_name: "CrowdStrike"
description: "Tailpipe plugin for collecting and querying CrowdStrike Falcon Data Replicator (FDR) data."
og_description: "Collect CrowdStrike FDR data and query it instantly with SQL! Open source CLI. No DB required."
og_image: "/images/plugins/l-teles/crowdstrike-social-graphic.png"
---

# CrowdStrike + Tailpipe

[Tailpipe](https://tailpipe.io) is an open-source CLI tool that allows you to collect logs and query them with SQL.

[CrowdStrike Falcon](https://www.crowdstrike.com/) is an endpoint protection platform. Its Falcon Data Replicator (FDR) delivers raw sensor telemetry and periodic inventory snapshots to an S3 bucket.

The [CrowdStrike Plugin for Tailpipe](https://hub.tailpipe.io/plugins/l-teles/crowdstrike) allows you to collect and query FDR data using SQL to hunt threats, investigate hosts, audit software and accounts, and more!

- Documentation: [Table definitions & examples](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables)
- Get involved: [Issues](https://github.com/l-teles/tailpipe-plugin-crowdstrike/issues)

## Getting Started

Install Tailpipe from the [downloads](https://tailpipe.io/downloads) page:

```sh
# MacOS
brew install turbot/tap/tailpipe
```

```sh
# Linux or Windows (WSL)
sudo /bin/sh -c "$(curl -fsSL https://tailpipe.io/install/tailpipe.sh)"
```

Install the plugin:

```sh
tailpipe plugin install l-teles/crowdstrike
```

Configure your [connection credentials](https://hub.tailpipe.io/plugins/l-teles/crowdstrike#connection-credentials), table partition, and data source ([examples](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_fdr_event#example-configurations)):

```sh
vi ~/.tailpipe/config/crowdstrike.tpc
```

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}

partition "crowdstrike_fdr_event" "my_events" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.fdr
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/data/"
  }
  tp_index = "cid"
}
```

Download, enrich, and save data from your source ([examples](https://tailpipe.io/docs/reference/cli/collect)):

```sh
tailpipe collect crowdstrike_fdr_event --from T-1d
```

Enter interactive query mode:

```sh
tailpipe query
```

Run a query:

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

```sh
+-------------------------+-------------+
| event_simple_name       | event_count |
+-------------------------+-------------+
| EndOfProcess            | 4381902     |
| ProcessRollup2          | 4012577     |
| NetworkConnectIP4       | 596214      |
| DnsRequest              | 441830      |
| ProcessRollup2Stats     | 188407      |
| DirectoryCreate         | 161293      |
| TlsClientHello          | 117046      |
| SyntheticProcessRollup2 | 109875      |
| AppProtocolDetected     | 98312       |
| CriticalFileAccessed    | 80764       |
+-------------------------+-------------+
```

## Tables

| Table | Contents |
|---|---|
| [crowdstrike_fdr_event](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_fdr_event) | Primary FDR events: sensor telemetry and external-API events. |
| [crowdstrike_aid_master](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_aid_master) | AIDMaster snapshots: sensor, OS and hardware metadata per host. |
| [crowdstrike_app_info](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_app_info) | AppInfo snapshots: installed-application inventory. |
| [crowdstrike_managed_asset](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_managed_asset) | ManagedAssets snapshots: network interface and gateway info per host. |
| [crowdstrike_user_info](https://hub.tailpipe.io/plugins/l-teles/crowdstrike/tables/crowdstrike_user_info) | UserInfo snapshots: account inventory per host. |

FDR data contains personal data (IP and MAC addresses, hostnames, usernames, SIDs). It is collected as-is, so restrict access to your local Tailpipe data directory accordingly.

## Connection Credentials

The `crowdstrike` connection holds the AWS credentials used to read your FDR bucket. It follows the standard AWS credential chain: static keys, then a named profile, then environment variables, then SSO, IRSA or an instance role.

### Arguments

| Name            | Type   | Required | Description |
|-----------------|--------|----------|-------------|
| `access_key`    | String | No       | AWS access key ID. Must be set together with `secret_key`. |
| `endpoint_url`  | String | No       | Custom S3 endpoint URL (e.g. for local testing). |
| `profile`       | String | No       | AWS CLI profile to use for credentials and configuration. |
| `region`        | String | No       | AWS region of the bucket. Required for `*-s3alias` access-point aliases and re-signing proxies; recommended otherwise. |
| `secret_key`    | String | No       | AWS secret access key. Must be set together with `access_key`. |
| `session_token` | String | No       | AWS session token for temporary credentials, used with `access_key` and `secret_key`. |

### AWS Profile Credentials

Named profiles are the most common setup, and also cover SSO and role assumption:

```hcl
connection "crowdstrike" "fdr" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"
}
```

For SSO profiles, run `aws sso login` before `tailpipe collect`; Tailpipe cannot re-authenticate you when credentials expire.

### Credentials from Environment Variables

With no `profile` or keys set, the connection uses the standard AWS environment variables:

```sh
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_REGION=eu-central-1
```

Prefer a profile, SSO or an instance role over static keys in HCL, and scope the IAM principal to `s3:GetObject` and `s3:ListBucket` on your tenant prefix only.
