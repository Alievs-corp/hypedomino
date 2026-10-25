# server

Go game server: WebSocket gateway (binary protobuf), rooms, matchmaking, accounts, economy. PostgreSQL and Redis.

Game rules live in one pure package, `server/internal/rules`, implementing [docs/rules.md](../docs/rules.md) (Epic 3). The service skeleton is created in #43.
