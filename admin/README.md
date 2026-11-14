# admin

Next.js admin panel (TypeScript, Tailwind) for content, economy, live-ops and player support. Every designer-tunable value is managed here.

The panel holds no permission logic of its own: the admin API decides every permission (`docs/admin.md` AP-2, matrix in `server/internal/rbac`).

## Run

```
cd admin
pnpm install
pnpm dev
pnpm typecheck
pnpm build
```

Node 24 and pnpm 11 (`packageManager` in `package.json`).

## Sign-in and navigation

The home page shows the Google sign-in button (Google Identity Services, popup) when there is no session. The button gets a single-use nonce from `POST /admin/nonce`; the ID token and that nonce go to `POST /admin/session` from a server action, and the returned token is stored in the `hd_admin_session` cookie (`HttpOnly`, `SameSite=Strict`, `Secure` in production, expiring with the session, `docs/admin.md` section 9). The browser never holds the token in script. Sign-out revokes the session with `DELETE /admin/session` and drops the cookie. A refused sign-in shows the reason (`account_refused`, `not_company_account`, an expired token or nonce) and prepares a new nonce.

Every page reads `GET /admin/me` once per request (`currentAdmin` in `src/admin/session.ts`). The header lists only the screens of `src/admin/navigation.ts` whose object the account may view, and each screen checks the same rule before it calls the API, which still refuses on its own (AP-7). A session the API rejects shows the sign-in again.

An `S` action needs a sign-in no older than `AdminStepUpMinutes` (`docs/admin.md` 9). When the API answers `step_up_required`, the screen shows `StepUpPrompt`: the same Google button, whose ID token and a new nonce go to `POST /admin/session/step-up` from a server action. The session keeps its token, and a sign-in with another account is refused (`different_account`). Once it succeeds, the action is sent again with what the admin entered. `GoogleButton` holds the shared Google Identity Services code for both. Calls to the admin API forward the browser's `X-Forwarded-For` and `User-Agent`, so the audit source is the admin's browser, not the web server.

## Audit log

`/audit` (A-38) searches `GET /admin/audit` by actor, object type and id, action, reason text and a time range, newest first, 50 per page; "older entries" continues below the last id shown (`before_id`). The form takes and shows times in UTC (AP-8): a `datetime-local` value is sent as that minute in UTC. Each entry shows the actor with the roles it used, environment, reason, reference, config version, source (address, user agent, session) and its before and after values. Which entries a role sees is decided by the API (`docs/admin.md` section 7); a malformed filter shows a notice instead of results.

## Cosmetics

