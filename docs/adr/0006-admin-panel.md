# 0006. Admin panel: Next.js, TypeScript, Tailwind

Status: Accepted
Date: 2026-10-25
Issue: #10

## Context
Designers and operators must change every tunable number (CLAUDE.md: no tunable constant in code), manage the asset manifest with rarity and status colour, run promo codes, seasons and passes, and review purchases. This surface has different users, release cadence and risk from the game server.

## Decision
- `admin/` is a Next.js app in TypeScript with Tailwind.
- It is a separate deployable from the game server. It does not open game WebSockets and holds no game logic.
- It talks to a distinct admin API (separate route group and credentials, never player tokens). The admin API validates and writes config to PostgreSQL; the game server reads config as a versioned snapshot and applies it without redeploy.
- Config is versioned: every publish creates a new version, each match records the version it used, and a rollback is a publish of an earlier version.
- Every admin write is authenticated, authorised by role and written to an audit log (who, what, before, after). Roles and the login method are decided in the admin spec.
- Secrets come from environment only.

## Alternatives considered and why rejected
- Admin pages inside the game server: fewer deployables, but it mixes a high-risk write surface with the latency-critical process and ties release cadence together.
- Retool or another low-code tool: fast to start, but the asset manifest, versioned config and audit workflows need custom UI, and the tool would sit in front of production data with its own permission model.
- Unity editor tooling only: operators are not in the editor and changes must go live without a build.
- Config in repo files only: designers need to tune live and non-engineers cannot open PRs. Repo files remain for defaults and schema.
- Another SPA framework (plain React, Vue): workable. Next.js is chosen for server-side session handling, API routes and a TypeScript stack that matches the rest of the toolchain.

## Consequences
- Easier: independent deploys, a UI toolkit for forms and tables, shared TypeScript types where an API schema is generated.
- Harder: a config schema contract between admin and server must be maintained and validated on both sides. Invalid config is rejected at publish time, not discovered in a match.
- Compromise of the admin panel is compromise of the economy; access control and audit are security-critical (ADR 0007).
