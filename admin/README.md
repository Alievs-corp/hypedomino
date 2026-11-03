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

## Shop offers

`/shop` (A-17, `docs/admin.md` 5.5) edits the published catalog of `GET /admin/config/shop`.
- **Offers:** for the selected offer, its price in the currencies its kind and section allow, and its window. The window opens and optionally closes, shown and entered in UTC (AP-8). A window that closes before it opens blocks the publish.
- **Catalog-wide:** crystal pack amounts, and chest pools and odds.
- **Rotations:** plan the offers of a daily or weekly period ahead. The slot counts and reset hour come from the server (`rotation` in `GET /admin/config/shop`). A daily period starts at the reset hour, a weekly one on a Monday at the reset hour. Candidates are the offers of the period's rotation section; checked offers take the slots in the order they were checked. A rotation with no offers or more offers than slots blocks the publish.
- **Publish** (`PUT /admin/config/shop`) takes a reason. The API checks the catalog like the game servers, and they open and close offers on server time once they load the version.
- **History:** the last 50 versions, each with its offer count and the offers it added or changed against the version before.
- **Rollback:** roles with publish on shop offers and rollback on snapshots can roll back to an earlier version. It needs a reason and a fresh sign-in (`POST /admin/config/shop/rollback`), using the same `Rollback` component as match rules.

## Players

`/players` (A-30 lookup, `docs/admin.md` 6.1) searches `GET /admin/players?q=` by account id, exact email or nickname prefix, up to 20 results. A search writes no audit entry. `/players/{id}` shows the record from `GET /admin/players/{id}`: account state, country, emails (sign-in identities only for `super_admin`), the temporary ban while it runs, balances, level, the forced-rename flag, nickname history, and the newest 50 ledger entries, inventory items, reward grants, purchases and matches. The API writes a `view` entry for every opening (AP-6), so the links to a record do not prefetch it: a prefetch would log a view nobody made.

Roles with `act_on_player` (support and `super_admin`) get the support actions of 6.2 on a record that is not deleted. Each one needs a reason and a fresh sign-in, and `StepUpPrompt` covers the second.
- **Coins:** a whole, non-zero amount; a minus sign debits. The form keeps one `request_id` until the adjustment is applied, so a resend after step-up cannot apply it twice (`repeated`). An amount above `AdminGrantApprovalCoins` becomes a proposal for a second admin (202).
- **Temporary ban:** ban or lift, whichever applies now. The ban is refused with `not_configured` while `HD_TEMP_BAN_HOURS` is unset.
- **Force rename:** offered while the player is not already flagged.

The page reloads the record after each change.

## Approvals

`/approvals` (two-person rule, `docs/admin.md` section 8) is in the nav for every signed-in admin. It lists what `GET /admin/approvals` returns: pending proposals the admin's roles may decide, and their own. A row shows the kind, the object (a player links to the record), the change without its bookkeeping ids, the reason, the proposer with their roles, and the expiry. An admin's own proposal says another admin decides. On the others, Approve and Reject take an optional note. Approve needs a fresh sign-in and uses `StepUpPrompt`, and the list reloads after a decision.

## Match rules

`/config/rules` is the `rules` part of A-26 (`docs/admin.md` 5.14). It lists the keys of `GET /admin/config/rules` with their unit, allowed range, default and current value. Roles with publish on config values may edit them; a value outside the range or not a whole number blocks the publish before the API sees it, and the API checks the bounds again. A publish (`PUT /admin/config/rules`) sends only the values that differ from their defaults, with the reason, and the page then shows the new `config_version`. Below, the last 50 published versions show their non-default values. A value applies to matches that open after the game servers pick it up (`HD_CONFIG_POLL`). Each earlier version has a rollback for roles with publish on config values and rollback on snapshots. It needs a reason and a fresh sign-in, and it publishes that version's values as the next version (`POST /admin/config/rules/rollback`).

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
