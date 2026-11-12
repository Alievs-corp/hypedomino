# Backup, restore and disaster plan

What is backed up, how often, where it goes, how fast it comes back, and how that is proved (#188). Decision record: ADR 0013. Hosting: Hetzner, Germany (`docs/privacy.md` section 2, ADR 0010). Release gate: GN-15 in `docs/launch-runbook.md`.

## 1. What each store holds

| Store | Holds | Loss means | Policy |
|---|---|---|---|
| PostgreSQL | Accounts, ledger, purchases, inventory, ranks, pass progress, matches, promo codes, admin config and its snapshots (ADR 0005) | Money-like data lost. Not acceptable | Continuous WAL archiving plus daily base backup (section 3) |
| Redis | Only ephemeral and rebuildable state (section 4) | Players requeue and reconnect; no balance or purchase is touched | No persistence, no backup |
| Object store, content bucket | Addressables bundles, catalogs, admin-uploaded cosmetic assets (`docs/content-delivery.md`) | Broken downloads; admin-uploaded assets are not reproducible from a build | Versioning plus a daily copy to a mirror bucket (section 5) |
| Object store, state bucket | OpenTofu state (ADR 0010) | Infrastructure cannot be changed safely | Bucket versioning (already required by ADR 0010) |
| Secrets, WAL-G encryption key | `docs/environments.md` section 3 | Backups unreadable | Section 3.4 |

## 2. Targets

| Target | Value | Reasoning |
|---|---|---|
| RPO, PostgreSQL | 5 minutes of committed writes | `archive_timeout = 60` ships a WAL segment at least once a minute, so a lost database host costs about 1 to 2 minutes of writes in practice; 5 minutes leaves room for archive retries and a slow upload. A purchase inside that window is recovered from the store receipt history (section 7.3), never by guessing |
| RTO, PostgreSQL | 1 hour from the decision to restore until the servers take traffic | Base backup download plus WAL replay for a launch-size database takes minutes (the drill measures it, section 8); the rest is provisioning a node, the post-restore steps and verification. The game has no value while the database is down, so a tight target is cheap to hold |
| RPO, Redis | Not applicable | Nothing in it is authoritative |
| RTO, Redis | 15 minutes | Start an empty instance; pods reconnect on their own, leaderboards rebuild from PostgreSQL on first read |
| RPO, content bucket | 24 hours | The mirror is daily; anything newer is re-uploaded from the build or the admin panel |
| RTO, content bucket | 4 hours | Point the CDN origin at the mirror; no data movement |
| Whole-region loss | Out of scope at launch | One region (ADR 0010 phase 1). The backup bucket is in another Hetzner location than the database, so the data survives; rebuilding the cluster from OpenTofu is the 4 to 8 hour path. Revisit with phase 2 |

Targets are reviewed after every drill. A drill result worse than the target is a defect, not a footnote.

## 3. PostgreSQL

### 3.1 Mechanism
- WAL archiving: `archive_mode = on`, `archive_timeout = 60`, `archive_command = 'wal-g wal-push %p'` (`deploy/k8s/backup/walg-config.yaml`, ConfigMap `postgres-archive-config`).
- Base backup: CronJob `walg-base-backup`, daily 02:17 UTC, `wal-g backup-push`. It only reads the data volume and is pinned to the node of the PostgreSQL pod with pod affinity.
- Compression `zstd`. Encryption on the client with libsodium (`WALG_LIBSODIUM_KEY`, 32 bytes, hex).
- Storage: an S3-compatible bucket used for backups only, in a different Hetzner location than the database, with its own access key that the game servers and the admin API never hold.

### 3.2 Retention
- `BackupRetentionDays` = 35 (key defined in `docs/privacy.md` section 4). CronJob `walg-retention` runs `wal-g delete retain FULL 35` daily, which removes base backups beyond 35 and the WAL they alone needed. The recovery window is therefore 35 days.
- 35 is the 30-day rights-request clock plus a 5-day margin. A backup older than that must not exist, because a deleted account would come back with it.
- Size is checked after the first month; if the bucket cost is out of proportion the window is shortened by changing this number here and in `walg-retention.yaml` together.

### 3.3 Monitoring and alerts (wired by #28)
- `pg_stat_archiver.failed_count` increasing, or `last_archived_time` older than 5 minutes: page. A stuck archive fills the data volume with WAL.
- No successful `backup-push` in 26 hours: page.
- `wal-g backup-list` (printed by the retention job) is the weekly human check; the T-1d runbook line "Backups verified fresh" is this list.

### 3.4 Keys and access
- Names only, never values: Secret `walg-credentials` (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `WALG_LIBSODIUM_KEY`) and Secret `content-mirror-credentials`.
- The encryption key is also escrowed offline (the owner's and the server lead's password vaults). Without it every backup is unreadable; losing every copy equals losing the backups.
- The PostgreSQL role `walg` has only what `backup-push` needs; this is verified on staging (section 9). The drill uses the superuser because its database is throwaway.

## 4. Redis persistence policy

What Redis holds in this code (`server/internal`):

| Data | Package | Authoritative copy | After total loss |
|---|---|---|---|
| Room routing, node registry | `cluster` | Memory of the pods, re-registered by heartbeat | Pods re-register within `HD_NODE_TTL` |
| Matchmaking tickets, leader locks, placement | `matchqueue` | The client's queue request | Players requeue |
| Presence | `friends` | The open WebSocket | Rebuilt as clients heartbeat |
| Lobby invites and joins | `lobbynet` | Short-lived messages | Lost invites are re-sent by the player |
| Leaderboards | `leaderboard` | PostgreSQL (`Rebuild`, run by `ensure` when a board is missing) | Rebuilt lazily from PostgreSQL |
| Rate-limit buckets | `ratelimit` | None | Buckets refill; a short burst of extra requests is possible |
| Sign-in and attestation nonces | `auth`, `integrity` | None | One replay inside the nonce TTL is possible after a loss; bounded by the token's own expiry and accepted |

Policy:
- `save ""` and `appendonly no`: no RDB, no AOF, no backup. A restored snapshot is worse than an empty instance because it brings back routes to pods that no longer exist, queue tickets of players who left, and presence of offline accounts.
- `maxmemory-policy noeviction`: a full instance refuses writes visibly instead of silently dropping a lock or a route.
- Config: `deploy/k8s/backup/redis-config.yaml`.
- A Redis restart or loss is an availability incident (players requeue), not a data incident. Review rule for any new use of Redis: if its loss would lose money or progress, the data belongs in PostgreSQL. Making Redis state authoritative requires changing this section and ADR 0005 first.

## 5. Object storage
- Content bucket: versioning on; CronJob `content-mirror` runs `rclone copy --immutable` daily into a mirror bucket in another location. `--immutable` fails loudly if a published key changed, which `docs/content-delivery.md` CD-3 forbids anyway. Remotes `content` and `mirror` are defined through `RCLONE_CONFIG_*` variables in Secret `content-mirror-credentials`, together with `CONTENT_BUCKET` and `MIRROR_BUCKET`.
- Hashed bundles and catalogs are reproducible from a tagged client build (`.github/workflows/release-build.yml`); the mirror matters most for admin-uploaded assets.
- The backup and mirror buckets are never writable by the game servers or the admin API.

## 6. Disaster scenarios

| Scenario | Detect | Action | Target |
|---|---|---|---|
| Database host or volume lost | `/readyz` fails, alert | Section 7.1 | RTO 1 h, RPO 5 min |
| Bad migration or deploy corrupted data | Support reports, ledger reconciliation | Section 7.2 to a time just before the change | RTO 1 h, writes after that time are lost |
| Ledger mistake (wrong grant) | Support, audit | Not a restore. Correct with a compensating ledger entry (`docs/admin.md` AP-1) | Minutes |
| Redis lost | Presence and queues empty | Section 7.4 | RTO 15 min |
| Backup bucket unavailable | Archive alert | PostgreSQL keeps running and queues WAL on disk; fix before the data volume reaches 50 percent | Hours |
| Backup bucket compromised | Provider alert, audit | Rotate keys and treat as a breach (`docs/launch-runbook.md` section 6); data is encrypted and the key is held elsewhere | Per runbook |
| Region lost | Provider status | Out of scope at launch (section 2) | Hours to days |

## 7. Restore runbook

Decision: the incident commander with the product owner (`docs/launch-runbook.md` section 5.4 and 6). A restore always goes to a new node and a new volume; the damaged one is kept untouched until the restore is verified.

### 7.1 Restore to the latest state
1. Stop the game servers (scale the Deployment to 0). Writes during a restore would be lost or double-applied.
2. Provision the replacement PostgreSQL node and an empty data volume.
3. `wal-g backup-list` shows what exists; fetch the newest: `wal-g backup-fetch $PGDATA LATEST`.
4. `touch $PGDATA/recovery.signal`; start PostgreSQL with `restore_command = 'wal-g wal-fetch %f %p'` and `recovery_target_action = promote` (no target means replay to the end of the archive).
5. Wait until `pg_is_in_recovery()` is false.
6. Continue at section 7.3.

### 7.2 Restore to a point in time
Same as 7.1, adding `recovery_target_time = '<UTC timestamp>'` in step 4. Pick the time from the audit log of the bad change, minus a margin. A time chosen too early is repaired by restoring again with a later one, because base backups and WAL are not consumed by a restore. `recovery_target_inclusive` stays at its default. The local drill performs exactly this procedure.

### 7.3 After every restore, before traffic
1. Re-apply account deletions completed after the restore point (`docs/privacy.md` section 8.2 step 5, L-19). The source is a deletion log that survives the restore; it does not exist yet (section 9).
2. Reconcile purchases after the restore point: ask the stores for the transactions since that time and run them through the normal idempotent grant path (a receipt already recorded is a no-op, ADR 0007). This repairs the RPO window.
3. Reconcile ledger sums against balances and check the schema version (`docs/environments.md` M-5).
4. Empty Redis (section 7.4); stale routing must not outlive a restore.
5. Scale the servers up; confirm `/readyz`, one sign-in, one match, one store sandbox purchase.
6. Take a fresh base backup at once, so the new timeline has a restore point of its own.
7. Write the incident record: times, target, writes lost, who decided.

### 7.4 Redis
Start an empty instance with `deploy/k8s/backup/redis-config.yaml`; restart the game pods so they re-register their nodes and rooms.

### 7.5 Content bucket
Point the CDN origin at the mirror bucket; re-upload anything newer from the release build or the admin panel.

## 8. Restore drill

### 8.1 Local drill
`bash tools/backup-drill/drill.sh` needs only Docker. It:
1. starts an S3-compatible object store (SeaweedFS, standing in for the Hetzner bucket) and creates the bucket;
2. starts PostgreSQL 16 with WAL archiving through WAL-G, compressed and encrypted as in production;
3. writes 1000 rows, takes a base backup, writes 500 rows, records a target time, writes 250 more rows, and forces the WAL to be archived;
4. destroys the database and restores into a fresh volume from the base backup and WAL with `recovery_target_time` set to the recorded time;
5. fails unless exactly the 1000 base rows and the 500 replayed rows exist, none of the 250 later rows exist, and the database accepts writes after promotion.

It prints the restore time. CI job `backup-drill` runs it on every change to `tools/backup-drill`, `deploy/k8s/backup` or the workflow, and checks that the manifests render. First run on 2026-11-12: 1500 rows restored, 0 after the target, fetch, replay and promote in about 6 seconds for the toy database, about 35 seconds end to end.

### 8.2 Staging drill (open)
Not possible until staging exists (#25). Procedure to run then, recorded as GN-15 evidence:
1. Load a production-sized synthetic dataset on staging (`docs/test-strategy.md` 7.2) and let it run 24 hours so base backups and WAL exist in the real backup bucket.
2. Declare a restore: stop the servers, run section 7.2 against the real bucket on a new node, start the clock.
3. Run section 7.3, including the deletion step.
4. Record RTO from the declaration to the first successful match, RPO from the last restored write, the base backup size and WAL volume per day, and every manual step not in this document.
5. Correct this document and the targets from what was measured.

### 8.3 Cadence
- Local drill: every change to the files above.
- Staging drill: before launch (GN-15), then every quarter and after any change of PostgreSQL major version, WAL-G version or hosting.
- Backup freshness: weekly (`wal-g backup-list`) and at T-1d before a release.

## 9. Open items
- Staging restore drill and measured RTO (needs staging, #25). Part of #188.
- A deletion log that survives a restore, so 7.3 step 1 can be executed. Today a deletion is a state change in PostgreSQL (`server/internal/account/deletion.go`) and is lost with the data it was written to. Proposal: append each completed deletion (tombstone id, time) to an object in the backup bucket. Needs its own issue; L-19 stays open until then.
- Publishing the image `ghcr.io/alievs-corp/hypedomino-postgres-walg` from `tools/backup-drill/Dockerfile`, and the PostgreSQL manifests the CronJobs assume (a `postgres` pod, PVC `postgres-data`, Service `postgres`), which are not in `deploy/k8s` yet.
- Real bucket, endpoint, credentials, encryption key and mirror bucket (owner provides, section 3.4).
- Least-privilege `walg` role verified against the real database.
- Alerts of section 3.3 (#28).
