# AFK and leaver penalties

Penalties for players who stop playing a match (R-51) or leave it (R-52). The server alone detects offenses, applies penalties and keeps the record; the client only shows them. Every count, weight, percentage and duration below is admin config (section 7). This document fixes the structure and the order of escalation, not the values.

## 1. Scope

- Applies to ranked 1v1 and 2v2 matches (R-53).
- Does not apply to custom rooms (R-55): nothing is recorded and nothing is penalised there.
- A match that is cancelled by the server (server fault, region outage) records no offense for anyone.

## 2. Offenses

| Offense | Detected when | Weight (config) |
|---|---|---|
| AFK, returned | The player is marked AFK (R-51), takes back control by reconnecting and is in control when the match ends | `AfkReturnedWeight` |
| AFK, not returned | The player is marked AFK (R-51) and the bot is still playing for them when the match ends | `AfkWeight` |
| Leave | The player leaves the match (R-52): quits from the match screen, or stays disconnected longer than `ReconnectGraceSeconds` | `LeaveWeight` |

- One match records at most one offense per player, the heaviest one.
- A disconnect shorter than `ReconnectGraceSeconds`, after which the player plays on, is not an offense. Turns missed while disconnected still count as timeouts toward R-51.
- The server cannot tell a dropped network from a deliberate quit, so it does not try: what counts is whether the player came back in time. Repeated drops are handled by the ladder like any other offense.

## 3. Offense level

- Every offense stays active for `OffenseWindowDays` from the match end and then expires.
- `CleanMatchesToForgive` completed matches without an offense in a row remove the oldest active offense early.
- A player's offense level is the sum of the weights of their active offenses. The level picks the step of the ladder.

## 4. Penalty ladder

Each step applies from the moment the offense is recorded. A higher step includes everything in the steps below it.

| Step | From level (config) | Penalty |
|---|---|---|
| 1. Warning | `WarningLevel` | A notice after the match: what was detected, the current level and the next step |
| 2. Queue delay | `QueueDelayLevel` | Every matchmaking entry waits `QueueDelaySeconds` before the search starts, for the next `QueueDelayMatches` completed matches |
| 3. Matchmaking ban | `BanLevel` | No matchmaking for `BanHours`. The duration is multiplied by `BanEscalationFactor` for each further ban inside the window. Custom rooms stay open |
| 4. Review | `ReviewLevel` | The account is flagged for manual review in the admin panel; the ban from step 3 stays until an admin decides |

## 5. Match result for the offender and the others

| Who | Rating | Rewards |
|---|---|---|
| Offender | The match counts as a loss whatever the result, plus an extra rating loss of `OffenseRatingPenalty` | No coins, pass XP, mission or medal progress from that match. AFK, returned: rewards are kept, the rating penalty still applies |
| Partner of the offender (2v2), if their team loses | The rating loss is reduced by `PartnerLossReductionPercent` | Normal |
| Opponents | Normal result | Normal |

- When both partners of a team commit an offense, neither gets the partner reduction.
- When the offender's team wins (the bot and the partner finish the match), the win still counts as a loss for the offender only.

## 6. Record, notice and appeal

- Every offense is stored with the match id, the seat, the timeout, reconnect and leave events with server timestamps, and the config version used.
- The player sees each active offense, its expiry, their current step and how to leave it.
- An admin can remove an offense or lift a penalty, with a mandatory reason. Every admin change is audited.
- Applying penalties is idempotent: replaying the same match end never records a second offense.

## 7. Config

| Field | Meaning |
|---|---|
| `ReconnectGraceSeconds` | How long a disconnected player may stay away before it counts as a leave |
| `AfkReturnedWeight`, `AfkWeight`, `LeaveWeight` | Weight of each offense |
| `OffenseWindowDays` | How long an offense stays active |
| `CleanMatchesToForgive` | Clean completed matches in a row that remove the oldest active offense |
| `WarningLevel`, `QueueDelayLevel`, `BanLevel`, `ReviewLevel` | Offense level at which each step starts |
| `QueueDelaySeconds`, `QueueDelayMatches` | Queue delay length and how many matches it lasts |
| `BanHours`, `BanEscalationFactor` | First ban length and its growth for each further ban in the window |
| `OffenseRatingPenalty` | Extra rating loss for the offender |
| `PartnerLossReductionPercent` | How much smaller the partner's rating loss is |

Values are set in the admin panel before launch. A match always uses the config version that was active when it started.
