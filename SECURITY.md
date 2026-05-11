# Security Policy

## Supported Versions

The latest tagged release is the only supported version. Older versions receive no fixes; please upgrade.

| Version | Supported |
| ------- | --------- |
| latest  | ✅        |
| < latest | ❌       |

## Reporting a Vulnerability

**Do not open a public GitHub issue for security-impacting findings.**

Use GitHub's private vulnerability reporting:

1. Visit the repository's **Security** tab.
2. Click **Report a vulnerability** (the green button under *Advisories*).
3. Include:
   - Affected version(s) / commit SHA.
   - Reproduction steps (minimal config + the FDR file shape that triggers it, if relevant — please sanitise PII).
   - Impact assessment: what an attacker controlling the input or environment could achieve.
   - Any fix or mitigation you have in mind.

I will acknowledge within **72 hours** and aim to publish a fix within **14 days** for high-severity findings. You will be credited in the advisory unless you ask to remain anonymous.

## Scope

In scope:
- Vulnerabilities in this plugin's Go source code under `crowdstrike/`, `crowdstrikeconfig/`, `sources/`, and `tables/`.
- Vulnerabilities introduced by direct dependencies in `go.mod` that we can mitigate (replace directive, version bump, validation at our boundary).

Out of scope:
- Issues in CrowdStrike's FDR product itself — report those to CrowdStrike directly.
- Issues in the Tailpipe core or SDK — report to [turbot/tailpipe](https://github.com/turbot/tailpipe).
- Tenant-data sensitivity concerns: this plugin ingests PII and security telemetry as-is from FDR; redaction is the operator's responsibility. See `README.md` for the PII inventory.

## Hardening recommendations for operators

- Run the plugin with the **least-privileged IAM principal** that can `s3:GetObject`/`s3:ListBucket` on the tenant prefix only — not bucket-wide.
- Use an AWS profile / SSO / IRSA over static `access_key`/`secret_key` in HCL where possible.
- Restrict read access to the local DuckLake parquet store (`~/.tailpipe/data`) — it contains the same PII as the source bucket.
- Keep the Go toolchain current; the `toolchain` directive in `go.mod` pins the minimum patched version for known stdlib vulns.
