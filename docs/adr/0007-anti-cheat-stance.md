# 0007. Anti-cheat stance: zero trust in the client

Status: Accepted
Date: 2026-10-25
Issue: #10

## Context
Modified APKs, memory editors, proxies and replayed traffic are expected. Currency, ranks and purchases make cheating profitable. Client-side detection is bypassable, so the design must make client tampering useless rather than merely detectable. Detailed controls are tracked in Epic 18 and will be written in `docs/anti-cheat.md` (#201, future work, not part of this ADR).

## Decision
- The client sends intents only: for example play tile X at end Y, pass, claim reward Z. It never sends results, balances, prices, scores or inventory.
- The server validates every intent against authoritative state through `server/internal/rules` (ADR 0003). An illegal intent is rejected and counted.
- No balance, price, reward, inventory or score is accepted from, or authoritatively stored on, the client. Client-side copies are display caches replaced by server state.
- Hidden information stays on the server: a client receives only the tiles its player holds (R-4) and public board state.
- Purchases are granted only after server-side receipt validation against the store. The store transaction id is unique in PostgreSQL, so a replayed receipt grants nothing twice (ADR 0005).
- All time-based logic (daily rewards, pass windows, turn timers, cooldowns, season boundaries) uses server time. The device clock is never read for decisions.
- Rate limits and per-account anomaly counters (Redis) guard intents, logins, promo redemptions and purchases.
- Client integrity signals (platform attestation, build signature) may be collected as risk input. They are never proof, and never replace a server check.

## Alternatives considered and why rejected
- Client-side validation with obfuscation: delays tampering but does not prevent it, and any value the client decides can be forged.
- Peer-to-peer or client-authoritative play: cannot be made safe.
- The client sends results and the server verifies by replay: more complex than sending intents, with a larger attack surface.
- A heavy third-party client anti-cheat as the main defence: same weakness as obfuscation. It may be added later as a signal.
- Sending the full deal to all clients and trusting them to hide it: memory inspection reveals it.

## Consequences
- Easier to reason about: a compromised client can only send intents the rules allow, so the worst case is legal play using information it already holds.
- Harder: every client feature that shows a number needs a server round-trip or a server-pushed value, and optimistic UI must reconcile with server truth. The server path latency counts against the 100 ms budget.
- Collusion in 2v2 and automated play are not solved by this stance; they are addressed by behavioural detection in `docs/anti-cheat.md` (#201).
