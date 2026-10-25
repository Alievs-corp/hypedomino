# 0008. Measurement plan for performance budgets

Status: Accepted
Date: 2026-10-25
Issue: #10

## Context
CLAUDE.md sets the budgets and requires a measurement in any PR that touches the affected path. This ADR fixes what measures each budget and where. Procedures, device lists, thresholds and report format are defined in `docs/performance.md` (#11).

## Decision

| Budget (CLAUDE.md) | What measures it | Where |
|---|---|---|
| 60 fps stable on mid-range Android, 120 fps option | Frame time capture over a scripted match on a reference device set | Device run, result attached to the PR |
| Cold start under 3 s | Launch to interactive on a reference device, cold process | Device run |
| Match scene draw calls under 150 | Frame Debugger and Profiler counters in the match scene | Editor and device |
| 0 B GC alloc per frame in a match | Profiler GC alloc over a scripted match; play-mode allocation check where possible | Editor test and device |
| Install under 150 MB | Size of the release build artifact per platform | CI build step |
| p99 move round-trip under 100 ms in region | Server histogram from intent receipt to broadcast, plus client-observed round trip in a load test | Load test, production metrics |
| 10k concurrent rooms per node | Load test with simulated clients on one node, tracking CPU, memory, GC pause and p99 | Load test |

Rules for use:
- A budget with no automated measurement is checked manually and recorded in the PR until automation exists.
- A PR touching client hot paths, the rules package, the protocol or the connection layer includes a before and after number from the matching row.
- Hot paths in `server/internal/rules` have Go benchmarks with allocation reporting.
- Budgets come only from CLAUDE.md. This ADR adds no thresholds.

## Alternatives considered and why rejected
- Measure only before release: regressions are cheaper to fix when introduced.
- Editor-only measurements: editor numbers do not represent mid-range Android.
- A single aggregate performance score: it hides which budget regressed.

## Consequences
- Easier: every budget has a named measuring mechanism and a place to look.
- Harder: needs a reference device set and a load-test harness, tracked as follow-up issues. Until then, manual records are required.
- If `docs/performance.md` disagrees with this table, fix one of them in the same PR.
