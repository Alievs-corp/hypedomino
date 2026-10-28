# server

Go game server: WebSocket gateway (binary protobuf), rooms, matchmaking, accounts, economy. PostgreSQL and Redis.

Game rules live in one pure package, `server/internal/rules`, implementing [docs/rules.md](../docs/rules.md) (Epic 3).

## Run

```
cd server && go run ./cmd/server
```

The binary reads its configuration from the environment only (CLAUDE.md: secrets and tunables never live in code) and logs JSON lines to stdout. `SIGTERM` or `SIGINT` starts a drain: `/readyz` turns 503 so the load balancer stops routing, no new room opens and no new WebSocket is accepted, live matches play to their end (bounded by `HD_DRAIN_TIMEOUT`), then connections are closed with 1001 `server shutdown` and the listener stops (`docs/launch-runbook.md` 5.1: never kill a node with live rooms).

| Variable | Default | Meaning |
|---|---|---|
| `HD_LISTEN_ADDR` | `:8080` | TCP address of the HTTP listener (`/healthz`, `/readyz`, `/ws`) |
| `HD_LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `HD_ORIGINS` | none | Comma separated origin host patterns allowed to open a WebSocket besides the request host |
| `HD_SHUTDOWN_TIMEOUT` | `10s` | Time given to close connections and the listener after the drain |
| `HD_DRAIN_TIMEOUT` | `15m` | Longest wait for live matches to finish; rooms still running after it are cancelled |
| `HD_MIN_APP_VERSION` | `1.0.0` | Clients below it get `UPDATE_REQUIRED` (admin config in #210) |
| `HD_CATALOG_VERSION` | empty | Catalog version announced in `ServerHello` |
| `HD_MAX_FRAME_BYTES`, `HD_WRITE_QUEUE`, `HD_HELLO_TIMEOUT`, `HD_WRITE_TIMEOUT`, `HD_CLOSE_TIMEOUT`, `HD_COMPRESSION_THRESHOLD`, `HD_INTENTS_PER_SECOND`, `HD_INTENT_BURST`, `HD_PING_INTERVAL`, `HD_PONG_TIMEOUT` | `gateway.DefaultConfig` | Gateway limits (`proto/README.md`, Gateway limits) |
| `HD_RECONNECT_GRACE` | `60s` | `ReconnectGraceSeconds` of `docs/penalties.md` |

`/healthz` answers 200 while the process runs. `/readyz` answers 200 while the server accepts work and 503 from the first drain step on. A bad value in the environment stops the start with every bad variable named, exit code 2.

## Container and deploy

`server/Dockerfile` builds a static binary into a distroless image that runs as user `65532` (ADR 0010). Kubernetes manifests are in `deploy/k8s`: `base` is shared, an overlay per environment sets the image tag, replicas and the `server-env` secret. Locally, with Docker and kind:

```
docker build -t hypedomino-server:ci server
bash tools/deploy/container-smoke.sh hypedomino-server:ci
kind create cluster --name hd
bash tools/deploy/kind-rollout.sh hd hypedomino-server:ci
```

`terminationGracePeriodSeconds` in `deploy/k8s/base/server.yaml` must stay above the 10 s `preStop` plus `HD_DRAIN_TIMEOUT` plus `HD_SHUTDOWN_TIMEOUT`; change them together.

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
| `internal/custom` | Custom rooms (R-55): room codes, seat and team choice, ready and start by the creator, creator removes and adds players before or during the match; starts a `room.KindCustom` table (#49) |
| `internal/penalties` | AFK and leaver penalties over `rules.PresenceReport`: offense detection, the ranked and casual ladders, bans, queue delays and result mapping, all pure and idempotent by match id (`docs/penalties.md`, #52) |
| `internal/app` | Composition root: typed config from the environment, JSON logger, HTTP mux with health endpoints, room registry and the drain sequence (#43) |
| `internal/botclient` | Go protocol client over the generated contract: hello, intents with sequence numbers, a legal default move for any view, used by the in-process scenario tests and later the load bot (`docs/test-strategy.md` section 3) |
| `internal/loadbot` | Load generator over `botclient`: rooms kept at a target count, human cadence, random drops with resume, per-move round-trip histograms overall and per window, pass/fail against the budget; plus the load-mode `Handler` that seats a token `<room>/<mode>/<seat>` without login (`docs/performance.md` section 6, #53) |
| `cmd/server` | The game server binary |
| `internal/db` | PostgreSQL access: `Open` (pgx pool), the embedded goose migrations and `Migrator` (up, down to, status, version) under a session advisory lock; `dbtest` starts PostgreSQL 16 in a container, or uses `HD_TEST_DATABASE_URL`, and hands every test a fresh migrated database (#55) |
| `cmd/migrate` | `up`, `down -to N`, `status`, `version` against `HD_DATABASE_URL`; run once per deploy from one runner (`docs/environments.md` 6) |
| `cmd/loadbot` | `serve` (a node with the load-mode handler), `run` (the generator against a node) and `local` (both in one process, with heap samples) |

The gateway takes a `gateway.Handler`: it receives the accepted `ClientHello` and returns a `gateway.Session` that gets decoded intents with their sequence numbers. `gateway.JoinRoom` and `gateway.ResumeRoom` bind a session to a room seat. A closed connection disconnects the seat without leaving it (`docs/penalties.md`: a short disconnect is not a leave); turns missed meanwhile are timeouts.

Sign-in: the server issues a nonce (`Verifier.IssueNonce`), the client passes it to the provider (Google: as is; Apple: its SHA-256 hex, set per provider in `idtoken.Config`) and sends the ID token with the raw nonce back. `Verifier.Verify` accepts the token once; the nonce is consumed, so a replayed token is rejected. The in-memory `MemoryNonces` holds nonces on one node until the Redis store of the data layer replaces it.

`session.Manager` is the production handler. `Create(account)` issues an opaque 128-bit resume token (login, #57 and #58, calls it); `Seat(token, room, seat)` binds the session to a room, now or later (matchmaking, #48). On connect the token is looked up, a previous live connection of the same session is closed with 1008 `session replaced` (newest device wins, no grace timer for the replaced one), and the `ServerHello` carries the last sequence the room acknowledged, so the client resends anything above it. The first attach joins the seat, every later one reconnects and receives the full `MatchView`. When a connection drops, `ReconnectGrace` starts; if the session has not reconnected when it ends, the seat is left (R-52, the bot plays until the player returns). `ReconnectGraceSeconds` is admin config (`docs/penalties.md` section 8); the code default is 60 s until the owner sets it. Intents from a session without a seat are answered `ILLEGAL_INTENT`. Sessions live in process memory; the Redis store with TTLs (ADR 0005) is #51.

Penalties: at match end the room's `View.Presence` report of each seat goes to `penalties.Record.MatchEnd` with the queue, match id and config version; the returned `Notice` is what the player sees. `Record.QueueDelay` becomes the `SearchFrom` offset of the next `matchmaking.Entry`, `Record.Banned` gates ranked enqueue (ban or review), and `penalties.Results` tells the rating and reward engines how to treat each seat. `penalties.DefaultConfig` holds placeholder numbers until the admin panel publishes the real ones; the record is kept in memory by its caller until the data layer (#50, #55) stores it.

## Database

```
cd server
HD_DATABASE_URL=postgres://user:secret@host:5432/hypedomino go run ./cmd/migrate up
HD_DATABASE_URL=... go run ./cmd/migrate status
HD_DATABASE_URL=... go run ./cmd/migrate down -to 0      # dev and CI only (M-1)
```

Migrations are goose SQL files in `internal/db/migrations`, numbered, each with an Up and a Down section, embedded into `cmd/migrate` and `cmd/server`. `migrate up` takes a session advisory lock so concurrent runners apply each version once. `00001_core.sql` creates `accounts`, `auth_identities`, `profiles`, `inventory_items`, `wallets`, `ledger_entries` (idempotency key unique per account, purchase keys unique across accounts, balances never negative), and `matches` with `match_participants` partitioned by month of `started_at` with a default partition and `match_partitions_ensure`. Integration tests run against a real PostgreSQL 16: `go test ./internal/db` starts `postgres:16-alpine` through testcontainers on the local Docker, or connects to `HD_TEST_DATABASE_URL` (a URL with a role that may create databases) when set; `-short` skips them. PostgreSQL 16 is the major version of every environment (`docs/environments.md` E-6).

## Load test

```
cd server
go run ./cmd/loadbot serve -addr :8080                 # on the node under test, same HD_* variables as cmd/server
go run ./cmd/loadbot run -target ws://node:8080/ws -rooms 10000 -2v2 50 -ramp 5m -hold 30m -out report.json
go run ./cmd/loadbot local -rooms 1000 -hold 2m       # one process, developer machine
```

`serve` is the game server with the load-mode handler instead of the session manager: any client whose token is `<room>/<mode>/<seat>` is seated, a room is opened on its first token and a second connection with the same token resumes the seat with the acked sequence it sends. It is never deployed to a player-facing environment. `run` keeps `-rooms` rooms active (a finished match is replaced at once), seats every player with a `botclient`, waits `-think-min` to `-think-max` before the first action of each turn and `-think-follow` before a further draw or placement in the same turn (R-49 times the whole turn), drops each connection `-drops` times per minute on average and resumes it, and timestamps every intent from send to the first broadcast that acknowledges it. The report has mean, p50, p95, p99 and max overall, the steady state after `-warmup`, every `-window`, drops, reconnects, rejected intents, dial and read failures, bytes per intent and, in `local` mode, heap samples. The exit code is 1 when the steady-state p99 or any window after the warm-up is not under `-budget` (100 ms), the error rate is not under 0.1% or the heap grew more than 5% in the steady state. A 12-hour soak is `-hold 12h`; the full target is `-rooms 10000` from a host with the file descriptor limit raised above 40,000.
