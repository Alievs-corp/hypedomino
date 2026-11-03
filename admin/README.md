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

Every page reads `GET /admin/me` once per request (`currentAdmin` in `src/admin/session.ts`). The header lists only the screens of `src/admin/navigation.ts` whose object the account may view, and each screen checks the same rule before it calls the API, which still refuses on its own (AP-7). A session the API rejects shows the sign-in again. Calls to the admin API forward the browser's `X-Forwarded-For` and `User-Agent`, so the audit source is the admin's browser, not the web server.

## Audit log

`/audit` (A-38) searches `GET /admin/audit` by actor, object type and id, action, reason text and a time range, newest first, 50 per page; "older entries" continues below the last id shown (`before_id`). The form takes and shows times in UTC (AP-8): a `datetime-local` value is sent as that minute in UTC. Each entry shows the actor with the roles it used, environment, reason, reference, config version, source (address, user agent, session) and its before and after values. Which entries a role sees is decided by the API (`docs/admin.md` section 7); a malformed filter shows a notice instead of results.

## Match rules

`/config/rules` is the `rules` part of A-26 (`docs/admin.md` 5.14). It lists the keys of `GET /admin/config/rules` with their unit, allowed range, default and current value. Roles with publish on config values may edit them; a value outside the range or not a whole number blocks the publish before the API sees it, and the API checks the bounds again. A publish (`PUT /admin/config/rules`) sends only the values that differ from their defaults, with the reason, and the page then shows the new `config_version`. Below, the last 50 published versions show their non-default values. A value applies to matches that open after the game servers pick it up (`HD_CONFIG_POLL`). Rollback needs a fresh sign-in (step-up), which the panel does not offer yet, so it is not on this screen.

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
