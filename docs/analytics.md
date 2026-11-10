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

Match end, purchase, reward, mission and pass events are written by the server from its own records (`docs/privacy.md` 6.4). For a player without `analytics` they carry no analytics id and only feed aggregates. They are not built yet.

| Event | Source |
|---|---|
| `match_end` | Match record (D-12) |
| `purchase_complete` | Validated purchase (D-17) |
| `reward_grant` | Ledger entry (D-16) |
| `mission_complete` | Mission progress (D-16) |
| `pass_level` | Pass progress (D-16) |

## 4. Sending

`Core/Analytics/AnalyticsQueue` holds tracked events in memory while the device is offline. `Flush` hands at most `AnalyticsBatchSize` events to an `IAnalyticsSender`; a failed send keeps them for the next flush. When the queue holds `AnalyticsQueueLimit` events, the oldest is dropped. The caller flushes every `AnalyticsFlushSeconds` and when the app pauses.

## 5. Config

| Field | Meaning |
|---|---|
| `AnalyticsBatchSize` | Most events in one send |
| `AnalyticsQueueLimit` | Most events held while offline |
| `AnalyticsFlushSeconds` | Time between flushes |

The values are owner decisions not yet made; they come from admin config.
