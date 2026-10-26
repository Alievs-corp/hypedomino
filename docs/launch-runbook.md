# Launch checklist and runbook

How a release goes from candidate to players, and what happens when it goes wrong: the go/no-go checklist, the launch timeline, deploy, staged rollout and rollback procedures, incident management, on-call, player communication and support. It applies to the soft launch (#196), to the global launch and, from section 5 on, to every later release.

Scope boundaries: `docs/roadmap.md` section 4 is the release gate and is not restated here; this document adds the operational evidence needed to pass it. Budgets are in `docs/performance.md`. Kill switches, the build allowlist and ban responses are in `docs/anti-cheat.md` and #210. Architecture is ADR 0002 to 0006.

Infrastructure does not exist yet. Procedures are written against the planned architecture; every concrete command, URL and tool name is marked `{filled by #N}` and is added by that issue. Numbers marked *default* are starting values: operational ones change by PR to this file, game ones are admin config (ADR 0006). Roles, never names; one person may hold several roles, except where a section forbids it.

## 1. Roles

| Role | Owns |
|---|---|
| Product owner | Final go/no-go decision, waivers (an owner comment on the issue concerned, `docs/roadmap.md` section 4), player compensation decisions |
| Release manager | This checklist, the release issue, the timeline, store consoles (submission, rollout, halt) |
| Server lead | Server and data deploys, server rollback, capacity, database restore (#188) |
| Client lead | Release builds, client hotfixes, forced minimum version |
| Live-ops lead | Production admin config, catalog, banners, maintenance mode, kill switches (#145, #146, #210) |
| QA lead | Release candidate regression and device smoke (#190, #191) |
| Security lead | Red team result (#212), anti-cheat config, security incidents, secret rotation |
| Support lead | Support channel, FAQ, templates, store review replies, ticket-to-incident escalation |
| Analyst | Health and KPI dashboards during launch (#147, #181, #196) |

Incident roles (section 6) are assigned per incident: incident commander, comms lead, scribe, responders.

## 2. Release record

Every release has one GitHub issue labelled `area:release` that links its release candidate commit and tag. The release manager copies section 3 into it, links each item's evidence, and records the go/no-go decision as an owner comment. A release without this issue does not ship.

## 3. Go/no-go checklist

Rules: every item is green, or waived by an owner comment on the issue concerned. Missing or unrecorded evidence is red. Evidence is recorded against the release candidate commit or the production environment it ships to, not an earlier build.

### 3.1 Release gate (`docs/roadmap.md` section 4)

| Id | Item | Owner | Evidence |
|---|---|---|---|
| GN-1 | Gate 1: no open `priority:p0` or `priority:p1` issue | Release manager | Output of both `gh issue list` commands from the gate, pasted in the release issue |
| GN-2 | Gate 2: performance budgets pass on the device matrix | Client lead, Server lead | Measurement reports in #191, #192, #161, #164 per `docs/performance.md` |
| GN-3 | Gate 3: content gate | Live-ops lead | `green`, non-stale production content gate report run after the config freeze starts, or a `red` one with an owner waiver on #198 (`docs/launch-content-gate.md` 4.3, 4.4) |
| GN-4 | Gate 4: legal documents published and linked | Product owner | Links open from the store pages and in-app settings (#195, #187) |
| GN-5 | Gate 5: red team has no open critical finding | Security lead | #212 report |
| GN-6 | Gate 6: this checklist green; backup and restore exercised; closed beta and soft launch KPI tracking done or waived | Release manager | This list, #188 drill record, #193 and #196 closed or waived |
| GN-7 | Gate 7: `CI complete` green on the release commit | Release manager | Link to the run on the tagged commit |

### 3.2 Operational readiness

| Id | Item | Owner | Evidence | Link |
|---|---|---|---|---|
| GN-8 | Production environment separate from staging; secrets only in the secret manager; secret scan green | Server lead | IaC apply log, secret scan run | #24 |
| GN-9 | Server deploy drill on staging: matches in progress finish or move during a deploy; rollback measured under 5 minutes | Server lead | Drill record with timings | #25, #43 |
| GN-10 | Catalog switch drill: previous Addressables catalog re-activated; an older client still loads with its catalog | Live-ops lead | Drill record | #26, ADR 0002 |
| GN-11 | Release builds are produced and signed by the pipeline from the tag; version and build numbers recorded; symbols and mapping files uploaded | Client lead | Pipeline run, build numbers in the release issue | #27, #182 |
| GN-12 | Dashboards live in production; alerts route to the on-call pager; one test alert reached the primary | Server lead | Screenshot of the test page and its acknowledgement | #28 |
| GN-13 | Crash reporting shows release health per app version; a test crash from the release build is visible | Client lead | Dashboard link | #182 |
| GN-14 | Funnel events (install, login, first match, purchase) visible on the launch dashboard | Analyst | Dashboard link | #181, #147, #196 |
| GN-15 | Backup restore drill done; RTO and RPO recorded | Server lead | Drill record | #188 |
| GN-16 | Rate limits and WebSocket abuse protection active in production | Server lead | Config dump and a test against staging | #184 |
| GN-17 | Kill switch drill in production: purchase and ranked switches each off and on, a test build blocked and unblocked, each under 5 minutes; the forced-update screen shown on a device | Live-ops lead, Client lead | Drill record with timings, screen recording | #210, #73, ADR 0004 |
| GN-18 | Maintenance mode and an in-game banner shown on a device in all four launch languages | Live-ops lead | Screen recordings | #146, `docs/localization.md` |
| GN-19 | Production config reviewed and published, including every value in `docs/anti-cheat.md` section 7; config rollback drill (publish the previous version) applies to new matches | Live-ops lead, Security lead | Config version id, audit log entries | #145, ADR 0006 |
| GN-20 | Admin accounts least privilege; audit log on; hardening passed | Security lead | Account and role list, #148 report | #134, #148 |
| GN-21 | Support tools work in production: player lookup, grant with reason, ban and mute; moderation and appeal queue staffed | Support lead | Audit log entries of test actions on a staff account | #144, #186 |
| GN-22 | Store products match admin shop config; a licensed test purchase on the release build is granted once and restores | Live-ops lead | Ledger entry, restore record | #106, #107, #139, #204 |
| GN-23 | Store listings complete, age ratings and data safety forms submitted, release approved and held for manual release | Release manager | Console screenshots | #194, #187 |
| GN-24 | Release candidate regression passed; smoke passed on every device tier | QA lead | Test run report per `docs/test-strategy.md` section 8 | #190, #191 |
| GN-25 | On-call schedule published from T-1d to T+14d; contact sheet complete; escalation path tested end to end | Release manager | Schedule link, test page record | section 7 |
| GN-26 | Support channel live; FAQ and templates published in all four languages | Support lead | Links, translation review | sections 8, 9 |
| GN-27 | Status page live with incident and maintenance components | Support lead | Link `{filled by #28}` | section 8 |
| GN-28 | At least two people hold release, halt and pause rights in both store consoles | Release manager | Console user lists | section 5 |

## 4. Launch timeline

T-0 is the moment the release is made available in the stores. Server and config changes ship before T-0; the client never depends on a server change that is not already deployed everywhere (ADR 0004: additive protocol within a major version).

| When | Action | Owner |
|---|---|---|
| T-7d | Release candidate tagged from `main`; only fixes for release-blocking issues merge after this. Builds submitted for review with manual release (App Store) and managed publishing (Play) so approval does not publish. First full pass of section 3; open items get owners and dates in the release issue | Release manager |
| T-7d | Load and soak test re-run against the production config version (#192) | Server lead |
| T-5d | Support templates, FAQ and banner texts sent for translation (`docs/localization.md`) | Support lead |
| T-3d | Server release deployed to production (section 5.1). Go/no-go dry run: every item has evidence or a named blocker | Server lead, Release manager |
| T-3d | Config freeze: production admin publishes only through the release manager until T+72h, except incident mitigation | Live-ops lead |
| T-3d | Content gate report run in production (GN-3); run again after any later content or config publish, since that makes it stale | Live-ops lead |
| T-1d | Go/no-go meeting. Product owner decides go, no-go or slip; decision recorded in the release issue. War room channel opened, on-call launch shifts start | Product owner |
| T-1d | Backups verified fresh; production smoke with the store-approved build | Server lead, QA lead |
| T-2h | Final health check of all dashboards; status page set to operational; on-call primary and secondary confirmed online | Release manager |
| T-0 | Release made available in both stores; announcement published | Release manager, Support lead |
| T+15m | First installs, logins and matches visible on the launch dashboard | Analyst |
| T+1h | Health review against section 4.1; then hourly until T+6h | Release manager |
| T+24h | Daily review: section 4.1 signals, support ticket themes, store reviews, KPIs. Decide whether to continue, hold or widen availability (#196) | Product owner |
| T+48h | Daily review | Product owner |
| T+72h | Launch exit: back to the normal on-call rotation, config freeze lifted, retrospective scheduled, this runbook updated with what was learned | Release manager |

### 4.1 Launch health signals

| Signal | Source | Threshold |
|---|---|---|
| Crash-free sessions per app version | #182 | At least 99.5% (the #193 beta exit level) and not below the previous version |
| Move round-trip p99 | #28 | Under 100 ms (`docs/performance.md`) |
| Login success, matchmaking wait, error rate, database connections | #28 | Alert thresholds `{filled by #28}` |
| Receipt validation success per store | #106, #28 | Alert thresholds `{filled by #28}` |
| Android vitals crash and ANR rates | Play Console | Below the console's bad behaviour thresholds |
| Support tickets per symptom | Support tool | Section 9.3 escalation rule |

A threshold breach with player impact is an incident (section 6).

## 5. Procedures

### 5.1 Server deploy

1. The release is a tagged commit with `CI complete` green; the image is built once and promoted from staging to production unchanged. `{filled by #25}`
2. Migrations follow expand then contract: a migration never breaks the previous server version, so the binary can roll back without touching the schema (ADR 0005). A contract step ships only after the previous version is no longer deployed.
3. Canary: new rooms are routed to one new node first, *default* 10% of new rooms for 30 minutes, while rooms in progress stay on old nodes (ADR 0003: a room lives on one node).
4. Watch section 4.1 on the canary. On a breach, stop and roll back (5.4).
5. Roll the remaining nodes: new rooms go to new nodes, old nodes drain until their rooms finish, then stop (#25, #43). Never kill a node with live rooms.
6. Record the deployed version and config version in the release issue.

Order with a client release: server first (backward compatible with every client at or above the minimum version), then client.

### 5.2 Client rollout

First store release: percentage rollouts are not used. Exposure is limited by country availability (soft launch regions, #196) and by server-side controls (feature flags, matchmaking, kill switches).

Every later update:

| Store | Mechanism | Default plan |
|---|---|---|
| Google Play | Staged rollout by percentage, halt and resume in Play Console | *Default* 1%, 5%, 20%, 50%, 100%, at least 24 hours per step |
| App Store | Phased release over 7 days with a schedule fixed by Apple (1%, 2%, 5%, 10%, 20%, 50%, 100% of users with automatic updates); can be paused, total pause up to 30 days, or released to everyone early | Phased release on; release to all early only after the Play rollout reaches 50% without incident |

Phased release covers automatic updates only: new installs and manual updates get the new version at once. Feature flags (#145) are the real limit for risky features on iOS.

Advance a step only when, for the new version: crash-free sessions meet section 4.1, no open SEV1 or SEV2 is caused by it, and support tickets show no new symptom for it. Otherwise hold; on a SEV1 or SEV2 caused by the version, halt (Play) and pause (App Store) at once.

### 5.3 Config and content publish

- Admin config is published as a new version; matches started after the publish use it and every match records its version (ADR 0005, ADR 0006). Publishing is audited (#134).
- Content goes live by activating a new Addressables catalog version (ADR 0002, #26).
- During the config freeze (section 4) every publish is approved by the release manager and noted in the release issue.

### 5.4 Rollback

Mitigate first, find the root cause after. Pick the fastest step that stops the harm.

| What broke | Step | Target | Notes |
|---|---|---|---|
| Server release | Redeploy the previous image with the same drain as 5.1 | Under 5 minutes (#25) | Works because migrations are expand only during a release. Never roll back below the version a live client needs; fix forward instead |
| Database data | Restore from point-in-time recovery | RTO and RPO from #188 | Only for corruption or loss, decided by the incident commander with the product owner. Ledger mistakes are corrected with ledger entries, never by restore or direct SQL (#101, #144) |
| Config | Publish the previous config version in admin | Under 5 minutes | Applies to new matches; matches in progress finish on their recorded version |
| Content | Re-activate the previous catalog version | Under 5 minutes | Old content URLs stay valid because files are immutable (ADR 0002) |
| Client feature | Turn off its feature flag or kill switch (5.5) | Under 5 minutes (#210) | First choice for any client-side defect |
| Client build | Halt the Play rollout and pause the App Store phased release | Immediate in the consoles | Players who already updated keep the version |
| Client build, no flag helps | Ship a hotfix with a higher version and build number, request expedited App Store review; then raise the minimum version once the fix is available in both stores | Hours to days, store dependent | Raising the minimum version before a fixed build exists locks those players out |
| Leaked, modded or compromised build | Remove its build hash from the allowlist or raise the minimum version | Under 5 minutes (#210) | Security only; players on it see the forced-update screen (ADR 0004) |

Store builds cannot be withdrawn from devices. That is why every client feature with risk ships behind a flag.

### 5.5 Kill switches and server-side controls

| Control | Effect | Changed by | Defined in |
|---|---|---|---|
| Purchase kill switch | Store purchases disabled in the shop; already validated receipts are still granted | Live-ops lead or incident commander, with the `live_ops` or `super_admin` admin role and the two-person rule of `docs/admin.md` section 8 | #210 |
| Ranked kill switch | Ranked queues closed; matches in progress finish | Same | #210 |
| Minimum version and build allowlist | Older or unlisted builds get `UPDATE_REQUIRED` | Same, with the client lead | #210, ADR 0004 |
| Maintenance mode | Client shows the maintenance screen; no new matches | Same | #146 |
| Feature flags | Any flagged feature off or on without a release | Same | #145 |
| Promo code disable | One code stops redeeming | Live-ops lead | #142, `docs/anti-cheat.md` V-11 |

Every change is audited with a reason (#134). Switching back follows the same path and is noted in the incident record.

## 6. Incident management

### 6.1 Severity

| Severity | Definition | Examples |
|---|---|---|
| SEV1 | Most players cannot play, or money, data or security is at risk | Login or matchmaking down; a path that mints currency or items; data loss; leaked secret; mass crash of the current version |
| SEV2 | A large group cannot play or pay, or a budget is broken for everyone | One region or one platform down; purchases failing on one store; p99 move round-trip over budget for 15 minutes (*default*) |
| SEV3 | A feature is broken with a workaround, or a small group is affected | Leaderboard stale; one device model crashing; banner text missing in one language |
| SEV4 | Cosmetic or no player impact | Typo; internal dashboard broken |

When unsure, pick the higher severity and lower it later.

### 6.2 Response targets

*Default* values, changed by PR to this file.

| Severity | Acknowledge | Incident commander named | First public update | Update cadence | Postmortem |
|---|---|---|---|---|---|
| SEV1 | 5 min | 15 min | 30 min | Every 30 min | Required, within 5 business days |
| SEV2 | 15 min | 30 min | 60 min | Every 60 min | Required, within 5 business days |
| SEV3 | Next business day | Not required | Only if players ask | On change | Optional |
| SEV4 | Backlog | Not required | None | None | No |

### 6.3 Roles

| Role | Does | Does not |
|---|---|---|
| Incident commander | Owns the incident: severity, decisions, roles, when to roll back, when it is resolved. The on-call primary until handed over | Debug or type fixes in SEV1 |
| Comms lead | Status page, banner, store replies, internal updates on the cadence in 6.2, using section 8 | Promise times or compensation not approved by the incident commander or product owner |
| Scribe | Timestamped log of signals, decisions and actions in the incident channel; drafts the postmortem timeline | Act on systems |
| Responders | Investigate and mitigate as asked by the incident commander | Change production without telling the incident commander |

In SEV1 the incident commander, comms lead and scribe are three different people.

### 6.4 Flow

1. Detect: alert (#28), crash spike (#182), ticket pattern (9.3), store reviews, internal report.
2. Declare: anyone may declare. Open the incident channel `{filled by #28}` and set the severity.
3. Assign roles (6.3).
4. Mitigate with section 5.4 and 5.5 before the root cause is known.
5. Communicate on the cadence in 6.2 (section 8).
6. Resolve when the signals in 4.1 are back to normal for 30 minutes (*default*); post the resolved update.
7. Postmortem: blameless; timeline, impact (players, duration, money), root cause, what helped, what hurt. Every action item becomes an issue with `type:`, `area:` and `priority:` labels.

### 6.5 Special cases

- Economy integrity (currency or items created outside legal play): purchase kill switch on, evidence preserved, ledger reconciliation (#202). Corrections go through the ledger with a reason through admin tools (#144). Compensation is a product owner decision.
- Security (leaked secret, admin compromise, modded build at scale): security lead joins as responder. Rotate secrets (#24), block builds (#210), disable affected admin accounts (#134). Personal data breaches follow the notification duties in #187 (GDPR: 72 hours to the supervisory authority).
- Cheating wave: responses follow `docs/anti-cheat.md` section 3; public messages never reveal detection details (#186, #209).
- Store outage (billing or review): nothing on our side to roll back; purchase kill switch if receipts fail, comms only.

## 7. On-call

| Item | Rule |
|---|---|
| Rotation | Weekly, a primary and a secondary, handoff on a fixed weekday with a written note of open issues |
| Coverage | Production server, data, admin config and release consoles. The primary carries the pager `{filled by #28}` |
| Launch window | T-1d to T+72h: primary and secondary both on shift, release manager reachable, shifts no longer than 12 hours (*default*) |
| Acknowledge | Within the section 6.2 target; no acknowledgement in 10 minutes (*default*) pages the secondary |
| Escalation | Primary, then secondary, then the server or client lead for the area, then the product owner. The product owner is told of every SEV1. The security lead joins every security or economy incident |
| Access | Every person on the rotation has production read access, admin roles for 5.5 and the runbook; store console rights per GN-28 |
| Contact sheet | Pager, phone and backup channel for every role, kept outside the systems it is needed to repair `{filled by #28}` |

## 8. Communication

Channels: status page `{filled by #28}`, in-game banner and maintenance screen (#146), store review replies, push only for SEV1 lasting more than an hour (*default*, #132). English is the source; every template is published in all four launch languages (`docs/localization.md`) before T-3d. Times are given in UTC.

Rules: state what players see and what to do, not internal causes. No blame on vendors or stores. No time we cannot hold: give the next update time instead. No compensation before the product owner decides. No detection detail.

### 8.1 Status page

| State | Text |
|---|---|
| Investigating | Some players cannot {symptom}. We are looking into it. Next update by {time} UTC. |
| Identified | We found the cause of {symptom} and are fixing it. Next update by {time} UTC. |
| Monitoring | A fix is live for {symptom}. We are watching to make sure it holds. |
| Resolved | {symptom} is fixed since {time} UTC. Thanks for your patience. |
| Maintenance | Planned maintenance from {start} to {end} UTC. Matches cannot start during this time. |

### 8.2 In-game banner and maintenance screen

| Use | Text |
|---|---|
| Degraded | Some players are having trouble with {feature}. We are on it. |
| Purchases paused | The shop is paused for a moment. Purchases you already made are safe. |
| Ranked paused | Ranked play is paused for a moment. Casual matches are open. |
| Maintenance screen | We are making HypeDomino better. Back at {end} UTC. |
| Update required | A new version is ready. Update to keep playing. |

"Purchases you already made are safe" is shown only when the incident commander has confirmed it.

### 8.3 Store review replies

| Case | Text |
|---|---|
| Crash or bug | Thanks for telling us. We are fixing this. Please update to the latest version, and if it still happens, use Settings > Support so we can look at your case. |
| Purchase not received | Thanks for reporting this. Every purchase is checked with the store. Please use Settings > Support > Purchases and include your store order number, and we will sort it out. |
| Ban or penalty | Penalties follow our fair play rules. If you think it was a mistake, please send an appeal from Settings > Support. |
| During an incident | Sorry for the trouble. We know about this issue and are working on it. Follow {status page} for updates. |

## 9. Support

### 9.1 Channels

| Channel | Use |
|---|---|
| In-game support form (Settings > Support) | Primary channel. Attaches player id, app version, build, platform, OS, device model, language and last match id automatically; nothing more personal than #187 allows |
| Support email `{filled by #144}` | Players who cannot open the game |
| Store review replies | Public answers with section 8.3; move the case to the form |

All actions on accounts go through admin tools with a reason in the audit log (#144). Refunds are handled by the stores; revoked purchases are reversed through the ledger (#204).

### 9.2 Ticket categories and first response

| Category | Handled by | First response (*default*) |
|---|---|---|
| Purchase missing or wrong | Support, live-ops lead for ledger corrections | 24 h |
| Account and login | Support, server lead | 24 h |
| Ban, penalty or appeal | Support with the moderation queue (#186) | Per #186 |
| Bug or crash | Support, triaged to an issue | 48 h |
| Privacy and account deletion | Support, product owner | Within the legal deadline (#187, #63) |
| Other | Support | 48 h |

### 9.3 Escalation to on-call

Page the on-call primary when 5 tickets with the same new symptom arrive within 60 minutes (*default*), when any ticket reports currency or items appearing or disappearing, or when a ticket suggests a security or privacy breach.

### 9.4 FAQ structure

One page per category, each answer in all four launch languages and taken from its source document, never written from memory.

| Category | Questions it answers | Source |
|---|---|---|
| How to play | Rules, scoring, 365 target, 1v1 and 2v2 | `docs/rules.md` |
| Account and login | Sign-in methods, switching devices, lost account | #56, #57, #58 |
| Purchases | Missing purchase, restore, refunds, pass purchase | #106, #107, #125, #195 |
| Coins and rewards | How coins and rewards are earned, daily rewards, promo codes | #7, #103, #108, #131 |
| Ranked and seasons | Ranks, placement, season reset, leaderboards | `docs/ranking.md` |
| Penalties and fair play | AFK, leaving, bans, appeals | `docs/penalties.md`, #186 |
| Connection | Reconnecting, lag, timer behaviour | #45, #87, `docs/rules.md` R-49 to R-52 |
| Devices and performance | Supported devices, graphics settings, storage | `docs/performance.md`, #157, #159 |
| Privacy and data | Data use, analytics opt-out, account deletion | #187, #63 |
| Languages | Changing the language | `docs/localization.md` |

## 10. Open items

- Every `{filled by #N}` placeholder: deploy and rollback commands (#25), environments and secret rotation (#24), catalog activation (#26), build and upload pipeline (#27), dashboards, alerting, pager, incident channel, status page and contact sheet (#28), support email and tooling (#144).
- Where the FAQ is hosted (in-app view or web page) is not decided.
- All *default* numbers are unvalidated until the closed beta (#193) and soft launch (#196) produce data.
- The go/no-go checklist turns green only at launch; this document is complete when the release issue records it so.
