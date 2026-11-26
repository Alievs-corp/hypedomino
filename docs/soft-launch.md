# Soft launch

How the game is released in a few countries before the global launch (#196): how the countries are chosen, the phases, how fixes ship fast, which KPIs decide whether to expand, hold or stop, and the weekly KPI report that feeds those decisions. Epic 17 (#189).

**Status: draft, awaiting owner approval.** Every rule here is a proposal the owner approves by merging the pull request that adds or changes it. A number that the repository has not already decided is not set here: it is an owner placeholder written `<SL-n>` and listed in section 8. Until the owner fills it, the rule that needs it runs without a threshold, as `docs/live-ops-90-days.md` OD-4 already decides for the KPI targets.

Owned elsewhere and only referenced here: go/no-go checklist, timeline, deploy, rollout, rollback, kill switches, incidents and support (`docs/launch-runbook.md`), beta phases, signals and exit (`docs/closed-beta.md`), calendar, balance process and weekly review after the global launch (`docs/live-ops-90-days.md`), personal data and store countries (`docs/privacy.md`), prices per country (`docs/economy.md` 7.4), KPI screens (`docs/admin.md` A-36, `admin/README.md` KPIs and cohorts), store countries in the consoles (`docs/store-listing.md` ST-19), server regions (ADR 0010). Roles are those of `docs/launch-runbook.md` section 1.

## 1. Where the soft launch sits

| Stage | Ends with | Defined in |
|---|---|---|
| Closed beta CB-1 to CB-3 | Exit record on #193; beta game state reset; beta builds blocked | `docs/closed-beta.md` 6.1, 7.1 |
| Soft launch (this document) | Owner decision to go global (section 5.4) | Here |
| Global launch, Day 0 | Season 1 starts; soft launch rank and pass progress reset, coins, cosmetics and medals stay (OD-1) | `docs/live-ops-90-days.md` section 2 |

- The soft launch is the first store release. `docs/launch-runbook.md` applies to it in full: release issue, go/no-go, T-0 timeline and the launch window to T+72h. Exposure is limited by the store country list, not by percentage rollout (`docs/launch-runbook.md` 5.2).
- It runs on the `production` environment in the single region Germany (ADR 0010 phase 1), with the real economy (`docs/closed-beta.md` EC-1).
- It has its own ranked season, created without rank rewards, and its own pass season; both end at Day 0 (`docs/live-ops-90-days.md` section 2, `docs/ranking.md` 7.1 `-no-rewards`).
- Legal documents are published before T-0 (go/no-go GN-4). Drafts: [`docs/legal/`](legal/README.md).

## 2. Region choice

### 2.1 Hard constraints

A country that fails one of these is not a soft launch country.

| Id | Constraint | Source |
|---|---|---|
| SR-1 | It is a store country: every country except Russia | `docs/privacy.md` 9.3, OQ-2 |
| SR-2 | The controller duties for players there are met before T-0: EU and UK representatives where needed (L-3); for Turkey, VERBIS through a Turkish representative (L-4) and the KVKK standard contract with every processor (L-9); for Azerbaijan, if the controller is an Azerbaijani entity, the information system registration and licence answer (L-29) | `docs/privacy.md` section 13, 12.2 |
| SR-3 | The stores can charge there, and the country has a price in the published `store_prices` table, so real-money offers are shown and revenue can be measured | `docs/economy.md` 7.4 |
| SR-4 | A launch language (`az`, `en`, `tr`, `ru`) is a language most players there read, and the support FAQ and templates exist in it | `docs/localization.md` section 1, `docs/launch-runbook.md` GN-26 |
| SR-5 | The move round-trip from players there to Germany fits the budget: p99 under 100 ms (`CLAUDE.md`, `docs/performance.md`). The server metric is not split by player country, so a candidate is checked before the choice with test devices on mobile networks in that country | ADR 0010 |

### 2.2 Weighed criteria

| Id | Criterion | Why |
|---|---|---|
| SW-1 | Enough players at the same time to fill ranked queues, which never use bots, and few enough that a defect reaches a contained group | `docs/ranking.md` section 9; `docs/matchmaking.md` |
| SW-2 | Device mix covering the tiers of `docs/performance.md` section 1 on both platforms | Performance budgets are proven on real devices |
| SW-3 | Spending behaviour close to the global launch markets; the `store_prices` tier of the country tells how prices compare | `docs/economy.md` 7.4, OQ-8 |
| SW-4 | Little media and store-chart spillover into the main launch markets | A soft launch result should not spend the global launch's first impression |
| SW-5 | Both Google Play and the App Store are commonly used | iOS and Android KPIs are both needed |

### 2.3 Known facts about candidates

Only facts already in the repository. The choice is owner decision `<SL-1>`.

| Candidate | Language | Price tier | Legal state |
|---|---|---|---|
| Azerbaijan | `az` | `mid` | AZ-1 no localisation needed; AZ-2 to AZ-4 if the controller is an Azerbaijani entity (L-29, waits on L-1) |
| Turkey | `tr` | `mid` | Blocked until L-4 (VERBIS through a representative) and L-9 (standard contracts) hold |
| EU member states | `en` (and `ru` where spoken) | Per `store_prices` | GDPR baseline; EU representative if the controller has no EU establishment (L-3) |

## 3. Phases

| Id | Phase | Entry | What happens | Length |
|---|---|---|---|---|
| SL-0 | Readiness | Closed beta exit recorded (#193) | Beta reset (`docs/closed-beta.md` 6.1); soft launch ranked and pass seasons published; store countries set to the soft launch list (ST-19); purchase kill switch per `<SL-3>`; release issue opened and `docs/launch-runbook.md` section 3 worked through | Until go |
| SL-1 | Launch window | Go decision (`docs/launch-runbook.md` T-1d) | T-0 to T+72h as `docs/launch-runbook.md` section 4: config freeze, hourly then daily health reviews | 72 hours |
| SL-2 | Baseline | T+72h | KPIs collected and reported weekly (section 6) without thresholds; only fixes and incident mitigation ship. At the end the owner sets the KPI targets (`docs/live-ops-90-days.md` OD-4 to OD-6) | The first two economy weeks of data (OD-4) |
| SL-3 | Iterate | Targets set | Weekly report and review (section 6); balance changes through `docs/live-ops-90-days.md` 6.2; fixes through section 4 | `<SL-2>` |
| SL-4 | Decision | End of SL-3, or a stop condition (5.4) | Expand, hold or stop (5.4) | One weekly review |
| SL-5 | Expansion | Expand decision for more countries | The new countries pass section 2; the store country list grows; SL-1 rules apply to the change | Back to SL-3 |

A phase advances only by a product owner decision at the weekly review, recorded on #196.

## 4. Hot fixes

Mitigate first with the fastest step that stops the harm, in this order (`docs/launch-runbook.md` 5.4):

1. Config: publish the previous config version, or a corrected one. New matches use it at once; no release.
2. Feature flag or kill switch (`docs/launch-runbook.md` 5.5): purchase, ranked, maintenance mode, build allowlist, promo code disable.
3. Server: a fix merged with `CI complete` green, deployed through `Server deploy` with the canary and drain of `docs/launch-runbook.md` 5.1; rollback under 5 minutes if it fails.
4. Client: a hotfix build with a higher version and build number from a tag (`Release build`), expedited App Store review requested, rolled out as `docs/launch-runbook.md` 5.2. The minimum version is raised only after the fixed build is available in both stores.

- Severity, response targets and the incident flow are `docs/launch-runbook.md` section 6. A SEV1 or SEV2 gets a postmortem with action issues.
- Every fix is an issue with a `priority:` label, one branch and one PR (`CLAUDE.md`). A fix found in the soft launch says "Found in soft launch (#196)" and its build number in the issue body, never player data.
- Hot fixes ship outside any release cadence. The regular cadence in the soft launch is `<SL-7>`.

## 5. Decision criteria

### 5.1 What is measured today

| KPI | Definition | Where | Population | State |
|---|---|---|---|---|
| Completed matches per day | Ended, not cancelled matches per economy day of their end and mode | Admin `/analytics/kpis` (`GET /admin/analytics/kpis`) | All players | Computed |
| New players per day and week | Size of the daily and weekly cohort: analytics ids with their first `app_open` in that economy day or week | Same | Players with `analytics` consent | Computed |
| D1 retention | Share of a daily or weekly cohort with an `app_open` on day 1 | Same | Players with `analytics` consent | Computed |
| D7 retention | Same, on day 7 | Same | Players with `analytics` consent | Computed |
| Matches per active player per day | Completed matches divided by daily active accounts | Not computed | All players | Waits on the DAU definition (#852, #147) |
| ARPDAU | Revenue per daily active account | Not computed | All players | Waits on the revenue price source decision (#852, #147): purchase records keep no price or currency (`docs/analytics.md` section 3), and on the DAU definition |
| Crash-free sessions and users | Sessions or users without a fatal error, per platform and app version | Sentry release health (#182) | Players with `diagnostics` consent | Needs owner input OI-6 |
| Android vitals | Crash rate and ANR rate | Play Console | Devices sharing usage data with Google | Available at T-0 |
| Move round-trip p99 | `move_roundtrip_seconds` | Server `/metrics` (#28) | All matches | Computed; dashboards need OI-7 |
| Login, queue wait, receipt validation | `hd_signins_total`, `hd_matchmaking_wait_seconds`, `hd_receipt_validations_total` | Server `/metrics` | All | Computed; thresholds are OI-8 |
| Match completion | Matches with a result divided by matches started, per mode (`docs/closed-beta.md` BS-8) | Not computed | All | No metric yet |
| Economy monitors | Sink ratio, median balance, account daily earn | Admin `/analytics` (#110) | All | Computed |

Caveats that every report states:

- Cohorts and retention count only players who turned `analytics` on, and count analytics ids, not accounts. An id is replaced when a player withdraws consent and differs per device, so one person can appear in more than one cohort. Whether cohorts move to accounts is open on #852.
- The soft launch countries are small. A share from a cohort below `<SL-8>` players is reported but not used for a decision.
- Paid acquisition, if the owner buys any (`<SL-9>`), cannot be attributed to a campaign: the game ships no attribution SDK (`docs/privacy.md` P-3). Cohorts measure all installs of a week together.

### 5.2 Health gates

Thresholds already decided in the repository. Each is measured per platform and per soft launch week.

| Id | Gate | Threshold | Source |
|---|---|---|---|
| G-1 | Crash-free sessions per app version | At least 99.5% | `docs/launch-runbook.md` 4.1, `docs/closed-beta.md` BX-1 |
| G-2 | Android vitals | Below the Play Console bad behaviour thresholds | `docs/launch-runbook.md` 4.1 |
| G-3 | Move round-trip p99 | Under 100 ms | `CLAUDE.md`, `docs/performance.md` |
| G-4 | Match completion per mode | At least 99.5% (*default*); not gated until the metric exists | `docs/closed-beta.md` BX-3 |
| G-5 | Economy integrity | No unexplained ledger mismatch or hard signal; no open economy integrity incident | `docs/closed-beta.md` BX-6 |
| G-6 | Sink ratio | Inside `SinkRatioAlertPercent` (60% to 120%) | `docs/economy.md` section 10 |
| G-7 | Incidents | No open SEV1 or SEV2, no open `priority:p0` issue found in the soft launch | `docs/launch-runbook.md` 6.1 |

### 5.3 KPI targets

Set by the owner from the first two weeks of soft launch data (`docs/live-ops-90-days.md` OD-4). Until then they are placeholders and the report shows the values without a verdict.

| KPI | Target | Stop floor |
|---|---|---|
| D1 retention, weekly cohort | `<SL-4a>` | `<SL-5a>` |
| D7 retention, weekly cohort | `<SL-4b>` | `<SL-5b>` |
| Completed matches per day; per active player once DAU is defined | `<SL-4c>` | `<SL-5c>` |
| ARPDAU, once the price source and DAU are decided | `<SL-4d>` | `<SL-5d>` |

### 5.4 Decision rules

Taken at the weekly review (section 6) by the product owner and recorded on #196.

| Decision | When | Next |
|---|---|---|
| Fix first | Any health gate (5.2) is red | No expansion. Section 4; the gate must be green for a full week before any other decision |
| Expand | Every health gate green for `<SL-6>` consecutive weeks and every KPI target met or waived by the owner | More countries (SL-5) or the global launch go/no-go (`docs/launch-runbook.md` section 3, GN-6) |
| Hold and iterate | Gates green, a KPI below its target and above its stop floor | Stay in SL-3; the weekly review sets the next changes |
| Stop | A KPI below its stop floor for `<SL-2>` weeks of iteration | Owner decides: pause and rework before a new soft launch, or end the project |

- A region whose move round-trip breaks G-3 while the server is healthy is not a reason to stop: it is the trigger for a game-only cluster closer to players (ADR 0010 phase 2).
- The decision and the gate table are posted on #196 every week; KPI values follow `<SL-10>`.

## 6. Weekly KPI report

The analyst prepares it once per economy week (Monday 00:00 UTC to Monday 00:00 UTC), from the first week after T-0 until the global launch, and posts it before the weekly review. From Day 0 the same report feeds the weekly live-ops review (`docs/live-ops-90-days.md` 6.2 step 1, 7.3).

### 6.1 Steps

1. Admin `/analytics/kpis`, weeks set to the number of soft launch weeks (at most 13): read the cards and tables, then "Export CSV" (`kpis-<from>-<to>.csv`, columns `table,start,mode,metric,value`). Pending D1 and D7 values stay pending; they are never estimated.
2. Admin `/analytics`: economy monitors for the week and around each config publish of the week, with their CSV export.
3. Sentry release health: crash-free sessions and users per platform and app version; top new crash groups and their issues.
4. Play Console: Android vitals per version. App Store Connect: crashes per version.
5. Server dashboards: move round-trip p99, login success per provider, queue wait per queue, receipt validation per store (`docs/launch-runbook.md` 4.1).
6. Support tool: tickets per category and new symptoms; store review themes.
7. Fill the template (6.2), compare with 5.2 and 5.3, and write a proposed decision (5.4).
8. Store it where `<SL-10>` says; post the gate table and proposed decision on #196.

### 6.2 Template

| Part | Content |
|---|---|
| Week | Economy week, phase (section 3), soft launch countries, builds in use and their share |
| Gates | G-1 to G-7 with value, threshold and state, week over week |
| Engagement | New players per week, D1 and D7 per weekly cohort with cohort sizes, completed matches per day and mode |
| Revenue | ARPDAU once computable; until then "waits on #852" |
| Economy | Sink ratio, median balance, account daily earn alerts |
| Stability detail | Top crash groups, vitals, incidents of the week |
| Feedback | Ticket themes, store review themes |
| Changes | Config publishes, releases and hot fixes of the week |
| Proposal | Decision per 5.4 and the next week's priorities |

- The report holds aggregates only: no account id, analytics id, nickname, email or screenshot with personal data (`docs/privacy.md` P-7, `docs/closed-beta.md` TP-5). The repository is public.

## 7. Implementation and open items

| Item | Needed for | Owner of the work |
|---|---|---|
| DAU definition | Matches per active player, ARPDAU | Owner decision on #852 (#147) |
| Revenue price source | ARPDAU | Owner decision on #852 (#147), then #147 |
| Cohorts by analytics id or by account | Cohort counts | Owner decision on #852 (#147) |
| Match completion metric | G-4 | #28, #50 |
| Crash reporting account | G-1 | Owner input OI-6 |
| Dashboards, alert thresholds | G-3, step 5 | Owner inputs OI-7, OI-8 |
| Funnel (install, login, first match, purchase) | Go/no-go GN-14 | #181, #147; funnel steps open on #852 |

## 8. Owner decisions

| Id | Question | Proposed default |
|---|---|---|
| SL-1 | Soft launch countries | None proposed; section 2 is the test each candidate passes |
| SL-2 | Length of SL-3, and weeks below a stop floor before a stop decision | Not proposed |
| SL-3 | Real-money purchases in the soft launch | On (purchase kill switch off), since ARPDAU and receipt validation cannot be measured otherwise |
| SL-4 | KPI targets for D1, D7, matches per day and ARPDAU | Set from the first two weeks of data (OD-4) |
| SL-5 | Stop floors for the same KPIs | Set with SL-4 |
| SL-6 | Consecutive green weeks before an expand decision | Not proposed |
| SL-7 | Release cadence in the soft launch | The cadence of `docs/live-ops-90-days.md` OD-7 (server weekly, client every two weeks, Tuesday); hot fixes outside it |
| SL-8 | Minimum cohort size for a share to count in a decision | Not proposed |
| SL-9 | Paid user acquisition in the soft launch, and its budget | Not proposed; any campaign is measured without attribution (5.1) |
| SL-10 | Where the weekly report is kept and which figures go on #196 | As `docs/closed-beta.md` OQ-11: the full report in the team's private workspace; #196 gets the gate table and the decision, without engagement or revenue figures |
