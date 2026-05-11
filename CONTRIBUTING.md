# Contributing

Thanks for considering a contribution. This plugin is small and the bar is straightforward: add tests, keep the build green, follow the conventions below.

## Development environment

- Go 1.25+ (`toolchain` directive in `go.mod` auto-fetches a patched version)
- [Tailpipe](https://tailpipe.io/downloads) CLI (for end-to-end testing)
- Access to a CrowdStrike FDR S3 bucket (or local FDR-format sample files) if you want to run the remote-source path

```bash
git clone https://github.com/l-teles/tailpipe-plugin-crowdstrike
cd tailpipe-plugin-crowdstrike
go mod download
make install   # builds + drops the .plugin binary in ~/.tailpipe/plugins/...
```

## Running tests

```bash
go vet ./...
go test ./...          # all unit tests; runs in <5s
go test ./... -race    # with race detector
```

Each new table or source must come with:
- A `testdata/sample.jsonl` fixture (sanitised — no real cids, aids, IPs, MACs, hostnames, usernames, or emails — see the `stable_hex` pattern used in existing fixtures).
- A `_test.go` that exercises the extractor and `EnrichRow` end-to-end against the fixture.

## Code style

- Go: `gofmt` + `golangci-lint run`. Pre-commit hooks (`.pre-commit-config.yaml`) enforce both.
- No `*.md` planning docs unless explicitly asked.
- Comments only where the "why" isn't obvious from the code. Don't restate the "what".
- For row structs, hot identifiers stay typed columns; less-common fields live in `payload` (JSON).
- Every column tag must match the FDR wire-format key case-sensitively. Capitalisation mistakes silently drop fields on unmarshal.

## Commit messages

Conventional Commits:
- `feat: add crowdstrike_not_managed table`
- `fix: handle FDR primary events with empty ContextTimeStamp`
- `docs: clarify access-point alias requirement`
- `ci: pin actions/checkout to commit SHA`
- `deps: bump tailpipe-plugin-sdk to v0.10.0`
- `test: add fixture for ProcessRollup2 with command line`

## Pull-request flow

1. Fork → branch → push.
2. Open the PR. The PR template (`.github/pull_request_template.md`) covers what to fill in.
3. Label your PR with one of `breaking` / `feature` / `fix` / `chore` / `docs` so Release Drafter categorises it correctly.
4. CI must be green: `test`, `golangci-lint`, `security`, `codeql`. Status checks are required by the repo ruleset.
5. Squash-merge once approved.

## Adding a new table

1. New package under `tables/<table_name>/`.
2. Four files: `<table>.go` (row struct), `<table>_table.go` (Identifier + GetSourceMetadata + EnrichRow), `<table>_extractor.go` (uses `common.ExtractJSONLines`), `<table>_test.go`.
3. Sanitised `testdata/sample.jsonl`.
4. Register it in `crowdstrike/plugin.go`.
5. Add `docs/tables/<table_name>/index.md` with HCL examples (both `crowdstrike_s3_bucket` and `file` sources) plus a couple of useful SQL queries.

## Security-impacting changes

If you find or fix a vulnerability, do not open a public PR. Follow `SECURITY.md`.

## License

Apache-2.0. By contributing you agree your changes are licensed under the same terms.
