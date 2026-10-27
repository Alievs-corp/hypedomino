# server

Go game server: WebSocket gateway (binary protobuf), rooms, matchmaking, accounts, economy. PostgreSQL and Redis.

Game rules live in one pure package, `server/internal/rules`, implementing [docs/rules.md](../docs/rules.md) (Epic 3). The service skeleton is created in #43.

## Packages

| Package | What |
|---|---|
| `internal/rules` | Pure rules engine: tiles, board, turns, hands, match, presence (Epic 3) |
| `internal/bot` | Easy, medium and hard bot players over a `bot.Table` (#40) |
| `internal/replay` | Intent record of a match, binary format, replay and verification (#41) |
| `internal/room` | One match as a single-goroutine actor: seats, intents, turn timer, bot takeover (#46, #47) |
| `internal/gateway` | WebSocket gateway: handshake, `Envelope` codec, size and rate limits, bounded write queue, ping/pong dead connection detection, `MatchView` encoding of a room view (#44) |
| `internal/session` | Session manager: resume tokens, one live connection per session, room seat binding, reconnect with the acked sequence, reconnect grace before a disconnect counts as a leave (#45) |
| `internal/pb` | Generated protobuf code and the golden frame test (`proto/README.md`) |
| `internal/matchmaking` | Queue entries, MMR window growth, random 2v2 teams, casual-only bot fill: `Form` over one tick and the in-process `Pool` (`docs/matchmaking.md`, #48) |
| `internal/integrity` | Attestation verdicts: single-use nonces bound to account, session and purpose; Play Integrity payload checks behind a `Decoder`; App Attest attestation and assertion verification; the block, flag or warn decision with its risk signal (`docs/anti-cheat.md` 5.1, #205) |
| `internal/idtoken` | Google and Apple Sign-In ID token verification: RS256 signature against provider key sets cached by `Cache-Control`, issuer, audience, expiry with clock skew, single-use server nonce. Returns provider, subject and email for account lookup (#57) |

The gateway takes a `gateway.Handler`: it receives the accepted `ClientHello` and returns a `gateway.Session` that gets decoded intents with their sequence numbers. `gateway.JoinRoom` and `gateway.ResumeRoom` bind a session to a room seat. A closed connection disconnects the seat without leaving it (`docs/penalties.md`: a short disconnect is not a leave); turns missed meanwhile are timeouts.

Sign-in: the server issues a nonce (`Verifier.IssueNonce`), the client passes it to the provider (Google: as is; Apple: its SHA-256 hex, set per provider in `idtoken.Config`) and sends the ID token with the raw nonce back. `Verifier.Verify` accepts the token once; the nonce is consumed, so a replayed token is rejected. The in-memory `MemoryNonces` holds nonces on one node until the Redis store of the data layer replaces it.

`session.Manager` is the production handler. `Create(account)` issues an opaque 128-bit resume token (login, #57 and #58, calls it); `Seat(token, room, seat)` binds the session to a room, now or later (matchmaking, #48). On connect the token is looked up, a previous live connection of the same session is closed with 1008 `session replaced` (newest device wins, no grace timer for the replaced one), and the `ServerHello` carries the last sequence the room acknowledged, so the client resends anything above it. The first attach joins the seat, every later one reconnects and receives the full `MatchView`. When a connection drops, `ReconnectGrace` starts; if the session has not reconnected when it ends, the seat is left (R-52, the bot plays until the player returns). `ReconnectGraceSeconds` is admin config (`docs/penalties.md` section 8); the code default is 60 s until the owner sets it. Intents from a session without a seat are answered `ILLEGAL_INTENT`. Sessions live in process memory; the Redis store with TTLs (ADR 0005) is #51.
