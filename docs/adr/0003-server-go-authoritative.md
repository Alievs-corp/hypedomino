# 0003. Server: authoritative Go with a pure rules package

Status: Accepted
Date: 2026-10-25
Issue: #10

## Context
Budgets: p99 move round-trip under 100 ms in region; 10k concurrent rooms per node as a target. The game is competitive with currency and ranking, so cheating must gain nothing (ADR 0007). Rules are defined once in `docs/rules.md` (R-1 onward) and code must match exactly; hands are shuffled server-side with a cryptographically secure source (R-4).

## Decision
- The server is written in Go and is authoritative for deals, legality of moves, scoring, rewards, currency and time.
- All game rules live in one pure package, `server/internal/rules`: no I/O, no clock, no global state, randomness injected. Input is state plus intent; output is new state plus events, or an error. It is tested table-driven against vectors derived from `docs/rules.md`.
- A room is a goroutine-owned state machine fed by a channel. One room is never touched by two goroutines, so there are no locks on the hot path.
- Connection handling, matchmaking, persistence and admin config live outside `rules`. They call it; it never calls them.
- Scaling is by room: a room lives on one node, and routing metadata lives in Redis (ADR 0005).

## Alternatives considered and why rejected
- Nakama (Go, open source game backend): provides accounts, matchmaking, storage and realtime out of the box. Rejected because the authoritative match handler would still be our code, we would adopt its data model, auth and extension model, and the economy, pass and anti-cheat requirements need a custom schema anyway. Hosting complexity and upgrade coupling outweigh the saved boilerplate. Reconsider only if time-to-market becomes the dominant risk.
- Node.js (TypeScript): shares a language with the admin panel and is quick to write. Rejected because the single-threaded event loop and GC pauses make p99 latency harder to guarantee under room-count pressure, and multi-core use needs process-level sharding.
- Elixir/Erlang (BEAM): excellent for many lightweight stateful rooms and fault isolation, and technically a credible alternative. Rejected mainly on team familiarity, hiring pool and a smaller ecosystem for store receipt validation and tooling.
- C#/.NET: would let the rules package be shared with the Unity client. Rejected deliberately: sharing rules with the client invites client-side decisions, which CLAUDE.md forbids. Go also gives small static binaries, simple deployment and cheap goroutines for 10k rooms.
- Client-authoritative or peer-to-peer: rejected; incompatible with zero trust.

## Consequences
- Easier: a simple concurrency model, one static binary, low memory per room, a strong standard library for networking, fast tests of the rules package.
- Harder: no code sharing with the Unity client, so message definitions come from `proto/` (ADR 0004). Go GC pauses must be watched under load; hot paths avoid allocation and are benchmarked (ADR 0008).
- Rule changes follow the order: edit `docs/rules.md`, then `rules`, then vectors. A rule not in the doc is a bug.
- The per-node room target stays a target until a load test proves it (ADR 0008).