`/cosmetics` (A-5 and A-6, #136, `docs/admin.md` 5.1) edits the manifest of `GET /admin/config/cosmetics` without a code change. With no published manifest it starts from the built-in one.
- **Catalog:** every item with its rarity, availability, animation, sources and revision. It can be searched by id or source record and filtered by type, rarity, availability and source type.
- **Bulk edit:** rarity and availability over the checked items. The header box checks every listed item.
- **New item:** a type and a slug become `<type>.<slug>`; a taken id is refused. The id and the type never change, and a published item cannot be removed: retire or disable it.
- **Item form:** rarity, availability, art source, animation, the cultural review and the sources. Type fields follow the type and the animation flag: an emoji group, a table's ordered layers. Turning animation off drops the fields a static item must not have.
- **Server fields:** `revision`, `nameKey` and every address are set by the API on publish, so they are shown and never typed.
- **Cultural review:** only roles with the review action can set it (`cultural_review`). An item without it blocks the publish.
- **Checks:** a missing source, a malformed record id, a repeated source, a default item that is not live and static, a default action, a missing emoji group, bad table layers, or a type without a live default item marks the item and blocks the publish. Everything else, such as dropping a source another domain grants from, is refused by the API with its reason.
- **Publish** (`PUT /admin/config/cosmetics`) takes a reason and a confirmation step that lists added, changed and newly disabled items. It is shown only to roles with publish on cosmetic items.
- **History and rollback:** on `/admin/config/cosmetics/history` and `/admin/config/cosmetics/rollback`. A rollback brings content back under new revisions.

## Reward bundles

`/rewards` (#138, `docs/admin.md` 5.4) reads `GET /admin/reward-bundles`: every bundle the published config grants, with its contents, guide value and uses. It searches ids, items, medals and objects, and filters by the domain using a bundle or by ids holding differing content. Bundles live inside the objects that grant them, so an edit (`PUT /admin/reward-bundles/{id}`) writes the new content to every holder and publishes each affected domain in one transaction. The editor lists those objects and needs a confirmation and a reason. A domain's own rules still apply, so one refusal, such as a running pass season, stops the whole edit. A bundle held by rank rewards, daily login or promo codes, or an id holding differing content, is not editable here.

## Shop offers

`/shop` (A-17, `docs/admin.md` 5.5) edits the published catalog of `GET /admin/config/shop`.
- **Offers:** for the selected offer, its price in the currencies its kind and section allow, and its window. The window opens and optionally closes, shown and entered in UTC (AP-8). A window that closes before it opens blocks the publish, and so does a `mythic` item in an offer that is not time-limited: only a rotation offer, or a featured or event offer whose window closes, may sell one (`docs/economy.md` 3.1). A store product price also takes an optional note about the offer; switching to coins or crystals drops it. Country prices are not noted here: they live on the Store prices page.
- **Catalog-wide:** crystal pack amounts, and chest pools and odds.
- **Rotations:** plan the offers of a daily or weekly period ahead. The slot counts and reset hour come from the server (`rotation` in `GET /admin/config/shop`). A daily period starts at the reset hour, a weekly one on a Monday at the reset hour. Candidates are the offers of the period's rotation section; checked offers take the slots in the order they were checked. A rotation with no offers or more offers than slots blocks the publish.
- **Preview:** shows the offer ids each section of the draft would show at a chosen UTC moment, or now (`POST /admin/config/shop/preview`). Nothing is published.
- **Publish** (`PUT /admin/config/shop`) takes a reason. The API checks the catalog like the game servers, and they open and close offers on server time once they load the version.
- **History:** the last 50 versions, each with its offer count and the offers it added or changed against the version before.
- **Rollback:** roles with publish on shop offers and rollback on snapshots can roll back to an earlier version. It needs a reason and a fresh sign-in (`POST /admin/config/shop/rollback`), using the same `Rollback` component as match rules.

## Store prices

`/store-prices` (A-14, `docs/admin.md` 5.6, `docs/economy.md` 7.4) edits the `store_prices` table of `GET /admin/config/store-prices`: the USD reference price of each real-money product, tier percents, the tier of each store country (searchable by code or by its name in the panel language), and exact overrides per product and country. Before the first publish the page starts from the server's default.
- **Preview:** the target of every product in every priced country, override in bold, computed in the browser like `storeprice.Target` (`admin/src/storeprices/prices.ts`). Countries without a tier or override keep the store's own price and are not shown.
- **Publish** (`PUT /admin/config/store-prices`) takes a reason; the API refuses `RU`, an unknown tier, product or country code, a percent outside 1 to 100 and an override above its reference.
- **CSV for the consoles:** after a publish, one file per store (Google Play Console, App Store Connect) with `product_id,country_code,target_usd` rows of the published version's server-computed targets.
- **History and rollback** like the other config pages, under publish on store products and rollback on snapshots.

## Pass seasons

`/pass` (#140, `docs/admin.md` 5.7) edits the pass config of `GET /admin/config/pass` without a code change.
- **Settings:** track budgets, premium plus levels, match XP, the daily XP cap, the bot seat share and level prices apply to every season and may change at any time. An empty field keeps the server default.
- **New season:** clone a season or start an empty one, with a start and end in UTC. The new season takes the next `season_<n>`. A clone copies the curve and every coin and crystal reward under new bundle ids `bundle.pass_<season>_<track>_<level>`; cosmetics are left out, because each item must first list the new level as its `pass` source. An empty season gets an empty premium bundle per level.
- **Curve:** levels, XP base, step and max, with the XP to the next level and the running total per level as the server computes them. Changing the levels adds or trims premium levels and trims free levels above the last.
- **Level rewards:** per level, free coins and cosmetic ids, premium crystals and cosmetic ids, with the track totals against the budgets. An emptied free level is removed; a premium level must grant something. Editing a published bundle raises its revision once.
- **Running seasons:** the state (planned, running, ended) comes from server time (`rules.server_time`). Once a season has started its start, curve and level rewards are read-only; only the end of a running season may change. A planned season may be removed until it starts.
- **Checks:** a window that ends before it starts, an overlap, curve values below 1, XP max below XP base, levels not above premium plus levels, an empty premium level, a track over its budget or an invalid setting marks the field and blocks the publish. Item sources and mythic items are checked by the API.
- **Publish** takes a reason and a confirmation step that lists the added or changed seasons with their totals, dropped seasons and changed settings (`PUT /admin/config/pass`). It is shown only to roles with publish on pass seasons.
- **History and rollback:** as for shop offers, on `/admin/config/pass/history` and `/admin/config/pass/rollback`; each version lists the seasons and settings it changed.

## Ranked seasons

`/ranked` (A-19, #143, `docs/admin.md` 5.12) schedules seasons and edits the ranking config without a code change.
- **Seasons:** every season with its state (scheduled, live, ending, closed), window, reward flag and the progress of its season end per mode, all from `GET /admin/ranked/seasons`. A new season takes an id `season.<slug>` and starts at the end of the last one; its end defaults to `seasonLengthMonths` of the 1v1 config. A scheduled season may change its end, its reward flag and, if no season ends at its start, its start. The live season may only end later. Moving an end names the seasons that move with it. Only the last season can be removed, before it starts. Each change needs a reason and is audited by the API.
- **Leaderboards and coins:** the board size, the regions with their countries and `rankRewardCoins` apply to both modes.
- **Per mode:** structural values and the reset table are read-only while a season is open (`rules.open_season`). MMR, LP and placement values may change at any time.
- **Reset table:** generated from the tier drop, floor and ceiling as the server computes it. A row can be edited on its own and is marked as edited; a row set back to the generated rank stops being an override.
- **Rank rewards:** per final tier, a bundle of cosmetics and medals, `bundle.rank_<tier>` by default. Editing a published bundle raises its revision once and an emptied row removes the bundle. Item sources are checked by the API.
- **Publish** takes a reason and a confirmation step that lists the changed keys per mode (`PUT /admin/config/rating`) and warns about frozen keys. The first publish starts from the server defaults (`rules.defaults`). The abuse values and emblem brackets are kept as published; this screen does not edit them.
- **History and rollback:** on `/admin/config/rating/history` and `/admin/config/rating/rollback`; each version lists the keys it changed.

## Missions

`/missions` (#141, `docs/admin.md` 5.8) edits the mission catalog of `GET /admin/config/missions` without a code change. With no published catalog the editor starts with no missions. The API fills slot rewards, budgets and rerolls from its defaults, so the first publish needs a live daily and weekly mission for every default slot.
- **New mission:** a slug (lowercase letters, digits, underscores) becomes `mission.<slug>`; a taken id is refused. The type sets the fields: daily and weekly take a slot and a weight, achievement and event take a reward bundle `bundle.mission_<slug>`, an achievement may chain to another achievement, and an event takes an event id.
- **Condition:** the counted progress event, the target, and optional queue and mode filters. Events, types, queues and modes come from the server (`rules` in the response), so the editor offers only what the game servers count.
- **Reward:** periodic missions pay their slot's reward from the catalog. An achievement or event bundle grants coins, cosmetic ids and medal ids, and an event bundle also pass XP. Editing a published bundle raises its revision once; restoring the published content restores its revision.
- **Server fields:** `nameKey` and the mission revision are shown, never typed; the API sets them on publish.
- **Checks:** a target, slot or weight below 1, a slot outside the type's slots, a bundle granting nothing or a malformed event id marks the mission and blocks the publish. Everything else, such as a granted medal or cosmetic that does not list the mission as a source, is refused by the API with its reason.
- **Publish** (`PUT /admin/config/missions`) takes a reason and is shown only to roles with publish on missions. A published mission is never removed; set it to retired.
- **History and rollback:** as for shop offers, on `/admin/config/missions/history` and `/admin/config/missions/rollback`.

## Assets

`/assets` (#135, `docs/admin.md` 5.2) uploads the static PNG of a cosmetic item. The panel asks `POST /admin/assets/uploads` for a signed URL, the browser sends the file straight to the object store, and `POST /admin/assets/{id}/check` makes the server read it back and record format, alpha, width, height, bytes and SHA-256. A file that fails stays listed as rejected and is never shown; a new upload makes a new row. The file does not reach players yet.

The browser needs the store in two places: `HD_ADMIN_ASSET_ORIGIN` adds its origin to `img-src` and `connect-src` of the Content Security Policy, and the bucket's CORS rules must allow `PUT` from the panel's origin. Without `HD_ASSET_STORE_URL` on the admin API, the page lists rows and closes uploads.

## Daily login

`/daily-login` (#141, `docs/admin.md` 5.8, `docs/economy.md` 2.3) edits the coins of each day of the calendar and publishes through `/admin/config/daily-login`, with a reason, history and rollback. Until a calendar is published the server has none and claims are refused, which the page says. The number of days is not edited here: a 30-day calendar is not proposed for launch.

## Medals

`/medals` (#141, `docs/admin.md` 5.9) edits the medal catalog of `GET /admin/config/medals`. With no published catalog the editor starts empty and the first publish creates it.
- **Catalog-wide:** the number of profile showcase slots, a whole number of 0 or more.
- **New medal:** a slug (lowercase letters, digits, underscores) becomes `medal.<slug>`; a taken id is refused. It starts with the server's first rarity and availability and one source to fill.
- **Fields:** rarity and availability (live, retired, disabled) come from the server (`rules` in the response). Sources are a list of a type (mission or rank reward) and an id; a medal needs at least one, every source needs an id, and a source listed twice blocks the publish.
- **Server fields:** `nameKey`, `descriptionKey`, `staticAddress` and the revision are shown, never typed; the API sets them on publish.
- **Order with missions:** a medal must list every published mission that grants it, and the API refuses a medal publish that drops one. Publish a new medal listing its mission first, then the mission that grants it.
- **Publish** (`PUT /admin/config/medals`) takes a reason and is shown only to roles with publish on medals. A published medal is never removed; retire or disable it.
- **History and rollback:** as for shop offers, on `/admin/config/medals/history` and `/admin/config/medals/rollback`.

## Game strings

`/strings` (#214, `docs/admin.md` 5.15, A-27 and A-28) edits game strings over `/admin/strings`.
- **Missing translations:** per language, approved values over live keys, the percentage and the list of missing keys.
- **Keys:** search over keys, notes and texts, and a filter to the keys a language is missing. Languages sit side by side. A new key, a note, the max length and the placeholders (`name:type`, comma separated) are saved with `PUT /admin/strings/{key}`; ticking deprecated retires the key.
- **Values:** each cell saves its text with status `draft`, or `approved` for roles with edit and publish on game strings.
- **Import and export:** a language as `csv` or `json` for an external translator. The panel sends a file through a server action, so it takes files up to 1 MB; larger ones go straight to `POST /admin/strings/import`. A blocked import lists each failing key.
- **Bundles:** publish builds the next bundle of a language; a blocked publish lists each failing key. Activating a newer bundle publishes the next `localization` snapshot with a reason. Activating an older one is a rollback and needs rollback on snapshots and a fresh sign-in.

## Players

`/players` (A-30 lookup, `docs/admin.md` 6.1) searches `GET /admin/players?q=` by account id, exact email or nickname prefix, up to 20 results. A `super_admin` also gets a second form that finds the account of one sign-in identity, a provider (`google` or `apple`) and its subject, through `GET /admin/players?provider=&subject=` (D-2). A search writes no audit entry. `/players/{id}` shows the record from `GET /admin/players/{id}`: account state, country, emails (sign-in identities only for `super_admin`), the temporary ban while it runs, balances, level, the forced-rename flag, nickname history, and the newest 50 ledger entries, inventory items, reward grants, purchases and matches. The API writes a `view` entry for every opening (AP-6), so the links to a record do not prefetch it: a prefetch would log a view nobody made.

Roles with `act_on_player` (support and `super_admin`) get the support actions of 6.2 on a record that is not deleted. Each one needs a reason and a fresh sign-in, and `StepUpPrompt` covers the second.
- **Coins:** a whole, non-zero amount; a minus sign debits. The form keeps one `request_id` until the adjustment is applied, so a resend after step-up cannot apply it twice (`repeated`). An amount above `AdminGrantApprovalCoins` becomes a proposal for a second admin (202).
- **Temporary ban:** ban or lift, whichever applies now. The ban is refused with `not_configured` while `HD_TEMP_BAN_HOURS` is unset.
- **Grant:** a reward bundle (bundle id, coins, pass XP, cosmetic and medal ids) under a source picked by type and id, such as the event or campaign whose reward the player missed. Every cosmetic and medal must list that source, or the API refuses the grant and says which one. Like coins, the form keeps one `request_id` until the grant is applied. Above `AdminGrantApprovalCoins` coins or with a `mythic` item it becomes a proposal, which `/approvals` shows as bundle, grants and source.
- **Force rename:** offered while the player is not already flagged.

**Replay `/matches/{id}`** (D-13, for support and `super_admin`) opens from the Replay link in a player's matches. It shows:
- the match and its seats, with links to the players;
- every hand: the tiles dealt to each seat in seat order, each move with tile, end and points, and the hand's outcome.

A record that does not decode, is refused by the engine or ends on another result shows the problem above the hands. Each opening is a `view_replay` audit entry.

The page reloads the record after each change.

## Approvals

`/live-controls` (A-25, `docs/admin.md` section 4, #210) is in the nav for roles that may view live controls. It edits the purchase and ranked kill switches, the minimum app version and the build allowlist of `docs/anti-cheat.md` 5.2. A change needs a reason and a fresh sign-in. Staging applies it at once; production waits in `/approvals` for a second admin.

`/reviews` (A-31, `docs/admin.md` 6.4, #209) is in the nav for roles that may view queues. It lists the open reviews of `GET /admin/reviews` with their signals and evidence, and links each account to its player record. Clear lifts a shadowban and confirm keeps it (`docs/anti-cheat.md` 5.3). Both need a reason and a fresh sign-in. An admin who wants a temporary ban uses the player record first, then confirms.

`/approvals` (two-person rule, `docs/admin.md` section 8) is in the nav for every signed-in admin. It lists what `GET /admin/approvals` returns: pending proposals the admin's roles may decide, and their own. A row shows the kind, the object (a player links to the record), the change without its bookkeeping ids, the reason, the proposer with their roles, and the expiry. An admin's own proposal says another admin decides. On the others, Approve and Reject take an optional note. Approve needs a fresh sign-in and uses `StepUpPrompt`, and the list reloads after a decision.
A `promo_code` proposal reads as its code count, campaign, bundle and window, and links to `/promo`.

## Promo codes

`/promo` (A-15, `docs/admin.md` 5.11, #142) is in the nav for roles that may view promo codes. It needs `HD_PROMO_KEY` on the admin API; without it the screen says promo codes are off.

- **Usage:** per campaign, the codes, how many are active, pending approval or disabled, redemptions against the summed cap, and the window. A campaign links to its codes.
- **Codes:** the newest 200, or those of one campaign: type, state (an active code past its end shows `expired`), what it grants, window, redemptions and cap, creator and approver.
- **Create (`live_ops`, `super_admin`):** one code of any type, or a batch of up to 10,000 `single_use` codes. The bundle is a bundle id with coins, pass XP and cosmetic item ids. The window is entered in UTC. A `multi_use` code takes a cap and an optional chosen code. A `targeted` code starts without accounts. A reason is required.
  - **One code:** its plaintext is shown once.
  - **Batch:** downloads as `<id>.csv` from the browser. Nothing keeps the plaintext afterwards, beyond a "download again" button while the page stays open.
  - **Second admin:** a code above the coin threshold or with a `mythic` item says so and waits in `/approvals`.
- **Templates (`docs/launch-content-gate.md` 3.6):** every role that may view promo codes sees them: id, campaign, type, what they grant (or "set when cloned" without a bundle), creator and last change. In the create form, "Save as template" keeps the id, campaign, type and bundle with a reason; on an existing id it updates that template. "Start from a template" fills the form, and the code still needs its own window, cap or code. A template never redeems and is not listed with the codes.
- **Code page `/promo/{id}`:** the code's fields for every role that may view it. `support` also sees who redeemed and when, linked to the player. On a targeted code that is not disabled, `support` edits the account list: ids by line or comma, a reason and a fresh sign-in. The screen keeps the save off while an account that redeemed is missing (`docs/admin.md` 9.1).
- **Disable (`live_ops`):** a code, or the whole batch from its first listed code. It takes a reason and a fresh sign-in (`StepUpPrompt`) and is immediate.

## Match rules

`/config/rules` is the `rules` part of A-26 (`docs/admin.md` 5.14). It lists the keys of `GET /admin/config/rules` with their unit, allowed range, default and current value. Roles with publish on config values may edit them; a value outside the range or not a whole number blocks the publish before the API sees it, and the API checks the bounds again. A publish (`PUT /admin/config/rules`) sends only the values that differ from their defaults, with the reason, and the page then shows the new `config_version`. Below, the last 50 published versions show their non-default values. A value applies to matches that open after the game servers pick it up (`HD_CONFIG_POLL`). Each earlier version has a rollback for roles with publish on config values and rollback on snapshots. It needs a reason and a fresh sign-in, and it publishes that version's values as the next version (`POST /admin/config/rules/rollback`).

## Economy monitors

`/analytics` is the economy sink and source part of A-36 (#147, `docs/economy.md` section 10). Every role with view on dashboards sees it. It shows `GET /admin/economy/monitors` with the alerts raised on top: the 7-day window, coins created and spent, the sink ratio and the median balance, then the economy day with the accounts above the daily earn alert and the coins created per reason. A config version in the form adds the same monitors for the 7 days before and after its publish (`GET /admin/economy/monitors/compare`) as two more columns. Only aggregates are shown, never an account id.

## Security headers

`src/proxy.ts` makes a fresh 128-bit nonce for every page request and sends a Content Security Policy (`src/security/headers.ts`): scripts only with that nonce and `'strict-dynamic'`, no `'unsafe-inline'`, `'unsafe-eval'` only under `next dev`; no framing (`frame-ancestors 'none'`), forms only to the panel, no `<base>`, no plugins. Google sign-in may load its button frame, its stylesheet and its calls under `https://accounts.google.com/gsi/` and nothing wider. Next.js puts the nonce on its own scripts; the sign-in button adds the Google script with the nonce of the loaded document, so it keeps working after a client-side navigation, when the page's nonce no longer matches the document's policy. `next.config.ts` adds `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, `X-Frame-Options: DENY`, `Cross-Origin-Opener-Policy: same-origin-allow-popups` (the sign-in popup) and a `Permissions-Policy` to every response, and HSTS in production.

Cross-site requests: the session cookie is `SameSite=Strict`, every change goes through a server action, which Next.js accepts only when the `Origin` header matches the host, and no `GET` changes anything. If the access proxy in front of the panel rewrites the host, its public host goes into `serverActions.allowedOrigins`.

The admin API limits sign-in per address and requests per account (`docs/admin.md` 9.1); a sign-in refused with `rate_limited` shows `signin.rate_limited`.

## Hardening check

The #148 acceptance is Lighthouse above 90 and no critical vulnerability (`docs/test-strategy.md` 2.6). To repeat it locally:

1. Start PostgreSQL. In `server/`, run `go run ./cmd/migrate up`, then `go run ./cmd/admin-api -bootstrap-email root@example.com -bootstrap-name Root`, then run the admin API.
2. Google sign-in needs a real client, so a local-only session stands in for it: insert an `admin_sessions` row for that account with `token_hash = sha256('<local token>')`, and send `Cookie: hd_admin_session=<local token>`. Never do this outside a throwaway database.
3. `pnpm build`, then `pnpm start` with `HD_ADMIN_API_URL` pointing at the local API.
4. Run `lighthouse <url> --preset=desktop` and the default mobile run on `/` signed out, and on `/`, `/audit` and `/shop` with `--extra-headers '{"Cookie":"hd_admin_session=<local token>"}'`.
5. Run `pnpm audit` here and `govulncheck ./cmd/admin-api/...` in `server/`.

Last result: 2026-11-03, Lighthouse 13.6.0, headless Chrome on Windows, empty catalog and one audit entry. The scores are performance / accessibility / best practices.

| Page | Desktop | Mobile |
|---|---|---|
| `/` signed out (sign-in) | 100 / 100 / 100 | 93 / 100 / 100 |
| `/` signed in | 99 / 100 / 100 | 97 / 100 / 100 |
| `/audit` | 100 / 100 / 100 | 97 / 100 / 100 |
| `/shop` | 100 / 100 / 100 | 98 / 100 / 100 |

Mobile LCP is 0.7 to 0.8 s and CLS is 0. SEO scores 50 by design: every page is `noindex`, which is the failing audit. `pnpm audit` and `govulncheck` found no known vulnerabilities.

The same run probed the panel and the API:
- Every response carried the headers above.
- `/audit` without a session, or with a forged cookie, showed only the sign-in prompt.
- A server action posted with a foreign `Origin` was refused, and one with the panel's origin passed.
- The API answered 401 `session_rejected` without a bearer token, and the eleventh nonce from one address within a minute got 429 with `Retry-After`.

## Configuration

`HD_ADMIN_API_URL` and `HD_ADMIN_GOOGLE_CLIENT_ID` (`docs/environments.md` 2.2), read on the server only.

## Colours

`src/app/globals.css` maps the Tailwind theme to `docs/design-tokens.json` (`color.background`, `color.surface`, `color.text`, `color.role.primary`). Change the token file first, then the theme.

## Languages

The panel UI is in `az` (default) and `en` (`docs/admin.md` section 10). Texts live in `src/i18n/messages`, one file per language, with `en.ts` as the reference. Server components use `await serverTranslator()`, client components `useTranslator()`; both return the same `t(key)`. A missing or extra key in a language is a type error, so `pnpm typecheck` fails. The choice is kept in the `hd_admin_locale` cookie, and switching re-renders the page without a reload.
