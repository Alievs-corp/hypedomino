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

## Configuration

`HD_ADMIN_API_URL` and `HD_ADMIN_GOOGLE_CLIENT_ID` (`docs/environments.md` 2.2), read on the server only.

## Colours

`src/app/globals.css` maps the Tailwind theme to `docs/design-tokens.json` (`color.background`, `color.surface`, `color.text`, `color.role.primary`). Change the token file first, then the theme.

## Languages

The panel UI is in `az` (default) and `en` (`docs/admin.md` section 10). Texts live in `src/i18n/messages`, one file per language, with `en.ts` as the reference. Server components use `await serverTranslator()`, client components `useTranslator()`; both return the same `t(key)`. A missing or extra key in a language is a type error, so `pnpm typecheck` fails. The choice is kept in the `hd_admin_locale` cookie, and switching re-renders the page without a reload.
