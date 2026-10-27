# Architecture decision records

One file per decision, numbered `NNNN-short-title.md`. An accepted ADR is not edited to change its meaning; it is superseded by a new ADR that links back.

## Status values

| Status | Meaning |
|---|---|
| Proposed | Under discussion, not binding |
| Accepted | Binding; code and docs follow it |
| Superseded by NNNN | Replaced; kept for history |
| Deprecated | No longer applies, no replacement |

## Template

```markdown
# NNNN. Title

Status: Proposed | Accepted
Date: YYYY-MM-DD
Issue: #N

## Context
Forces and constraints. Link budgets from CLAUDE.md.

## Decision
What is chosen, stated as a rule.

## Alternatives considered and why rejected
- Option: reason.

## Consequences
What gets easier, what gets harder, what must be verified.
```

## Index

| ADR | Title | Status |
|---|---|---|
| [0001](0001-client-unity.md) | Client: Unity 6 LTS, URP 2D, IL2CPP | Accepted; renderer superseded by 0009 |
| [0002](0002-asset-delivery.md) | Asset delivery: Addressables, CDN, ASTC | Accepted |
| [0003](0003-server-go-authoritative.md) | Server: authoritative Go with a pure rules package | Accepted |
| [0004](0004-transport-and-protocol.md) | Transport and protocol: WebSocket with binary protobuf | Accepted |
| [0005](0005-data-postgres-redis.md) | Data: PostgreSQL source of truth, Redis for ephemeral state | Accepted |
| [0006](0006-admin-panel.md) | Admin panel: Next.js, TypeScript, Tailwind | Accepted |
| [0007](0007-anti-cheat-stance.md) | Anti-cheat stance: zero trust in the client | Accepted |
| [0008](0008-measurement-plan.md) | Measurement plan for performance budgets | Accepted |
| [0009](0009-characters-real-time-3d.md) | Characters: real-time 3D in the Universal Renderer | Accepted |

Related specs: `docs/rules.md` (game rules), `docs/performance.md` (#11), `docs/anti-cheat.md` (#201). The last two are separate issues and may not exist yet.
