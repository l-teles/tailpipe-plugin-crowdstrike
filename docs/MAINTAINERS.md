# Maintainer checklist

One-off settings that can't be encoded as files in the repo. Apply these the
first time the repo is created on GitHub. Keep this list in sync with the
repo's actual state.

## 1. Repo metadata

- **Description**: `Tailpipe plugin that ingests CrowdStrike Falcon Data Replicator (FDR) data from S3 and exposes it as SQL.`
- **Homepage**: (leave empty for now; point to docs site once one exists)
- **Topics** (Repository settings → About → Topics):
  - `tailpipe`
  - `tailpipe-plugin`
  - `crowdstrike`
  - `falcon`
  - `fdr`
  - `siem`
  - `duckdb`
  - `sql`
  - `go`

## 2. General settings (Settings → General)

- Default branch: **`main`**
- Issues: **enabled**
- Discussions / Projects / Wikis / Sponsorships: **disabled**
- Preserve this repository: **disabled** (until archived)
- **Pull requests**: squash only; delete branch on merge; auto-merge off; suggest updating PR branches.
- **Pushes**: ☑ Require contributors to sign off on web-based commits.

## 3. Security settings (Settings → Code security)

All free on a public repo:

- ☑ Private vulnerability reporting
- ☑ Dependency graph
- ☑ Dependabot alerts + security updates + version updates (config in `.github/dependabot.yml`)
- ☑ Secret scanning + push protection
- ☑ Code scanning with CodeQL (we ship `.github/workflows/codeql.yml`)

## 4. Branch / tag rulesets

> **Important — security note.** Rulesets describe *exactly which CI checks
> must pass to merge*, which bypass actors are allowed, and which branch
> patterns are protected. Publishing the JSON in a public repo gives an
> attacker a roadmap. **Keep the actual ruleset JSON outside this repo.**

Rulesets live in the maintainer's private store (e.g.
`~/.local/share/repo-private-config/<repo>/`). They are imported via:

```bash
gh api -X POST "repos/<OWNER>/<REPO>/rulesets" --input <private-path>/main-branch.ruleset.json
gh api -X POST "repos/<OWNER>/<REPO>/rulesets" --input <private-path>/tag-protection.ruleset.json
```

### Policy goal (describe-not-disclose)

The policy enforces, on `main`:
- No direct pushes (PR required, 1 approving review, code-owner review, last-push approval, stale-review dismissal, conversation-resolution).
- Linear history, no force-push, no deletion.
- All required-status-check contexts must succeed before merge (the specific list lives in the private ruleset JSON so attackers can't trivially target which workflows to subvert).
- Squash-only merge method.

On `v*` tags:
- No deletion, force-push, or update once published.

Required-status-check contexts must match the workflow `name:` values exactly. Re-export the ruleset JSON when a workflow is renamed.

## 5. Actions settings (Settings → Actions → General)

- **Actions permissions**: All actions and reusable workflows allowed; **require SHA pinning** for actions and reusable workflows (✅ on free plan).
- **Workflow permissions**: default `GITHUB_TOKEN` = read-only (workflows elevate per job).
- ☐ Allow GitHub Actions to create and approve pull requests.
- ☑ Require approval for first-time contributors' PR workflows from forks.

## 6. Environment-gated secrets (if/when we add integration tests against a real FDR bucket)

Create an Environment (e.g. `e2e-prod`) with:
- Deployment branch rule: only `main`.
- Required reviewers: at least one.
- AWS credentials live here, not at repo-wide scope.
- Reference via `environment: e2e-prod` on the relevant job.

## 7. Tagging a release

1. Open a PR that bumps `CHANGELOG.md`'s `[Unreleased]` section to a new version header.
2. After merge, tag from `main`:
   ```bash
   git tag -s vX.Y.Z -m "vX.Y.Z"
   git push origin vX.Y.Z
   ```
3. The `release.yml` workflow builds binaries via goreleaser and attaches them to a **draft** GitHub release.
4. Review the draft and publish it manually.
