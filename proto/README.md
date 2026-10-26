# proto

Shared protobuf contracts between client, server and admin. This directory is the only source of message definitions (ADR 0004). Generated code is never edited by hand.

## Layout

| Path | What |
|---|---|
| `buf.yaml` | Module, lint (`STANDARD`) and breaking (`FILE`) configuration |
| `buf.gen.yaml` | Code generation: pinned remote plugins and output directories |
| `hypedomino/v1/*.proto` | The `hypedomino.v1` package, one major version |
| `testdata/frames/*.json` | Golden `Envelope` frames, one per payload type, in protobuf JSON |
| `testdata/frames/*.bin` | The same frames in binary wire format, written by the Go test |

Generated output:

| Language | Plugin | Output | Used by |
|---|---|---|---|
| Go | `protocolbuffers/go` | `server/internal/pb/hypedomino/v1` (package `hypedominov1`) | Game server, bot client |
| C# | `protocolbuffers/csharp` | `client/Assets/HypeDomino/Generated/Proto` (namespace `Hypedomino.V1`) | Unity client |
| TypeScript | `bufbuild/es` | `admin/src/gen/proto/hypedomino/v1` | Admin panel |

## Commands

Run from the repository root. `buf` is required; Go is required for the golden frame test.

| Command | What |
|---|---|
| `buf lint proto` | Style and naming rules |
| `buf format --diff --exit-code proto` | Formatting |
| `buf breaking proto --against ".git#branch=main,subdir=proto"` | Wire and source compatibility against `main` |
| `npm run lint:proto` | No value field reaches a client-to-server message (below) |
| `npm run proto:gen` | Regenerate Go, C# and TypeScript. Commit the output |
| `cd server && go test ./internal/pb/...` | Decode every golden frame and compare with its JSON |
| `cd server && go test ./internal/pb/... -update` | Rewrite the binary golden frames from their JSON |

CI runs all of these on every change under `proto/` or a generated directory and fails when the committed generated code differs from the contracts.

## Protocol

- One WebSocket per session carries binary `Envelope` frames in both directions. Every frame has `protocol_version`, the client-assigned `sequence`, the server's `acked_sequence` and exactly one payload.
- The first client frame is `ClientHello`. The server answers `ServerHello` with `HELLO_STATUS_ACCEPTED` or `HELLO_STATUS_UPDATE_REQUIRED`; on the latter the client blocks play and shows the store update screen. Server time and the active catalog version arrive in the same message.
- A client-to-server payload is an intent: `PlaceTileIntent`, `PassIntent`, `DrawIntent`. A rejected intent is answered with `Error` carrying `rejected_sequence`.
- Tiles are `Tile { low, high }` with `low <= high` outside the board. Ends are `BoardEnd`; `TOP` and `BOTTOM` exist only once there is a spinner (`docs/rules.md` R-17, R-18).

## Versioning rules

- The package is versioned by major: `hypedomino.v1`. A breaking redesign is a new package served in parallel for a transition window and needs its own ADR.
- Inside a major version only additive changes are allowed. Field numbers are never reused or renumbered. A removed field is kept as `reserved` with its number and its name.
- Every enum has an `_UNSPECIFIED` zero value. Readers handle unknown enum values safely.
- `buf breaking` runs against the PR base and fails the `proto` CI job on a breaking change. Every `proto/` PR states whether it is additive.
- `buf.gen.yaml` pins plugin versions. Bumping one is its own PR and regenerates all three outputs.

## Client-to-server messages carry no values

The client sends intents only (`docs/anti-cheat.md` NT-1). `tools/proto-lint` enforces it:

- A message is client-to-server when its name ends with `Intent`, or it is `ClientHello`. New client-to-server messages must follow this naming.
- Those messages, and every message type reachable from their fields, must not have a field whose snake_case name contains a value token: `balance`, `coin`, `price`, `cost`, `reward`, `score`, `points`, `rating`, `mmr`, `lp`, `xp`, `rank`, `tier`, `inventory`, `owned`, `result` (singular and plural).
- Server-to-client messages may carry any of these. A quantity the client legitimately chooses (for example a pass level count) is named `count`.

## Allocation on the client

The generated C# parser reuses an existing instance for a singular message field (`if (tile_ == null) Tile = new Tile(); input.ReadMessage(Tile);`) but allocates a fresh instance for every `oneof` member on each parse. `Envelope.payload` is a `oneof`, so the client never parses an `Envelope` with the generated parser. It reads the three varints and the payload tag with `CodedInputStream` by hand, keeps one long-lived instance per payload type, and calls `input.ReadMessage(instance)` on the matching one. Sending works the same way: one `Envelope` per intent type is kept and only its scalar fields change. Repeated fields are cleared and refilled, not reallocated. This is a usage rule for `client/` (#67), checked by the allocation tests of `docs/test-strategy.md` section 2.5; generation cannot enforce it.
