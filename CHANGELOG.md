# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial plugin scaffold.
- Tables: `crowdstrike_fdr_event` (primary FDR events — sensor telemetry + external-API), `crowdstrike_aid_master`, `crowdstrike_app_info`, `crowdstrike_managed_assets`, `crowdstrike_user_info`.
- Sources: `crowdstrike_s3_bucket` (S3 bucket / access-point alias) and the SDK's built-in `file` source.
- Default grok layout covers both FDR variants: classic Hive-style (`batch=<uuid>/year=…/platform=…/part-*.txt.gz`) and the newer flat layout (`<uuid>/part-*.gz`).
- Unit tests across extractors, `EnrichRow`, and the grok layout expansion.
- Security: S3-key validator rejecting absolute paths, parent-directory segments, and NUL bytes before joining onto the local temp directory.
- Hub docs: `docs/index.md`, per-table `queries.md`, and a CC BY-NC-ND 4.0 licence for `docs/`.

### Changed
- **Breaking:** `crowdstrike_managed_assets` is renamed to `crowdstrike_managed_asset`.
- **Breaking:** timestamps, counts and flags are typed columns (`TIMESTAMP`, `BIGINT`, `BOOLEAN`) instead of strings. In `crowdstrike_fdr_event`, `timestamp_raw` and `utc_timestamp_raw` become `timestamp` and `utc_timestamp`. Delete and re-collect any existing partitions.
- Files are streamed line by line instead of loaded whole; malformed lines, lines over 16 MiB and records without a timestamp are reported as row errors instead of being skipped or stamped with the collection time.
- Release builds enable CGO with per-target cross compilers (required by go-duckdb) and run in the `goreleaser-cross` image.
