# Environments, secrets and migrations

The environments the game runs in, every environment variable the code reads, where secrets live and how they are rotated, the secret scan, and how database schema changes reach each environment (#24).

Scope boundaries: deploy, rollback and orchestration are #25 and `docs/launch-runbook.md` section 5. Backups and restore are #188. Observability is #28. Admin config (every number a designer tunes) is never an environment variable; it is versioned in admin (ADR 0006).

## 1. Environments

| Environment | Purpose | Data | Who reaches it |
|---|---|---|---|
| `dev` | One developer's machine: the local stack of #22 and the in-process tests | Synthetic seed only (`docs/test-strategy.md` 7.2) | The developer |
| `staging` | Release candidates, drills (`docs/launch-runbook.md` GN-9 to GN-19), device smoke, load and soak, admin change sets tried on staff test accounts (`docs/admin.md` 3.4) | Synthetic seed and staff test accounts; sized like production for load runs | Team accounts with staging roles |
| `production` | Players | Player data under `docs/privacy.md` | Players; the `eng` role and admin roles granted for production only (`docs/admin.md` section 2, `docs/privacy.md` section 3) |

Rules:
- E-1 Each environment has its own database, Redis, object storage bucket, admin accounts and secrets. No credential is valid in two environments.
- E-2 No production data is copied into `staging` or `dev` (`docs/test-strategy.md` 7.4). A production incident is reproduced from its replay record and config version.
- E-3 Staging-only paths (the bot identity path of `docs/test-strategy.md` section 3.1) are refused by the server at start when it runs as `production` (#57, #58).
- E-4 An image is built once and promoted from `staging` to `production` unchanged; environments differ only by environment variables and secrets (`docs/launch-runbook.md` 5.1, #25).
- E-5 Admin change sets move from `staging` to `production` by export and import, approved again in production (`docs/admin.md` 3.6).
- E-6 PostgreSQL and Redis run the same major version in CI, `staging` and `production` (`docs/test-strategy.md` 2.3). PostgreSQL is 16 (#55, `server/internal/db/dbtest` runs `postgres:16-alpine`); the Redis version is fixed with #51 and the hosting with #25.

Hosting: the primary data region is Germany (EU) with Hetzner as hosting provider (`docs/privacy.md` section 2, OQ-3). Orchestration and the node layout are not decided; they are the ADR of #25. Infrastructure as code is written with that ADR, so the IaC tool and its state location are open items (section 7).

## 2. Environment variables

The server reads its configuration from the environment only, with the `HD_` prefix. A variable with a default is optional; a bad value stops the start with every bad variable named and exit code 2. A secret is never given a default and never logged (`docs/privacy.md` P-9).

Today no code on `main` reads an environment variable. The server variables below come from the open #264 (service skeleton, #43) and are marked so; when it merges, `server/README.md` holds the same table and the two stay in sync.

### 2.1 Server runtime

| Name | Required | Service | Default | Example placeholder | Secret | Source |
|---|---|---|---|---|---|---|
| `HD_LISTEN_ADDR` | No | server | `:8080` | `:8080` | No | from #264 |
| `HD_LOG_LEVEL` | No | server | `info` | `info` | No | from #264 |
| `HD_ORIGINS` | No | server | none | `play.example.com,admin.example.com` | No | from #264 |
| `HD_SHUTDOWN_TIMEOUT` | No | server | `10s` | `10s` | No | from #264 |
| `HD_DRAIN_TIMEOUT` | No | server | `15m` | `15m` | No | from #264 |
| `HD_MIN_APP_VERSION` | No | server | `1.0.0` | `1.0.0` | No | from #264 |
| `HD_CATALOG_VERSION` | No | server | empty | `<catalog-version>` | No | from #264 |
| `HD_MAX_FRAME_BYTES` | No | server | `1024` | `1024` | No | from #264 |
| `HD_WRITE_QUEUE` | No | server | `32` | `32` | No | from #264 |
| `HD_HELLO_TIMEOUT` | No | server | `5s` | `5s` | No | from #264 |
| `HD_WRITE_TIMEOUT` | No | server | `5s` | `5s` | No | from #264 |
| `HD_CLOSE_TIMEOUT` | No | server | `2s` | `2s` | No | from #264 |
| `HD_COMPRESSION_THRESHOLD` | No | server | `256` | `256` | No | from #264 |
| `HD_INTENTS_PER_SECOND` | No | server | `10` | `10` | No | from #264 |
| `HD_INTENT_BURST` | No | server | `20` | `20` | No | from #264 |
| `HD_PING_INTERVAL` | No | server | `15s` | `15s` | No | from #264 (default from #262) |
| `HD_PONG_TIMEOUT` | No | server | `5s` | `5s` | No | from #264 (default from #262) |
| `HD_RECONNECT_GRACE` | No | server | `60s` | `60s` | No | from #264 |

The gateway limits and the reconnect grace are environment variables only until admin config serves them (#210, `docs/penalties.md` section 8); then the admin value wins and the variable is removed in the same PR.

### 2.2 CI

| Name | Required | Service | Example placeholder | Secret | Stored as |
|---|---|---|---|---|---|
| `UNITY_LICENSE` | Yes, for the `client` job | CI, Unity test runner | `<unity-license-file-contents>` | Yes | Repository secret |
| `UNITY_EMAIL` | Yes, for the `client` job | CI, Unity test runner | `<unity-account-email>` | Yes | Repository secret |
| `UNITY_PASSWORD` | Yes, for the `client` job | CI, Unity test runner | `<unity-account-password>` | Yes | Repository secret |

### 2.3 Not read yet

The admin panel and the client read no environment variable. Database and Redis connections (#55, #51), session token signing keys (#58), ID token verification settings (#57), store receipt validation credentials (#106) and crash reporting (#182) add their variables when the code that reads them lands. That PR adds the row here with all columns filled and the secret to section 3 for each environment.

## 3. Where secrets live

| Secret use | Lives in | Reaches the consumer as |
|---|---|---|
| CI only (Unity license) | GitHub Actions repository secrets | Job environment variables, masked in logs |
| `staging` runtime | GitHub Actions environment secrets of the `staging` environment | Environment variables injected by the staging deploy (#25) |
| `production` runtime | GitHub Actions environment secrets of the `production` environment, deployable only from protected refs | Environment variables injected by the production deploy (#25) |
| `dev` | A local `.env` file, ignored by `.gitignore` | The local stack (#22) |

Rules:
- S-1 No secret is committed: not in code, docs, fixtures, images, `.env` files or CI files. `.gitignore` ignores `.env`, `.env.*` (except `.env.example`), keystores, signing certificates, provisioning profiles and `service-account*.json`. Only `.env.example` with placeholders may be committed.
- S-2 Test keys are generated at test time, never committed (CLAUDE.md).
- S-3 A secret is read by the process from its environment at start and kept in memory; it is never written to the image, a log, a metric label, a crash report or a replay record.
- S-4 `production` secrets are visible to the `production` GitHub environment only; workflows for pull requests never receive them.
- S-5 Values are entered in the GitHub settings or with `gh secret set --env <environment>` reading from standard input, never typed into a command line, chat or issue.

## 4. Secret scan

The `secret-scan` job of `.github/workflows/ci.yml` runs on every pull request and every push to `main`, whatever paths changed. It runs gitleaks, pinned by version and by the SHA-256 of the release archive, with findings redacted in the log:

- On a pull request it scans every commit from the merge base to the head of the branch.
- On a push to `main` it scans the full git history.

`CI complete` fails unless `secret-scan` succeeded, so a finding blocks the merge. A finding is never allowlisted to pass: a real secret is rotated (section 5); a false positive (generated code, test vectors) gets a narrow path allowlist in `.gitleaks.toml`, reviewed in its PR, never a disabled rule.

Local run, before pushing a branch that touched credentials or config:

```bash
gitleaks git --redact --no-banner .
```

## 5. Rotation

The security lead owns rotation (`docs/launch-runbook.md` section 1). Every rotation is recorded in an issue labelled `area:infra` with the secret name, environment, reason and time, never the value.

When: on any suspected leak, when a person with access to it leaves the team or changes role, when the issuer expires it, and at the regular cadence (an owner decision, section 7).

Procedure for one secret in one environment:
1. Create the new value at its issuer (database role password, key pair, store or vendor credential). Keep the old one valid.
2. Store the new value in the environment's secrets (section 3).
3. Deploy, so every node runs with the new value (#25). Signing keys rotate with an overlap: the new key signs, the old key still verifies until every token it signed has expired (#58).
4. Verify on that environment: health and readiness pass, logins and matches work, no authentication errors in the logs (#28).
5. Revoke the old value at the issuer.
6. Rotate `staging` first, then `production`, unless the leak is in `production`.

On a leak: it is a SEV1 incident (`docs/launch-runbook.md` 6.1, 6.5). Revoke first when the exposed value is in active use by someone else, accepting the outage, then follow the procedure. A secret found in git history is rotated; rewriting history does not undo the exposure and is not a substitute. A personal data breach follows `docs/privacy.md` and #187.

## 6. Database migrations

Schema changes are versioned migrations reviewed in PRs (ADR 0005). The tool is goose (#55): SQL files under `server/internal/db/migrations`, embedded in the binaries, applied by `server/cmd/migrate` under a PostgreSQL session advisory lock. PostgreSQL major version: 16 (E-6).

Rules:
- M-1 Forward only in `staging` and `production`. A mistake is fixed by a new migration. Down migrations, if the tool writes them, run only in `dev` and CI.
- M-2 Expand, then contract (`docs/launch-runbook.md` 5.1). A migration never breaks the server version currently deployed, so the binary rolls back without touching the schema. A contract step ships in a later release, after the previous version is no longer deployed anywhere.
- M-3 A migration is never edited after it merged; its version number is unique and ordered.
- M-4 Every migration PR passes the migrations suite of `docs/test-strategy.md` 2.3 in CI: up from empty to head on a fresh database, and the previous server version working on the new schema. Before a release the expand step is drilled on a staging copy (`docs/test-strategy.md` RG-6).
- M-5 Migrations run once per deploy, from one runner holding the tool's lock, before nodes of the new version take traffic; server nodes never migrate on start concurrently (#25).
- M-6 Order per release: `staging` first, a deploy drill on it, then `production` with the same migration files.
- M-7 The migration role may change the schema; the server's runtime role may not. Each has its own secret per environment.
- M-8 Ledger, purchase and inventory rows are never corrected by a migration or direct SQL; corrections are ledger entries through admin tools (`docs/launch-runbook.md` 5.4, #144).
- M-9 `matches` and `match_participants` are partitioned by month of `started_at`. `match_partitions_ensure(from, months)` creates the months ahead and is idempotent; a deploy or a scheduled job calls it for the coming six months, and a `_default` partition catches any row outside the created months so a write never fails. A partition is created before rows for its month exist.

## 7. Open items

- Infrastructure as code: tool, state location and the resources it creates wait for the orchestration ADR of #25. GN-8 evidence (IaC apply log) depends on it.
- Staging automatic deploy (acceptance criterion of #24) depends on #25.
- The GitHub environments `staging` and `production`, their protection rules and their secrets are created by a repository admin when the first deploy exists (#25).
- Regular rotation cadence per secret type: owner decision.
- Encryption at rest of PostgreSQL volumes and backups is a hosting setting delivered with the IaC (#25, #188, `docs/privacy.md` 9.1).
- Migration tool: #55.
