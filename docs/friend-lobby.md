# Friend lobby and custom matches

How friends play together: the friend lobby, team setup and the start of a custom match (R-55). Ranked and casual matches never use any of this: matchmaking forms the match and it opens straight into the loading screen (R-53, R-57). Owner decision 2026-10-28.

Owned elsewhere and only referenced: game rules of custom play (`docs/rules.md` R-55, R-57, R-58), friends and blocking (#153), emoji ownership and limits (#150, #151), character models and clips (ADR 0009, `docs/cosmetics.md`), the loading screen (#85).

## 1. Flow

| Stage | What the players see | Who moves it on |
|---|---|---|
| FL-1 Lobby | A 3D scene with every member's equipped character standing in it, their names and frames, the invite panel, actions and emoji | The creator, with Start |
| FL-2 Team setup | Two team columns with the seats of the chosen mode, the friends panel to invite into free seats | The creator, with Start match |
| FL-3 Loading | The loading screen of every match (R-57, #85) | The server, when everyone is ready or `MatchLoadingSeconds` have passed |
| FL-4 Match | The match scene | The match end |
| FL-5 Back to lobby | Every member who is still in the room returns to FL-1 with the result shown | Automatic |

A member can leave the room in FL-1 and FL-2. Leaving during FL-3 and FL-4 is a leave under R-52 with no penalty (R-55).

## 2. Lobby (FL-1)

- Members: the creator plus invited friends, at most 4, the seat count of the largest mode (R-5). The room has a code for joining without an invite (#268).
- Scene: each member stands as their equipped character, high detail where the graphics setting allows (ADR 0009), with nickname, frame and rank badge. A member who has no animated character stands as the static fallback.
- Invite: from the friends panel (#153), an online friend gets an in-game invite; anyone else gets a link that opens the game into the room (deep link) or the room code. A blocked player cannot invite or be invited (#153).
- Actions: a member taps an action and their character plays it, seen by every member. The action wheel holds the equipped character's own clips (`docs/cosmetics.md`, character `clips`) plus the owned `action` items the member has equipped, up to `LobbyActionSlots` (admin config, #278). An action is an intent; the server relays it and refuses a clip the character does not have or an action the member does not own.
- Emoji: a member sends an owned emoji (#150); it shows over their character for every member. The server refuses an emoji the member does not own and applies the same spam limit as in a match (#151).
- Leadership: the creator controls the room. When the creator leaves, control passes to the member who joined earliest; the room closes when the last member leaves.

## 3. Team setup (FL-2)

- The creator picks the mode, 1v1 or 2v2, and places members into seats: N and S are one team, E and W the other (R-6); X and Y in 1v1 (R-8). Members may ask for a seat by tapping it; the creator decides (R-55).
- Free seats can still be filled by invite from the friends panel. Members who get no seat in 1v1 wait in the lobby until the match ends and see its result.
- Start match needs every seat of the mode filled (#268). Then the match opens with the loading phase (FL-3).
- The creator can go back to the lobby from team setup; seats are kept.

## 4. What the server decides

- Membership, leadership, seats, mode, the start and the return to the lobby are server state; the client sends intents only (CLAUDE.md, zero trust).
- Actions and emoji are relayed, not trusted: ownership of action items and emoji and clip existence are checked on every intent.
- A member whose match has not ended cannot join another room or queue (R-58).
- Nothing in a custom room grants coins, pass XP, missions, medals or rating (R-55, `docs/economy.md` 2.1).

## 5. Decisions and open items

- Decided 2026-10-28: lobby actions are also their own cosmetic type, `action` (`docs/cosmetics.md`), sold and earned like other cosmetics. Their shop price weight (`GuidePriceTypePercent` in `docs/economy.md`) and the value of `LobbyActionSlots` are owner decisions still open; until then actions come only from pass, missions, events, promos and rank rewards.
- Decided 2026-10-28: no bot seats in custom matches. A 2v2 needs four players; with fewer, the room plays 1v1.
