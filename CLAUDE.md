# HypeDomino

Real-time online multiplayer mobile domino game. Hyperactive, colorful, effect-heavy UI. Performance is the top priority.

## Source of truth
- Game rules: `docs/rules.md` (Epic 1). Code must match it exactly. Change the doc first, then the code.
- Economy, rarity, ranking, pass specs: `docs/` (Epic 1). Plan and scope: GitHub epics and sub-issues.
- Never mention any third-party game name in code, docs, issues, commits or assets. Describe behavior on its own terms.

## Monorepo
- `client/` Unity 6 LTS, URP 2D, C#, Addressables
- `server/` Go, WebSocket (binary protobuf), PostgreSQL, Redis
- `admin/` Next.js, TypeScript, Tailwind
- `proto/` shared protobuf contracts
- `docs/` specs and ADRs
- `assets-src/` source art (Blender, Adobe, AI-generated), never shipped directly
- `tools/` scripts and pipelines

## Code rules
- Write code only. No comments, no docstrings, no XML doc, no TODO markers. Names must explain intent.
- Exception: license headers required by a dependency, and config files where a format needs a comment.
- Small functions, early returns, no dead code, no speculative abstractions.
- Server is authoritative. Client never decides scores, legality of moves, rewards or currency.
- All game rules live in one pure Go package (`server/internal/rules`) with table-driven tests. The client only renders and sends intents.
- No allocations in hot paths on the client (Update loops, input, animation). Pool everything that spawns repeatedly.
- Every number that designers may tune comes from admin-managed config, not constants.
- Secrets only from environment. Never commit keys, tokens or service accounts.

## Performance budgets
- Client: stable 60 fps on mid-range Android, 120 fps option on capable devices; cold start under 3 s; match scene draw calls under 150; GC alloc 0 B/frame during a match.
- Install size under 150 MB; extra content via Addressables/CDN on demand.
- Server: p99 move round-trip under 100 ms in region; 10k concurrent rooms per node target.
- Any change touching these paths must include a measurement in the PR.

## Workflow
- One issue = one branch = one PR. Branch: `feat/<issue>-<slug>`, `fix/<issue>-<slug>`.
- Conventional commits (`feat:`, `fix:`, `perf:`, `test:`, `chore:`, `docs:`).
- PR description: what changed, how measured/tested, `Closes #N` only if the issue is fully done, else `Part of #N`.
- CI must be green before merge. Never disable or skip tests to get green.
- Rules, scoring and economy logic require tests before merge.
- Issues carry labels: `type:*`, `area:*`, `priority:*`, `epic` for epics. Sub-issues are native GitHub sub-issues of their epic.

## Skills (use them)
- caveman (https://github.com/juliusbrussee/caveman): terse agent replies and compressed context to save tokens. Default communication style in sessions.
- ponytail (https://github.com/DietrichGebert/ponytail): lazy senior dev mindset. Prefer deleting and reusing over writing. Challenge every new abstraction and dependency.
- andrej-karpathy-skills (https://github.com/multica-ai/andrej-karpathy-skills): think before coding, simplicity first, surgical changes, goal-driven verification.
- graphify (https://github.com/Graphify-Labs/graphify): build and query the codebase knowledge graph (`/graphify`) before broad grep/reads. Rebuild after large structural changes.

## Asset pipeline
- Concept and raw art: ChatGPT (browser), then cleanup in Adobe (Photoshop, Illustrator, After Effects) or Blender.
- Characters and animated avatars: Blender/After Effects to sprite sheets or skeletal 2D rigs, packed into atlases, compressed (ASTC) and shipped via Addressables.
- Every asset has a rarity, a status color and a manifest entry managed from the admin panel.
- Static fallback exists for every animated asset.

## Language
- Code, identifiers, commits, docs: English.
- Issues and user-facing planning notes: Azerbaijani, technical terms in English.
