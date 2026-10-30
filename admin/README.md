# admin

Next.js admin panel (TypeScript, Tailwind) for content, economy, live-ops and player support. Every designer-tunable value is managed here.

The panel holds no permission logic of its own: the admin API decides every permission (`docs/admin.md` AP-2, matrix in `server/internal/rbac`). Sign-in, sessions and the audit log follow in #134.

## Run

```
cd admin
pnpm install
pnpm dev
pnpm typecheck
pnpm build
```

Node 24 and pnpm 11 (`packageManager` in `package.json`).

## Colours

`src/app/globals.css` maps the Tailwind theme to `docs/design-tokens.json` (`color.background`, `color.surface`, `color.text`, `color.role.primary`). Change the token file first, then the theme.

## Languages

The panel UI is in `az` (default) and `en` (`docs/admin.md` section 10). Texts live in `src/i18n/messages`, one file per language, with `en.ts` as the reference. Server components use `await serverTranslator()`, client components `useTranslator()`; both return the same `t(key)`. A missing or extra key in a language is a type error, so `pnpm typecheck` fails. The choice is kept in the `hd_admin_locale` cookie, and switching re-renders the page without a reload.
