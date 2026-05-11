<!--
Thanks for the PR! Please:
1. Label this PR with exactly one of:
   breaking · feature · fix · security · docs · chore · refactor · test · ci · deps
   (drives the Release Drafter category and the next version bump).
2. Fill in the sections below.
-->

## Summary

<!-- One or two sentences on WHAT changed and WHY. -->

## Changes

<!--
Bulleted list of user-visible changes only. Do not enumerate every file —
the diff already does that.
-->
- 

## Compatibility

<!-- Any breaking changes to HCL config, SQL column names, or behaviour? -->
- [ ] No breaking changes
- [ ] Breaking — described above, with migration notes

## Tests

<!--
What did you test? Unit tests are required for any new table/extractor/source.
For end-to-end changes, mention if you ran `tailpipe collect` against a real
or local fixture.
-->
- [ ] `go test ./...` passes locally
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run` clean
- [ ] New tests added for new behaviour (or N/A)
- [ ] Sample fixture sanitised (no real cids / aids / IPs / MACs / hostnames / usernames / emails)

## Security

- [ ] No new external dependencies (or: justified in description)
- [ ] No secrets added; new env / HCL fields documented in `README.md` / table docs
- [ ] If the PR touches the S3 source or any path-joining code, the `validateArtifactKey` test still covers the new path
