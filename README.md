# CrowdStrike Plugin for Tailpipe

[![tests](https://github.com/l-teles/tailpipe-plugin-crowdstrike/actions/workflows/test.yml/badge.svg)](https://github.com/l-teles/tailpipe-plugin-crowdstrike/actions/workflows/test.yml)
[![security](https://github.com/l-teles/tailpipe-plugin-crowdstrike/actions/workflows/security.yml/badge.svg)](https://github.com/l-teles/tailpipe-plugin-crowdstrike/actions/workflows/security.yml)
[![release](https://img.shields.io/github/v/release/l-teles/tailpipe-plugin-crowdstrike?include_prereleases&sort=semver)](https://github.com/l-teles/tailpipe-plugin-crowdstrike/releases)
[![license](https://img.shields.io/github/license/l-teles/tailpipe-plugin-crowdstrike)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/l-teles/tailpipe-plugin-crowdstrike)](https://goreportcard.com/report/github.com/l-teles/tailpipe-plugin-crowdstrike)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/l-teles/tailpipe-plugin-crowdstrike/badge)](https://scorecard.dev/viewer/?uri=github.com/l-teles/tailpipe-plugin-crowdstrike)

[Tailpipe](https://tailpipe.io) is an open-source CLI that collects logs and exposes them as SQL. This plugin reads [CrowdStrike Falcon Data Replicator](https://www.crowdstrike.com/) (FDR) data from an S3 bucket — primary sensor / external-API events plus the periodic secondary lookup snapshots — and surfaces them as five SQL tables.

## Tables

| Table | What it holds |
|---|---|
| `crowdstrike_fdr_event` | Primary FDR events. Sensor telemetry (`ProcessRollup2`, `EndOfProcess`, `DnsRequest`, …) and external-API events (`Event_ModuleSummaryInfoEvent`, `Event_AuthActivityAuditEvent`, …) share this table. Hot identifiers are typed columns; the full event JSON is preserved in a `payload` column. |
| `crowdstrike_aid_master` | AIDMaster — one row per agent (host) with sensor / OS / hardware metadata. |
| `crowdstrike_app_info` | AppInfo — installed-application inventory. |
| `crowdstrike_managed_assets` | ManagedAssets — network interface / gateway info per managed agent. |
| `crowdstrike_user_info` | UserInfo — local-account inventory per host. |

`NotManaged` is intentionally absent in v1 (no reference data to validate the schema against).

## Sources

| Source | Description |
|---|---|
| `crowdstrike_s3_bucket` | Reads `.txt.gz` / `.gz` files from the FDR bucket (or its `*-s3alias` access-point alias). Authenticates via the standard AWS credential chain. Default grok layout matches both the classic Hive-style (`batch=<uuid>/year=…/platform=…/`) and the newer flat (`<uuid>/`) FDR layouts. |
| `file` | SDK-provided local-file source. Use it to replay FDR files downloaded out-of-band, for testing or air-gapped review. |

## Requirements

- [Tailpipe](https://tailpipe.io/downloads) v0.7+
- Go 1.25+ (only if building from source; `toolchain` directive in `go.mod` auto-fetches a patched version)
- Read access to a CrowdStrike FDR S3 bucket — recommended: an IAM principal scoped to `s3:GetObject` + `s3:ListBucket` on the tenant prefix only

## Quick start

```bash
git clone https://github.com/l-teles/tailpipe-plugin-crowdstrike
cd tailpipe-plugin-crowdstrike
make install      # drops the plugin under ~/.tailpipe/plugins/hub.tailpipe.io/plugins/l-teles/crowdstrike@latest
```

Configure credentials any way the AWS SDK can find them — named profile, `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY`, SSO, IRSA, instance role. Then drop a `crowdstrike.tpc` into `~/.tailpipe/config/`:

```hcl
connection "crowdstrike" "default" {
  profile = "crowdstrike-fdr"
  region  = "eu-central-1"  # required for *-s3alias buckets — the bucket-region probe doesn't work for access-point aliases
}

partition "crowdstrike_fdr_event" "prod" {
  source "crowdstrike_s3_bucket" {
    connection = connection.crowdstrike.default
    bucket     = "cs-lion-cannon-XXXXXX-s3alias"
    prefix     = "<tenant-id>/data/"
  }
}

partition "crowdstrike_aid_master"     "prod" { source "crowdstrike_s3_bucket" { connection = connection.crowdstrike.default  bucket = "cs-lion-cannon-XXXXXX-s3alias"  prefix = "<tenant-id>/fdrv2/aidmaster/" } }
partition "crowdstrike_app_info"       "prod" { source "crowdstrike_s3_bucket" { connection = connection.crowdstrike.default  bucket = "cs-lion-cannon-XXXXXX-s3alias"  prefix = "<tenant-id>/fdrv2/appinfo/" } }
partition "crowdstrike_managed_assets" "prod" { source "crowdstrike_s3_bucket" { connection = connection.crowdstrike.default  bucket = "cs-lion-cannon-XXXXXX-s3alias"  prefix = "<tenant-id>/fdrv2/managedassets/" } }
partition "crowdstrike_user_info"      "prod" { source "crowdstrike_s3_bucket" { connection = connection.crowdstrike.default  bucket = "cs-lion-cannon-XXXXXX-s3alias"  prefix = "<tenant-id>/fdrv2/userinfo/" } }
```

Replay a local dump instead:

```hcl
partition "crowdstrike_fdr_event" "local" {
  source "file" {
    paths       = ["/path/to/fdr-samples"]
    file_layout = "%{DATA}.gz"
  }
}
```

Then collect and query:

```bash
tailpipe collect crowdstrike_fdr_event.prod --from T-1d
tailpipe query "select event_simple_name, count(*) from crowdstrike_fdr_event group by 1 order by 2 desc limit 20"
```

Per-table docs and example queries live under [`docs/tables/`](docs/tables/).

## Notes

- **Wire format** — every value in FDR JSON is delivered as a string (timestamps, integers, floats included). All row columns are `*string`; cast at query time, e.g. `cast(payload->>'$.RawProcessId' as bigint)` or `to_timestamp(cast(time as bigint))`.
- **Timestamps** — sensor `ContextTimeStamp` is epoch-seconds with optional fractional ms (`"1778159119.283"`); sensor `timestamp` and external-API `UTCTimestamp` are epoch-milliseconds; external-API `timestamp` is RFC3339. `tp_timestamp` resolves the best available.
- **`payload` JSON column** — every table carries one. Anything not promoted to a typed column stays queryable via `payload->>'$.Field'`.
- **PII** — `aip`, `LocalAddressIP4`, `UserName`, `User`, `MAC`, `ExternalIP`, `UserSid_readable`, and others are personally identifying. They are ingested as-is. Restrict access to the local DuckLake store and downstream queries; see [SECURITY.md](SECURITY.md).
- **Operator hardening** — use an IAM principal scoped to the tenant prefix only, not the whole bucket. Prefer profile / SSO / IRSA over static keys in HCL.

## Contributing & security

- Bugs and features → [GitHub Issues](https://github.com/l-teles/tailpipe-plugin-crowdstrike/issues).
- Code contributions → [CONTRIBUTING.md](CONTRIBUTING.md).
- Vulnerabilities → [SECURITY.md](SECURITY.md) (private vuln reporting, please).

## License

Apache-2.0. See [LICENSE](LICENSE).
