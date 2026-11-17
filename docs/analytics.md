# Analytics events

First-party product analytics (#181). Privacy rules are `docs/privacy.md` section 6.4 and data item D-24; this document lists the events and how the client sends them.

## 1. Rules

- Nothing is tracked, queued or written to disk before the `analytics` consent purpose is on (`docs/privacy.md` 6.2). An event tracked before consent is dropped, not held.
- Withdrawing consent drops every queued event at once and stops sending; the caller rotates the analytics id.
- An event carries the analytics id, its name, one optional label, one optional integer value and the device time. Never an account id, email, nickname, IP, install id or free text.
- Events go only to the game's own server. No third-party analytics SDK ships.
- Analytics never decides anything: revenue and economy KPIs (#147) come from the ledger and purchase records, not from these events.

## 2. Client events

| Event | Label | Value | Funnel step |
|---|---|---|---|
| `app_open` | none | 0 | Install (first `app_open` of an analytics id), session |
| `sign_in` | Provider: `google` or `apple` | 0 | Login |
| `nickname_set` | none | 0 | Login |
| `screen_view` | Screen id | 0 | |
| `queue_join` | Queue mode | 0 | First match |
| `purchase_start` | Product id | 0 | Purchase |
| `purchase_cancel` | Product id | 0 | Purchase |

The client sends only names on this list (`AnalyticsEvents.IsClientEvent`); any other name is ignored.

## 3. Server events

Match end, purchase, reward, mission and pass events are written by the server from its own records (`docs/privacy.md` 6.4). For a player without `analytics` they carry no analytics id and only feed aggregates. Only `match_end` is built.

| Event | Source |
|---|---|
| `match_end` | Match record (D-12) |
| `purchase_complete` | Validated purchase (D-17) |
| `reward_grant` | Ledger entry (D-16) |
| `mission_complete` | Mission progress (D-16) |
| `pass_level` | Pass progress (D-16) |

`match_end`: `matchstore.Finish` writes the match outcome to the `analytics` outbox topic in the same transaction as the result, for every match kind (`docs/matchmaking.md` section 4). The game server's relay on that topic (`HD_OUTBOX_POLL`) writes one `match_end` row per human player of the outcome, cancelled matches included, to `analytics_events`. The server cannot read the `analytics` consent yet (#157), so every row has no analytics id, no device time, no label and value 0; its time is the match end. `analytics_sources` keeps the topic and key of every applied outbox row, so a redelivered match writes nothing twice; it holds no account and no link to the events. The label and value of `match_end` are not decided.

## 4. Sending

`Core/Analytics/AnalyticsQueue` holds tracked events in memory while the device is offline. `Flush` hands at most `AnalyticsBatchSize` events to an `IAnalyticsSender`; a failed send keeps them for the next flush. When the queue holds `AnalyticsQueueLimit` events, the oldest is dropped. The caller flushes every `AnalyticsFlushSeconds` and when the app pauses.

## 5. Server ingest

`POST /analytics/events` (`server/internal/analytics`) takes the access token and `{analyticsId, events: [{name, label, value, clientTimeMs}]}` as JSON.

- `analyticsId` is a random UUID in its 36-character form. The first account that sends it owns it (`analytics_ids`); the same id from another account answers 409 `analytics.id_taken` and stores nothing. This table is the only link from an analytics id to an account (`docs/privacy.md` P-7) and is what deletion and export use.
- An event is kept only if its name is on the client list in section 2, its value is 0, `clientTimeMs` is positive and its label matches the event: none for `app_open` and `nickname_set`, `google` or `apple` for `sign_in`, and for the others an id of at most 64 letters, digits, `_`, `.` or `-` starting with a letter or digit. Any other event is dropped and the rest of the batch is kept, so a client that knows a newer event does not lose the batch. Server event names are never accepted from the client.
- The answer is 200 `{accepted}`, the number of events stored. A body that is not valid JSON, is larger than 64 KiB or has no valid analytics id answers 400 `analytics.malformed`; a 400 or 409 batch will never be accepted, so the sender drops it rather than retrying.
- `analytics_events` stores the analytics id, name, label, value, the device time as sent and the server receive time. The source IP, the account id and the request headers are never stored.
- Erasure: `analytics_ids` is an erase-rows table in account deletion and its events go with it. Export: the `analytics` section lists the account's events with their analytics id.
- The global per-IP and per-account rate limits (`HD_RATE_*`) cover the route.
- Retention: every run of the hourly deletion job (`server/cmd/account-deletion run`, `docs/backup-restore.md` 3.7) deletes the `analytics_events` rows whose server receive time is older than `HD_ANALYTICS_RAW_RETENTION`, default 9480 h, which is `AnalyticsRawRetentionDays` (395) of `docs/privacy.md` section 4. The device time is never used for it. `analytics_ids` rows stay: they are the account's link for deletion and export and go with the account. The purge runs even when the deletion log is unreachable, and a failure of either fails the job. The same run deletes the server events of section 3 by their time and the `analytics_sources` rows older than the cutoff.

Not built yet: the server events of section 3 other than `match_end`, aggregates without ids (none exist yet, so nothing is kept past the purge), a server check of the `analytics` consent record (the client gate is the only gate until #157 writes it), and the client sender.

## 6. Config

| Field | Meaning |
|---|---|
| `AnalyticsBatchSize` | Most events in one send |
| `AnalyticsQueueLimit` | Most events held while offline |
| `AnalyticsFlushSeconds` | Time between flushes |

The values are owner decisions not yet made; they come from admin config.
