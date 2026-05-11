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
- Discussions: **disabled** (re-enable if user demand emerges)
- Projects: **disabled**
- Wikis: **disabled**
- Sponsorships: **disabled**
- Preserve this repository: **disabled** (until archived)
- Template repository: **disabled**
- **Pull Requests**:
  - ☑ Allow squash merging (default commit title: PR title)
  - ☐ Allow merge commits
  - ☐ Allow rebase merging
  - ☑ Always suggest updating pull request branches
  - ☑ Allow auto-merge **only after** rulesets in §4 are active
  - ☑ Automatically delete head branches
- **Pushes**:
  - ☑ Require contributors to sign off on web-based commits (`web_commit_signoff_required: true`)

## 3. Security settings (Settings → Code security)

- ☑ **Private vulnerability reporting**
- ☑ **Dependency graph**
- ☑ **Dependabot alerts**
- ☑ **Dependabot security updates**
- ☑ **Dependabot version updates** (config in `.github/dependabot.yml`)
- ☑ **Secret scanning**
- ☑ **Push protection** (block commits containing secrets)
- ☑ **Code scanning** with **CodeQL** (default setup or via `.github/workflows/codeql.yml` — we ship the latter)

## 4. Branch / tag rulesets (Settings → Rules → Rulesets → New ruleset → Import)

Two ruleset JSONs ship in the repo:

```bash
gh api -X POST "repos/l-teles/tailpipe-plugin-crowdstrike/rulesets" \
  --input .github/rulesets/main-branch.json

gh api -X POST "repos/l-teles/tailpipe-plugin-crowdstrike/rulesets" \
  --input .github/rulesets/tag-protection.json
```

After import, verify each shows **Status: Active** in the UI. Required status-check contexts must match the workflow `name:` values exactly — keep `main-branch.json` in sync if a workflow is renamed.

## 5. Actions settings (Settings → Actions → General)

- **Actions permissions**: Allow `l-teles` actions and reusable workflows; for third-party: ☑ Allow actions created by GitHub + ☑ Allow specified actions and reusable workflows (paste the SHAs from our workflows if you want to be strict).
- **Require SHA pinning** for actions and reusable workflows: ☑ (the API equivalent is `sha_pinning_required: true`).
- **Workflow permissions**: ☑ **Read repository contents and packages permissions** (default `GITHUB_TOKEN` is read-only; each workflow elevates per job).
- **Allow GitHub Actions to create and approve pull requests**: ☐ (off — defence in depth)
- **Fork PR workflows from outside collaborators**: ☑ Require approval for first-time contributors.

## 6. Environment-gated secrets (if/when we add integration tests against real FDR)

If we ever add CI that touches a real S3 bucket / AWS account:

- Create an **Environment** named `e2e-prod` (Settings → Environments).
- Add deployment branch rule: only `main`.
- Add required reviewers (at least one).
- Store the AWS credentials there, not in repo-wide secrets.
- Reference via `environment: e2e-prod` on the relevant job.

## 7. Email forwarding for private vuln reports (optional)

If you want PVR submissions to also reach a personal email, add a **Security advisory contact** in Settings → Security advisories. Not strictly necessary — GitHub already notifies repo admins.
