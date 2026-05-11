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
