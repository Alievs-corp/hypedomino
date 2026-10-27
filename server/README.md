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
| `internal/gateway` | WebSocket gateway: handshake, `Envelope` codec, size and rate limits, bounded write queue, `MatchView` encoding of a room view (#44). Session lookup and reconnect grace are #45 |
| `internal/pb` | Generated protobuf code and the golden frame test (`proto/README.md`) |
| `internal/matchmaking` | Queue entries, MMR window growth, random 2v2 teams, casual-only bot fill: `Form` over one tick and the in-process `Pool` (`docs/matchmaking.md`, #48) |

The gateway takes a `gateway.Handler`: it receives the accepted `ClientHello` and returns a `gateway.Session` that gets decoded intents with their sequence numbers. `gateway.JoinRoom` and `gateway.ResumeRoom` bind a session to a room seat. A closed connection disconnects the seat without leaving it (`docs/penalties.md`: a short disconnect is not a leave); turns missed meanwhile are timeouts.
